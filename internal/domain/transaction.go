package domain

import "time"

// Transaction представляет транзакцию
type Transaction struct {
	ID          int       `json:"id" db:"id"`
	FromAccount *int      `json:"from_account" db:"from_account"`
	ToAccount   *int      `json:"to_account" db:"to_account"`
	Amount      float64   `json:"amount" db:"amount"`
	Type        string    `json:"type" db:"type"`
	Status      string    `json:"status" db:"status"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// TransactionType определяет типы транзакций
const (
	TransactionTypeDeposit  = "deposit"
	TransactionTypeWithdraw = "withdraw"
	TransactionTypeTransfer = "transfer"
	TransactionTypePayment  = "payment"
	TransactionTypeCredit   = "credit"
)

// TransactionStatus определяет статусы транзакций
const (
	TransactionStatusPending   = "pending"
	TransactionStatusCompleted = "completed"
	TransactionStatusFailed    = "failed"
	TransactionStatusCancelled = "cancelled"
)
