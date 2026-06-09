package servicecatalog

import (
	"context"
	"errors"
	"strings"
	"time"

	domaincatalog "3za-digital/internal/domain/catalog"
	"3za-digital/internal/dto"
	"3za-digital/internal/integrations/h2h"
	interfaceappconfig "3za-digital/internal/interfaces/appconfig"
	interfacecatalog "3za-digital/internal/interfaces/catalog"
	interfaceprovider "3za-digital/internal/interfaces/provider"
	"3za-digital/pkg/filter"
	"3za-digital/pkg/money"
	"3za-digital/utils"
)

var (
	ErrUnsupportedProductType = errors.New("unsupported product type")
	ErrProviderUnavailable    = errors.New("provider unavailable")
	ErrCatalogSyncLocked      = errors.New("catalog sync is already running")
)

const (
	smmServicePriceMaxAgeConfigKey = "pricing.smm_service_price_max_age"
	syncWaitTimeout                = 10 * time.Second
	syncWaitInterval               = 250 * time.Millisecond
)

type CatalogService struct {
	Repo            interfacecatalog.RepoCatalogInterface
	ProviderFactory interfaceprovider.ClientFactory
	ConfigService   interfaceappconfig.ServiceAppConfigInterface
	SyncState       interfacecatalog.SyncStateStoreInterface
}

func NewCatalogService(repo interfacecatalog.RepoCatalogInterface, providerFactory interfaceprovider.ClientFactory, configServices ...interfaceappconfig.ServiceAppConfigInterface) *CatalogService {
	service := &CatalogService{
		Repo:            repo,
		ProviderFactory: providerFactory,
	}
	if len(configServices) > 0 {
		service.ConfigService = configServices[0]
	}
	return service
}

func (s *CatalogService) WithSyncStateStore(syncState interfacecatalog.SyncStateStoreInterface) *CatalogService {
	s.SyncState = syncState
	return s
}

