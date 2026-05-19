package repositoryprovider

import (
	"context"

	domainprovider "3za-digital/internal/domain/provider"
	interfaceprovider "3za-digital/internal/interfaces/provider"
	repositorygeneric "3za-digital/internal/repositories/generic"
	"3za-digital/pkg/filter"
	"3za-digital/utils"

	"gorm.io/gorm"
)

type repo struct {
	db *gorm.DB
}

func NewProviderRepo(db *gorm.DB) interfaceprovider.RepoProviderInterface {
	return &repo{db: db}
}

func (r *repo) StoreBalanceSnapshot(ctx context.Context, snapshot domainprovider.BalanceSnapshot) error {
	if snapshot.Id == "" {
		snapshot.Id = utils.CreateUUID()
	}
	return r.db.WithContext(ctx).Create(&snapshot).Error
}

func (r *repo) StoreAPILog(ctx context.Context, log domainprovider.APILog) error {
	if log.Id == "" {
		log.Id = utils.CreateUUID()
	}
	return r.db.WithContext(ctx).Create(&log).Error
}

func (r *repo) GetAPILogs(ctx context.Context, params filter.BaseParams) ([]domainprovider.APILog, int64, error) {
	generic := repositorygeneric.New[domainprovider.APILog](r.db)
	return generic.GetAll(ctx, params, repositorygeneric.QueryOptions{
		Search: repositorygeneric.BuildSearchFunc("endpoint", "request_ref", "error_message"),
		AllowedFilters: []string{
			"provider",
			"product_type",
			"endpoint",
			"request_ref",
			"response_status",
		},
		FilterSanitizer: filter.WhitelistStringFilter,
		AllowedOrderColumns: []string{
			"endpoint",
			"request_ref",
			"response_status",
			"duration_ms",
			"created_at",
		},
		DefaultOrders: []string{"created_at DESC"},
	})
}
