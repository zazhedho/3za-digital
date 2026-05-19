package interfaceprovider

import (
	"context"

	"3za-digital/internal/integrations/h2h"
)

type Client interface {
	GetBalance(ctx context.Context) (*h2h.BalanceResponse, error)
	GetPriceList(ctx context.Context, req h2h.PriceListRequest) (*h2h.PriceListResponse, error)
	CreateOrder(ctx context.Context, req h2h.CreateOrderRequest) (*h2h.CreateOrderResponse, error)
	GetOrderStatus(ctx context.Context, refID string) (*h2h.OrderStatusResponse, error)
}

type ClientFactory func() (Client, error)
