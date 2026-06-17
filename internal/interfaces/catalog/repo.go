package interfacecatalog

import (
	"context"
	"time"

	domaincatalog "3za-digital/internal/domain/catalog"
	interfacegeneric "3za-digital/internal/interfaces/generic"
	"3za-digital/pkg/filter"
)

type RepoCatalogInterface interface {
	interfacegeneric.GenericRepository[domaincatalog.ProviderService]

	GetAll(ctx context.Context, params filter.BaseParams) ([]domaincatalog.ProviderService, int64, error)
	UpsertServices(ctx context.Context, services []domaincatalog.ProviderService) error
	DeactivateStaleServices(ctx context.Context, provider, productType, platform, category string, syncedAtThreshold time.Time) error
}
