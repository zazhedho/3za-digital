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
)

type CatalogService struct {
	Repo            interfacecatalog.RepoCatalogInterface
	ProviderFactory interfaceprovider.ClientFactory
	ConfigService   interfaceappconfig.ServiceAppConfigInterface
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

func (s *CatalogService) GetAll(ctx context.Context, params filter.BaseParams) ([]domaincatalog.ProviderService, int64, error) {
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

	return dto.SyncCatalogResponse{
		Provider:    domaincatalog.ProviderH2H,
		ProductType: productType,
		Total:       priceList.Total,
		Synced:      len(services),
	}, nil
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
