package servicecatalog

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	domainappconfig "3za-digital/internal/domain/appconfig"
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

func TestCatalogServiceEnsureFreshSyncsWhenStale(t *testing.T) {
	stale := time.Now().Add(-25 * time.Hour)
	repo := &mockCatalogRepo{
		list: []domaincatalog.ProviderService{
			{Id: "svc-1", ProductType: domaincatalog.ProductTypeSMM, Provider: domaincatalog.ProviderH2H, SyncedAt: &stale},
		},
		total: 1,
	}
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

	if err := service.EnsureFresh(context.Background(), domaincatalog.ProductTypeSMM); err != nil {
		t.Fatalf("EnsureFresh returned error: %v", err)
	}
	if len(repo.upserted) != 1 {
		t.Fatalf("expected sync to run once, got %d", len(repo.upserted))
	}
	if client.req.Type != domaincatalog.ProductTypeSMM {
		t.Fatalf("expected lazy sync to request smm, got %s", client.req.Type)
	}
}

func TestCatalogServiceEnsureFreshSkipsWhenFresh(t *testing.T) {
	fresh := time.Now()
	repo := &mockCatalogRepo{
		list: []domaincatalog.ProviderService{
			{Id: "svc-1", ProductType: domaincatalog.ProductTypeSMM, Provider: domaincatalog.ProviderH2H, SyncedAt: &fresh},
		},
		total: 1,
	}
	client := &mockProviderClient{
		response: &h2h.PriceListResponse{
			Status: true,
			Total:  1,
		},
	}
	service := NewCatalogService(repo, func() (interfaceprovider.Client, error) {
		return client, nil
	})

	if err := service.EnsureFresh(context.Background(), domaincatalog.ProductTypeSMM); err != nil {
		t.Fatalf("EnsureFresh returned error: %v", err)
	}
	if len(repo.upserted) != 0 {
		t.Fatalf("expected no sync when fresh, got %d", len(repo.upserted))
	}
	if client.req.Type != "" {
		t.Fatalf("expected no provider call when fresh, got %s", client.req.Type)
	}
}

func TestCatalogServiceGetAllAppliesMarkupConfig(t *testing.T) {
	service := NewCatalogService(
		&mockCatalogRepo{
			list: []domaincatalog.ProviderService{
				{
					Id:          "svc-1",
					ProductType: domaincatalog.ProductTypeSMM,
					Price:       "1000.00",
					RawResponse: []byte(`{"price_per_1k":"1000"}`),
				},
				{
					Id:          "svc-2",
					ProductType: domaincatalog.ProductTypeSMM,
					Price:       "2000.00",
				},
			},
			total: 2,
		},
		nil,
		&mockCatalogConfigService{values: map[string]string{
			"pricing.product_markup_percent.smm": "5",
		}},
	)

	services, total, err := service.GetAll(context.Background(), filter.BaseParams{})
	if err != nil {
		t.Fatalf("GetAll returned error: %v", err)
	}
	if total != 2 || len(services) != 2 {
		t.Fatalf("expected two services, total=%d len=%d", total, len(services))
	}
	if services[0].Price != "1050.00" {
		t.Fatalf("expected marked up price 1050.00, got %s", services[0].Price)
	}
	if services[1].Price != "2100.00" {
		t.Fatalf("expected marked up price 2100.00, got %s", services[1].Price)
	}
	if len(services[0].RawResponse) != 0 {
		t.Fatalf("expected raw provider response to be hidden, got %s", string(services[0].RawResponse))
	}
}

func TestCatalogServiceGetAllReadsMarkupConfigOncePerProductType(t *testing.T) {
	config := &mockCatalogConfigService{values: map[string]string{
		"pricing.product_markup_percent.smm": "5",
	}}
	service := NewCatalogService(
		&mockCatalogRepo{
			list: []domaincatalog.ProviderService{
				{Id: "svc-1", ProductType: domaincatalog.ProductTypeSMM, Price: "1000.00"},
				{Id: "svc-2", ProductType: domaincatalog.ProductTypeSMM, Price: "2000.00"},
			},
			total: 2,
		},
		nil,
		config,
	)

	if _, _, err := service.GetAll(context.Background(), filter.BaseParams{}); err != nil {
		t.Fatalf("GetAll returned error: %v", err)
	}
	if config.calls["pricing.product_markup_percent.smm"] != 1 {
		t.Fatalf("expected one smm markup config read, got %d", config.calls["pricing.product_markup_percent.smm"])
	}
	if config.calls["pricing.default_markup_percent"] != 0 {
		t.Fatalf("expected no default markup config read, got %d", config.calls["pricing.default_markup_percent"])
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
	list     []domaincatalog.ProviderService
	total    int64
}

func (m *mockCatalogRepo) Store(ctx context.Context, data domaincatalog.ProviderService) error {
	return nil
}

func (m *mockCatalogRepo) GetByID(ctx context.Context, id string) (domaincatalog.ProviderService, error) {
	return domaincatalog.ProviderService{}, nil
}

func (m *mockCatalogRepo) GetAll(ctx context.Context, params filter.BaseParams) ([]domaincatalog.ProviderService, int64, error) {
	return m.list, m.total, nil
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

type mockCatalogConfigService struct {
	values map[string]string
	calls  map[string]int
}

func (m *mockCatalogConfigService) GetAll(ctx context.Context, params filter.BaseParams) ([]domainappconfig.AppConfig, int64, error) {
	return nil, 0, nil
}

func (m *mockCatalogConfigService) GetByID(ctx context.Context, id string) (domainappconfig.AppConfig, error) {
	return domainappconfig.AppConfig{}, nil
}

func (m *mockCatalogConfigService) GetByKey(ctx context.Context, configKey string) (domainappconfig.AppConfig, error) {
	return domainappconfig.AppConfig{}, nil
}

func (m *mockCatalogConfigService) Update(ctx context.Context, id string, req dto.UpdateAppConfig) (domainappconfig.AppConfig, error) {
	return domainappconfig.AppConfig{}, nil
}

func (m *mockCatalogConfigService) GetString(ctx context.Context, configKey string, fallback string) (string, error) {
	if m.calls == nil {
		m.calls = make(map[string]int)
	}
	m.calls[configKey]++
	if value, ok := m.values[configKey]; ok {
		return value, nil
	}
	return fallback, nil
}

func (m *mockCatalogConfigService) GetBool(ctx context.Context, configKey string, fallback bool) (bool, error) {
	return fallback, nil
}

func (m *mockCatalogConfigService) GetInt(ctx context.Context, configKey string, fallback int) (int, error) {
	return fallback, nil
}

func (m *mockCatalogConfigService) GetDuration(ctx context.Context, configKey string, fallback time.Duration) (time.Duration, error) {
	return fallback, nil
}

func (m *mockCatalogConfigService) DecodeJSON(ctx context.Context, configKey string, target interface{}) error {
	return nil
}

func (m *mockCatalogConfigService) IsEnabled(ctx context.Context, configKey string, fallback bool) (bool, error) {
	return fallback, nil
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
