package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/vterdunov/learn-bank-app/internal/config"
	"github.com/vterdunov/learn-bank-app/internal/domain"
	"github.com/vterdunov/learn-bank-app/internal/repository"
)

// SchedulerServiceImpl реализация SchedulerService
type SchedulerServiceImpl struct {
	creditRepo      repository.CreditRepository
	paymentRepo     repository.PaymentScheduleRepository
	accountRepo     repository.AccountRepository
	transactionRepo repository.TransactionRepository
	userRepo        repository.UserRepository
	emailService    EmailService
	logger          *slog.Logger
	ticker          *time.Ticker
	stopChan        chan struct{}
	interval        time.Duration
	penaltyRate     float64
}

// NewSchedulerService создает новый экземпляр SchedulerService
func NewSchedulerService(
	cfg *config.Config,
	creditRepo repository.CreditRepository,
	paymentRepo repository.PaymentScheduleRepository,
	accountRepo repository.AccountRepository,
	transactionRepo repository.TransactionRepository,
	userRepo repository.UserRepository,
	emailService EmailService,
	logger *slog.Logger,
) SchedulerService {
	return &SchedulerServiceImpl{
		creditRepo:      creditRepo,
		paymentRepo:     paymentRepo,
		accountRepo:     accountRepo,
		transactionRepo: transactionRepo,
		userRepo:        userRepo,
		emailService:    emailService,
		logger:          logger,
		stopChan:        make(chan struct{}),
		interval:        cfg.Scheduler.Interval,
		penaltyRate:     cfg.Scheduler.PenaltyRate,
	}
}

// Start запускает шедулер
func (s *SchedulerServiceImpl) Start(ctx context.Context) error {
	s.logger.Info("Starting payment scheduler", "interval", s.interval)

	s.ticker = time.NewTicker(s.interval)

	// Запускаем первую обработку сразу
	go func() {
		if err := s.ProcessOverduePayments(ctx); err != nil {
			s.logger.Error("Failed to process overdue payments on startup", "error", err)
		}
	}()

	// Запускаем периодическую обработку
	go func() {
		for {
			select {
			case <-s.ticker.C:
				if err := s.ProcessOverduePayments(ctx); err != nil {
					s.logger.Error("Failed to process overdue payments", "error", err)
				}
			case <-s.stopChan:
				s.logger.Info("Scheduler stopped")
				return
			case <-ctx.Done():
				s.logger.Info("Scheduler stopped due to context cancellation")
				return
			}
		}
	}()

	return nil
}

// Stop останавливает шедулер
func (s *SchedulerServiceImpl) Stop() {
	if s.ticker != nil {
		s.ticker.Stop()
	}
	close(s.stopChan)
}

// ProcessOverduePayments обрабатывает просроченные платежи
func (s *SchedulerServiceImpl) ProcessOverduePayments(ctx context.Context) error {
	s.logger.Info("Starting overdue payments processing")

	// Получаем все просроченные платежи
	overduePayments, err := s.paymentRepo.GetOverduePayments(ctx)
	if err != nil {
		return fmt.Errorf("failed to get overdue payments: %w", err)
	}

	s.logger.Info("Found overdue payments", "count", len(overduePayments))

	var processedCount, failedCount int

	for _, payment := range overduePayments {
		if err := s.processOverduePayment(ctx, payment); err != nil {
			s.logger.Error("Failed to process overdue payment",
				"payment_id", payment.ID,
				"credit_id", payment.CreditID,
				"error", err)
			failedCount++
		} else {
			processedCount++
		}
	}

	s.logger.Info("Overdue payments processing completed",
		"processed", processedCount,
		"failed", failedCount,
		"total", len(overduePayments))

	return nil
}

