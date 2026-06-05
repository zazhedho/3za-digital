package serviceprovider

import (
	"context"
	"encoding/json"
	"testing"

	domainprovider "3za-digital/internal/domain/provider"
	"3za-digital/internal/integrations/h2h"
	interfaceprovider "3za-digital/internal/interfaces/provider"
	"3za-digital/pkg/filter"
	"3za-digital/utils"
)

func TestProviderServiceGetH2HBalanceUsesNestedBalance(t *testing.T) {
	repo := &mockProviderRepo{}
	service := NewProviderService(repo, func() (interfaceprovider.Client, error) {
		return &mockProviderClient{
			balance: &h2h.BalanceResponse{
				Status:  true,
				Message: "success",
				Balance: utils.FlexibleNumber("3409"),
				Raw:     json.RawMessage(`{"status":true,"data":{"balance":3409}}`),
			},
		}, nil
	})

	snapshot, err := service.GetH2HBalance(context.Background())
	if err != nil {
		t.Fatalf("GetH2HBalance returned error: %v", err)
	}
	if snapshot.Balance != "3409.00" {
		t.Fatalf("expected balance 3409.00, got %s", snapshot.Balance)
	}
	if repo.snapshot.Balance != "3409.00" {
		t.Fatalf("expected stored balance 3409.00, got %s", repo.snapshot.Balance)
	}
}

type mockProviderClient struct {
	balance *h2h.BalanceResponse
	err     error
}

func (m *mockProviderClient) GetBalance(ctx context.Context) (*h2h.BalanceResponse, error) {
	return m.balance, m.err
}

func (m *mockProviderClient) GetPriceList(ctx context.Context, req h2h.PriceListRequest) (*h2h.PriceListResponse, error) {
	return nil, nil
}

func (m *mockProviderClient) CreateOrder(ctx context.Context, req h2h.CreateOrderRequest) (*h2h.CreateOrderResponse, error) {
	return nil, nil
}

func (m *mockProviderClient) GetOrderStatus(ctx context.Context, refID string) (*h2h.OrderStatusResponse, error) {
	return nil, nil
}

type mockProviderRepo struct {
	snapshot domainprovider.BalanceSnapshot
}

func (m *mockProviderRepo) StoreBalanceSnapshot(ctx context.Context, snapshot domainprovider.BalanceSnapshot) error {
	m.snapshot = snapshot
	return nil
}

func (m *mockProviderRepo) StoreAPILog(ctx context.Context, log domainprovider.APILog) error {
	return nil
}

func (m *mockProviderRepo) GetAPILogs(ctx context.Context, params filter.BaseParams) ([]domainprovider.APILog, int64, error) {
	return nil, 0, nil
}
