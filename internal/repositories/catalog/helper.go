package repositorycatalog

import (
	"crypto/sha256"
)

func lastSyncKey(productType string) string {
	return "catalog:sync:last:h2h:" + productType
}

func syncLockKey(productType string) string {
	return "catalog:sync:lock:h2h:" + productType
}

func syncLockDBKey(productType string) int64 {
	sum := sha256.Sum256([]byte(syncLockKey(productType)))
	var value int64
	for _, part := range sum[:8] {
		value = (value << 8) | int64(part)
	}
	return value & (1<<63 - 1)
}
