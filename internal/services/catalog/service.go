package servicecatalog

import (
	"context"
	"encoding/json"
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
	productType = normalizeProductType(productType)
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

func mapH2HService(productType string, item h2h.Service, syncedAt time.Time) domaincatalog.ProviderService {
	minQuantity := int64(item.MinQuantity.Int())
	maxQuantity := int64(item.MaxQuantity.Int())
	rawResponse := item.Raw
	if len(rawResponse) == 0 {
		rawResponse = mustJSON(item)
	}

	return domaincatalog.ProviderService{
		Id:                utils.CreateUUID(),
		Provider:          domaincatalog.ProviderH2H,
		ProductType:       productType,
		ProviderServiceID: item.ProviderServiceID(),
		Name:              item.Name,
		Category:          item.Category,
		Brand:             item.Brand,
		Platform:          firstNonEmpty(item.Platform, productType),
		MinQuantity:       quantityPointer(minQuantity),
		MaxQuantity:       quantityPointer(maxQuantity),
		Price:             firstNonEmpty(item.Price.String(), "0"),
		Metadata:          mustJSON(map[string]string{"source": "h2h"}),
		RawResponse:       rawResponse,
		IsActive:          serviceIsActive(item.Status.String()),
		SyncedAt:          &syncedAt,
		UpdatedAt:         &syncedAt,
	}
}

func normalizeProductType(productType string) string {
	return strings.ToLower(strings.TrimSpace(productType))
}

func isSupportedProductType(productType string) bool {
	switch productType {
	case domaincatalog.ProductTypeSMM,
		domaincatalog.ProductTypePulsa,
		domaincatalog.ProductTypePPOB,
		domaincatalog.ProductTypeGame,
		domaincatalog.ProductTypeEWallet:
		return true
	default:
		return false
	}
}

func quantityPointer(value int64) *int64 {
	if value <= 0 {
		return nil
	}
	return &value
}

func serviceIsActive(status string) bool {
	status = strings.ToLower(strings.TrimSpace(status))
	return status == "" || status == "active" || status == "1" || status == "available" || status == "true"
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func mustJSON(value interface{}) json.RawMessage {
	body, err := json.Marshal(value)
	if err != nil {
		return json.RawMessage(`{}`)
	}
	return body
}

var _ interfacecatalog.ServiceCatalogInterface = (*CatalogService)(nil)
