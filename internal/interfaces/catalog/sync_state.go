package interfacecatalog

import (
	"context"
	"errors"
	"time"
)

var ErrSyncLocked = errors.New("catalog sync is already running")

type SyncLockInterface interface {
	Release(ctx context.Context)
}

type SyncStateStoreInterface interface {
	LastSyncedAt(ctx context.Context, productType string) (time.Time, bool, error)
	StoreLastSyncedAt(ctx context.Context, productType string, syncedAt time.Time)
	AcquireSyncLock(ctx context.Context, productType string) (SyncLockInterface, error)
}
