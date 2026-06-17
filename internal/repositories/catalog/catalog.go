package repositorycatalog

import (
	"context"

	domaincatalog "3za-digital/internal/domain/catalog"
	interfacecatalog "3za-digital/internal/interfaces/catalog"
	repositorygeneric "3za-digital/internal/repositories/generic"
	"3za-digital/pkg/filter"

	"gorm.io/gorm"
)

type repo struct {
	*repositorygeneric.GenericRepository[domaincatalog.ProviderService]
}

func NewCatalogRepo(db *gorm.DB) interfacecatalog.RepoCatalogInterface {
	return &repo{GenericRepository: repositorygeneric.New[domaincatalog.ProviderService](db)}
}

func (r *repo) GetAll(ctx context.Context, params filter.BaseParams) ([]domaincatalog.ProviderService, int64, error) {
	return r.GenericRepository.GetAll(ctx, params, repositorygeneric.QueryOptions{
		Search: repositorygeneric.BuildTokenSearchFunc("name", "provider_service_id", "category", "brand", "platform"),
		AllowedFilters: []string{
			"provider",
			"product_type",
			"provider_service_id",
			"category",
			"brand",
			"platform",
			"is_active",
		},
		FilterSanitizer: filter.WhitelistStringFilter,
		AllowedOrderColumns: []string{
			"provider_service_id",
			"name",
			"category",
			"brand",
			"platform",
			"price",
			"min_quantity",
			"max_quantity",
			"synced_at",
			"created_at",
			"updated_at",
		},
		DefaultOrders: []string{"platform ASC", "category ASC", "name ASC"},
	})
}

func (r *repo) UpsertServices(ctx context.Context, services []domaincatalog.ProviderService) error {
	return r.GenericRepository.Upsert(ctx, services,
		[]string{"provider", "product_type", "provider_service_id"},
		[]string{
			"name",
			"category",
			"brand",
			"platform",
			"min_quantity",
			"max_quantity",
			"price",
			"metadata",
			"raw_response",
			"is_active",
			"synced_at",
			"updated_at",
			"deleted_at",
		},
	)
}

func (r *repo) DeactivateStaleServices(ctx context.Context, productType string, syncedAtThreshold string) error {
	return r.DB.WithContext(ctx).
		Model(&domaincatalog.ProviderService{}).
		Where("product_type = ? AND (synced_at IS NULL OR synced_at < ?) AND is_active = ?", productType, syncedAtThreshold, true).
		Update("is_active", false).Error
}
