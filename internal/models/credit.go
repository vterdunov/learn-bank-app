package models

import "time"

// Credit представляет кредит
type Credit struct {
	ID             int       `json:"id" db:"id"`
	UserID         int       `json:"user_id" db:"user_id"`
	AccountID      int       `json:"account_id" db:"account_id"`
	Amount         float64   `json:"amount" db:"amount"`
	InterestRate   float64   `json:"interest_rate" db:"interest_rate"`
	TermMonths     int       `json:"term_months" db:"term_months"`
	MonthlyPayment float64   `json:"monthly_payment" db:"monthly_payment"`
	RemainingDebt  float64   `json:"remaining_debt" db:"remaining_debt"`
	Status         string    `json:"status" db:"status"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// CreateCreditRequest представляет запрос на создание кредита
type CreateCreditRequest struct {
	AccountID    int     `json:"account_id"`
	Amount       float64 `json:"amount"`
	TermMonths   int     `json:"term_months"`
	InterestRate float64 `json:"interest_rate"`
}

// CreditStatus определяет статусы кредита
const (
	CreditStatusActive    = "active"
	CreditStatusPaidOff   = "paid_off"
	CreditStatusOverdue   = "overdue"
	CreditStatusCancelled = "cancelled"
)
