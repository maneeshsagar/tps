package dto

type AccountResponse struct {
	AccountID int64  `json:"account_id"`
	Balance   string `json:"balance"`
}

type TransactionResponse struct {
	TransactionID        string `json:"transaction_id"`
	SourceAccountID      int64  `json:"source_account_id"`
	DestinationAccountID int64  `json:"destination_account_id"`
	Amount               string `json:"amount"`
}

type AsyncTransactionResponse struct {
	TransactionID string `json:"transaction_id"`
	Status        string `json:"status"`
}

type AsyncStatusResponse struct {
	TransactionID string `json:"transaction_id"`
	FromAccount   int64  `json:"from_account"`
	ToAccount     int64  `json:"to_account"`
	Amount        string `json:"amount"`
	Status        string `json:"status"`
	Error         string `json:"error,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type HealthResponse struct {
	Status string `json:"status"`
}
