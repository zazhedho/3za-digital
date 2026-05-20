package serviceorder

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	domainaudit "3za-digital/internal/domain/audit"
	domaincatalog "3za-digital/internal/domain/catalog"
	domainorder "3za-digital/internal/domain/order"
	domainwallet "3za-digital/internal/domain/wallet"
	"3za-digital/internal/dto"
	"3za-digital/internal/integrations/h2h"
	interfaceappconfig "3za-digital/internal/interfaces/appconfig"
	interfaceaudit "3za-digital/internal/interfaces/audit"
	interfaceorder "3za-digital/internal/interfaces/order"
	interfaceprovider "3za-digital/internal/interfaces/provider"
	"3za-digital/pkg/filter"
	"3za-digital/pkg/money"
	"3za-digital/utils"

	"gorm.io/gorm"
)

var (
	ErrInvalidOrderRequest       = errors.New("invalid order request")
	ErrInactiveService           = errors.New("service is inactive")
	ErrUnsupportedService        = errors.New("service is not supported for this product")
	ErrQuantityBelowMinimum      = errors.New("quantity is below service minimum")
	ErrQuantityAboveMaximum      = errors.New("quantity is above service maximum")
	ErrOrderAlreadyFinal         = errors.New("order already has final status")
	ErrProviderClientUnavailable = errors.New("provider client unavailable")
	ErrProviderEmptyResponse     = errors.New("provider returned empty response")
)

type OrderService struct {
	Repo            interfaceorder.RepoOrderInterface
	ProviderFactory interfaceprovider.ClientFactory
	ConfigService   interfaceappconfig.ServiceAppConfigInterface
	AuditService    interfaceaudit.ServiceAuditInterface
}

func NewOrderService(repo interfaceorder.RepoOrderInterface, providerFactory interfaceprovider.ClientFactory, configServices ...interfaceappconfig.ServiceAppConfigInterface) *OrderService {
	service := &OrderService{
		Repo:            repo,
		ProviderFactory: providerFactory,
	}
	if len(configServices) > 0 {
		service.ConfigService = configServices[0]
	}
	return service
}

func (s *OrderService) WithAuditService(auditService interfaceaudit.ServiceAuditInterface) *OrderService {
	s.AuditService = auditService
	return s
}

func (s *OrderService) GetAll(ctx context.Context, params filter.BaseParams) ([]domainorder.Order, int64, error) {
	return s.Repo.GetAll(ctx, params)
}

func (s *OrderService) GetByID(ctx context.Context, id string) (domainorder.Order, error) {
	return s.Repo.GetByID(ctx, id)
}

func (s *OrderService) GetStatusLogs(ctx context.Context, orderID string) ([]domainorder.OrderStatusLog, error) {
	if _, err := s.Repo.GetByID(ctx, orderID); err != nil {
		return nil, err
	}
	return s.Repo.GetStatusLogs(ctx, orderID)
}

func (s *OrderService) CreateOrder(ctx context.Context, productType string, req dto.CreateOrderRequest, createdBy string) (domainorder.Order, error) {
	req.Target = strings.TrimSpace(req.Target)
	createdBy = strings.TrimSpace(createdBy)
	if req.Target == "" || req.Quantity <= 0 || strings.TrimSpace(req.ServiceID) == "" || createdBy == "" {
		return domainorder.Order{}, ErrInvalidOrderRequest
	}
	productType = utils.NormalizeKey(productType)

	service, err := s.Repo.GetServiceByID(ctx, req.ServiceID)
	if err != nil {
		return domainorder.Order{}, err
	}
	if err := validateService(service, productType, req.Quantity); err != nil {
		return domainorder.Order{}, err
	}
	providerCharge, amount, profit, err := s.calculatePrice(ctx, productType, service.Price)
	if err != nil {
		return domainorder.Order{}, err
	}

	order := domainorder.Order{
		Id:                utils.CreateUUID(),
		Provider:          domaincatalog.ProviderH2H,
		ProductType:       productType,
		RefID:             generateRefID(productType),
		ServiceID:         new(service.Id),
		ProviderServiceID: service.ProviderServiceID,
		Target:            req.Target,
		Quantity:          new(req.Quantity),
		Status:            domainorder.StatusPending,
		Amount:            amount,
		ProviderCharge:    providerCharge,
		Profit:            profit,
		Metadata:          utils.MustJSON(map[string]string{"source": "api"}),
		ProviderResponse:  utils.EmptyJSON(),
		CreatedBy:         utils.StringPtrIfNotEmpty(createdBy),
		UpdatedAt:         new(time.Now()),
	}

	order, err = s.Repo.CreateWithStatusLogAndWalletDebit(ctx, order, domainorder.OrderStatusLog{
		Id:               utils.CreateUUID(),
		NewStatus:        domainorder.StatusPending,
		ProviderResponse: utils.EmptyJSON(),
	}, createdBy)
	if err != nil {
		return domainorder.Order{}, err
	}

	client, err := s.providerClient()
	if err != nil {
		_ = s.failAndRefund(ctx, order, err, utils.EmptyJSON())
		return order, err
	}

	providerResp, err := client.CreateOrder(ctx, h2h.CreateOrderRequest{
		Type:        productType,
		ServiceCode: service.ProviderServiceID,
		Target:      req.Target,
		Quantity:    int(req.Quantity),
		RefID:       order.RefID,
	})
	if err != nil {
		_ = s.failAndRefund(ctx, order, err, utils.EmptyJSON())
		return order, err
	}
	if providerResp == nil {
		_ = s.failAndRefund(ctx, order, ErrProviderEmptyResponse, utils.EmptyJSON())
		return order, ErrProviderEmptyResponse
	}

	return s.applyCreateOrderResponse(ctx, order, providerResp)
}

