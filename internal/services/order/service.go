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
	"3za-digital/internal/dto"
	"3za-digital/internal/integrations/h2h"
	interfaceorder "3za-digital/internal/interfaces/order"
	interfaceprovider "3za-digital/internal/interfaces/provider"
	"3za-digital/pkg/filter"
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
}

func NewOrderService(repo interfaceorder.RepoOrderInterface, providerFactory interfaceprovider.ClientFactory) *OrderService {
	return &OrderService{
		Repo:            repo,
		ProviderFactory: providerFactory,
	}
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

func (s *OrderService) CreateSMMOrder(ctx context.Context, req dto.CreateSMMOrderRequest, createdBy string) (domainorder.Order, error) {
	req.Target = strings.TrimSpace(req.Target)
	if req.Target == "" || req.Quantity <= 0 || strings.TrimSpace(req.ServiceID) == "" {
		return domainorder.Order{}, ErrInvalidOrderRequest
	}

	service, err := s.Repo.GetServiceByID(ctx, req.ServiceID)
	if err != nil {
		return domainorder.Order{}, err
	}
	if err := validateSMMService(service, req.Quantity); err != nil {
		return domainorder.Order{}, err
	}

	now := time.Now()
	serviceID := service.Id
	quantity := req.Quantity
	order := domainorder.Order{
		Id:                utils.CreateUUID(),
		Provider:          domaincatalog.ProviderH2H,
		ProductType:       domaincatalog.ProductTypeSMM,
		RefID:             generateRefID(domaincatalog.ProductTypeSMM),
		ServiceID:         &serviceID,
		ProviderServiceID: service.ProviderServiceID,
		Target:            req.Target,
		Quantity:          &quantity,
		Status:            domainorder.StatusPending,
		Amount:            "0",
		ProviderCharge:    "0",
		Profit:            "0",
		Metadata:          mustJSON(map[string]string{"source": "api"}),
		ProviderResponse:  json.RawMessage(`{}`),
		CreatedBy:         normalizeOptionalString(createdBy),
		UpdatedAt:         &now,
	}

	order, err = s.Repo.CreateWithStatusLog(ctx, order, domainorder.OrderStatusLog{
		Id:               utils.CreateUUID(),
		NewStatus:        domainorder.StatusPending,
		ProviderResponse: json.RawMessage(`{}`),
	})
	if err != nil {
		return domainorder.Order{}, err
	}

	client, err := s.providerClient()
	if err != nil {
		_ = s.markOrderFailed(ctx, order, err, json.RawMessage(`{}`))
		return order, err
	}

	providerResp, err := client.CreateOrder(ctx, h2h.CreateOrderRequest{
		Type:        h2h.ProductTypeSMM,
		ServiceCode: service.ProviderServiceID,
		Target:      req.Target,
		Quantity:    int(req.Quantity),
		RefID:       order.RefID,
	})
	if err != nil {
		_ = s.markOrderFailed(ctx, order, err, json.RawMessage(`{}`))
		return order, err
	}
	if providerResp == nil {
		_ = s.markOrderFailed(ctx, order, ErrProviderEmptyResponse, json.RawMessage(`{}`))
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
	now := time.Now()
	newStatus := normalizeProviderStatus(providerResp.ProviderStatus)
	if newStatus == "" {
		newStatus = domainorder.StatusProcessing
	}

	order.Status = newStatus
	order.ProviderCharge = defaultNumber(providerResp.Charge.String())
	order.Amount = order.ProviderCharge
	order.Profit = "0"
	order.StartCount = optionalInt64(providerResp.StartCount.String())
	order.Remains = optionalInt64(providerResp.Remains.String())
	order.ProviderResponse = providerResp.Raw
	order.UpdatedAt = &now

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

	return order, nil
}

func (s *OrderService) applyStatusResponse(ctx context.Context, order domainorder.Order, providerResp *h2h.OrderStatusResponse) (domainorder.Order, error) {
	oldStatus := order.Status
	now := time.Now()
	newStatus := normalizeProviderStatus(providerResp.ProviderStatus)
	if newStatus == "" {
		newStatus = order.Status
	}

	order.Status = newStatus
	if providerResp.Charge.String() != "" {
		order.ProviderCharge = providerResp.Charge.String()
		order.Amount = order.ProviderCharge
	}
	if startCount := optionalInt64(providerResp.StartCount.String()); startCount != nil {
		order.StartCount = startCount
	}
	if remains := optionalInt64(providerResp.Remains.String()); remains != nil {
		order.Remains = remains
	}
	order.ProviderResponse = providerResp.Raw
	order.UpdatedAt = &now

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

	return order, nil
}

func (s *OrderService) markOrderFailed(ctx context.Context, order domainorder.Order, cause error, providerResponse json.RawMessage) error {
	oldStatus := order.Status
	now := time.Now()
	order.Status = domainorder.StatusFailed
	order.Metadata = mustJSON(map[string]string{
		"source":        "api",
		"error_message": cause.Error(),
	})
	order.ProviderResponse = providerResponse
	order.UpdatedAt = &now

	return s.Repo.UpdateWithStatusLog(ctx, order, domainorder.OrderStatusLog{
		OldStatus:        oldStatus,
		NewStatus:        domainorder.StatusFailed,
		ProviderStatus:   domainorder.StatusFailed,
		ProviderResponse: providerResponse,
	})
}

func validateSMMService(service domaincatalog.ProviderService, quantity int64) error {
	if service.Provider != domaincatalog.ProviderH2H || service.ProductType != domaincatalog.ProductTypeSMM {
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
		errors.Is(err, ErrOrderAlreadyFinal)
}

func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

var _ interfaceorder.ServiceOrderInterface = (*OrderService)(nil)
