package servicecatalog

import (
	domaincatalog "3za-digital/internal/domain/catalog"
	"3za-digital/internal/integrations/h2h"
	"3za-digital/utils"
	"time"
)

func mapH2HService(productType string, item h2h.Service, syncedAt time.Time) domaincatalog.ProviderService {
	minQuantity := int64(item.MinQuantity.Int())
	maxQuantity := int64(item.MaxQuantity.Int())
	rawResponse := item.Raw
	if len(rawResponse) == 0 {
		rawResponse = utils.MustJSON(item)
	}

	return domaincatalog.ProviderService{
		Id:                utils.CreateUUID(),
		Provider:          domaincatalog.ProviderH2H,
		ProductType:       productType,
		ProviderServiceID: item.ProviderServiceID(),
		Name:              item.Name,
		Category:          item.Category,
		Brand:             item.Brand,
		Platform:          utils.FirstNonEmptyString(item.Platform, productType),
		MinQuantity:       utils.Int64PtrIfPositive(minQuantity),
		MaxQuantity:       utils.Int64PtrIfPositive(maxQuantity),
		Price:             catalogPrice(item),
		Metadata:          utils.MustJSON(map[string]string{"source": "h2h", "price_unit": priceUnit(productType)}),
		RawResponse:       rawResponse,
		IsActive:          serviceIsActive(item.Status.String()),
		SyncedAt:          &syncedAt,
		UpdatedAt:         &syncedAt,
	}
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

func catalogPrice(item h2h.Service) string {
	return utils.FirstNonEmptyString(item.Price.String(), item.PricePer1K.String(), "0")
}

func priceUnit(productType string) string {
	if productType == domaincatalog.ProductTypeSMM {
		return "per_1000"
	}
	return "unit"
}

func serviceIsActive(status string) bool {
	status = utils.NormalizeKey(status)
	return status == "" || status == "active" || status == "1" || status == "available" || status == "true"
}
