package repositorydashboard

import (
	"context"
	"strings"

	domaindashboard "3za-digital/internal/domain/dashboard"
	interfacedashboard "3za-digital/internal/interfaces/dashboard"

	"gorm.io/gorm"
)

type repo struct {
	db *gorm.DB
}

func NewDashboardRepo(db *gorm.DB) interfacedashboard.RepoDashboardInterface {
	return &repo{db: db}
}

func (r *repo) GetSummary(ctx context.Context, productType string, userID string) (domaindashboard.Summary, error) {
	var row struct {
		TotalOrders         int64
		PendingOrders       int64
		ProcessingOrders    int64
		CompletedOrders     int64
		PartialOrders       int64
		FailedOrders        int64
		CancelledOrders     int64
		TotalAmount         string
		TotalProviderCharge string
		TotalProfit         string
	}

	query := r.db.WithContext(ctx).Table("orders").Where("deleted_at IS NULL")
	if productType != "" {
		query = query.Where("product_type = ?", productType)
	}
	if strings.TrimSpace(userID) != "" {
		query = query.Where("created_by = ?", strings.TrimSpace(userID))
	}

	err := query.Select(`
		COUNT(*) AS total_orders,
		COUNT(*) FILTER (WHERE status = 'pending') AS pending_orders,
		COUNT(*) FILTER (WHERE status = 'processing') AS processing_orders,
		COUNT(*) FILTER (WHERE status = 'completed') AS completed_orders,
		COUNT(*) FILTER (WHERE status = 'partial') AS partial_orders,
		COUNT(*) FILTER (WHERE status = 'failed') AS failed_orders,
		COUNT(*) FILTER (WHERE status = 'cancelled') AS cancelled_orders,
		COALESCE(SUM(amount) FILTER (WHERE status = 'completed'), 0)::text AS total_amount,
		COALESCE(SUM(provider_charge) FILTER (WHERE status = 'completed'), 0)::text AS total_provider_charge,
		COALESCE(SUM(profit) FILTER (WHERE status = 'completed'), 0)::text AS total_profit
	`).Scan(&row).Error
	if err != nil {
		return domaindashboard.Summary{}, err
	}

	return domaindashboard.Summary{
		TotalOrders:         row.TotalOrders,
		PendingOrders:       row.PendingOrders,
		ProcessingOrders:    row.ProcessingOrders,
		CompletedOrders:     row.CompletedOrders,
		PartialOrders:       row.PartialOrders,
		FailedOrders:        row.FailedOrders,
		CancelledOrders:     row.CancelledOrders,
		TotalAmount:         row.TotalAmount,
		TotalProviderCharge: row.TotalProviderCharge,
		TotalProfit:         row.TotalProfit,
	}, nil
}
