package domain

import "errors"

var (
	ErrInsufficientBalance   = errors.New("insufficient balance")
	ErrAccountNotFound       = errors.New("account not found")
	ErrAccountAlreadyExists  = errors.New("account already exists")
	ErrInvalidAccountID      = errors.New("invalid account ID")
	ErrInvalidAmount         = errors.New("invalid amount")
	ErrSameAccount           = errors.New("same account")
	ErrLockAcquisitionFailed = errors.New("lock acquisition failed")
	ErrTransactionNotFound   = errors.New("transaction not found")
)
