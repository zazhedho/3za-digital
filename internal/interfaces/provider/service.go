package interfaceprovider

import (
	"context"

	domainprovider "3za-digital/internal/domain/provider"
	"3za-digital/pkg/filter"
)

type ServiceProviderInterface interface {
	GetH2HBalance(ctx context.Context) (domainprovider.BalanceSnapshot, error)
	GetAPILogs(ctx context.Context, params filter.BaseParams) ([]domainprovider.APILog, int64, error)
}
