package ports

import "context"

// Transaction is a marker interface for db transactions
type Transaction interface {
}

type TransactionManager interface {
	WithTransaction(ctx context.Context, fn func(tx Transaction) error) error
}
