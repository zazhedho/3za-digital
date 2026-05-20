package serviceorder

import (
	"context"
	"encoding/json"
	"testing"

	domaincatalog "3za-digital/internal/domain/catalog"
	domainorder "3za-digital/internal/domain/order"
	"3za-digital/internal/dto"
	"3za-digital/internal/integrations/h2h"
	interfaceprovider "3za-digital/internal/interfaces/provider"
	"3za-digital/pkg/filter"
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
			Charge:         h2h.FlexibleNumber("1200"),
			StartCount:     h2h.FlexibleNumber("10"),
			Remains:        h2h.FlexibleNumber("90"),
			Raw:            json.RawMessage(`{"status":true}`),
		},
	}
	service := NewOrderService(repo, func() (interfaceprovider.Client, error) {
		return provider, nil
	})

	order, err := service.CreateOrder(context.Background(), domaincatalog.ProductTypeSMM, dto.CreateOrderRequest{
		ServiceID: "service-id",
		Target:    "https://instagram.com/3zadigital",
		Quantity:  100,
	}, "user-id")
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

	_, err := service.CreateOrder(context.Background(), domaincatalog.ProductTypeSMM, dto.CreateOrderRequest{
		ServiceID: "service-id",
		Target:    "target",
		Quantity:  1,
	}, "user-id")
	if err != ErrQuantityBelowMinimum {
		t.Fatalf("expected ErrQuantityBelowMinimum, got %v", err)
	}
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
	service domaincatalog.ProviderService
	created domainorder.Order
	updated domainorder.Order
	logs    []domainorder.OrderStatusLog
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

func (m *mockOrderRepo) RefundWalletForOrder(ctx context.Context, order domainorder.Order, amount string, description string) error {
	return nil
}
