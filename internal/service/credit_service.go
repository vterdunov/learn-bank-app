package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/vterdunov/learn-bank-app/internal/domain"
	"github.com/vterdunov/learn-bank-app/internal/repository"
)

var (
	ErrCreditNotFound      = errors.New("credit not found")
	ErrInvalidCreditAmount = errors.New("invalid credit amount")
	ErrInvalidCreditTerm   = errors.New("invalid credit term")
	ErrCreditAlreadyExists = errors.New("active credit already exists for this account")
	ErrInsufficientIncome  = errors.New("insufficient income for credit")
)

// creditService реализует интерфейс CreditService
type creditService struct {
	creditRepo          repository.CreditRepository
	paymentScheduleRepo repository.PaymentScheduleRepository
	accountRepo         repository.AccountRepository
	transactionRepo     repository.TransactionRepository
	cbrService          CBRService
	logger              *slog.Logger
}

// NewCreditService создает новый экземпляр сервиса кредитования
func NewCreditService(
	creditRepo repository.CreditRepository,
	paymentScheduleRepo repository.PaymentScheduleRepository,
	accountRepo repository.AccountRepository,
	transactionRepo repository.TransactionRepository,
	cbrService CBRService,
	logger *slog.Logger,
) CreditService {
	return &creditService{
		creditRepo:          creditRepo,
		paymentScheduleRepo: paymentScheduleRepo,
		accountRepo:         accountRepo,
		transactionRepo:     transactionRepo,
		cbrService:          cbrService,
		logger:              logger,
	}
}

// CalculateAnnuityPayment рассчитывает аннуитетный платеж (делегирует в доменную логику)
func (s *creditService) CalculateAnnuityPayment(principal, rate float64, months int) float64 {
	return domain.CalculateAnnuityPayment(principal, rate, months)
}

