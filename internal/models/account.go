package models

import "time"

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
