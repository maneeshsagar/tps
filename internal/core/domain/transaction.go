package domain

import (
	"time"

	"github.com/google/uuid"
)

// Transaction represents a money transfer in the domain
type Transaction struct {
	ID                   uuid.UUID
	SourceAccountID      int64
	DestinationAccountID int64
	Amount               int64
	CreatedAt            time.Time
}
