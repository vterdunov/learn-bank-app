package models

import "time"

// PaymentSchedule представляет график платежей по кредиту
type PaymentSchedule struct {
	ID               int        `json:"id" db:"id"`
	CreditID         int        `json:"credit_id" db:"credit_id"`
	PaymentNumber    int        `json:"payment_number" db:"payment_number"`
	DueDate          time.Time  `json:"due_date" db:"due_date"`
	PaymentAmount    float64    `json:"payment_amount" db:"payment_amount"`
	PrincipalAmount  float64    `json:"principal_amount" db:"principal_amount"`
	InterestAmount   float64    `json:"interest_amount" db:"interest_amount"`
	RemainingBalance float64    `json:"remaining_balance" db:"remaining_balance"`
	Status           string     `json:"status" db:"status"`
	PaidDate         *time.Time `json:"paid_date" db:"paid_date"`
	PenaltyAmount    float64    `json:"penalty_amount" db:"penalty_amount"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

// PaymentStatus определяет статусы платежей
const (
	PaymentStatusPending   = "pending"
	PaymentStatusPaid      = "paid"
	PaymentStatusOverdue   = "overdue"
	PaymentStatusCancelled = "cancelled"
)

// CreditAnalytics представляет аналитику по кредиту
type CreditAnalytics struct {
	TotalCredits    int     `json:"total_credits"`
	TotalDebt       float64 `json:"total_debt"`
	MonthlyPayments float64 `json:"monthly_payments"`
	OverduePayments int     `json:"overdue_payments"`
}

// MonthlyStatistics представляет месячную статистику
type MonthlyStatistics struct {
	Income   float64 `json:"income"`
	Expenses float64 `json:"expenses"`
	Balance  float64 `json:"balance"`
	Month    int     `json:"month"`
	Year     int     `json:"year"`
}

// BalancePrediction представляет прогноз баланса
type BalancePrediction struct {
	Date              time.Time `json:"date"`
	PredictedBalance  float64   `json:"predicted_balance"`
	ScheduledPayments float64   `json:"scheduled_payments"`
}
