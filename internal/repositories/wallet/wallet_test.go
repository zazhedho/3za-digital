package repositorywallet

import (
	"errors"
	"testing"
	"time"

	domainwallet "3za-digital/internal/domain/wallet"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func newWalletMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
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

func TestAssertMainBalanceLimitIncludesLockedBalance(t *testing.T) {
	db, mock := newWalletMockDB(t)
	repo := &repo{db: db}

	mock.ExpectQuery(`SELECT \* FROM "wallets" WHERE is_active = \$1 AND "wallets"\."deleted_at" IS NULL FOR UPDATE`).
		WithArgs(true).
		WillReturnRows(walletRows().
			AddRow("wallet-1", "user-1", "60.00", "20.00", "IDR", true, time.Now(), nil, nil))

	err := repo.assertMainBalanceLimitTx(db, "30.00", "100.00")
	if !errors.Is(err, domainwallet.ErrInsufficientMainBalance) {
		t.Fatalf("expected ErrInsufficientMainBalance, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestAssertMainBalanceLimitAllowsExactLiabilityLimit(t *testing.T) {
	db, mock := newWalletMockDB(t)
	repo := &repo{db: db}

	mock.ExpectQuery(`SELECT \* FROM "wallets" WHERE is_active = \$1 AND "wallets"\."deleted_at" IS NULL FOR UPDATE`).
		WithArgs(true).
		WillReturnRows(walletRows().
			AddRow("wallet-1", "user-1", "60.00", "10.00", "IDR", true, time.Now(), nil, nil))

	if err := repo.assertMainBalanceLimitTx(db, "30.00", "100.00"); err != nil {
		t.Fatalf("expected exact limit to pass, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func walletRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id",
		"user_id",
		"balance",
		"locked_balance",
		"currency",
		"is_active",
		"created_at",
		"updated_at",
		"deleted_at",
	})
}
