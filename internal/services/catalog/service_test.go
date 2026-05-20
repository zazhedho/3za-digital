package servicecatalog

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	domaincatalog "3za-digital/internal/domain/catalog"
	"3za-digital/internal/dto"
	"3za-digital/internal/integrations/h2h"
	interfaceprovider "3za-digital/internal/interfaces/provider"
	"3za-digital/pkg/filter"
)

func TestCatalogServiceSyncSMM(t *testing.T) {
	repo := &mockCatalogRepo{}
	client := &mockProviderClient{
		response: &h2h.PriceListResponse{
			Status: true,
			Total:  1,
			Services: []h2h.Service{
				mustH2HService(t, map[string]interface{}{
					"id":           "1001",
					"name":         "Instagram Followers",
					"category":     "Followers",
					"platform":     "instagram",
					"price_per_1k": "1000",
					"min":          "10",
					"max":          "10000",
					"status":       "active",
				}),
			},
		},
	}
	service := NewCatalogService(repo, func() (interfaceprovider.Client, error) {
		return client, nil
	})

	result, err := service.Sync(context.Background(), domaincatalog.ProductTypeSMM, dto.SyncCatalogRequest{
		Platform: "instagram",
	})
	if err != nil {
		t.Fatalf("Sync returned error: %v", err)
	}

	if result.Synced != 1 {
		t.Fatalf("expected synced 1, got %d", result.Synced)
	}
	if client.req.Type != domaincatalog.ProductTypeSMM {
		t.Fatalf("expected product type smm, got %s", client.req.Type)
	}
	if len(repo.upserted) != 1 {
		t.Fatalf("expected 1 upserted service, got %d", len(repo.upserted))
	}

	synced := repo.upserted[0]
	if synced.Provider != domaincatalog.ProviderH2H {
		t.Fatalf("expected provider h2h, got %s", synced.Provider)
	}
	if synced.ProviderServiceID != "1001" {
		t.Fatalf("expected provider service id 1001, got %s", synced.ProviderServiceID)
	}
	if synced.Price != "1000" {
		t.Fatalf("expected price 1000, got %s", synced.Price)
	}
	if synced.MinQuantity == nil || *synced.MinQuantity != 10 {
		t.Fatalf("expected min quantity 10, got %#v", synced.MinQuantity)
	}
	if !synced.IsActive {
		t.Fatal("expected active service")
	}
}

func TestCatalogServiceSyncRejectsUnsupportedProductType(t *testing.T) {
	service := NewCatalogService(&mockCatalogRepo{}, func() (interfaceprovider.Client, error) {
		return &mockProviderClient{}, nil
	})

	_, err := service.Sync(context.Background(), "unknown", dto.SyncCatalogRequest{})
	if !errors.Is(err, ErrUnsupportedProductType) {
		t.Fatalf("expected ErrUnsupportedProductType, got %v", err)
	}
}

type mockProviderClient struct {
	req      h2h.PriceListRequest
	response *h2h.PriceListResponse
	err      error
}

func (m *mockProviderClient) GetPriceList(ctx context.Context, req h2h.PriceListRequest) (*h2h.PriceListResponse, error) {
	m.req = req
	return m.response, m.err
}

func (m *mockProviderClient) GetBalance(ctx context.Context) (*h2h.BalanceResponse, error) {
	return nil, nil
}

func (m *mockProviderClient) CreateOrder(ctx context.Context, req h2h.CreateOrderRequest) (*h2h.CreateOrderResponse, error) {
	return nil, nil
}

func (m *mockProviderClient) GetOrderStatus(ctx context.Context, refID string) (*h2h.OrderStatusResponse, error) {
	return nil, nil
}

type mockCatalogRepo struct {
	upserted []domaincatalog.ProviderService
}

func (m *mockCatalogRepo) Store(ctx context.Context, data domaincatalog.ProviderService) error {
	return nil
}

func (m *mockCatalogRepo) GetByID(ctx context.Context, id string) (domaincatalog.ProviderService, error) {
	return domaincatalog.ProviderService{}, nil
}

func (m *mockCatalogRepo) GetAll(ctx context.Context, params filter.BaseParams) ([]domaincatalog.ProviderService, int64, error) {
	return nil, 0, nil
}

func (m *mockCatalogRepo) Update(ctx context.Context, data domaincatalog.ProviderService) error {
	return nil
}

func (m *mockCatalogRepo) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *mockCatalogRepo) UpsertServices(ctx context.Context, services []domaincatalog.ProviderService) error {
	m.upserted = services
	return nil
}

func mustH2HService(t *testing.T, payload map[string]interface{}) h2h.Service {
	t.Helper()

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal service: %v", err)
	}

	var service h2h.Service
	if err := json.Unmarshal(body, &service); err != nil {
		t.Fatalf("unmarshal service: %v", err)
	}
	service.Raw = body
	return service
}
