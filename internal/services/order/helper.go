package serviceorder

import (
	domainwallet "3za-digital/internal/domain/wallet"
	"3za-digital/internal/integrations/h2h"
	"encoding/json"
	"errors"
	"net/url"

	domaincatalog "3za-digital/internal/domain/catalog"
	domainorder "3za-digital/internal/domain/order"
	"3za-digital/utils"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

func validateService(service domaincatalog.ProviderService, productType string, quantity int64) error {
	if service.Provider != domaincatalog.ProviderH2H || service.ProductType != productType {
		return ErrUnsupportedService
	}
	if !service.IsActive {
		return ErrInactiveService
	}
	if service.MinQuantity != nil && quantity < *service.MinQuantity {
		return ErrQuantityBelowMinimum
	}
	if service.MaxQuantity != nil && quantity > *service.MaxQuantity {
		return ErrQuantityAboveMaximum
	}
	return nil
}

func validateOrderTarget(productType string, target string) error {
	if productType != domaincatalog.ProductTypeSMM {
		return nil
	}
	parsed, err := url.ParseRequestURI(strings.TrimSpace(target))
	if err != nil || parsed == nil {
		return ErrInvalidOrderRequest
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return ErrInvalidOrderRequest
	}
	if parsed.Host == "" {
		return ErrInvalidOrderRequest
	}
	return nil
}

func normalizeProviderStatus(status string) string {
	status = utils.NormalizeKey(status)
	status = strings.ReplaceAll(status, " ", "_")
	switch status {
	case domainorder.StatusPending:
		return domainorder.StatusPending
	case "process", "in_progress", domainorder.StatusProcessing:
		return domainorder.StatusProcessing
	case "success", "done", domainorder.StatusCompleted:
		return domainorder.StatusCompleted
	case domainorder.StatusPartial:
		return domainorder.StatusPartial
	case "error", "reject", "rejected", domainorder.StatusFailed:
		return domainorder.StatusFailed
	case "cancel", domainorder.StatusCancelled:
		return domainorder.StatusCancelled
	default:
		return ""
	}
}

func normalizeCreateProviderStatus(status string) string {
	normalized := normalizeProviderStatus(status)
	if isFinalStatus(normalized) && normalized != domainorder.StatusCompleted {
		return domainorder.StatusProcessing
	}
	return normalized
}

func isRefundableFinalStatus(status string) bool {
	switch status {
	case domainorder.StatusFailed, domainorder.StatusCancelled:
		return true
	default:
		return false
	}
}

func isFinalStatus(status string) bool {
	switch status {
	case domainorder.StatusCompleted, domainorder.StatusPartial, domainorder.StatusFailed, domainorder.StatusCancelled:
		return true
	default:
		return false
	}
}

func generateRefID(productType string) string {
	id := strings.ToUpper(strings.ReplaceAll(utils.CreateUUID(), "-", ""))
	return fmt.Sprintf("%s-%s", strings.ToUpper(productType), id)
}

func providerErrorRaw(err error) json.RawMessage {
	if apiErr, ok := errors.AsType[*h2h.APIError](err); ok && len(apiErr.Raw) > 0 {
		return apiErr.Raw
	}
	return utils.EmptyJSON()
}

func mergeOrderMetadata(raw json.RawMessage, values map[string]string) json.RawMessage {
	metadata := orderMetadata(raw)
	for key, value := range values {
		metadata[key] = value
	}
	return utils.MustJSON(metadata)
}

func orderMetadata(raw json.RawMessage) map[string]string {
	metadata := map[string]string{}
	if len(raw) == 0 {
		return metadata
	}
	_ = json.Unmarshal(raw, &metadata)
	return metadata
}

func IsPublicError(err error) bool {
	return errors.Is(err, ErrInvalidOrderRequest) ||
		errors.Is(err, ErrInactiveService) ||
		errors.Is(err, ErrUnsupportedService) ||
		errors.Is(err, ErrQuantityBelowMinimum) ||
		errors.Is(err, ErrQuantityAboveMaximum) ||
		errors.Is(err, ErrOrderAlreadyFinal) ||
		errors.Is(err, ErrServicePriceStale) ||
		errors.Is(err, domainwallet.ErrInactiveWallet) ||
		errors.Is(err, domainwallet.ErrInsufficientBalance) ||
		errors.Is(err, domainwallet.ErrInvalidAmount)
}

func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
