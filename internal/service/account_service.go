package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/vterdunov/learn-bank-app/internal/domain"
	"github.com/vterdunov/learn-bank-app/internal/repository"
)

var (
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrAccountNotFound   = errors.New("account not found")
	ErrInvalidAmount     = errors.New("invalid amount")
	ErrAccountBlocked    = errors.New("account is blocked")
)

// accountService реализует интерфейс AccountService
type accountService struct {
	accountRepo     repository.AccountRepository
	transactionRepo repository.TransactionRepository
	accessControl   domain.AccessControlService
	logger          *slog.Logger
}

// NewAccountService создает новый экземпляр сервиса счетов
func NewAccountService(
	accountRepo repository.AccountRepository,
	transactionRepo repository.TransactionRepository,
	accessControl domain.AccessControlService,
	logger *slog.Logger,
) AccountService {
	return &accountService{
		accountRepo:     accountRepo,
		transactionRepo: transactionRepo,
		accessControl:   accessControl,
		logger:          logger,
	}
}

// CreateAccount создает новый банковский счет
func (s *accountService) CreateAccount(ctx context.Context, userID int, req CreateAccountRequest) (*domain.Account, error) {
	// Валидация валюты
	if req.Currency == "" {
		req.Currency = "RUB" // По умолчанию рубли
	}

	if req.Currency != "RUB" {
		s.logger.Warn("Unsupported currency", "currency", req.Currency, "user_id", userID)
		return nil, fmt.Errorf("unsupported currency: %s. Only RUB is supported", req.Currency)
	}

	// Генерация номера счета
	accountNumber := s.generateAccountNumber()

	// Создание счета
	account := &domain.Account{
		UserID:    userID,
		Number:    accountNumber,
		Balance:   0.0,
		Currency:  req.Currency,
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.accountRepo.Create(ctx, account); err != nil {
		s.logger.Error("Failed to create account", "user_id", userID, "error", err)
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	s.logger.Info("Account created successfully",
		"account_id", account.ID,
		"user_id", userID,
		"number", account.Number,
		"currency", account.Currency)

	return account, nil
}

// GetUserAccounts возвращает все счета пользователя
func (s *accountService) GetUserAccounts(ctx context.Context, userID int) ([]*domain.Account, error) {
	accounts, err := s.accountRepo.GetByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user accounts", "user_id", userID, "error", err)
		return nil, fmt.Errorf("failed to get user accounts: %w", err)
	}

	s.logger.Debug("Retrieved user accounts", "user_id", userID, "count", len(accounts))
	return accounts, nil
}

// DepositMoney пополняет баланс счета с проверкой прав доступа
func (s *accountService) DepositMoney(ctx context.Context, userID, accountID int, amount float64) error {
	// 1. Проверка прав доступа (domain logic)
	if err := s.accessControl.CanAccessAccount(ctx, userID, accountID); err != nil {
		s.logger.Warn("Access denied for deposit", "user_id", userID, "account_id", accountID)
		if domain.IsAccessDeniedError(err) {
			return &ServiceError{Code: http.StatusForbidden, Message: err.Error()}
		}
		return err
	}

	// 2. Валидация суммы
	if amount <= 0 {
		s.logger.Warn("Invalid deposit amount", "account_id", accountID, "amount", amount)
		return ErrInvalidAmount
	}

	// 3. Получение счета
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		s.logger.Error("Account not found for deposit", "account_id", accountID, "error", err)
		return ErrAccountNotFound
	}

	// 4. Проверка статуса счета
	if account.Status != "active" {
		s.logger.Warn("Account is not active", "account_id", accountID, "status", account.Status)
		return ErrAccountBlocked
	}

	// 5. Пополнение баланса
	newBalance := account.Balance + amount
	if err := s.accountRepo.UpdateBalance(ctx, accountID, newBalance); err != nil {
		s.logger.Error("Failed to update balance", "account_id", accountID, "error", err)
		return fmt.Errorf("failed to update balance: %w", err)
	}

	// 6. Создание записи о транзакции
	transaction := &domain.Transaction{
		FromAccount: nil, // Пополнение извне
		ToAccount:   &accountID,
		Amount:      amount,
		Type:        "deposit",
		Status:      "completed",
		Description: "Account deposit",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.transactionRepo.Create(ctx, transaction); err != nil {
		s.logger.Error("Failed to create transaction record", "account_id", accountID, "amount", amount, "error", err)
		// Не возвращаем ошибку, так как операция пополнения уже выполнена
	}

	s.logger.Info("Account deposit successful",
		"user_id", userID,
		"account_id", accountID,
		"amount", amount,
		"new_balance", newBalance)

	return nil
}

// WithdrawMoney списывает средства со счета с проверкой прав доступа
func (s *accountService) WithdrawMoney(ctx context.Context, userID, accountID int, amount float64) error {
	// 1. Проверка прав доступа (domain logic)
	if err := s.accessControl.CanAccessAccount(ctx, userID, accountID); err != nil {
		s.logger.Warn("Access denied for withdrawal", "user_id", userID, "account_id", accountID)
		if domain.IsAccessDeniedError(err) {
			return &ServiceError{Code: http.StatusForbidden, Message: err.Error()}
		}
		return err
	}

	// 2. Валидация суммы
	if amount <= 0 {
		s.logger.Warn("Invalid withdrawal amount", "account_id", accountID, "amount", amount)
		return ErrInvalidAmount
	}

	// 3. Получение счета
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		s.logger.Error("Account not found for withdrawal", "account_id", accountID, "error", err)
		return ErrAccountNotFound
	}

	// 4. Проверка статуса счета
	if account.Status != "active" {
		s.logger.Warn("Account is not active", "account_id", accountID, "status", account.Status)
		return ErrAccountBlocked
	}

	// 5. Проверка достаточности средств
	if account.Balance < amount {
		s.logger.Warn("Insufficient funds",
			"account_id", accountID,
			"balance", account.Balance,
			"requested", amount)
		return ErrInsufficientFunds
	}

	// 6. Списание средств
	newBalance := account.Balance - amount
	if err := s.accountRepo.UpdateBalance(ctx, accountID, newBalance); err != nil {
		s.logger.Error("Failed to update balance", "account_id", accountID, "error", err)
		return fmt.Errorf("failed to update balance: %w", err)
	}

	// 7. Создание записи о транзакции
	transaction := &domain.Transaction{
		FromAccount: &accountID,
		ToAccount:   nil, // Списание во внешнюю систему
		Amount:      amount,
		Type:        "withdrawal",
		Status:      "completed",
		Description: "Account withdrawal",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.transactionRepo.Create(ctx, transaction); err != nil {
		s.logger.Error("Failed to create transaction record", "account_id", accountID, "amount", amount, "error", err)
		// Не возвращаем ошибку, так как операция списания уже выполнена
	}

	s.logger.Info("Account withdrawal successful",
		"user_id", userID,
		"account_id", accountID,
		"amount", amount,
		"new_balance", newBalance)

	return nil
}

// TransferMoney выполняет перевод между счетами с проверкой прав доступа
func (s *accountService) TransferMoney(ctx context.Context, userID, fromAccountID, toAccountID int, amount float64) error {
	// 1. Проверка прав доступа к исходящему счету (domain logic)
	if err := s.accessControl.CanAccessAccount(ctx, userID, fromAccountID); err != nil {
		s.logger.Warn("Access denied for transfer from account", "user_id", userID, "from_account_id", fromAccountID)
		if domain.IsAccessDeniedError(err) {
			return &ServiceError{Code: http.StatusForbidden, Message: err.Error()}
		}
		return err
	}

	// 2. Валидация суммы
	if amount <= 0 {
		s.logger.Warn("Invalid transfer amount", "from_account_id", fromAccountID, "to_account_id", toAccountID, "amount", amount)
		return ErrInvalidAmount
	}

	// 3. Валидация что счета разные
	if fromAccountID == toAccountID {
		s.logger.Warn("Cannot transfer to the same account", "account_id", fromAccountID)
		return fmt.Errorf("cannot transfer to the same account")
	}

	// 4. Выполнение перевода через репозиторий (транзакция)
	if err := s.accountRepo.Transfer(ctx, fromAccountID, toAccountID, amount); err != nil {
		s.logger.Error("Transfer failed",
			"from_account_id", fromAccountID,
			"to_account_id", toAccountID,
			"amount", amount,
			"error", err)
		return fmt.Errorf("transfer failed: %w", err)
	}

	// 5. Создание записи о транзакции
	transaction := &domain.Transaction{
		FromAccount: &fromAccountID,
		ToAccount:   &toAccountID,
		Amount:      amount,
		Type:        "transfer",
		Status:      "completed",
		Description: "Account transfer",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.transactionRepo.Create(ctx, transaction); err != nil {
		s.logger.Error("Failed to create transaction record", "from_account_id", fromAccountID, "to_account_id", toAccountID, "amount", amount, "error", err)
		// Не возвращаем ошибку, так как операция перевода уже выполнена
	}

	s.logger.Info("Transfer successful",
		"user_id", userID,
		"from_account_id", fromAccountID,
		"to_account_id", toAccountID,
		"amount", amount)

	return nil
}

// generateAccountNumber генерирует уникальный номер счета
func (s *accountService) generateAccountNumber() string {
	// Генерируем 16-значный номер счета
	bytes := make([]byte, 8)
	_, err := rand.Read(bytes)
	if err != nil {
		// Fallback на timestamp если rand не работает
		return fmt.Sprintf("40817810%d", time.Now().Unix())
	}

	var number uint64
	for i, b := range bytes {
		number |= uint64(b) << (8 * i)
	}

	// Формат: 40817810 + 12 цифр (банк.счет РФ)
	return fmt.Sprintf("40817810%012d", number%1000000000000)
}

// ServiceError кастомная ошибка сервиса с HTTP статус кодом
type ServiceError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *ServiceError) Error() string {
	return e.Message
}