func (s *OrderService) RefreshStatus(ctx context.Context, id string) (domainorder.Order, error) {
	order, err := s.Repo.GetByID(ctx, id)
	if err != nil {
		return domainorder.Order{}, err
	}
	if isFinalStatus(order.Status) {
		return order, ErrOrderAlreadyFinal
	}

	client, err := s.providerClient()
	if err != nil {
		return order, err
	}

	providerResp, err := client.GetOrderStatus(ctx, order.RefID)
	if err != nil {
		return order, err
	}
	if providerResp == nil {
		return order, ErrProviderEmptyResponse
	}

	return s.applyStatusResponse(ctx, order, providerResp)
}

func (s *OrderService) providerClient() (interfaceprovider.Client, error) {
	if s.ProviderFactory == nil {
		return nil, ErrProviderClientUnavailable
	}
	return s.ProviderFactory()
}

func (s *OrderService) applyCreateOrderResponse(ctx context.Context, order domainorder.Order, providerResp *h2h.CreateOrderResponse) (domainorder.Order, error) {
	oldStatus := order.Status
	newStatus := normalizeProviderStatus(providerResp.ProviderStatus)
	if newStatus == "" {
		newStatus = domainorder.StatusProcessing
	}

	order.Status = newStatus
	if charge := utils.StringOrDefault(providerResp.Charge.String(), "0"); charge != "0" {
		order.ProviderCharge = money.NormalizeOrZero(charge)
		order.Profit = money.SubOrZero(order.Amount, order.ProviderCharge)
	}
	order.StartCount = utils.Int64PtrFromString(providerResp.StartCount.String())
	order.Remains = utils.Int64PtrFromString(providerResp.Remains.String())
	order.ProviderResponse = providerResp.Raw
	order.UpdatedAt = new(time.Now())

	log := domainorder.OrderStatusLog{
		OldStatus:        oldStatus,
		NewStatus:        order.Status,
		ProviderStatus:   providerResp.ProviderStatus,
		ProviderResponse: providerResp.Raw,
	}
	if oldStatus == order.Status {
		log.NewStatus = ""
	}

	if err := s.Repo.UpdateWithStatusLog(ctx, order, log); err != nil {
		return domainorder.Order{}, err
	}
	if order.Status == domainorder.StatusFailed {
		_ = s.refundWallet(ctx, order, "provider failed order "+order.RefID)
	}

	return order, nil
}

func (s *OrderService) applyStatusResponse(ctx context.Context, order domainorder.Order, providerResp *h2h.OrderStatusResponse) (domainorder.Order, error) {
	oldStatus := order.Status
	newStatus := normalizeProviderStatus(providerResp.ProviderStatus)
	if newStatus == "" {
		newStatus = order.Status
	}

	order.Status = newStatus
	if providerResp.Charge.String() != "" {
		order.ProviderCharge = money.NormalizeOrZero(providerResp.Charge.String())
		order.Profit = money.SubOrZero(order.Amount, order.ProviderCharge)
	}
	if startCount := utils.Int64PtrFromString(providerResp.StartCount.String()); startCount != nil {
		order.StartCount = startCount
	}
	if remains := utils.Int64PtrFromString(providerResp.Remains.String()); remains != nil {
		order.Remains = remains
	}
	order.ProviderResponse = providerResp.Raw
	order.UpdatedAt = new(time.Now())

	log := domainorder.OrderStatusLog{
		OldStatus:        oldStatus,
		NewStatus:        order.Status,
		ProviderStatus:   providerResp.ProviderStatus,
		ProviderResponse: providerResp.Raw,
	}
	if oldStatus == order.Status {
		log.NewStatus = ""
	}

	if err := s.Repo.UpdateWithStatusLog(ctx, order, log); err != nil {
		return domainorder.Order{}, err
	}
	if order.Status == domainorder.StatusFailed && oldStatus != domainorder.StatusFailed {
		_ = s.refundWallet(ctx, order, "provider failed order "+order.RefID)
	}

	return order, nil
}

