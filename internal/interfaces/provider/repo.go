package interfaceprovider

import (
	"context"

	domainprovider "3za-digital/internal/domain/provider"
	"3za-digital/pkg/filter"
)

type RepoProviderInterface interface {
	StoreBalanceSnapshot(ctx context.Context, snapshot domainprovider.BalanceSnapshot) error
	StoreAPILog(ctx context.Context, log domainprovider.APILog) error
	GetAPILogs(ctx context.Context, params filter.BaseParams) ([]domainprovider.APILog, int64, error)
}
