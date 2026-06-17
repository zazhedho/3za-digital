package repositoryorder

import (
	"strings"
	"testing"

	domainorder "3za-digital/internal/domain/order"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func newOrderDryRunDB(t *testing.T) *gorm.DB {
	t.Helper()

	sqlDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{DryRun: true})
	if err != nil {
		t.Fatalf("gorm open: %v", err)
	}

	return db
}

func TestOrderSearchFuncIncludesServiceNameAndProviderResponse(t *testing.T) {
	db := newOrderDryRunDB(t).Model(&domainorder.Order{})

	query := orderSearchFunc(db, "TikTok Followers [ Max 10M ] | INV0001 🚀")
	var rows []domainorder.Order
	query.Find(&rows)

	sql := query.Statement.SQL.String()
	if !strings.Contains(sql, "orders.provider_response::text") {
		t.Fatalf("expected provider response text search, got %q", sql)
	}
	if !strings.Contains(sql, "provider_services ps") {
		t.Fatalf("expected service name subquery search, got %q", sql)
	}
	if !strings.Contains(sql, " AND ") {
		t.Fatalf("expected token search AND groups, got %q", sql)
	}
}
