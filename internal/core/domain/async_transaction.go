package domain

import (
	"time"

	"github.com/google/uuid"
)

type TxStatus string

const (
	TxStatusPending   TxStatus = "pending"
	TxStatusCompleted TxStatus = "completed"
	TxStatusFailed    TxStatus = "failed"
)

type AsyncTransaction struct {
	ID          uuid.UUID
	FromAccount int64
	ToAccount   int64
	Amount      int64
	Status      TxStatus
	Error       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
