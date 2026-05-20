package serviceorder

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	domaincatalog "3za-digital/internal/domain/catalog"
	domainorder "3za-digital/internal/domain/order"
	domainwallet "3za-digital/internal/domain/wallet"
	"3za-digital/internal/dto"
	"3za-digital/internal/integrations/h2h"
	interfaceappconfig "3za-digital/internal/interfaces/appconfig"
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
	productType = strings.ToLower(strings.TrimSpace(productType))

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
		Metadata:          mustJSON(map[string]string{"source": "api"}),
		ProviderResponse:  json.RawMessage(`{}`),
		CreatedBy:         normalizeOptionalString(createdBy),
		UpdatedAt:         new(time.Now()),
	}

	order, err = s.Repo.CreateWithStatusLogAndWalletDebit(ctx, order, domainorder.OrderStatusLog{
		Id:               utils.CreateUUID(),
		NewStatus:        domainorder.StatusPending,
		ProviderResponse: json.RawMessage(`{}`),
	}, createdBy)
	if err != nil {
		return domainorder.Order{}, err
	}

	client, err := s.providerClient()
	if err != nil {
		_ = s.failAndRefund(ctx, order, err, json.RawMessage(`{}`))
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
		_ = s.failAndRefund(ctx, order, err, json.RawMessage(`{}`))
		return order, err
	}
	if providerResp == nil {
		_ = s.failAndRefund(ctx, order, ErrProviderEmptyResponse, json.RawMessage(`{}`))
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
	if charge := defaultNumber(providerResp.Charge.String()); charge != "0" {
		order.ProviderCharge = normalizeMoney(charge)
		order.Profit = subtractMoney(order.Amount, order.ProviderCharge)
	}
	order.StartCount = optionalInt64(providerResp.StartCount.String())
	order.Remains = optionalInt64(providerResp.Remains.String())
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
		_ = s.Repo.RefundWalletForOrder(ctx, order, order.Amount, "provider failed order "+order.RefID)
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
		order.ProviderCharge = normalizeMoney(providerResp.Charge.String())
		order.Profit = subtractMoney(order.Amount, order.ProviderCharge)
	}
	if startCount := optionalInt64(providerResp.StartCount.String()); startCount != nil {
		order.StartCount = startCount
	}
	if remains := optionalInt64(providerResp.Remains.String()); remains != nil {
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
		_ = s.Repo.RefundWalletForOrder(ctx, order, order.Amount, "provider failed order "+order.RefID)
	}

	return order, nil
}

func (s *OrderService) failAndRefund(ctx context.Context, order domainorder.Order, cause error, providerResponse json.RawMessage) error {
	if err := s.markOrderFailed(ctx, order, cause, providerResponse); err != nil {
		return err
	}
	return s.Repo.RefundWalletForOrder(ctx, order, order.Amount, "provider create order failed "+order.RefID)
}

func (s *OrderService) markOrderFailed(ctx context.Context, order domainorder.Order, cause error, providerResponse json.RawMessage) error {
	oldStatus := order.Status
	order.Status = domainorder.StatusFailed
	order.Metadata = mustJSON(map[string]string{
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

func validateService(service domaincatalog.ProviderService, productType string, quantity int64) error {
	if service.Provider != domaincatalog.ProviderH2H || service.ProductType != productType {
		return ErrUnsupportedService
	}
	if !service.IsActive {
		return ErrInactiveService
	}
	if service.MinQuantity != nil && quantity < *service.MinQuantity {
		return ErrQuantityBelowMinimum
	}
	if service.MaxQuantity != nil && quantity > *service.MaxQuantity {
		return ErrQuantityAboveMaximum
	}
	return nil
}

func normalizeProviderStatus(status string) string {
	status = strings.ToLower(strings.TrimSpace(status))
	status = strings.ReplaceAll(status, " ", "_")
	switch status {
	case domainorder.StatusPending:
		return domainorder.StatusPending
	case "process", "in_progress", domainorder.StatusProcessing:
		return domainorder.StatusProcessing
	case "success", "done", domainorder.StatusCompleted:
		return domainorder.StatusCompleted
	case domainorder.StatusPartial:
		return domainorder.StatusPartial
	case "error", "reject", "rejected", domainorder.StatusFailed:
		return domainorder.StatusFailed
	case "cancel", domainorder.StatusCancelled:
		return domainorder.StatusCancelled
	default:
		return ""
	}
}

func isFinalStatus(status string) bool {
	switch status {
	case domainorder.StatusCompleted, domainorder.StatusPartial, domainorder.StatusFailed, domainorder.StatusCancelled:
		return true
	default:
		return false
	}
}

func generateRefID(productType string) string {
	id := strings.ToUpper(strings.ReplaceAll(utils.CreateUUID(), "-", ""))
	return fmt.Sprintf("%s-%s", strings.ToUpper(productType), id)
}

func normalizeOptionalString(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func optionalInt64(value string) *int64 {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return nil
	}
	return &parsed
}

func defaultNumber(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "0"
	}
	return value
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

func normalizeMoney(value string) string {
	normalized, err := money.Normalize(value)
	if err != nil {
		return "0.00"
	}
	return normalized
}

func subtractMoney(left string, right string) string {
	result, err := money.Sub(left, right)
	if err != nil {
		return "0.00"
	}
	return result
}

func mustJSON(value interface{}) json.RawMessage {
	body, err := json.Marshal(value)
	if err != nil {
		return json.RawMessage(`{}`)
	}
	return body
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