// CreateCredit создает новый кредит с графиком платежей
func (s *creditService) CreateCredit(ctx context.Context, req domain.CreateCreditRequest) (*domain.Credit, error) {
	// Валидация входных данных через доменную логику
	if err := req.Validate(); err != nil {
		s.logger.Warn("Invalid credit request", "error", err)
		return nil, err
	}

	// Проверяем существование счета
	account, err := s.accountRepo.GetByID(ctx, req.AccountID)
	if err != nil {
		s.logger.Error("Account not found for credit", "account_id", req.AccountID, "error", err)
		return nil, ErrAccountNotFound
	}

	// Проверяем статус счета
	if account.Status != "active" {
		s.logger.Warn("Cannot create credit for inactive account", "account_id", req.AccountID, "status", account.Status)
		return nil, ErrAccountBlocked
	}

	// Получаем ключевую ставку ЦБ РФ
	baseRate, err := s.cbrService.GetKeyRate(ctx)
	if err != nil {
		s.logger.Warn("Failed to get CBR key rate, using fallback", "error", err)
		baseRate = 16.0 // Fallback ставка
	}

	// Добавляем маржу банка (например, +5%)
	creditRate := baseRate + 5.0

	s.logger.Info("Credit rate calculated",
		"base_rate", baseRate,
		"credit_rate", creditRate,
		"account_id", req.AccountID)

	// Рассчитываем аннуитетный платеж
	monthlyPayment := s.CalculateAnnuityPayment(req.Amount, creditRate, req.TermMonths)

	// Создаем кредит
	credit := &domain.Credit{
		AccountID:      req.AccountID,
		Amount:         req.Amount,
		InterestRate:   creditRate,
		TermMonths:     req.TermMonths,
		MonthlyPayment: monthlyPayment,
		RemainingDebt:  req.Amount,
		Status:         "active",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.creditRepo.Create(ctx, credit); err != nil {
		s.logger.Error("Failed to create credit", "account_id", req.AccountID, "error", err)
		return nil, fmt.Errorf("failed to create credit: %w", err)
	}

	// Создаем график платежей
	if err := s.createPaymentSchedule(ctx, credit); err != nil {
		s.logger.Error("Failed to create payment schedule", "credit_id", credit.ID, "error", err)
		// Не возвращаем ошибку, так как кредит уже создан
	}

	// Зачисляем кредитные средства на счет
	newBalance := account.Balance + req.Amount
	if err := s.accountRepo.UpdateBalance(ctx, req.AccountID, newBalance); err != nil {
		s.logger.Error("Failed to update balance after credit", "account_id", req.AccountID, "error", err)
		return nil, fmt.Errorf("failed to update account balance: %w", err)
	}

	// Создаем транзакцию о выдаче кредита
	transaction := &domain.Transaction{
		FromAccount: nil, // Кредит от банка
		ToAccount:   &req.AccountID,
		Amount:      req.Amount,
		Type:        "credit",
		Status:      "completed",
		Description: fmt.Sprintf("Credit disbursement (Credit ID: %d)", credit.ID),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.transactionRepo.Create(ctx, transaction); err != nil {
		s.logger.Error("Failed to create credit transaction", "credit_id", credit.ID, "error", err)
		// Не возвращаем ошибку, так как кредит уже выдан
	}

	s.logger.Info("Credit created successfully",
		"credit_id", credit.ID,
		"account_id", req.AccountID,
		"amount", req.Amount,
		"rate", creditRate,
		"monthly_payment", monthlyPayment,
		"new_balance", newBalance)

	return credit, nil
}

// GetCreditSchedule возвращает график платежей по кредиту
func (s *creditService) GetCreditSchedule(ctx context.Context, creditID int) ([]*domain.PaymentSchedule, error) {
	// Проверяем существование кредита
	credit, err := s.creditRepo.GetByID(ctx, creditID)
	if err != nil {
		s.logger.Error("Credit not found for schedule", "credit_id", creditID, "error", err)
		return nil, ErrCreditNotFound
	}

	// Получаем график платежей
	schedule, err := s.paymentScheduleRepo.GetByCreditID(ctx, creditID)
	if err != nil {
		s.logger.Error("Failed to get payment schedule", "credit_id", creditID, "error", err)
		return nil, fmt.Errorf("failed to get payment schedule: %w", err)
	}

	s.logger.Debug("Retrieved payment schedule",
		"credit_id", creditID,
		"payments_count", len(schedule),
		"credit_status", credit.Status)

	return schedule, nil
}

// ProcessOverduePayments обрабатывает просроченные платежи
func (s *creditService) ProcessOverduePayments(ctx context.Context) error {
	s.logger.Info("Starting overdue payments processing")

	// Получаем все просроченные платежи
	overduePayments, err := s.paymentScheduleRepo.GetOverduePayments(ctx)
	if err != nil {
		s.logger.Error("Failed to get overdue payments", "error", err)
		return fmt.Errorf("failed to get overdue payments: %w", err)
	}

	processed := 0
	failed := 0

	for _, payment := range overduePayments {
		if err := s.processOverduePayment(ctx, payment); err != nil {
			s.logger.Error("Failed to process overdue payment",
				"payment_id", payment.ID,
				"credit_id", payment.CreditID,
				"error", err)
			failed++
		} else {
			processed++
		}
	}

	s.logger.Info("Overdue payments processing completed",
		"processed", processed,
		"failed", failed,
		"total", len(overduePayments))

	return nil
}

// createPaymentSchedule создает график платежей для кредита
func (s *creditService) createPaymentSchedule(ctx context.Context, credit *domain.Credit) error {
	remainingPrincipal := credit.Amount

	for month := 1; month <= credit.TermMonths; month++ {
		// Используем доменную логику для расчета разбивки платежа
		principalPayment, interestPayment := credit.CalculatePaymentBreakdown(month, remainingPrincipal)

		// Дата платежа
		paymentDate := credit.CreatedAt.AddDate(0, month, 0)

		payment := &domain.PaymentSchedule{
			CreditID:         credit.ID,
			PaymentNumber:    month,
			DueDate:          paymentDate,
			PaymentAmount:    credit.MonthlyPayment,
			PrincipalAmount:  principalPayment,
			InterestAmount:   interestPayment,
			RemainingBalance: math.Round((remainingPrincipal-principalPayment)*100) / 100,
			Status:           "pending",
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}

		if err := s.paymentScheduleRepo.Create(ctx, payment); err != nil {
			return fmt.Errorf("failed to create payment %d: %w", month, err)
		}

		remainingPrincipal -= principalPayment
	}

	s.logger.Info("Payment schedule created successfully",
		"credit_id", credit.ID,
		"payments_count", credit.TermMonths)

	return nil
}

// processOverduePayment обрабатывает один просроченный платеж
func (s *creditService) processOverduePayment(ctx context.Context, payment *domain.PaymentSchedule) error {
	// Получаем кредит
	credit, err := s.creditRepo.GetByID(ctx, payment.CreditID)
	if err != nil {
		return fmt.Errorf("failed to get credit: %w", err)
	}

	// Получаем счет
	account, err := s.accountRepo.GetByID(ctx, credit.AccountID)
	if err != nil {
		return fmt.Errorf("failed to get account: %w", err)
	}

	// Рассчитываем штраф (10% от суммы платежа)
	penalty := payment.PaymentAmount * 0.10
	totalAmount := payment.PaymentAmount + penalty

	// Проверяем достаточность средств
	if account.Balance < totalAmount {
		// Недостаточно средств - только начисляем штраф
		payment.PenaltyAmount += penalty
		payment.Status = "overdue_with_penalty"
		payment.UpdatedAt = time.Now()

		if err := s.paymentScheduleRepo.Update(ctx, payment); err != nil {
			return fmt.Errorf("failed to update payment with penalty: %w", err)
		}

		s.logger.Warn("Insufficient funds for overdue payment, penalty added",
			"payment_id", payment.ID,
			"credit_id", payment.CreditID,
			"account_id", credit.AccountID,
			"balance", account.Balance,
			"required", totalAmount,
			"penalty", penalty)

		return nil
	}

	// Списываем средства
	newBalance := account.Balance - totalAmount
	if err := s.accountRepo.UpdateBalance(ctx, credit.AccountID, newBalance); err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}

	// Обновляем платеж
	payment.PenaltyAmount += penalty
	payment.Status = "paid"
	now := time.Now()
	payment.PaidDate = &now
	payment.UpdatedAt = time.Now()

	if err := s.paymentScheduleRepo.Update(ctx, payment); err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	// Обновляем остаток по кредиту через доменную логику
	credit.UpdateRemainingDebt(payment.PrincipalAmount)

	if err := s.creditRepo.Update(ctx, credit); err != nil {
		return fmt.Errorf("failed to update credit: %w", err)
	}

	// Создаем транзакцию
	transaction := &domain.Transaction{
		FromAccount: &credit.AccountID,
		ToAccount:   nil, // Платеж банку
		Amount:      totalAmount,
		Type:        "credit_payment",
		Status:      "completed",
		Description: fmt.Sprintf("Overdue credit payment with penalty (Payment ID: %d)", payment.ID),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.transactionRepo.Create(ctx, transaction); err != nil {
		s.logger.Error("Failed to create overdue payment transaction",
			"payment_id", payment.ID,
			"error", err)
	}

	s.logger.Info("Overdue payment processed successfully",
		"payment_id", payment.ID,
		"credit_id", payment.CreditID,
		"account_id", credit.AccountID,
		"amount", payment.PaymentAmount,
		"penalty", penalty,
		"new_balance", newBalance)

	return nil
}
