package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/maneeshsagar/tps/internal/core/domain"
	"gorm.io/gorm"
)

type TransactionModel struct {
	ID                   uuid.UUID `gorm:"primaryKey;column:id;type:uuid"`
	SourceAccountID      int64     `gorm:"column:source_account_id;index"`
	DestinationAccountID int64     `gorm:"column:destination_account_id;index"`
	Amount               int64     `gorm:"column:amount"`
	CreatedAt            time.Time `gorm:"column:created_at"`
}

func (TransactionModel) TableName() string {
	return "transactions"
}

type TransactionRepo struct {
	db *gorm.DB
}

func NewTransactionRepo(db *gorm.DB) *TransactionRepo {
	return &TransactionRepo{db}
}

func (r *TransactionRepo) Create(tx *domain.Transaction) error {
	m := TransactionModel{
		ID:                   tx.ID,
		SourceAccountID:      tx.SourceAccountID,
		DestinationAccountID: tx.DestinationAccountID,
		Amount:               tx.Amount,
		CreatedAt:            tx.CreatedAt,
	}
	return r.db.Create(&m).Error
}
