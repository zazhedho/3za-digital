package repositorycatalog

import (
	"context"
	"testing"
	"time"

	domaincatalog "3za-digital/internal/domain/catalog"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func newCatalogMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB, PreferSimpleProtocol: true}), &gorm.Config{SkipDefaultTransaction: true})
	if err != nil {
		t.Fatalf("gorm open: %v", err)
	}

	return db, mock
}

func TestDeactivateStaleServicesScopesByProviderProductAndRequestedFilters(t *testing.T) {
	db, mock := newCatalogMockDB(t)
	repo := NewCatalogRepo(db)
	threshold := time.Now()

	mock.ExpectExec(`UPDATE "provider_services" SET "is_active"=\$1,"updated_at"=\$2 WHERE \(provider = \$3 AND product_type = \$4 AND \(synced_at IS NULL OR synced_at < \$5\) AND is_active = \$6\) AND platform = \$7 AND category = \$8 AND "provider_services"."deleted_at" IS NULL`).
		WithArgs(false, sqlmock.AnyArg(), domaincatalog.ProviderH2H, domaincatalog.ProductTypeSMM, threshold, true, "instagram", "followers").
		WillReturnResult(sqlmock.NewResult(0, 2))

	err := repo.DeactivateStaleServices(
		context.Background(),
		domaincatalog.ProviderH2H,
		domaincatalog.ProductTypeSMM,
		"instagram",
		"followers",
		threshold,
	)
	if err != nil {
		t.Fatalf("DeactivateStaleServices returned error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}
