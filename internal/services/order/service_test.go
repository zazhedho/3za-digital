package serviceorder

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"3za-digital/internal/authscope"
	domaincatalog "3za-digital/internal/domain/catalog"
	domainorder "3za-digital/internal/domain/order"
	"3za-digital/internal/dto"
	"3za-digital/internal/integrations/h2h"
	interfaceprovider "3za-digital/internal/interfaces/provider"
	"3za-digital/pkg/filter"
	"3za-digital/utils"
)

func TestOrderServiceCreateOrderWithSMMType(t *testing.T) {
	minQty := int64(10)
	maxQty := int64(1000)
	repo := &mockOrderRepo{
		service: domaincatalog.ProviderService{
			Id:                "service-id",
			Provider:          domaincatalog.ProviderH2H,
			ProductType:       domaincatalog.ProductTypeSMM,
			ProviderServiceID: "1001",
			Name:              "Instagram Followers",
			Price:             "1200",
			MinQuantity:       &minQty,
			MaxQuantity:       &maxQty,
			IsActive:          true,
		},
	}
	provider := &mockOrderProvider{
		createResp: &h2h.CreateOrderResponse{
			Status:         true,
			RefID:          "ref",
			ProviderStatus: "processing",
			Charge:         utils.FlexibleNumber("1200"),
			StartCount:     utils.FlexibleNumber("10"),
			Remains:        utils.FlexibleNumber("90"),
			Raw:            json.RawMessage(`{"status":true}`),
		},
	}
	service := NewOrderService(repo, func() (interfaceprovider.Client, error) {
		return provider, nil
	})

	order, err := service.CreateOrder(orderTestActorContext("user-id"), domaincatalog.ProductTypeSMM, dto.CreateOrderRequest{
		ServiceID: "service-id",
		Target:    "https://instagram.com/3zadigital",
		Quantity:  100,
	})
	if err != nil {
		t.Fatalf("CreateOrder returned error: %v", err)
	}

	if provider.createReq.Type != h2h.ProductTypeSMM {
		t.Fatalf("expected provider type smm, got %s", provider.createReq.Type)
	}
	if provider.createReq.ServiceCode != "1001" {
		t.Fatalf("expected service code 1001, got %s", provider.createReq.ServiceCode)
	}
	if order.Status != domainorder.StatusProcessing {
		t.Fatalf("expected processing status, got %s", order.Status)
	}
	if order.ProviderCharge != "1200.00" {
		t.Fatalf("expected provider charge 1200.00, got %s", order.ProviderCharge)
	}
	if repo.created.Status != domainorder.StatusPending {
		t.Fatalf("expected initial pending order, got %s", repo.created.Status)
	}
	if repo.updated.Status != domainorder.StatusProcessing {
		t.Fatalf("expected updated processing order, got %s", repo.updated.Status)
	}
}

func TestOrderServiceCreateOrderProviderErrorKeepsProcessingWithoutRefund(t *testing.T) {
	repo := &mockOrderRepo{service: validSMMTestService()}
	providerErr := &h2h.APIError{
		Message: "provider timeout",
		Raw:     json.RawMessage(`{"status":false,"message":"timeout"}`),
	}
	service := NewOrderService(repo, func() (interfaceprovider.Client, error) {
		return &mockOrderProvider{createErr: providerErr}, nil
	})

	order, err := service.CreateOrder(orderTestActorContext("user-id"), domaincatalog.ProductTypeSMM, dto.CreateOrderRequest{
		ServiceID: "service-id",
		Target:    "https://instagram.com/3zadigital",
		Quantity:  100,
	})
	if !errors.Is(err, providerErr) {
		t.Fatalf("expected provider error, got %v", err)
	}
	if order.Status != domainorder.StatusPending {
		t.Fatalf("expected returned order still pending, got %s", order.Status)
	}
	if repo.updated.Status != domainorder.StatusProcessing {
		t.Fatalf("expected stored order processing, got %s", repo.updated.Status)
	}
	if repo.refundCount != 0 {
		t.Fatalf("expected no refund on unknown provider submission, got %d", repo.refundCount)
	}
}

