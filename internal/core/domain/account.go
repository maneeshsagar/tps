package domain

// Account represents a bank account in the domain
type Account struct {
	AccountID int64
	Balance   int64
}

func (a *Account) CanDebit(amount int64) bool {
	return a.Balance >= amount
}

func (a *Account) Debit(amount int64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	if !a.CanDebit(amount) {
		return ErrInsufficientBalance
	}
	a.Balance -= amount
	return nil
}

func (a *Account) Credit(amount int64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	a.Balance += amount
	return nil
}
