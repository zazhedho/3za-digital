package repositorycatalog

import (
	"context"
	"database/sql"
	"time"

	interfacecatalog "3za-digital/internal/interfaces/catalog"
	"3za-digital/utils"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	syncLockTTL = 5 * time.Minute
)

var redisUnlockScript = redis.NewScript(`
if redis.call("get", KEYS[1]) == ARGV[1] then
	return redis.call("del", KEYS[1])
end
return 0
`)

type syncStateStore struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewCatalogSyncStateStore(db *gorm.DB, redisClient *redis.Client) interfacecatalog.SyncStateStoreInterface {
	return &syncStateStore{db: db, redis: redisClient}
}

func (s *syncStateStore) LastSyncedAt(ctx context.Context, productType string) (time.Time, bool, error) {
	if s.redis != nil {
		value, err := s.redis.Get(ctx, lastSyncKey(productType)).Result()
		if err == nil {
			syncedAt, parseErr := time.Parse(time.RFC3339Nano, value)
			if parseErr == nil {
				return syncedAt, true, nil
			}
		}
	}

	syncedAt, found, err := s.lastSyncedAtFromDB(ctx, productType)
	if err != nil {
		return time.Time{}, false, err
	}
	if found {
		s.StoreLastSyncedAt(ctx, productType, syncedAt)
	}
	return syncedAt, found, nil
}

func (s *syncStateStore) StoreLastSyncedAt(ctx context.Context, productType string, syncedAt time.Time) {
	if s.redis == nil {
		return
	}
	_ = s.redis.Set(ctx, lastSyncKey(productType), syncedAt.UTC().Format(time.RFC3339Nano), 0).Err()
}

func (s *syncStateStore) AcquireSyncLock(ctx context.Context, productType string) (interfacecatalog.SyncLockInterface, error) {
	lock := &syncLock{
		redis:     s.redis,
		redisKey:  syncLockKey(productType),
		redisVal:  utils.CreateUUID(),
		dbLockKey: syncLockDBKey(productType),
	}

	if s.redis != nil {
		locked, err := s.redis.SetNX(ctx, lock.redisKey, lock.redisVal, syncLockTTL).Result()
		if err == nil && !locked {
			return nil, interfacecatalog.ErrSyncLocked
		}
		if err == nil && locked {
			lock.redisLocked = true
		}
	}

	if s.db == nil {
		return lock, nil
	}

	sqlDB, err := s.db.DB()
	if err != nil {
		lock.Release(context.Background())
		return nil, err
	}

	conn, err := sqlDB.Conn(ctx)
	if err != nil {
		lock.Release(context.Background())
		return nil, err
	}

	locked, err := tryAdvisoryLock(ctx, conn, lock.dbLockKey)
	if err != nil {
		_ = conn.Close()
		lock.Release(context.Background())
		return nil, err
	}
	if !locked {
		_ = conn.Close()
		lock.Release(context.Background())
		return nil, interfacecatalog.ErrSyncLocked
	}
	lock.dbConn = conn

	return lock, nil
}

func (s *syncStateStore) lastSyncedAtFromDB(ctx context.Context, productType string) (time.Time, bool, error) {
	if s.db == nil {
		return time.Time{}, false, nil
	}

	var value sql.NullTime
	err := s.db.WithContext(ctx).
		Table("provider_services").
		Select("MAX(synced_at)").
		Where("provider = ? AND product_type = ? AND deleted_at IS NULL", "h2h", productType).
		Scan(&value).Error
	if err != nil {
		return time.Time{}, false, err
	}
	if !value.Valid {
		return time.Time{}, false, nil
	}
	return value.Time, true, nil
}

type syncLock struct {
	redis       *redis.Client
	redisLocked bool
	redisKey    string
	redisVal    string
	dbLockKey   int64
	dbConn      *sql.Conn
}

func (l *syncLock) Release(ctx context.Context) {
	if l == nil {
		return
	}
	if l.dbConn != nil {
		_, _ = l.dbConn.ExecContext(ctx, "SELECT pg_advisory_unlock($1)", l.dbLockKey)
		_ = l.dbConn.Close()
		l.dbConn = nil
	}
	if l.redisLocked && l.redis != nil {
		_ = redisUnlockScript.Run(ctx, l.redis, []string{l.redisKey}, l.redisVal).Err()
	}
}

func tryAdvisoryLock(ctx context.Context, conn *sql.Conn, key int64) (bool, error) {
	var locked bool
	if err := conn.QueryRowContext(ctx, "SELECT pg_try_advisory_lock($1)", key).Scan(&locked); err != nil {
		return false, err
	}
	return locked, nil
}