func TestOrderServiceCreateOrderFailedProviderStatusWaitsForRefreshWithoutRefund(t *testing.T) {
	repo := &mockOrderRepo{service: validSMMTestService()}
	provider := &mockOrderProvider{
		createResp: &h2h.CreateOrderResponse{
			Status:         true,
			ProviderStatus: "failed",
			Raw:            json.RawMessage(`{"status":true,"data":{"status":"failed"}}`),
		},
	}
	service := NewOrderService(repo, func() (interfaceprovider.Client, error) {
		return provider, nil
	})

	order, err := service.CreateOrder(orderTestActorContext("user-id"), domaincatalog.ProductTypeSMM, dto.CreateOrderRequest{
		ServiceID: "service-id",
		Target:    "https://instagram.com/3zadigital",
		Quantity:  100,
	})
	if err != nil {
		t.Fatalf("CreateOrder returned error: %v", err)
	}
	if order.Status != domainorder.StatusProcessing {
		t.Fatalf("expected processing status until refresh, got %s", order.Status)
	}
	if repo.refundCount != 0 {
		t.Fatalf("expected no refund on create response status, got %d", repo.refundCount)
	}
}

func TestOrderServiceRefreshStatusFailedRefundsWallet(t *testing.T) {
	repo := &mockOrderRepo{
		created: domainorder.Order{
			Id:             "order-id",
			Provider:       domaincatalog.ProviderH2H,
			ProductType:    domaincatalog.ProductTypeSMM,
			RefID:          "SMM-REF",
			Status:         domainorder.StatusProcessing,
			Amount:         "1500.00",
			ProviderCharge: "1000.00",
			Profit:         "500.00",
			Metadata:       utils.EmptyJSON(),
			CreatedBy:      utils.StringPtrIfNotEmpty("user-id"),
		},
	}
	provider := &mockOrderProvider{
		statusResp: &h2h.OrderStatusResponse{
			Status:         true,
			ProviderStatus: "failed",
			Raw:            json.RawMessage(`{"status":true,"data":{"status":"failed"}}`),
		},
	}
	service := NewOrderService(repo, func() (interfaceprovider.Client, error) {
		return provider, nil
	})

	order, err := service.RefreshStatus(orderTestActorContext("user-id"), "order-id")
	if err != nil {
		t.Fatalf("RefreshStatus returned error: %v", err)
	}
	if order.Status != domainorder.StatusFailed {
		t.Fatalf("expected failed status, got %s", order.Status)
	}
	if repo.refundCount != 1 {
		t.Fatalf("expected one refund on final failed status, got %d", repo.refundCount)
	}
}

func TestOrderServiceCreateOrderRejectsBelowMinimum(t *testing.T) {
	minQty := int64(10)
	repo := &mockOrderRepo{
		service: domaincatalog.ProviderService{
			Id:                "service-id",
			Provider:          domaincatalog.ProviderH2H,
			ProductType:       domaincatalog.ProductTypeSMM,
			ProviderServiceID: "1001",
			MinQuantity:       &minQty,
			IsActive:          true,
		},
	}
	service := NewOrderService(repo, func() (interfaceprovider.Client, error) {
		return &mockOrderProvider{}, nil
	})

	_, err := service.CreateOrder(orderTestActorContext("user-id"), domaincatalog.ProductTypeSMM, dto.CreateOrderRequest{
		ServiceID: "service-id",
		Target:    "https://instagram.com/3zadigital",
		Quantity:  1,
	})
	if !errors.Is(err, ErrQuantityBelowMinimum) {
		t.Fatalf("expected ErrQuantityBelowMinimum, got %v", err)
	}
}

func TestOrderServiceCreateOrderRejectsInvalidSMMTargetURL(t *testing.T) {
	service := NewOrderService(&mockOrderRepo{}, func() (interfaceprovider.Client, error) {
		return &mockOrderProvider{}, nil
	})

	_, err := service.CreateOrder(orderTestActorContext("user-id"), domaincatalog.ProductTypeSMM, dto.CreateOrderRequest{
		ServiceID: "service-id",
		Target:    "instagram.com/3zadigital",
		Quantity:  100,
	})
	if !errors.Is(err, ErrInvalidOrderRequest) {
		t.Fatalf("expected ErrInvalidOrderRequest, got %v", err)
	}
}

