package ports

import "github.com/maneeshsagar/tps/internal/core/domain"

type TransactionRepository interface {
	Create(tx *domain.Transaction) error
	WithTx(tx Transaction) TransactionRepository
}
