package dto

type CreateAccountRequest struct {
	AccountID      int64  `json:"account_id" `
	InitialBalance string `json:"initial_balance" `
}

type CreateTransactionRequest struct {
	SourceAccountID      int64  `json:"source_account_id"`
	DestinationAccountID int64  `json:"destination_account_id"`
	Amount               string `json:"amount" binding:"required"`
}