func (s *OrderService) failAndRefund(ctx context.Context, order domainorder.Order, cause error, providerResponse json.RawMessage) error {
	if err := s.markOrderFailed(ctx, order, cause, providerResponse); err != nil {
		return err
	}
	return s.refundWallet(ctx, order, "provider create order failed "+order.RefID)
}

func (s *OrderService) refundWallet(ctx context.Context, order domainorder.Order, description string) error {
	refunded, err := s.Repo.RefundWalletForOrder(ctx, order, order.Amount, description)
	if err != nil {
		s.writeRefundAudit(ctx, order, domainaudit.StatusFailed, description, err)
		return err
	}
	if refunded {
		s.writeRefundAudit(ctx, order, domainaudit.StatusSuccess, description, nil)
	}
	return nil
}

func (s *OrderService) writeRefundAudit(ctx context.Context, order domainorder.Order, status string, message string, err error) {
	if s.AuditService == nil {
		return
	}
	actorID := ""
	if order.CreatedBy != nil {
		actorID = *order.CreatedBy
	}
	errMessage := ""
	if err != nil {
		errMessage = err.Error()
	}
	_ = s.AuditService.Store(ctx, domainaudit.AuditEvent{
		ActorUserID:  actorID,
		Action:       domainaudit.ActionCreate,
		Resource:     "wallet_refund",
		ResourceID:   order.Id,
		Status:       status,
		Message:      message,
		ErrorMessage: errMessage,
		AfterData: map[string]interface{}{
			"order_id": order.Id,
			"ref_id":   order.RefID,
			"amount":   order.Amount,
		},
	})
}

func (s *OrderService) markOrderFailed(ctx context.Context, order domainorder.Order, cause error, providerResponse json.RawMessage) error {
	oldStatus := order.Status
	order.Status = domainorder.StatusFailed
	order.Metadata = utils.MustJSON(map[string]string{
		"source":        "api",
		"error_message": cause.Error(),
	})
	order.ProviderResponse = providerResponse
	order.UpdatedAt = new(time.Now())

	return s.Repo.UpdateWithStatusLog(ctx, order, domainorder.OrderStatusLog{
		OldStatus:        oldStatus,
		NewStatus:        domainorder.StatusFailed,
		ProviderStatus:   domainorder.StatusFailed,
		ProviderResponse: providerResponse,
	})
}

func (s *OrderService) calculatePrice(ctx context.Context, productType string, providerPrice string) (string, string, string, error) {
	markupPercent := "0"
	if s.ConfigService != nil {
		productKey := "pricing.product_markup_percent." + productType
		if value, err := s.ConfigService.GetString(ctx, productKey, ""); err != nil {
			return "", "", "", err
		} else if strings.TrimSpace(value) != "" {
			markupPercent = value
		} else if value, err := s.ConfigService.GetString(ctx, "pricing.default_markup_percent", "0"); err != nil {
			return "", "", "", err
		} else {
			markupPercent = value
		}
	}

	return money.MarkupAmount(providerPrice, markupPercent)
}

func IsPublicError(err error) bool {
	return errors.Is(err, ErrInvalidOrderRequest) ||
		errors.Is(err, ErrInactiveService) ||
		errors.Is(err, ErrUnsupportedService) ||
		errors.Is(err, ErrQuantityBelowMinimum) ||
		errors.Is(err, ErrQuantityAboveMaximum) ||
		errors.Is(err, ErrOrderAlreadyFinal) ||
		errors.Is(err, domainwallet.ErrInactiveWallet) ||
		errors.Is(err, domainwallet.ErrInsufficientBalance) ||
		errors.Is(err, domainwallet.ErrInvalidAmount)
}

func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

var _ interfaceorder.ServiceOrderInterface = (*OrderService)(nil)
