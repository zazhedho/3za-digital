package servicedashboard

import (
	"context"
	"strings"

	domaindashboard "3za-digital/internal/domain/dashboard"
	interfacedashboard "3za-digital/internal/interfaces/dashboard"
)

type DashboardService struct {
	Repo interfacedashboard.RepoDashboardInterface
}

func NewDashboardService(repo interfacedashboard.RepoDashboardInterface) *DashboardService {
	return &DashboardService{Repo: repo}
}

func (s *DashboardService) GetSummary(ctx context.Context, productType string) (domaindashboard.Summary, error) {
	return s.Repo.GetSummary(ctx, strings.ToLower(strings.TrimSpace(productType)))
}

var _ interfacedashboard.ServiceDashboardInterface = (*DashboardService)(nil)
