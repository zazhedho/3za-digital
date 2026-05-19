package interfacedashboard

import (
	"context"

	domaindashboard "3za-digital/internal/domain/dashboard"
)

type ServiceDashboardInterface interface {
	GetSummary(ctx context.Context, productType string) (domaindashboard.Summary, error)
}
