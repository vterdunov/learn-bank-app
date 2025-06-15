package domain

import (
	"errors"
	"math"
	"time"
)

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

// CalculateAnnuityPayment рассчитывает аннуитетный платеж для кредита
func (c *Credit) CalculateAnnuityPayment() float64 {
	return CalculateAnnuityPayment(c.Amount, c.InterestRate, c.TermMonths)
}

// CalculateAnnuityPayment рассчитывает аннуитетный платеж по формуле
func CalculateAnnuityPayment(principal, rate float64, months int) float64 {
	if principal <= 0 || rate < 0 || months <= 0 {
		return 0
	}

	// Преобразуем годовую ставку в месячную
	monthlyRate := rate / 100 / 12

	// Формула аннуитетного платежа:
	// PMT = P * (r * (1 + r)^n) / ((1 + r)^n - 1)
	// где P - основная сумма, r - месячная ставка, n - количество месяцев

	if monthlyRate == 0 {
		// Если ставка 0%, то просто делим сумму на количество месяцев
		return principal / float64(months)
	}

	numerator := monthlyRate * math.Pow(1+monthlyRate, float64(months))
	denominator := math.Pow(1+monthlyRate, float64(months)) - 1

	payment := principal * (numerator / denominator)

	// Округляем до 2 знаков после запятой
	return math.Round(payment*100) / 100
}

// CalculatePaymentBreakdown рассчитывает разбивку платежа на основной долг и проценты
func (c *Credit) CalculatePaymentBreakdown(paymentNumber int, remainingPrincipal float64) (principalAmount, interestAmount float64) {
	monthlyRate := c.InterestRate / 100 / 12

	// Рассчитываем проценты с остатка
	interestAmount = remainingPrincipal * monthlyRate

	// Основной долг = аннуитетный платеж - проценты
	principalAmount = c.MonthlyPayment - interestAmount

	// Для последнего платежа корректируем сумму основного долга
	if paymentNumber == c.TermMonths {
		principalAmount = remainingPrincipal
		interestAmount = c.MonthlyPayment - principalAmount
	}

	return math.Round(principalAmount*100) / 100, math.Round(interestAmount*100) / 100
}

// IsActive проверяет, активен ли кредит
func (c *Credit) IsActive() bool {
	return c.Status == CreditStatusActive
}

// IsPaidOff проверяет, погашен ли кредит
func (c *Credit) IsPaidOff() bool {
	return c.Status == CreditStatusPaidOff || c.RemainingDebt <= 0
}

// GetTotalCost возвращает общую стоимость кредита
func (c *Credit) GetTotalCost() float64 {
	return c.MonthlyPayment * float64(c.TermMonths)
}

// GetTotalInterest возвращает общую сумму процентов
func (c *Credit) GetTotalInterest() float64 {
	return c.GetTotalCost() - c.Amount
}

// GetEffectiveRate возвращает эффективную процентную ставку
func (c *Credit) GetEffectiveRate() float64 {
	if c.Amount <= 0 {
		return 0
	}
	totalInterest := c.GetTotalInterest()
	return (totalInterest / c.Amount) * 100
}

// UpdateRemainingDebt обновляет остаток задолженности
func (c *Credit) UpdateRemainingDebt(paidPrincipal float64) {
	c.RemainingDebt -= paidPrincipal
	if c.RemainingDebt <= 0 {
		c.RemainingDebt = 0
		c.Status = CreditStatusPaidOff
	}
	c.UpdatedAt = time.Now()
}

// CreateCreditRequest представляет запрос на создание кредита
type CreateCreditRequest struct {
	AccountID    int     `json:"account_id"`
	Amount       float64 `json:"amount"`
	TermMonths   int     `json:"term_months"`
	InterestRate float64 `json:"interest_rate"`
}

// Validate валидирует запрос на создание кредита
func (r *CreateCreditRequest) Validate() error {
	if r.Amount <= 0 {
		return ErrInvalidCreditAmount
	}
	if r.TermMonths <= 0 || r.TermMonths > 360 { // максимум 30 лет
		return ErrInvalidCreditTerm
	}
	if r.AccountID <= 0 {
		return ErrInvalidAccountID
	}
	return nil
}

// CreditStatus определяет статусы кредита
const (
	CreditStatusActive    = "active"
	CreditStatusPaidOff   = "paid_off"
	CreditStatusOverdue   = "overdue"
	CreditStatusCancelled = "cancelled"
)

// Domain errors
var (
	ErrInvalidCreditAmount   = errors.New("invalid credit amount")
	ErrInvalidCreditTerm     = errors.New("invalid credit term")
	ErrInvalidAccountID      = errors.New("invalid account ID")
	ErrInvalidInterestRate   = errors.New("invalid interest rate")
	ErrInvalidMonthlyPayment = errors.New("invalid monthly payment")
	ErrInvalidRemainingDebt  = errors.New("invalid remaining debt")
	ErrInvalidCreditStatus   = errors.New("invalid credit status")
)

// Validate валидирует кредит
func (c *Credit) Validate() error {
	if c.Amount <= 0 {
		return ErrInvalidCreditAmount
	}
	if c.Amount > 100000000 { // максимум 100 млн
		return ErrInvalidCreditAmount
	}
	if c.InterestRate < 0 || c.InterestRate > 100 {
		return ErrInvalidInterestRate
	}
	if c.TermMonths <= 0 || c.TermMonths > 360 {
		return ErrInvalidCreditTerm
	}
	if c.MonthlyPayment <= 0 {
		return ErrInvalidMonthlyPayment
	}
	if c.RemainingDebt < 0 {
		return ErrInvalidRemainingDebt
	}

	validStatuses := []string{
		CreditStatusActive,
		CreditStatusPaidOff,
		CreditStatusOverdue,
		CreditStatusCancelled,
	}
	isValidStatus := false
	for _, validStatus := range validStatuses {
		if c.Status == validStatus {
			isValidStatus = true
			break
		}
	}
	if !isValidStatus {
		return ErrInvalidCreditStatus
	}

	return nil
}
