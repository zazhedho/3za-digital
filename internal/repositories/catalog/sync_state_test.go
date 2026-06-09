package repositorycatalog

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	redismock "github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	domaincatalog "3za-digital/internal/domain/catalog"
	interfacecatalog "3za-digital/internal/interfaces/catalog"
)

func TestSyncStateLastSyncedAtUsesRedis(t *testing.T) {
	client, mock := redismock.NewClientMock()
	store := NewCatalogSyncStateStore(nil, client)
	now := time.Now().UTC()

	mock.ExpectGet(lastSyncKey(domaincatalog.ProductTypeSMM)).SetVal(now.Format(time.RFC3339Nano))

	got, found, err := store.LastSyncedAt(context.Background(), domaincatalog.ProductTypeSMM)
	if err != nil || !found || !got.Equal(now) {
		t.Fatalf("last synced: got=%s found=%v err=%v", got, found, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("redis expectations: %v", err)
	}
}

func TestSyncStateLastSyncedAtFallsBackToDBAndRefillsRedis(t *testing.T) {
	db, sqlMock := newCatalogSyncStateMockDB(t)
	client, redisMock := redismock.NewClientMock()
	store := NewCatalogSyncStateStore(db, client)
	now := time.Now().UTC()

	redisMock.ExpectGet(lastSyncKey(domaincatalog.ProductTypeSMM)).SetErr(redis.Nil)
	sqlMock.ExpectQuery(`SELECT MAX\(synced_at\) FROM "provider_services"`).
		WithArgs(domaincatalog.ProviderH2H, domaincatalog.ProductTypeSMM).
		WillReturnRows(sqlmock.NewRows([]string{"max"}).AddRow(now))
	redisMock.Regexp().ExpectSet(lastSyncKey(domaincatalog.ProductTypeSMM), `.+`, 0).SetVal("OK")

	got, found, err := store.LastSyncedAt(context.Background(), domaincatalog.ProductTypeSMM)
	if err != nil || !found || !got.Equal(now) {
		t.Fatalf("last synced: got=%s found=%v err=%v", got, found, err)
	}
	if err := redisMock.ExpectationsWereMet(); err != nil {
		t.Fatalf("redis expectations: %v", err)
	}
	if err := sqlMock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestSyncStateLastSyncedAtFallsBackToDBWhenRedisUnavailable(t *testing.T) {
	db, sqlMock := newCatalogSyncStateMockDB(t)
	client, redisMock := redismock.NewClientMock()
	store := NewCatalogSyncStateStore(db, client)
	now := time.Now().UTC()

	redisMock.ExpectGet(lastSyncKey(domaincatalog.ProductTypeSMM)).SetErr(errors.New("redis unavailable"))
	sqlMock.ExpectQuery(`SELECT MAX\(synced_at\) FROM "provider_services"`).
		WithArgs(domaincatalog.ProviderH2H, domaincatalog.ProductTypeSMM).
		WillReturnRows(sqlmock.NewRows([]string{"max"}).AddRow(now))
	redisMock.Regexp().ExpectSet(lastSyncKey(domaincatalog.ProductTypeSMM), `.+`, 0).SetErr(errors.New("redis unavailable"))

	got, found, err := store.LastSyncedAt(context.Background(), domaincatalog.ProductTypeSMM)
	if err != nil || !found || !got.Equal(now) {
		t.Fatalf("last synced: got=%s found=%v err=%v", got, found, err)
	}
	if err := redisMock.ExpectationsWereMet(); err != nil {
		t.Fatalf("redis expectations: %v", err)
	}
	if err := sqlMock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestSyncStateAcquireSyncLockUsesDBAdvisoryLock(t *testing.T) {
	db, sqlMock := newCatalogSyncStateMockDB(t)
	store := NewCatalogSyncStateStore(db, nil)
	lockKey := syncLockDBKey(domaincatalog.ProductTypeSMM)

	sqlMock.ExpectQuery(`SELECT pg_try_advisory_lock\(\$1\)`).
		WithArgs(lockKey).
		WillReturnRows(sqlmock.NewRows([]string{"pg_try_advisory_lock"}).AddRow(true))
	sqlMock.ExpectExec(`SELECT pg_advisory_unlock\(\$1\)`).
		WithArgs(lockKey).
		WillReturnResult(sqlmock.NewResult(0, 1))

	lock, err := store.AcquireSyncLock(context.Background(), domaincatalog.ProductTypeSMM)
	if err != nil {
		t.Fatalf("AcquireSyncLock returned error: %v", err)
	}
	lock.Release(context.Background())

	if err := sqlMock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestSyncStateAcquireSyncLockReturnsLockedWhenDBLockBusy(t *testing.T) {
	db, sqlMock := newCatalogSyncStateMockDB(t)
	store := NewCatalogSyncStateStore(db, nil)
	lockKey := syncLockDBKey(domaincatalog.ProductTypeSMM)

	sqlMock.ExpectQuery(`SELECT pg_try_advisory_lock\(\$1\)`).
		WithArgs(lockKey).
		WillReturnRows(sqlmock.NewRows([]string{"pg_try_advisory_lock"}).AddRow(false))

	_, err := store.AcquireSyncLock(context.Background(), domaincatalog.ProductTypeSMM)
	if !errors.Is(err, interfacecatalog.ErrSyncLocked) {
		t.Fatalf("expected ErrSyncLocked, got %v", err)
	}
	if err := sqlMock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func newCatalogSyncStateMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB, PreferSimpleProtocol: true}), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm open: %v", err)
	}
	return db, mock
}
