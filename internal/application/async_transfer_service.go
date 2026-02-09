package application

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/maneeshsagar/tps/internal/core/domain"
	"github.com/maneeshsagar/tps/internal/core/ports"
)

const (
	TopicTransactions    = "transactions"
	TopicTransactionsDLQ = "transactions-dlq"
	MaxRetries           = 3
)

type TransferMessage struct {
	ID     string `json:"id"`
	From   int64  `json:"from"`
	To     int64  `json:"to"`
	Amount int64  `json:"amount"`
	Retry  int    `json:"retry,omitempty"`
}

// TransferService only submits transfer requests to a queue and updates their status.
// The actual transfer logic is handled by the consumer.
func (s *TransferService) SubmitTransfer(ctx context.Context, from, to, amount int64) (uuid.UUID, error) {

	id := uuid.New()
	now := time.Now()

	tx := &domain.AsyncTransaction{
		ID:          id,
		FromAccount: from,
		ToAccount:   to,
		Amount:      amount,
		Status:      domain.TxStatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.asynctxns.Create(tx); err != nil {
		s.log.Error("failed to create async transaction", "id", id, "err", err)
		return uuid.Nil, err
	}

	msg := TransferMessage{
		ID:     id.String(),
		From:   from,
		To:     to,
		Amount: amount,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		s.log.Error("failed to marshal transfer message", "id", id, "err", err)
		s.asynctxns.UpdateStatus(id, domain.TxStatusFailed, "internal error")
		return uuid.Nil, err
	}

	s.log.Debug("Sending transfer message to queue", "id", id, "data", string(data))

	err = s.producer.Publish(ctx, TopicTransactions, ports.Message{
		Key:   id.String(),
		Value: data,
	})

	if err != nil {
		s.log.Error("failed to publish transfer message", "id", id, "err", err)
		s.asynctxns.UpdateStatus(id, domain.TxStatusFailed, "failed to queue")
		return uuid.Nil, err
	}

	s.log.Info("submitted transfer", "id", id, "from", from, "to", to, "amount", amount)

	return id, nil
}

// GetStatus returns the current status of submitted transaction
func (s *TransferService) GetStatus(ctx context.Context, id uuid.UUID) (*domain.AsyncTransaction, error) {
	return s.asynctxns.GetByID(id)
}

// ProcessTransfer is called by the consumer to process the transfer message.
// It updates the transaction status based on the outcome.
// It implements retry logic for transient errors and marks business errors as failed without retrying.
func (s *TransferService) ProcessTransfer(ctx context.Context, msg TransferMessage) error {
	id, err := uuid.Parse(msg.ID)
	if err != nil {
		s.log.Error("invalid message, sending to DLQ", "id", msg.ID, "err", err)
		s.sendToDLQ(ctx, msg, "invalid message id")
		return nil
	}

	s.log.Info("started processing transfer", "id", id, "from", msg.From, "to", msg.To, "amount", msg.Amount, "retry", msg.Retry)

	_, err = s.Transfer(ctx, msg.From, msg.To, msg.Amount)
	if err != nil {
		// business errors - no retry, mark as failed
		if isBusinessError(err) {
			s.log.Error("transfer failed", "id", msg.ID, "err", err)
			s.asynctxns.UpdateStatus(id, domain.TxStatusFailed, err.Error())
			return nil
		}

		// transient error - retry or send to DLQ
		if msg.Retry >= MaxRetries {
			s.log.Error("max retries exceeded, sending to DLQ", "id", msg.ID, "retry", msg.Retry)
			s.asynctxns.UpdateStatus(id, domain.TxStatusFailed, "max retries exceeded")
			s.sendToDLQ(ctx, msg, err.Error())
			return nil
		}

		// retry
		msg.Retry++
		s.log.Warn("retrying transfer", "id", msg.ID, "retry", msg.Retry, "err", err)
		s.requeue(ctx, msg)
		return nil
	}

	s.asynctxns.UpdateStatus(id, domain.TxStatusCompleted, "")
	s.log.Info("transfer completed", "id", msg.ID)
	return nil
}

// helper methods for retry and DLQ handling
func (s *TransferService) requeue(ctx context.Context, msg TransferMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		s.log.Error("failed to marshal transfer message for retry", "id", msg.ID, "err", err)
		return
	}

	s.log.Debug("Requeuing transfer message", "id", msg.ID, "retry", msg.Retry, "data", string(data))
	requeMsg := ports.Message{
		Key:   msg.ID,
		Value: data,
	}

	err = s.producer.Publish(ctx, TopicTransactions, requeMsg)

	if err != nil {
		s.log.Error("failed to requeue", "id", msg.ID, "err", err)
	}
}

type DeadLaterQueueMessage struct {
	TransferMessage
	Reason string `json:"reason"`
}

// sendToDLQ sends the failed message to a Dead Letter Queue for further analysis
func (s *TransferService) sendToDLQ(ctx context.Context, msg TransferMessage, reason string) {

	dlqMsg := DeadLaterQueueMessage{
		TransferMessage: msg,
		Reason:          reason,
	}

	data, err := json.Marshal(dlqMsg)
	if err != nil {
		s.log.Error("failed to marshal DLQ message", "id", msg.ID, "err", err)
		return
	}

	s.log.Debug("Sending message to DLQ", "id", msg.ID, "reason", reason, "data", string(data))

	dequeMsg := ports.Message{
		Key:   msg.ID,
		Value: data,
	}

	err = s.producer.Publish(ctx, TopicTransactionsDLQ, dequeMsg)
	if err != nil {
		s.log.Error("failed to send to DLQ", "id", msg.ID, "err", err)
	}
}

// isBusinessError checks if the error is a known business error that should not be retried
func isBusinessError(err error) bool {
	return errors.Is(err, domain.ErrInsufficientBalance) ||
		errors.Is(err, domain.ErrAccountNotFound) ||
		errors.Is(err, domain.ErrInvalidAmount) ||
		errors.Is(err, domain.ErrSameAccount)
}
