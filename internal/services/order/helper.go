package serviceorder

import (
	domaincatalog "3za-digital/internal/domain/catalog"
	domainorder "3za-digital/internal/domain/order"
	"3za-digital/utils"
	"fmt"
	"strings"
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

func normalizeProviderStatus(status string) string {
	status = strings.ToLower(strings.TrimSpace(status))
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
