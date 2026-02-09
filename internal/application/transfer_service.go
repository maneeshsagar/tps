package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/maneeshsagar/tps/internal/core/domain"
	"github.com/maneeshsagar/tps/internal/core/ports"
	"github.com/maneeshsagar/tps/logger"
)

type TransferResult struct {
	TransactionID uuid.UUID
}

type TransferServiceIntf interface {
	CreateAccount(ctx context.Context, id, balance int64) error
	GetAccount(ctx context.Context, id int64) (*domain.Account, error)
	Transfer(ctx context.Context, from, to, amount int64) (*TransferResult, error)
	SubmitTransfer(ctx context.Context, from, to, amount int64) (uuid.UUID, error)
	GetStatus(ctx context.Context, id uuid.UUID) (*domain.AsyncTransaction, error)
	ProcessTransfer(ctx context.Context, msg TransferMessage) error
}

type TransferService struct {
	accounts  ports.AccountRepository
	txns      ports.TransactionRepository
	asynctxns ports.AsyncTransactionRepository
	db        ports.TransactionManager
	locks     ports.LockManager
	producer  ports.MessageProducer
	log       logger.Logger
}

func NewTransferService(
	accounts ports.AccountRepository,
	txns ports.TransactionRepository,
	asynctxns ports.AsyncTransactionRepository,
	db ports.TransactionManager,
	locks ports.LockManager,
	producer ports.MessageProducer,
	log logger.Logger,
) TransferServiceIntf {
	return &TransferService{accounts, txns, asynctxns, db, locks, producer, log}
}

// transfer money between two accounts
func (s *TransferService) Transfer(ctx context.Context, fromAccountID, toAccountID, amount int64) (*TransferResult, error) {
	// check if transfer amount is zero
	if amount <= 0 {
		return nil, domain.ErrInvalidAmount
	}
	// check if souce and destination accounts are same
	if fromAccountID == toAccountID {
		return nil, domain.ErrSameAccount
	}

	// acquire locks on both accounts to prevent concurrent modifications
	unlock, err := s.locks.LockAccounts(ctx, []int64{fromAccountID, toAccountID}, 10*time.Second)
	if err != nil {
		return nil, err
	}

	// Release locks after transfer attempt (success or failure) on function exit
	defer unlock()

	txID := uuid.New()
	err = s.db.WithTransaction(ctx, func(tx ports.Transaction) error {

		// started the transaction and got a transactional context, now get transactional repositories
		acctRepo := s.accounts.WithTx(tx)
		txnRepo := s.txns.WithTx(tx)

		from, err := acctRepo.GetByID(fromAccountID)
		if err != nil {
			return err
		}
		to, err := acctRepo.GetByID(toAccountID)
		if err != nil {
			return err
		}

		if err := from.Debit(amount); err != nil {
			return err
		}
		if err := to.Credit(amount); err != nil {
			return err
		}

		rec := &domain.Transaction{
			ID:                   txID,
			SourceAccountID:      fromAccountID,
			DestinationAccountID: toAccountID,
			Amount:               amount,
			CreatedAt:            time.Now(),
		}
		if err := txnRepo.Create(rec); err != nil {
			return err
		}
		if err := acctRepo.Update(from); err != nil {
			return err
		}
		return acctRepo.Update(to)
	})

	if err != nil {
		s.log.Error("transfer failed", "err", err)
		return nil, err
	}

	return &TransferResult{TransactionID: txID}, nil
}

// create a new account
func (s *TransferService) CreateAccount(ctx context.Context, id, balance int64) error {
	if id <= 0 {
		return domain.ErrInvalidAccountID
	}
	if balance < 0 {
		return domain.ErrInvalidAmount
	}

	acct := &domain.Account{AccountID: id, Balance: balance}
	return s.accounts.Create(acct)
}

// get an account by id
func (s *TransferService) GetAccount(ctx context.Context, id int64) (*domain.Account, error) {
	return s.accounts.GetByID(id)
}
