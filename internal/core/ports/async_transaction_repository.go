package ports

import (
	"github.com/google/uuid"
	"github.com/maneeshsagar/tps/internal/core/domain"
)

type AsyncTransactionRepository interface {
	Create(tx *domain.AsyncTransaction) error
	GetByID(id uuid.UUID) (*domain.AsyncTransaction, error)
	UpdateStatus(id uuid.UUID, status domain.TxStatus, errMsg string) error
}
