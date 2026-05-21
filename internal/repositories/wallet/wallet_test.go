package repositorywallet

import (
	"context"
	"errors"
	"testing"
	"time"

	domainwallet "3za-digital/internal/domain/wallet"
	"3za-digital/pkg/filter"

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

func TestGetDepositsPreloadsUserSummaryForAdminList(t *testing.T) {
	db, mock := newWalletMockDB(t)
	repo := &repo{db: db}
	now := time.Now()

	mock.ExpectQuery(`SELECT count\(\*\) FROM "deposit_requests" WHERE status = \$1 AND "deposit_requests"\."deleted_at" IS NULL`).
		WithArgs(domainwallet.DepositStatusPending).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT \* FROM "deposit_requests" WHERE status = \$1 AND "deposit_requests"\."deleted_at" IS NULL ORDER BY created_at DESC LIMIT \$2`).
		WithArgs(domainwallet.DepositStatusPending, 50).
		WillReturnRows(depositRows().
			AddRow("deposit-1", "user-1", "150000.00", domainwallet.DepositStatusPending, domainwallet.DepositMethodManualAdmin, "", "", "", nil, nil, []byte("{}"), nil, now, nil, nil))
	mock.ExpectQuery(`SELECT "id","name","email","phone","role","avatar_url" FROM "users" WHERE "users"\."id" = \$1`).
		WithArgs("user-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "phone", "role", "avatar_url"}).
			AddRow("user-1", "Member One", "member@example.com", "08123456789", "member", ""))

	deposits, total, err := repo.GetDeposits(context.Background(), "", filter.BaseParams{
		Filters:        map[string]interface{}{"status": domainwallet.DepositStatusPending},
		OrderBy:        "created_at",
		OrderDirection: "desc",
		Page:           1,
		Limit:          50,
	})
	if err != nil {
		t.Fatalf("GetDeposits returned error: %v", err)
	}
	if total != 1 || len(deposits) != 1 {
		t.Fatalf("unexpected deposits total=%d len=%d", total, len(deposits))
	}
	if deposits[0].User == nil || deposits[0].User.Name != "Member One" || deposits[0].User.Email != "member@example.com" {
		t.Fatalf("expected user summary, got %#v", deposits[0].User)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestGetDepositsDoesNotPreloadUserForScopedMemberList(t *testing.T) {
	db, mock := newWalletMockDB(t)
	repo := &repo{db: db}
	now := time.Now()

	mock.ExpectQuery(`SELECT count\(\*\) FROM "deposit_requests" WHERE user_id = \$1 AND status = \$2 AND "deposit_requests"\."deleted_at" IS NULL`).
		WithArgs("user-1", domainwallet.DepositStatusPending).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT \* FROM "deposit_requests" WHERE user_id = \$1 AND status = \$2 AND "deposit_requests"\."deleted_at" IS NULL ORDER BY created_at DESC LIMIT \$3`).
		WithArgs("user-1", domainwallet.DepositStatusPending, 50).
		WillReturnRows(depositRows().
			AddRow("deposit-1", "user-1", "150000.00", domainwallet.DepositStatusPending, domainwallet.DepositMethodManualAdmin, "", "", "", nil, nil, []byte("{}"), nil, now, nil, nil))

	deposits, total, err := repo.GetDeposits(context.Background(), "user-1", filter.BaseParams{
		Filters:        map[string]interface{}{"status": domainwallet.DepositStatusPending},
		OrderBy:        "created_at",
		OrderDirection: "desc",
		Page:           1,
		Limit:          50,
	})
	if err != nil {
		t.Fatalf("GetDeposits returned error: %v", err)
	}
	if total != 1 || len(deposits) != 1 {
		t.Fatalf("unexpected deposits total=%d len=%d", total, len(deposits))
	}
	if deposits[0].User != nil {
		t.Fatalf("expected no user summary for scoped member list, got %#v", deposits[0].User)
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

func depositRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id",
		"user_id",
		"amount",
		"status",
		"method",
		"provider",
		"payment_reference",
		"payment_url",
		"expired_at",
		"paid_at",
		"metadata",
		"created_by",
		"created_at",
		"updated_at",
		"deleted_at",
	})
}