func (s *CatalogService) GetAll(ctx context.Context, params filter.BaseParams) ([]domaincatalog.ProviderService, int64, error) {
	if productType, ok := params.Filters["product_type"].(string); ok && utils.NormalizeKey(productType) == domaincatalog.ProductTypeSMM {
		if err := s.EnsureFresh(ctx, domaincatalog.ProductTypeSMM); err != nil {
			return nil, 0, err
		}
	}

	services, total, err := s.Repo.GetAll(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	markupByProduct := make(map[string]string)
	for index := range services {
		productType := services[index].ProductType
		markupPercent, ok := markupByProduct[productType]
		if !ok {
			markupPercent, err = s.markupPercent(ctx, productType)
			if err != nil {
				return nil, 0, err
			}
			markupByProduct[productType] = markupPercent
		}

		_, amount, _, err := money.MarkupAmount(services[index].Price, markupPercent)
		if err != nil {
			return nil, 0, err
		}
		services[index].Price = amount
		services[index].RawResponse = nil
	}
	return services, total, nil
}

func (s *CatalogService) Sync(ctx context.Context, productType string, req dto.SyncCatalogRequest) (dto.SyncCatalogResponse, error) {
	productType = utils.NormalizeKey(productType)
	if !isSupportedProductType(productType) {
		return dto.SyncCatalogResponse{}, ErrUnsupportedProductType
	}

	lock, err := s.acquireSyncLock(ctx, productType)
	if err != nil {
		return dto.SyncCatalogResponse{}, err
	}
	defer lock.Release(context.Background())

	return s.syncUnlocked(ctx, productType, req)
}

func (s *CatalogService) syncUnlocked(ctx context.Context, productType string, req dto.SyncCatalogRequest) (dto.SyncCatalogResponse, error) {
	if s.ProviderFactory == nil {
		return dto.SyncCatalogResponse{}, ErrProviderUnavailable
	}

	client, err := s.ProviderFactory()
	if err != nil {
		return dto.SyncCatalogResponse{}, err
	}

	priceList, err := client.GetPriceList(ctx, h2h.PriceListRequest{
		Type:     productType,
		Platform: strings.TrimSpace(req.Platform),
		Category: strings.TrimSpace(req.Category),
	})
	if err != nil {
		return dto.SyncCatalogResponse{}, err
	}

	now := time.Now()
	services := make([]domaincatalog.ProviderService, 0, len(priceList.Services))
	for _, item := range priceList.Services {
		services = append(services, mapH2HService(productType, item, now))
	}

	if err := s.Repo.UpsertServices(ctx, services); err != nil {
		return dto.SyncCatalogResponse{}, err
	}
	s.storeLastSyncedAt(ctx, productType, now)

	return dto.SyncCatalogResponse{
		Provider:    domaincatalog.ProviderH2H,
		ProductType: productType,
		Total:       priceList.Total,
		Synced:      len(services),
	}, nil
}

func (s *CatalogService) EnsureFresh(ctx context.Context, productType string) error {
	productType = utils.NormalizeKey(productType)
	if productType != domaincatalog.ProductTypeSMM {
		return nil
	}

	maxAge := 24 * time.Hour
	if s.ConfigService != nil {
		value, err := s.ConfigService.GetDuration(ctx, smmServicePriceMaxAgeConfigKey, maxAge)
		if err != nil {
			return err
		}
		if value > 0 {
			maxAge = value
		}
	}

	latestSyncedAt, found, err := s.lastSyncedAt(ctx, productType)
	if err != nil {
		return err
	}
	if found && time.Since(latestSyncedAt) < maxAge {
		return nil
	}

	_, err = s.Sync(ctx, productType, dto.SyncCatalogRequest{})
	if errors.Is(err, ErrCatalogSyncLocked) {
		return s.waitForFreshSync(ctx, productType, maxAge)
	}
	return err
}

var _ interfacecatalog.ServiceCatalogInterface = (*CatalogService)(nil)

func (s *CatalogService) markupPercent(ctx context.Context, productType string) (string, error) {
	markupPercent := "0"
	if s.ConfigService != nil {
		productKey := "pricing.product_markup_percent." + productType
		if value, err := s.ConfigService.GetString(ctx, productKey, ""); err != nil {
			return "", err
		} else if strings.TrimSpace(value) != "" {
			markupPercent = value
		} else if value, err := s.ConfigService.GetString(ctx, "pricing.default_markup_percent", "0"); err != nil {
			return "", err
		} else {
			markupPercent = value
		}
	}
	return markupPercent, nil
}

func (s *CatalogService) latestSyncedAt(ctx context.Context, productType string) (time.Time, bool, error) {
	params := filter.BaseParams{
		Filters: map[string]interface{}{
			"provider":     domaincatalog.ProviderH2H,
			"product_type": productType,
		},
		OrderBy:        "synced_at",
		OrderDirection: "desc",
		Page:           1,
		Limit:          1,
	}

	services, _, err := s.Repo.GetAll(ctx, params)
	if err != nil {
		return time.Time{}, false, err
	}
	if len(services) == 0 || services[0].SyncedAt == nil {
		return time.Time{}, false, nil
	}
	return *services[0].SyncedAt, true, nil
}

func (s *CatalogService) lastSyncedAt(ctx context.Context, productType string) (time.Time, bool, error) {
	if s.SyncState == nil {
		return s.latestSyncedAt(ctx, productType)
	}
	return s.SyncState.LastSyncedAt(ctx, productType)
}

func (s *CatalogService) storeLastSyncedAt(ctx context.Context, productType string, syncedAt time.Time) {
	if s.SyncState == nil {
		return
	}
	s.SyncState.StoreLastSyncedAt(ctx, productType, syncedAt)
}

func (s *CatalogService) acquireSyncLock(ctx context.Context, productType string) (interfacecatalog.SyncLockInterface, error) {
	if s.SyncState == nil {
		return noopSyncLock{}, nil
	}
	lock, err := s.SyncState.AcquireSyncLock(ctx, productType)
	if errors.Is(err, interfacecatalog.ErrSyncLocked) {
		return nil, ErrCatalogSyncLocked
	}
	return lock, err
}

func (s *CatalogService) waitForFreshSync(ctx context.Context, productType string, maxAge time.Duration) error {
	deadline := time.NewTimer(syncWaitTimeout)
	defer deadline.Stop()

	ticker := time.NewTicker(syncWaitInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-deadline.C:
			return ErrProviderUnavailable
		case <-ticker.C:
			syncedAt, found, err := s.lastSyncedAt(ctx, productType)
			if err != nil {
				return err
			}
			if found && time.Since(syncedAt) < maxAge {
				return nil
			}
		}
	}
}

type noopSyncLock struct{}

func (noopSyncLock) Release(ctx context.Context) {}