// processOverduePayment обрабатывает один просроченный платеж
func (s *SchedulerServiceImpl) processOverduePayment(ctx context.Context, payment *domain.PaymentSchedule) error {
	// Получаем информацию о кредите
	credit, err := s.creditRepo.GetByID(ctx, payment.CreditID)
	if err != nil {
		return fmt.Errorf("failed to get credit: %w", err)
	}

	// Получаем информацию о счете
	account, err := s.accountRepo.GetByID(ctx, credit.AccountID)
	if err != nil {
		return fmt.Errorf("failed to get account: %w", err)
	}

	// Рассчитываем штраф
	penaltyAmount := payment.PaymentAmount * s.penaltyRate / 100
	totalAmount := payment.PaymentAmount + penaltyAmount

	s.logger.Info("Processing overdue payment",
		"payment_id", payment.ID,
		"original_amount", payment.PaymentAmount,
		"penalty", penaltyAmount,
		"penalty_rate", s.penaltyRate,
		"total_amount", totalAmount,
		"account_balance", account.Balance)

	// Проверяем, достаточно ли средств для списания
	if account.Balance >= totalAmount {
		// Списываем средства со счета
		if err := s.processPaymentDeduction(ctx, account, payment, penaltyAmount, totalAmount); err != nil {
			return fmt.Errorf("failed to process payment deduction: %w", err)
		}

		s.logger.Info("Successfully processed overdue payment with penalty",
			"payment_id", payment.ID,
			"amount_deducted", totalAmount)
	} else {
		// Недостаточно средств - увеличиваем штраф
		if err := s.increasePenalty(ctx, payment, penaltyAmount); err != nil {
			return fmt.Errorf("failed to increase penalty: %w", err)
		}

		s.logger.Warn("Insufficient funds for overdue payment, penalty increased",
			"payment_id", payment.ID,
			"required", totalAmount,
			"available", account.Balance)
	}

	// Отправляем уведомление пользователю
	if err := s.sendOverdueNotification(ctx, credit.UserID, payment); err != nil {
		s.logger.Error("Failed to send overdue notification",
			"payment_id", payment.ID,
			"user_id", credit.UserID,
			"error", err)
		// Не возвращаем ошибку, чтобы не прерывать обработку платежа
	}

	return nil
}

// processPaymentDeduction списывает платеж со счета
func (s *SchedulerServiceImpl) processPaymentDeduction(
	ctx context.Context,
	account *domain.Account,
	payment *domain.PaymentSchedule,
	penaltyAmount, totalAmount float64,
) error {
	// Списываем средства со счета
	account.Balance -= totalAmount
	if err := s.accountRepo.UpdateBalance(ctx, account.ID, account.Balance); err != nil {
		return fmt.Errorf("failed to update account balance: %w", err)
	}

	// Обновляем статус платежа
	payment.Status = domain.PaymentStatusPaid
	now := time.Now()
	payment.PaidDate = &now
	payment.PenaltyAmount += penaltyAmount
	payment.UpdatedAt = now

	if err := s.paymentRepo.Update(ctx, payment); err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	// Создаем транзакцию
	transaction := &domain.Transaction{
		FromAccount: &account.ID,
		ToAccount:   nil, // Списание со счета
		Type:        "credit_payment",
		Amount:      totalAmount,
		Description: fmt.Sprintf("Credit payment #%d with penalty %.2f", payment.PaymentNumber, penaltyAmount),
		Status:      "completed",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.transactionRepo.Create(ctx, transaction); err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	return nil
}

// increasePenalty увеличивает штраф за просрочку
func (s *SchedulerServiceImpl) increasePenalty(ctx context.Context, payment *domain.PaymentSchedule, additionalPenalty float64) error {
	payment.PenaltyAmount += additionalPenalty
	payment.UpdatedAt = time.Now()

	if err := s.paymentRepo.Update(ctx, payment); err != nil {
		return fmt.Errorf("failed to update payment penalty: %w", err)
	}

	return nil
}

// sendOverdueNotification отправляет уведомление о просроченном платеже
func (s *SchedulerServiceImpl) sendOverdueNotification(ctx context.Context, userID int, payment *domain.PaymentSchedule) error {
	// Получаем информацию о пользователе
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Отправляем уведомление
	if err := s.emailService.SendOverdueNotification(user.Email, payment); err != nil {
		return fmt.Errorf("failed to send email notification: %w", err)
	}

	return nil
}
