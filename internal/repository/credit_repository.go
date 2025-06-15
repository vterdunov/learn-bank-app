package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vterdunov/learn-bank-app/internal/domain"
)

// CreditRepositoryImpl реализация CreditRepository
type CreditRepositoryImpl struct {
	db *pgxpool.Pool
}

// NewCreditRepository создает новый экземпляр CreditRepository
func NewCreditRepository(db *pgxpool.Pool) CreditRepository {
	return &CreditRepositoryImpl{db: db}
}

// Create создает новый кредит
func (r *CreditRepositoryImpl) Create(ctx context.Context, credit *domain.Credit) error {
	query := `
		INSERT INTO credits (user_id, account_id, amount, interest_rate, term_months, monthly_payment, remaining_debt, status, start_date, end_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id`

	now := time.Now()
	credit.CreatedAt = now
	credit.UpdatedAt = now

	// Рассчитываем дату окончания кредита
	startDate := now
	endDate := startDate.AddDate(0, credit.TermMonths, 0)

	err := r.db.QueryRow(ctx, query,
		credit.UserID,
		credit.AccountID,
		credit.Amount,
		credit.InterestRate,
		credit.TermMonths,
		credit.MonthlyPayment,
		credit.RemainingDebt,
		credit.Status,
		startDate,
		endDate,
		credit.CreatedAt,
		credit.UpdatedAt,
	).Scan(&credit.ID)

	return err
}

// GetByID получает кредит по ID
func (r *CreditRepositoryImpl) GetByID(ctx context.Context, id int) (*domain.Credit, error) {
	query := `
		SELECT id, user_id, account_id, amount, interest_rate, term_months, monthly_payment, remaining_debt, status, created_at, updated_at
		FROM credits
		WHERE id = $1`

	credit := &domain.Credit{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&credit.ID,
		&credit.UserID,
		&credit.AccountID,
		&credit.Amount,
		&credit.InterestRate,
		&credit.TermMonths,
		&credit.MonthlyPayment,
		&credit.RemainingDebt,
		&credit.Status,
		&credit.CreatedAt,
		&credit.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("credit not found")
		}
		return nil, err
	}

	return credit, nil
}

// GetByUserID получает все кредиты пользователя
func (r *CreditRepositoryImpl) GetByUserID(ctx context.Context, userID int) ([]*domain.Credit, error) {
	query := `
		SELECT id, user_id, account_id, amount, interest_rate, term_months, monthly_payment, remaining_debt, status, created_at, updated_at
		FROM credits
		WHERE user_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var credits []*domain.Credit
	for rows.Next() {
		credit := &domain.Credit{}
		err := rows.Scan(
			&credit.ID,
			&credit.UserID,
			&credit.AccountID,
			&credit.Amount,
			&credit.InterestRate,
			&credit.TermMonths,
			&credit.MonthlyPayment,
			&credit.RemainingDebt,
			&credit.Status,
			&credit.CreatedAt,
			&credit.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		credits = append(credits, credit)
	}

	return credits, nil
}

// GetByAccountID получает все кредиты счета
func (r *CreditRepositoryImpl) GetByAccountID(ctx context.Context, accountID int) ([]*domain.Credit, error) {
	query := `
		SELECT id, user_id, account_id, amount, interest_rate, term_months, monthly_payment, remaining_debt, status, created_at, updated_at
		FROM credits
		WHERE account_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var credits []*domain.Credit
	for rows.Next() {
		credit := &domain.Credit{}
		err := rows.Scan(
			&credit.ID,
			&credit.UserID,
			&credit.AccountID,
			&credit.Amount,
			&credit.InterestRate,
			&credit.TermMonths,
			&credit.MonthlyPayment,
			&credit.RemainingDebt,
			&credit.Status,
			&credit.CreatedAt,
			&credit.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		credits = append(credits, credit)
	}

	return credits, nil
}

// Update обновляет данные кредита
func (r *CreditRepositoryImpl) Update(ctx context.Context, credit *domain.Credit) error {
	query := `
		UPDATE credits
		SET amount = $2, interest_rate = $3, term_months = $4, monthly_payment = $5, remaining_debt = $6, status = $7, updated_at = $8
		WHERE id = $1`

	credit.UpdatedAt = time.Now()

	result, err := r.db.Exec(ctx, query,
		credit.ID,
		credit.Amount,
		credit.InterestRate,
		credit.TermMonths,
		credit.MonthlyPayment,
		credit.RemainingDebt,
		credit.Status,
		credit.UpdatedAt,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("credit not found")
	}

	return nil
}

// Delete удаляет кредит
func (r *CreditRepositoryImpl) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM credits WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("credit not found")
	}

	return nil
}

