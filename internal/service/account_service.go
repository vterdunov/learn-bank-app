package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
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
	logger          *slog.Logger
}

// NewAccountService создает новый экземпляр сервиса счетов
func NewAccountService(
	accountRepo repository.AccountRepository,
	transactionRepo repository.TransactionRepository,
	logger *slog.Logger,
) AccountService {
	return &accountService{
		accountRepo:     accountRepo,
		transactionRepo: transactionRepo,
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

// DepositMoney пополняет баланс счета
func (s *accountService) DepositMoney(ctx context.Context, accountID int, amount float64) error {
	// Валидация суммы
	if amount <= 0 {
		s.logger.Warn("Invalid deposit amount", "account_id", accountID, "amount", amount)
		return ErrInvalidAmount
	}

	// Получение счета
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		s.logger.Error("Account not found for deposit", "account_id", accountID, "error", err)
		return ErrAccountNotFound
	}

	// Проверка статуса счета
	if account.Status != "active" {
		s.logger.Warn("Account is not active", "account_id", accountID, "status", account.Status)
		return ErrAccountBlocked
	}

	// Пополнение баланса
	newBalance := account.Balance + amount
	if err := s.accountRepo.UpdateBalance(ctx, accountID, newBalance); err != nil {
		s.logger.Error("Failed to update balance", "account_id", accountID, "error", err)
		return fmt.Errorf("failed to update balance: %w", err)
	}

	// Создание записи о транзакции
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
		"account_id", accountID,
		"amount", amount,
		"new_balance", newBalance)

	return nil
}

// WithdrawMoney списывает средства со счета
func (s *accountService) WithdrawMoney(ctx context.Context, accountID int, amount float64) error {
	// Валидация суммы
	if amount <= 0 {
		s.logger.Warn("Invalid withdrawal amount", "account_id", accountID, "amount", amount)
		return ErrInvalidAmount
	}

	// Получение счета
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		s.logger.Error("Account not found for withdrawal", "account_id", accountID, "error", err)
		return ErrAccountNotFound
	}

	// Проверка статуса счета
	if account.Status != "active" {
		s.logger.Warn("Account is not active", "account_id", accountID, "status", account.Status)
		return ErrAccountBlocked
	}

	// Проверка достаточности средств
	if account.Balance < amount {
		s.logger.Warn("Insufficient funds",
			"account_id", accountID,
			"balance", account.Balance,
			"requested", amount)
		return ErrInsufficientFunds
	}

	// Списание средств
	newBalance := account.Balance - amount
	if err := s.accountRepo.UpdateBalance(ctx, accountID, newBalance); err != nil {
		s.logger.Error("Failed to update balance", "account_id", accountID, "error", err)
		return fmt.Errorf("failed to update balance: %w", err)
	}

	// Создание записи о транзакции
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
		"account_id", accountID,
		"amount", amount,
		"new_balance", newBalance)

	return nil
}

// TransferMoney переводит средства между счетами
func (s *accountService) TransferMoney(ctx context.Context, fromAccountID, toAccountID int, amount float64) error {
	// Валидация суммы
	if amount <= 0 {
		s.logger.Warn("Invalid transfer amount",
			"from_account", fromAccountID,
			"to_account", toAccountID,
			"amount", amount)
		return ErrInvalidAmount
	}

	// Проверка, что счета разные
	if fromAccountID == toAccountID {
		s.logger.Warn("Cannot transfer to the same account", "account_id", fromAccountID)
		return errors.New("cannot transfer to the same account")
	}

	// Используем транзакцию базы данных для атомарности
	if err := s.accountRepo.Transfer(ctx, fromAccountID, toAccountID, amount); err != nil {
		s.logger.Error("Transfer failed",
			"from_account", fromAccountID,
			"to_account", toAccountID,
			"amount", amount,
			"error", err)
		return fmt.Errorf("transfer failed: %w", err)
	}

	// Создание записи о транзакции
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
		s.logger.Error("Failed to create transaction record",
			"from_account", fromAccountID,
			"to_account", toAccountID,
			"amount", amount,
			"error", err)
		// Не возвращаем ошибку, так как перевод уже выполнен
	}

	s.logger.Info("Transfer completed successfully",
		"from_account", fromAccountID,
		"to_account", toAccountID,
		"amount", amount)

	return nil
}

// generateAccountNumber генерирует уникальный номер счета
func (s *accountService) generateAccountNumber() string {
	// Генерация номера счета в формате: 408 17 810 XXXXXXXXXX X
	// 408 - счета физических лиц в рублях
	// 17 - код валюты (RUB)
	// 810 - код страны (Россия)
	// XXXXXXXXXX - случайные цифры
	// X - контрольная цифра

	bankCode := "40817810" // Стандартный код для счетов физлиц в рублях

	// Генерируем 10 случайных цифр с использованием crypto/rand
	randomBytes := make([]byte, 10)
	if _, err := rand.Read(randomBytes); err != nil {
		// Fallback на timestamp если crypto/rand недоступен
		timestamp := time.Now().UnixNano()
		for i := range randomBytes {
			randomBytes[i] = byte(timestamp % 10)
			timestamp /= 10
		}
	}

	accountDigits := make([]byte, 10)
	for i := range accountDigits {
		accountDigits[i] = byte('0' + (randomBytes[i] % 10))
	}

	// Простая контрольная сумма (последняя цифра)
	sum := 0
	for _, digit := range accountDigits {
		sum += int(digit - '0')
	}
	checkDigit := sum % 10

	return bankCode + string(accountDigits) + strconv.Itoa(checkDigit)
}