func validSMMTestService() domaincatalog.ProviderService {
	minQty := int64(10)
	maxQty := int64(1000)
	return domaincatalog.ProviderService{
		Id:                "service-id",
		Provider:          domaincatalog.ProviderH2H,
		ProductType:       domaincatalog.ProductTypeSMM,
		ProviderServiceID: "1001",
		Name:              "Instagram Followers",
		Price:             "1200",
		MinQuantity:       &minQty,
		MaxQuantity:       &maxQty,
		IsActive:          true,
	}
}

func orderTestActorContext(userID string) context.Context {
	return authscope.WithContext(context.Background(), authscope.New(userID, "member", "member", nil))
}

type mockOrderProvider struct {
	createReq  h2h.CreateOrderRequest
	createResp *h2h.CreateOrderResponse
	createErr  error
	statusResp *h2h.OrderStatusResponse
	statusErr  error
}

func (m *mockOrderProvider) CreateOrder(ctx context.Context, req h2h.CreateOrderRequest) (*h2h.CreateOrderResponse, error) {
	m.createReq = req
	return m.createResp, m.createErr
}

func (m *mockOrderProvider) GetOrderStatus(ctx context.Context, refID string) (*h2h.OrderStatusResponse, error) {
	return m.statusResp, m.statusErr
}

func (m *mockOrderProvider) GetBalance(ctx context.Context) (*h2h.BalanceResponse, error) {
	return nil, nil
}

func (m *mockOrderProvider) GetPriceList(ctx context.Context, req h2h.PriceListRequest) (*h2h.PriceListResponse, error) {
	return nil, nil
}

type mockOrderRepo struct {
	service     domaincatalog.ProviderService
	created     domainorder.Order
	updated     domainorder.Order
	logs        []domainorder.OrderStatusLog
	refundCount int
}

func (m *mockOrderRepo) Store(ctx context.Context, data domainorder.Order) error {
	return nil
}

func (m *mockOrderRepo) GetByID(ctx context.Context, id string) (domainorder.Order, error) {
	if m.updated.Id != "" {
		return m.updated, nil
	}
	return m.created, nil
}

func (m *mockOrderRepo) GetAll(ctx context.Context, params filter.BaseParams) ([]domainorder.Order, int64, error) {
	return nil, 0, nil
}

func (m *mockOrderRepo) Update(ctx context.Context, data domainorder.Order) error {
	m.updated = data
	return nil
}

func (m *mockOrderRepo) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *mockOrderRepo) GetServiceByID(ctx context.Context, id string) (domaincatalog.ProviderService, error) {
	return m.service, nil
}

func (m *mockOrderRepo) GetStatusLogs(ctx context.Context, orderID string) ([]domainorder.OrderStatusLog, error) {
	return m.logs, nil
}

func (m *mockOrderRepo) CreateWithStatusLog(ctx context.Context, order domainorder.Order, log domainorder.OrderStatusLog) (domainorder.Order, error) {
	m.created = order
	m.logs = append(m.logs, log)
	return order, nil
}

func (m *mockOrderRepo) CreateWithStatusLogAndWalletDebit(ctx context.Context, order domainorder.Order, log domainorder.OrderStatusLog, userID string) (domainorder.Order, error) {
	m.created = order
	m.logs = append(m.logs, log)
	return order, nil
}

func (m *mockOrderRepo) UpdateWithStatusLog(ctx context.Context, order domainorder.Order, log domainorder.OrderStatusLog) error {
	m.updated = order
	if log.NewStatus != "" {
		m.logs = append(m.logs, log)
	}
	return nil
}

func (m *mockOrderRepo) RefundWalletForOrder(ctx context.Context, order domainorder.Order, amount string, description string) (bool, error) {
	m.refundCount++
	return true, nil
}
