package interfaceorder

import (
	"context"

	domaincatalog "3za-digital/internal/domain/catalog"
	domainorder "3za-digital/internal/domain/order"
	interfacegeneric "3za-digital/internal/interfaces/generic"
	"3za-digital/pkg/filter"
)

type RepoOrderInterface interface {
	interfacegeneric.GenericRepository[domainorder.Order]

	GetAll(ctx context.Context, params filter.BaseParams) ([]domainorder.Order, int64, error)
	GetServiceByID(ctx context.Context, id string) (domaincatalog.ProviderService, error)
	GetStatusLogs(ctx context.Context, orderID string) ([]domainorder.OrderStatusLog, error)
	CreateWithStatusLog(ctx context.Context, order domainorder.Order, log domainorder.OrderStatusLog) (domainorder.Order, error)
	CreateWithStatusLogAndWalletDebit(ctx context.Context, order domainorder.Order, log domainorder.OrderStatusLog, userID string) (domainorder.Order, error)
	UpdateWithStatusLog(ctx context.Context, order domainorder.Order, log domainorder.OrderStatusLog) error
	RefundWalletForOrder(ctx context.Context, order domainorder.Order, amount string, description string) (bool, error)
}
