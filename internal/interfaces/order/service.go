package interfaceorder

import (
	"context"

	domainorder "3za-digital/internal/domain/order"
	"3za-digital/internal/dto"
	"3za-digital/pkg/filter"
)

type ServiceOrderInterface interface {
	GetAll(ctx context.Context, params filter.BaseParams) ([]domainorder.Order, int64, error)
	GetByID(ctx context.Context, id string) (domainorder.Order, error)
	GetStatusLogs(ctx context.Context, orderID string) ([]domainorder.OrderStatusLog, error)
	CreateOrder(ctx context.Context, productType string, req dto.CreateOrderRequest, createdBy string) (domainorder.Order, error)
	RefreshStatus(ctx context.Context, id string) (domainorder.Order, error)
}
