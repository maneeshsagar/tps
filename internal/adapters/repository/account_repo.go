package repository

import (
	"errors"
	"strings"

	"github.com/maneeshsagar/tps/internal/core/domain"
	"gorm.io/gorm"
)

type AccountModel struct {
	AccountID int64 `gorm:"primaryKey;column:account_id"`
	Balance   int64 `gorm:"column:balance"`
}

func (AccountModel) TableName() string {
	return "accounts"
}

type AccountRepo struct {
	db *gorm.DB
}

func NewAccountRepo(db *gorm.DB) *AccountRepo {
	return &AccountRepo{db}
}

func (r *AccountRepo) GetByID(id int64) (*domain.Account, error) {
	var m AccountModel
	if err := r.db.First(&m, "account_id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrAccountNotFound
		}
		return nil, err
	}
	return &domain.Account{
		AccountID: m.AccountID,
		Balance:   m.Balance,
	}, nil
}

func (r *AccountRepo) Update(account *domain.Account) error {
	result := r.db.Model(&AccountModel{}).
		Where("account_id = ?", account.AccountID).
		Update("balance", account.Balance)

	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *AccountRepo) Create(account *domain.Account) error {
	m := AccountModel{
		AccountID: account.AccountID,
		Balance:   account.Balance,
	}

	if err := r.db.Create(&m).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate key") ||
			strings.Contains(err.Error(), "unique constraint") {
			return domain.ErrAccountAlreadyExists
		}
		return err
	}
	return nil
}
