package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/maneeshsagar/tps/internal/core/domain"
	"gorm.io/gorm"
)

type AsyncTransactionStatusModel struct {
	ID          string `gorm:"primaryKey;column:id"`
	FromAccount int64  `gorm:"column:from_account"`
	ToAccount   int64  `gorm:"column:to_account"`
	Amount      int64  `gorm:"column:amount"`
	Status      string `gorm:"column:status"`
	Error       string `gorm:"column:error"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (AsyncTransactionStatusModel) TableName() string {
	return "async_transactions_status"
}

type AsyncTransactionRepo struct {
	db *gorm.DB
}

func NewAsyncTransactionRepo(db *gorm.DB) *AsyncTransactionRepo {
	return &AsyncTransactionRepo{db}
}

func (r *AsyncTransactionRepo) Create(tx *domain.AsyncTransaction) error {
	model := &AsyncTransactionStatusModel{
		ID:          tx.ID.String(),
		FromAccount: tx.FromAccount,
		ToAccount:   tx.ToAccount,
		Amount:      tx.Amount,
		Status:      string(tx.Status),
		CreatedAt:   tx.CreatedAt,
		UpdatedAt:   tx.UpdatedAt,
	}
	return r.db.Create(model).Error
}

func (r *AsyncTransactionRepo) GetByID(id uuid.UUID) (*domain.AsyncTransaction, error) {
	var m AsyncTransactionStatusModel
	if err := r.db.Where("id = ?", id.String()).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrTransactionNotFound
		}
		return nil, err
	}
	uid, _ := uuid.Parse(m.ID)
	return &domain.AsyncTransaction{
		ID:          uid,
		FromAccount: m.FromAccount,
		ToAccount:   m.ToAccount,
		Amount:      m.Amount,
		Status:      domain.TxStatus(m.Status),
		Error:       m.Error,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}, nil
}

func (r *AsyncTransactionRepo) UpdateStatus(id uuid.UUID, status domain.TxStatus, errMsg string) error {
	return r.db.Model(&AsyncTransactionStatusModel{}).
		Where("id = ?", id.String()).
		Updates(map[string]interface{}{
			"status":     string(status),
			"error":      errMsg,
			"updated_at": time.Now(),
		}).Error
}
