package interfacecatalog

import (
	"context"

	domaincatalog "3za-digital/internal/domain/catalog"
	"3za-digital/internal/dto"
	"3za-digital/pkg/filter"
)

type ServiceCatalogInterface interface {
	GetAll(ctx context.Context, params filter.BaseParams) ([]domaincatalog.ProviderService, int64, error)
	Sync(ctx context.Context, productType string, req dto.SyncCatalogRequest) (dto.SyncCatalogResponse, error)
	EnsureFresh(ctx context.Context, productType string) error
}
