package repository

import (
	"context"

	"github.com/maneeshsagar/tps/internal/core/ports"
	"gorm.io/gorm"
)

type TxManager struct {
	db *gorm.DB
}

func NewTxManager(db *gorm.DB) *TxManager {
	return &TxManager{db: db}
}

// this function accepts a function that takes a Transaction and returns an error.
func (tm *TxManager) WithTransaction(ctx context.Context, fn func(tx ports.Transaction) error) error {

	dbTx := tm.db.WithContext(ctx).Begin()
	if dbTx.Error != nil {
		return dbTx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			dbTx.Rollback()
			panic(r)
		}
	}()

	if err := fn(dbTx); err != nil {
		dbTx.Rollback()
		return err
	}

	return dbTx.Commit().Error
}

func (r *AccountRepo) WithTx(tx ports.Transaction) ports.AccountRepository {
	gormTx, ok := tx.(*gorm.DB)
	if !ok {
		panic("WithTx: expected *gorm.DB")
	}
	return &AccountRepo{db: gormTx}
}

func (r *TransactionRepo) WithTx(tx ports.Transaction) ports.TransactionRepository {
	gormTx, ok := tx.(*gorm.DB)
	if !ok {
		panic("WithTx: expected *gorm.DB")
	}
	return &TransactionRepo{db: gormTx}
}
