package domain

import (
	"errors"
	"time"
)

// PaymentSchedule представляет график платежей по кредиту
type PaymentSchedule struct {
	ID               int        `json:"id" db:"id"`
	CreditID         int        `json:"credit_id" db:"credit_id"`
	PaymentNumber    int        `json:"payment_number" db:"payment_number"`
	DueDate          time.Time  `json:"due_date" db:"due_date"`
	PaymentAmount    float64    `json:"payment_amount" db:"payment_amount"`
	PrincipalAmount  float64    `json:"principal_amount" db:"principal_amount"`
	InterestAmount   float64    `json:"interest_amount" db:"interest_amount"`
	PenaltyAmount    float64    `json:"penalty_amount" db:"penalty_amount"`
	PaidAmount       float64    `json:"paid_amount" db:"paid_amount"`
	RemainingBalance float64    `json:"remaining_balance" db:"remaining_balance"`
	Status           string     `json:"status" db:"status"`
	PaidAt           *time.Time `json:"paid_at" db:"paid_at"`
	PaidDate         *time.Time `json:"paid_date" db:"paid_date"`
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

// Validation errors
var (
	ErrInvalidSchedulePaymentAmount = errors.New("invalid payment amount")
	ErrInvalidSchedulePaymentStatus = errors.New("invalid payment status")
	ErrInvalidSchedulePaymentNumber = errors.New("invalid payment number")
	ErrInvalidRemainingBalance      = errors.New("invalid remaining balance")
)

// Validate валидирует график платежей
func (ps *PaymentSchedule) Validate() error {
	if ps.PaymentNumber <= 0 {
		return ErrInvalidSchedulePaymentNumber
	}
	if ps.PaymentAmount <= 0 {
		return ErrInvalidSchedulePaymentAmount
	}
	if ps.PrincipalAmount < 0 {
		return errors.New("principal amount cannot be negative")
	}
	if ps.InterestAmount < 0 {
		return errors.New("interest amount cannot be negative")
	}
	if ps.PenaltyAmount < 0 {
		return errors.New("penalty amount cannot be negative")
	}
	if ps.PaidAmount < 0 {
		return errors.New("paid amount cannot be negative")
	}
	if ps.RemainingBalance < 0 {
		return ErrInvalidRemainingBalance
	}

	validStatuses := []string{
		PaymentStatusPending,
		PaymentStatusPaid,
		PaymentStatusOverdue,
		PaymentStatusCancelled,
	}
	isValidStatus := false
	for _, validStatus := range validStatuses {
		if ps.Status == validStatus {
			isValidStatus = true
			break
		}
	}
	if !isValidStatus {
		return ErrInvalidSchedulePaymentStatus
	}

	return nil
}
