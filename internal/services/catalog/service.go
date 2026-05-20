package servicecatalog

import (
	"context"
	"errors"
	"strings"
	"time"

	domaincatalog "3za-digital/internal/domain/catalog"
	"3za-digital/internal/dto"
	"3za-digital/internal/integrations/h2h"
	interfacecatalog "3za-digital/internal/interfaces/catalog"
	interfaceprovider "3za-digital/internal/interfaces/provider"
	"3za-digital/pkg/filter"
	"3za-digital/utils"
)

var (
	ErrUnsupportedProductType = errors.New("unsupported product type")
	ErrProviderUnavailable    = errors.New("provider unavailable")
)

type CatalogService struct {
	Repo            interfacecatalog.RepoCatalogInterface
	ProviderFactory interfaceprovider.ClientFactory
}

func NewCatalogService(repo interfacecatalog.RepoCatalogInterface, providerFactory interfaceprovider.ClientFactory) *CatalogService {
	return &CatalogService{
		Repo:            repo,
		ProviderFactory: providerFactory,
	}
}

func (s *CatalogService) GetAll(ctx context.Context, params filter.BaseParams) ([]domaincatalog.ProviderService, int64, error) {
	return s.Repo.GetAll(ctx, params)
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
