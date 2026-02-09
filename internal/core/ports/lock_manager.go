package ports

import (
	"context"
	"time"
)


type LockManager interface {
	// Lock acquires a lock for a key. Returns unlock function on success.
	Lock(ctx context.Context, key string, ttl time.Duration) (unlock func() error, err error)

	// LockAccounts locks multiple accounts in consistent order to prevent deadlocks.
	LockAccounts(ctx context.Context, accountIDs []int64, ttl time.Duration) (unlock func() error, err error)
}
