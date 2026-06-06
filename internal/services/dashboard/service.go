package servicedashboard

import (
	"context"
	"strings"

	"3za-digital/internal/authscope"
	domaindashboard "3za-digital/internal/domain/dashboard"
	interfacedashboard "3za-digital/internal/interfaces/dashboard"
	"3za-digital/utils"
)

type DashboardService struct {
	Repo interfacedashboard.RepoDashboardInterface
}

func NewDashboardService(repo interfacedashboard.RepoDashboardInterface) *DashboardService {
	return &DashboardService{Repo: repo}
}

func (s *DashboardService) GetSummary(ctx context.Context, productType string) (domaindashboard.Summary, error) {
	scope := authscope.FromContext(ctx)
	userID := ""
	if !scope.Has("smm_orders", "list_all") {
		userID = strings.TrimSpace(scope.UserID)
	}
	summary, err := s.Repo.GetSummary(ctx, utils.NormalizeKey(productType), userID)
	if err != nil {
		return domaindashboard.Summary{}, err
	}
	if !scope.Has("provider_balance", "view") {
		summary.TotalProviderCharge = "0"
		summary.TotalProfit = "0"
	}
	return summary, nil
}

var _ interfacedashboard.ServiceDashboardInterface = (*DashboardService)(nil)
