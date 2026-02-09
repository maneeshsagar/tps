package ports

import "github.com/maneeshsagar/tps/internal/core/domain"

type AccountRepository interface {
	GetByID(id int64) (*domain.Account, error)
	Update(account *domain.Account) error
	Create(account *domain.Account) error
	WithTx(tx Transaction) AccountRepository
}
