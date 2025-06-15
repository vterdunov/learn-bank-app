package domain

import (
	"errors"
	"time"
)

// Account представляет банковский счет
type Account struct {
	ID        int       `json:"id" db:"id"`
	UserID    int       `json:"user_id" db:"user_id"`
	Number    string    `json:"number" db:"number"`
	Balance   float64   `json:"balance" db:"balance"`
	Currency  string    `json:"currency" db:"currency"`
	Status    string    `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// CreateAccountRequest представляет запрос на создание счета
type CreateAccountRequest struct {
	Currency string `json:"currency"`
}

// DepositRequest представляет запрос на пополнение счета
type DepositRequest struct {
	Amount float64 `json:"amount"`
}

// WithdrawRequest представляет запрос на списание со счета
type WithdrawRequest struct {
	Amount float64 `json:"amount"`
}

// TransferRequest представляет запрос на перевод между счетами
type TransferRequest struct {
	FromAccountID int     `json:"from_account_id"`
	ToAccountID   int     `json:"to_account_id"`
	Amount        float64 `json:"amount"`
}

// AccountStatus определяет возможные статусы счета
const (
	AccountStatusActive  = "active"
	AccountStatusBlocked = "blocked"
	AccountStatusClosed  = "closed"
)

// Validation errors
var (
	ErrInvalidAccountBalance  = errors.New("invalid account balance")
	ErrInvalidAccountCurrency = errors.New("invalid account currency")
	ErrInvalidAccountStatus   = errors.New("invalid account status")
	ErrInvalidDepositAmount   = errors.New("invalid deposit amount")
	ErrInvalidWithdrawAmount  = errors.New("invalid withdraw amount")
	ErrInvalidTransferAmount  = errors.New("invalid transfer amount")
)

// Validate валидирует счет
func (a *Account) Validate() error {
	if a.Balance < 0 {
		return ErrInvalidAccountBalance
	}
	if a.Currency != "RUB" {
		return ErrInvalidAccountCurrency
	}
	if a.Status != AccountStatusActive && a.Status != AccountStatusBlocked && a.Status != AccountStatusClosed {
		return ErrInvalidAccountStatus
	}
	return nil
}

// Validate валидирует запрос на создание счета
func (r *CreateAccountRequest) Validate() error {
	if r.Currency != "RUB" {
		return ErrInvalidAccountCurrency
	}
	return nil
}

// Validate валидирует запрос на пополнение
func (r *DepositRequest) Validate() error {
	if r.Amount <= 0 {
		return ErrInvalidDepositAmount
	}
	if r.Amount > 1000000000 {
		return ErrInvalidDepositAmount
	}
	return nil
}

// Validate валидирует запрос на списание
func (r *WithdrawRequest) Validate() error {
	if r.Amount <= 0 {
		return ErrInvalidWithdrawAmount
	}
	if r.Amount > 1000000000 {
		return ErrInvalidWithdrawAmount
	}
	return nil
}

// Validate валидирует запрос на перевод
func (r *TransferRequest) Validate() error {
	if r.Amount <= 0 {
		return ErrInvalidTransferAmount
	}
	if r.Amount > 1000000000 {
		return ErrInvalidTransferAmount
	}
	if r.FromAccountID <= 0 || r.ToAccountID <= 0 {
		return errors.New("invalid account IDs")
	}
	if r.FromAccountID == r.ToAccountID {
		return errors.New("cannot transfer to the same account")
	}
	return nil
}
