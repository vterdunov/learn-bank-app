package domain

import (
	"errors"
	"time"
)

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

// Validation errors
var (
	ErrInvalidTransactionAmount = errors.New("invalid transaction amount")
	ErrInvalidTransactionType   = errors.New("invalid transaction type")
	ErrInvalidTransactionStatus = errors.New("invalid transaction status")
)

// Validate валидирует транзакцию
func (t *Transaction) Validate() error {
	if t.Amount <= 0 {
		return ErrInvalidTransactionAmount
	}
	if t.Amount > 1000000000 {
		return ErrInvalidTransactionAmount
	}

	validTypes := []string{
		TransactionTypeDeposit,
		TransactionTypeWithdraw,
		TransactionTypeTransfer,
		TransactionTypePayment,
		TransactionTypeCredit,
	}
	isValidType := false
	for _, validType := range validTypes {
		if t.Type == validType {
			isValidType = true
			break
		}
	}
	if !isValidType {
		return ErrInvalidTransactionType
	}

	validStatuses := []string{
		TransactionStatusPending,
		TransactionStatusCompleted,
		TransactionStatusFailed,
		TransactionStatusCancelled,
	}
	isValidStatus := false
	for _, validStatus := range validStatuses {
		if t.Status == validStatus {
			isValidStatus = true
			break
		}
	}
	if !isValidStatus {
		return ErrInvalidTransactionStatus
	}

	return nil
}