// UpdateRemainingDebt обновляет остаток долга по кредиту
func (r *CreditRepositoryImpl) UpdateRemainingDebt(ctx context.Context, id int, remainingDebt float64) error {
	query := `
		UPDATE credits
		SET remaining_debt = $2, updated_at = $3
		WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id, remainingDebt, time.Now())
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("credit not found")
	}

	return nil
}

// GetActiveCredits получает все активные кредиты
func (r *CreditRepositoryImpl) GetActiveCredits(ctx context.Context) ([]*domain.Credit, error) {
	query := `
		SELECT id, user_id, account_id, amount, interest_rate, term_months, monthly_payment, remaining_debt, status, created_at, updated_at
		FROM credits
		WHERE status = 'active' AND remaining_debt > 0
		ORDER BY created_at ASC`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var credits []*domain.Credit
	for rows.Next() {
		credit := &domain.Credit{}
		err := rows.Scan(
			&credit.ID,
			&credit.UserID,
			&credit.AccountID,
			&credit.Amount,
			&credit.InterestRate,
			&credit.TermMonths,
			&credit.MonthlyPayment,
			&credit.RemainingDebt,
			&credit.Status,
			&credit.CreatedAt,
			&credit.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		credits = append(credits, credit)
	}

	return credits, nil
}

// GetCreditAnalytics получает аналитику по кредитам пользователя
func (r *CreditRepositoryImpl) GetCreditAnalytics(ctx context.Context, userID int) (*domain.CreditAnalytics, error) {
	query := `
		SELECT
			COUNT(*) as total_credits,
			COALESCE(SUM(remaining_debt), 0) as total_debt,
			COALESCE(SUM(monthly_payment), 0) as monthly_payments,
			COUNT(CASE WHEN status = 'overdue' THEN 1 END) as overdue_payments
		FROM credits
		WHERE user_id = $1 AND status IN ('active', 'overdue')`

	analytics := &domain.CreditAnalytics{}
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&analytics.TotalCredits,
		&analytics.TotalDebt,
		&analytics.MonthlyPayments,
		&analytics.OverduePayments,
	)

	if err != nil {
		return nil, err
	}

	return analytics, nil
}

// PaymentScheduleRepositoryImpl реализация PaymentScheduleRepository
type PaymentScheduleRepositoryImpl struct {
	db *pgxpool.Pool
}

// NewPaymentScheduleRepository создает новый экземпляр PaymentScheduleRepository
func NewPaymentScheduleRepository(db *pgxpool.Pool) PaymentScheduleRepository {
	return &PaymentScheduleRepositoryImpl{db: db}
}

// Create создает новый платеж
func (r *PaymentScheduleRepositoryImpl) Create(ctx context.Context, payment *domain.PaymentSchedule) error {
	query := `
		INSERT INTO payment_schedules (credit_id, payment_number, due_date, payment_amount, principal_amount, interest_amount, status, penalty_amount, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id`

	now := time.Now()
	payment.CreatedAt = now
	payment.UpdatedAt = now

	err := r.db.QueryRow(ctx, query,
		payment.CreditID,
		payment.PaymentNumber,
		payment.DueDate,
		payment.PaymentAmount,
		payment.PrincipalAmount,
		payment.InterestAmount,
		payment.Status,
		payment.PenaltyAmount,
		payment.CreatedAt,
		payment.UpdatedAt,
	).Scan(&payment.ID)

	return err
}

// CreateBatch создает несколько платежей одновременно
func (r *PaymentScheduleRepositoryImpl) CreateBatch(ctx context.Context, payments []*domain.PaymentSchedule) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO payment_schedules (credit_id, payment_number, due_date, payment_amount, principal_amount, interest_amount, remaining_balance, status, penalty_amount, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	now := time.Now()
	for _, payment := range payments {
		payment.CreatedAt = now
		payment.UpdatedAt = now

		_, err := tx.Exec(ctx, query,
			payment.CreditID,
			payment.PaymentNumber,
			payment.DueDate,
			payment.PaymentAmount,
			payment.PrincipalAmount,
			payment.InterestAmount,
			payment.RemainingBalance,
			payment.Status,
			payment.PenaltyAmount,
			payment.CreatedAt,
			payment.UpdatedAt,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// GetByID получает платеж по ID
func (r *PaymentScheduleRepositoryImpl) GetByID(ctx context.Context, id int) (*domain.PaymentSchedule, error) {
	query := `
		SELECT id, credit_id, payment_number, due_date, payment_amount, principal_amount, interest_amount, remaining_balance, status, paid_date, penalty_amount, created_at, updated_at
		FROM payment_schedules
		WHERE id = $1`

	payment := &domain.PaymentSchedule{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&payment.ID,
		&payment.CreditID,
		&payment.PaymentNumber,
		&payment.DueDate,
		&payment.PaymentAmount,
		&payment.PrincipalAmount,
		&payment.InterestAmount,
		&payment.RemainingBalance,
		&payment.Status,
		&payment.PaidDate,
		&payment.PenaltyAmount,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("payment not found")
		}
		return nil, err
	}

	return payment, nil
}

// scanPaymentSchedules сканирует результаты запроса в слайс PaymentSchedule
func (r *PaymentScheduleRepositoryImpl) scanPaymentSchedules(rows pgx.Rows) ([]*domain.PaymentSchedule, error) {
	var payments []*domain.PaymentSchedule
	for rows.Next() {
		payment := &domain.PaymentSchedule{}
		err := rows.Scan(
			&payment.ID,
			&payment.CreditID,
			&payment.PaymentNumber,
			&payment.DueDate,
			&payment.PaymentAmount,
			&payment.PrincipalAmount,
			&payment.InterestAmount,
			&payment.RemainingBalance,
			&payment.Status,
			&payment.PaidDate,
			&payment.PenaltyAmount,
			&payment.CreatedAt,
			&payment.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		payments = append(payments, payment)
	}

	return payments, nil
}

// GetByCreditID получает все платежи по кредиту
func (r *PaymentScheduleRepositoryImpl) GetByCreditID(ctx context.Context, creditID int) ([]*domain.PaymentSchedule, error) {
	query := `
		SELECT id, credit_id, payment_number, due_date, payment_amount, principal_amount, interest_amount, remaining_balance, status, paid_date, penalty_amount, created_at, updated_at
		FROM payment_schedules
		WHERE credit_id = $1
		ORDER BY payment_number ASC`

	rows, err := r.db.Query(ctx, query, creditID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanPaymentSchedules(rows)
}

// Update обновляет платеж
func (r *PaymentScheduleRepositoryImpl) Update(ctx context.Context, payment *domain.PaymentSchedule) error {
	query := `
		UPDATE payment_schedules
		SET payment_amount = $2, principal_amount = $3, interest_amount = $4, remaining_balance = $5, status = $6, paid_date = $7, penalty_amount = $8, updated_at = $9
		WHERE id = $1`

	payment.UpdatedAt = time.Now()

	result, err := r.db.Exec(ctx, query,
		payment.ID,
		payment.PaymentAmount,
		payment.PrincipalAmount,
		payment.InterestAmount,
		payment.RemainingBalance,
		payment.Status,
		payment.PaidDate,
		payment.PenaltyAmount,
		payment.UpdatedAt,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("payment not found")
	}

	return nil
}

// Delete удаляет платеж
func (r *PaymentScheduleRepositoryImpl) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM payment_schedules WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("payment not found")
	}

	return nil
}

// GetOverduePayments получает просроченные платежи
func (r *PaymentScheduleRepositoryImpl) GetOverduePayments(ctx context.Context) ([]*domain.PaymentSchedule, error) {
	query := `
		SELECT ps.id, ps.credit_id, ps.payment_number, ps.due_date, ps.payment_amount, ps.principal_amount, ps.interest_amount, ps.remaining_balance, ps.status, ps.paid_date, ps.penalty_amount, ps.created_at, ps.updated_at
		FROM payment_schedules ps
		INNER JOIN credits c ON ps.credit_id = c.id
		WHERE ps.due_date < NOW()
		  AND ps.status = 'pending'
		  AND c.status = 'active'
		ORDER BY ps.due_date ASC`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanPaymentSchedules(rows)
}

// GetUpcomingPayments получает предстоящие платежи в ближайшие дни
func (r *PaymentScheduleRepositoryImpl) GetUpcomingPayments(ctx context.Context, days int) ([]*domain.PaymentSchedule, error) {
	query := `
		SELECT ps.id, ps.credit_id, ps.payment_number, ps.due_date, ps.payment_amount, ps.principal_amount, ps.interest_amount, ps.remaining_balance, ps.status, ps.paid_date, ps.penalty_amount, ps.created_at, ps.updated_at
		FROM payment_schedules ps
		INNER JOIN credits c ON ps.credit_id = c.id
		WHERE ps.due_date BETWEEN NOW() AND NOW() + INTERVAL '%d days'
		  AND ps.status = 'pending'
		  AND c.status = 'active'
		ORDER BY ps.due_date ASC`

	rows, err := r.db.Query(ctx, query, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanPaymentSchedules(rows)
}

// MarkAsPaid отмечает платеж как оплаченный
func (r *PaymentScheduleRepositoryImpl) MarkAsPaid(ctx context.Context, id int, paidDate time.Time) error {
	query := `
		UPDATE payment_schedules
		SET status = 'paid', paid_date = $2, updated_at = $3
		WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id, paidDate, time.Now())
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("payment not found")
	}

	return nil
}

// AddPenalty добавляет штраф к платежу
func (r *PaymentScheduleRepositoryImpl) AddPenalty(ctx context.Context, id int, penaltyAmount float64) error {
	query := `
		UPDATE payment_schedules
		SET penalty_amount = penalty_amount + $2, status = 'overdue', updated_at = $3
		WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id, penaltyAmount, time.Now())
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("payment not found")
	}

	return nil
}
