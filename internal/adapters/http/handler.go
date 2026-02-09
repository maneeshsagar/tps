package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/maneeshsagar/tps/internal/adapters/http/dto"
	"github.com/maneeshsagar/tps/internal/application"
	"github.com/maneeshsagar/tps/internal/core/domain"
	"github.com/maneeshsagar/tps/pkg/currency"
)

type Handler struct {
	svc application.TransferServiceIntf
}

func NewHandler(svc application.TransferServiceIntf) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, dto.HealthResponse{Status: "healthy"})
}

func (h *Handler) CreateAccount(c *gin.Context) {
	var req dto.CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid request"})
		return
	}

	balance, err := currency.RupeesToPaise(req.InitialBalance)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid balance format"})
		return
	}
	if balance < 0 {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "balance cannot be negative"})
		return
	}

	if err := h.svc.CreateAccount(c, req.AccountID, balance); err != nil {
		h.handleErr(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.AccountResponse{
		AccountID: req.AccountID,
		Balance:   currency.PaiseToRupees(balance),
	})
}

func (h *Handler) GetAccount(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("account_id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid account_id"})
		return
	}

	acc, err := h.svc.GetAccount(c, id)
	if err != nil {
		h.handleErr(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.AccountResponse{
		AccountID: acc.AccountID,
		Balance:   currency.PaiseToRupees(acc.Balance),
	})
}

func (h *Handler) CreateTransaction(c *gin.Context) {
	var req dto.CreateTransactionRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid request"})
		return
	}

	// check if source and destination accounts are same
	if req.SourceAccountID == req.DestinationAccountID {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "same account"})
		return
	}

	// convert amount to paise, this is best practice to avoid floating point arithmetic issues
	amount, err := currency.RupeesToPaise(req.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid amount format"})
		return
	}
	if amount <= 0 {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "amount must be positive"})
		return
	}

	result, err := h.svc.Transfer(c, req.SourceAccountID, req.DestinationAccountID, amount)
	if err != nil {
		h.handleErr(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.TransactionResponse{
		TransactionID:        result.TransactionID.String(),
		SourceAccountID:      req.SourceAccountID,
		DestinationAccountID: req.DestinationAccountID,
		Amount:               currency.PaiseToRupees(amount),
	})
}

func (h *Handler) CreateAsyncTransaction(c *gin.Context) {
	var req dto.CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid request"})
		return
	}

	if req.SourceAccountID == req.DestinationAccountID {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "same account"})
		return
	}

	amount, err := currency.RupeesToPaise(req.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid amount format"})
		return
	}
	if amount <= 0 {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "amount must be positive"})
		return
	}

	id, err := h.svc.SubmitTransfer(c, req.SourceAccountID, req.DestinationAccountID, amount)
	if err != nil {
		h.handleErr(c, err)
		return
	}

	c.JSON(http.StatusAccepted, dto.AsyncTransactionResponse{
		TransactionID: id.String(),
		Status:        string(domain.TxStatusPending),
	})
}

func (h *Handler) GetAsyncTransactionStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid transaction id"})
		return
	}

	tx, err := h.svc.GetStatus(c, id)
	if err != nil {
		h.handleErr(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.AsyncStatusResponse{
		TransactionID: tx.ID.String(),
		FromAccount:   tx.FromAccount,
		ToAccount:     tx.ToAccount,
		Amount:        currency.PaiseToRupees(tx.Amount),
		Status:        string(tx.Status),
		Error:         tx.Error,
	})
}

func (h *Handler) handleErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrAccountNotFound):
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "account not found"})
	case errors.Is(err, domain.ErrAccountAlreadyExists):
		c.JSON(http.StatusConflict, dto.ErrorResponse{Error: "account exists"})
	case errors.Is(err, domain.ErrInsufficientBalance):
		c.JSON(http.StatusUnprocessableEntity, dto.ErrorResponse{Error: "insufficient balance"})
	case errors.Is(err, domain.ErrLockAcquisitionFailed):
		c.JSON(http.StatusServiceUnavailable, dto.ErrorResponse{Error: "busy, retry"})
	case errors.Is(err, domain.ErrTransactionNotFound):
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "transaction not found"})
	default:
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal error"})
	}
}
