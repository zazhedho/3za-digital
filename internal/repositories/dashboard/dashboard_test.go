package repositorydashboard

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func newDashboardMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB, PreferSimpleProtocol: true}), &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		t.Fatalf("open gorm: %v", err)
	}
	return db, mock
}

func TestGetSummaryOnlySumsCompletedOrderRevenue(t *testing.T) {
	db, mock := newDashboardMockDB(t)
	repo := NewDashboardRepo(db)

	mock.ExpectQuery(`SUM\(amount\) FILTER \(WHERE status = 'completed'\).*SUM\(provider_charge\) FILTER \(WHERE status = 'completed'\).*SUM\(profit\) FILTER \(WHERE status = 'completed'\)`).
		WithArgs("smm", "user-1").
		WillReturnRows(sqlmock.NewRows([]string{
			"total_orders",
			"pending_orders",
			"processing_orders",
			"completed_orders",
			"partial_orders",
			"failed_orders",
			"cancelled_orders",
			"total_amount",
			"total_provider_charge",
			"total_profit",
		}).AddRow(3, 0, 0, 0, 0, 3, 0, "0", "0", "0"))

	summary, err := repo.GetSummary(context.Background(), "smm", "user-1")
	if err != nil {
		t.Fatalf("GetSummary returned error: %v", err)
	}
	if summary.TotalProfit != "0" || summary.TotalAmount != "0" || summary.TotalProviderCharge != "0" {
		t.Fatalf("expected zero revenue for non-completed orders, got %+v", summary)
	}
	if summary.FailedOrders != 3 {
		t.Fatalf("expected failed order count to remain 3, got %d", summary.FailedOrders)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
