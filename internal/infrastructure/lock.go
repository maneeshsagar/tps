package infrastructure

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/maneeshsagar/tps/internal/core/domain"
	"gorm.io/gorm"
)

type LockManager struct {
	db *gorm.DB
}

func NewLockManager(db *gorm.DB) *LockManager {
	return &LockManager{db: db}
}

func (m *LockManager) Lock(ctx context.Context, key string, ttl time.Duration) (func() error, error) {
	id := hashKey(key)
	deadline := time.Now().Add(ttl)
	wait := 5 * time.Millisecond

	for {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("%w: cancelled", domain.ErrLockAcquisitionFailed)
		}

		var ok bool
		err := m.db.WithContext(ctx).Raw("SELECT pg_try_advisory_lock(?)", id).Scan(&ok).Error
		if err != nil {
			return nil, fmt.Errorf("%w: %v", domain.ErrLockAcquisitionFailed, err)
		}
		if ok {
			return func() error {
				var released bool
				if err := m.db.Raw("SELECT pg_advisory_unlock(?)", id).Scan(&released).Error; err != nil {
					return err
				}
				if !released {
					return fmt.Errorf("lock %s not held", key)
				}
				return nil
			}, nil
		}

		if time.Now().After(deadline) {
			return nil, fmt.Errorf("%w: timeout", domain.ErrLockAcquisitionFailed)
		}
		time.Sleep(wait)
		wait = min(wait*2, 50*time.Millisecond)
	}
}

// LockAccounts locks accounts in sorted order to prevent deadlocks
func (m *LockManager) LockAccounts(ctx context.Context, ids []int64, ttl time.Duration) (func() error, error) {
	sorted := make([]int64, len(ids))
	copy(sorted, ids)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

	var unlocks []func() error
	for _, id := range sorted {
		unlock, err := m.Lock(ctx, fmt.Sprintf("account:%d", id), ttl)
		if err != nil {
			for _, fn := range unlocks {
				fn()
			}
			return nil, err
		}
		unlocks = append(unlocks, unlock)
	}

	return func() error {
		var firstErr error
		for _, fn := range unlocks {
			if err := fn(); err != nil && firstErr == nil {
				firstErr = err
			}
		}
		return firstErr
	}, nil
}

func hashKey(s string) int64 {
	var h int64
	for i, c := range s {
		h = h*31 + int64(c) + int64(i)
	}
	if h < 0 {
		h = -h
	}
	return h
}
