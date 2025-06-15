package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/vterdunov/learn-bank-app/internal/models"
	"github.com/vterdunov/learn-bank-app/internal/repository"
	"github.com/vterdunov/learn-bank-app/internal/utils"
)

var (
	ErrCardNotFound    = errors.New("card not found")
	ErrCardBlocked     = errors.New("card is blocked")
	ErrCardExpired     = errors.New("card is expired")
	ErrInvalidCardData = errors.New("invalid card data")
)

// cardService реализует интерфейс CardService
type cardService struct {
	cardRepo        repository.CardRepository
	accountRepo     repository.AccountRepository
	transactionRepo repository.TransactionRepository
	logger          *slog.Logger
	encryptionKey   []byte
}

// NewCardService создает новый экземпляр сервиса карт
func NewCardService(
	cardRepo repository.CardRepository,
	accountRepo repository.AccountRepository,
	transactionRepo repository.TransactionRepository,
	logger *slog.Logger,
) CardService {
	// Генерируем ключ шифрования (в продакшене должен браться из конфигурации)
	key, err := utils.GenerateRandomKey(32)
	if err != nil {
		logger.Error("Failed to generate encryption key", "error", err)
		// Fallback на фиксированный ключ для тестирования
		key = []byte("test-encryption-key-32-bytes-!!")
	}

	return &cardService{
		cardRepo:        cardRepo,
		accountRepo:     accountRepo,
		transactionRepo: transactionRepo,
		logger:          logger,
		encryptionKey:   key,
	}
}

// CreateCard создает новую банковскую карту
func (s *cardService) CreateCard(ctx context.Context, accountID int) (*models.Card, error) {
	// Проверяем существование счета
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		s.logger.Error("Account not found for card creation", "account_id", accountID, "error", err)
		return nil, ErrAccountNotFound
	}

	// Проверяем статус счета
	if account.Status != "active" {
		s.logger.Warn("Cannot create card for inactive account", "account_id", accountID, "status", account.Status)
		return nil, ErrAccountBlocked
	}

	// Генерируем номер карты по алгоритму Луна
	cardNumber, err := utils.GenerateCardNumber()
	if err != nil {
		s.logger.Error("Failed to generate card number", "account_id", accountID, "error", err)
		return nil, fmt.Errorf("failed to generate card number: %w", err)
	}

	// Генерируем CVV
	cvv, err := utils.GenerateCVV()
	if err != nil {
		s.logger.Error("Failed to generate CVV", "account_id", accountID, "error", err)
		return nil, fmt.Errorf("failed to generate CVV: %w", err)
	}

	// Хешируем CVV
	cvvHash, err := utils.HashCVV(cvv)
	if err != nil {
		s.logger.Error("Failed to hash CVV", "account_id", accountID, "error", err)
		return nil, fmt.Errorf("failed to hash CVV: %w", err)
	}

	// Генерируем срок действия карты
	expiryStr := utils.GenerateExpiryDate()
	expiryDate, err := time.Parse("01/06", expiryStr)
	if err != nil {
		s.logger.Error("Failed to parse generated expiry date", "expiry", expiryStr, "error", err)
		// Fallback - 4 года от текущей даты
		expiryDate = time.Now().AddDate(4, 0, 0)
	}

	// Шифруем данные карты
	encryptedNumber, encryptedExpiry, err := utils.EncryptCardData(cardNumber, expiryStr, s.encryptionKey)
	if err != nil {
		s.logger.Error("Failed to encrypt card data", "account_id", accountID, "error", err)
		return nil, fmt.Errorf("failed to encrypt card data: %w", err)
	}

	// Объединяем зашифрованные данные в одну строку для хранения
	encryptedDataStr := fmt.Sprintf("%x:%s", encryptedNumber.Data, encryptedNumber.HMAC)
	hmacStr := fmt.Sprintf("%x:%s", encryptedExpiry.Data, encryptedExpiry.HMAC)

	// Создаем карту
	card := &models.Card{
		AccountID:     accountID,
		EncryptedData: encryptedDataStr,
		HMAC:          hmacStr,
		CVVHash:       cvvHash,
		ExpiryDate:    expiryDate,
		Status:        "active",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.cardRepo.Create(ctx, card); err != nil {
		s.logger.Error("Failed to create card", "account_id", accountID, "error", err)
		return nil, fmt.Errorf("failed to create card: %w", err)
	}

	s.logger.Info("Card created successfully",
		"card_id", card.ID,
		"account_id", accountID,
		"expiry_date", card.ExpiryDate,
		"card_type", utils.GetCardType(cardNumber))

	return card, nil
}

// GetAccountCards возвращает все карты счета
func (s *cardService) GetAccountCards(ctx context.Context, accountID int) ([]*models.Card, error) {
	cards, err := s.cardRepo.GetByAccountID(ctx, accountID)
	if err != nil {
		s.logger.Error("Failed to get account cards", "account_id", accountID, "error", err)
		return nil, fmt.Errorf("failed to get account cards: %w", err)
	}

	s.logger.Debug("Retrieved account cards", "account_id", accountID, "count", len(cards))
	return cards, nil
}

// DecryptCardData расшифровывает данные карты
func (s *cardService) DecryptCardData(ctx context.Context, card *models.Card) (*CardData, error) {
	// Парсим зашифрованные данные номера карты
	numberParts := strings.Split(card.EncryptedData, ":")
	if len(numberParts) != 2 {
		s.logger.Error("Invalid encrypted card number format", "card_id", card.ID)
		return nil, ErrInvalidCardData
	}

	// Парсим зашифрованные данные даты истечения
	expiryParts := strings.Split(card.HMAC, ":")
	if len(expiryParts) != 2 {
		s.logger.Error("Invalid encrypted expiry date format", "card_id", card.ID)
		return nil, ErrInvalidCardData
	}

	// Воссоздаем EncryptedData структуры
	var encryptedNumber, encryptedExpiry utils.EncryptedData

	// Декодируем hex данные
	if _, err := fmt.Sscanf(numberParts[0], "%x", &encryptedNumber.Data); err != nil {
		s.logger.Error("Failed to decode card number data", "card_id", card.ID, "error", err)
		return nil, ErrInvalidCardData
	}
	encryptedNumber.HMAC = numberParts[1]

	if _, err := fmt.Sscanf(expiryParts[0], "%x", &encryptedExpiry.Data); err != nil {
		s.logger.Error("Failed to decode expiry date data", "card_id", card.ID, "error", err)
		return nil, ErrInvalidCardData
	}
	encryptedExpiry.HMAC = expiryParts[1]

	// Расшифровываем данные
	cardNumber, expiryStr, err := utils.DecryptCardData(&encryptedNumber, &encryptedExpiry, s.encryptionKey)
	if err != nil {
		s.logger.Error("Failed to decrypt card data", "card_id", card.ID, "error", err)
		return nil, fmt.Errorf("failed to decrypt card data: %w", err)
	}

	// Парсим дату истечения
	expiryDate, err := time.Parse("01/06", expiryStr)
	if err != nil {
		s.logger.Error("Failed to parse expiry date", "card_id", card.ID, "expiry", expiryStr, "error", err)
		return nil, fmt.Errorf("failed to parse expiry date: %w", err)
	}

	cardData := &CardData{
		Number:     cardNumber,
		ExpiryDate: expiryDate,
	}

	s.logger.Debug("Card data decrypted successfully", "card_id", card.ID)
	return cardData, nil
}

// ProcessPayment обрабатывает платеж с карты
func (s *cardService) ProcessPayment(ctx context.Context, cardID int, amount float64) error {
	// Валидация суммы
	if amount <= 0 {
		s.logger.Warn("Invalid payment amount", "card_id", cardID, "amount", amount)
		return ErrInvalidAmount
	}

	// Получаем карту
	card, err := s.cardRepo.GetByID(ctx, cardID)
	if err != nil {
		s.logger.Error("Card not found for payment", "card_id", cardID, "error", err)
		return ErrCardNotFound
	}

	// Проверяем статус карты
	if card.Status != "active" {
		s.logger.Warn("Card is not active", "card_id", cardID, "status", card.Status)
		return ErrCardBlocked
	}

	// Проверяем срок действия карты
	if time.Now().After(card.ExpiryDate) {
		s.logger.Warn("Card is expired", "card_id", cardID, "expiry_date", card.ExpiryDate)
		return ErrCardExpired
	}

	// Получаем счет карты
	account, err := s.accountRepo.GetByID(ctx, card.AccountID)
	if err != nil {
		s.logger.Error("Account not found for card payment", "card_id", cardID, "account_id", card.AccountID, "error", err)
		return ErrAccountNotFound
	}

	// Проверяем достаточность средств
	if account.Balance < amount {
		s.logger.Warn("Insufficient funds for card payment",
			"card_id", cardID,
			"account_id", card.AccountID,
			"balance", account.Balance,
			"amount", amount)
		return ErrInsufficientFunds
	}

	// Списываем средства со счета
	newBalance := account.Balance - amount
	if err := s.accountRepo.UpdateBalance(ctx, card.AccountID, newBalance); err != nil {
		s.logger.Error("Failed to update balance for card payment", "card_id", cardID, "error", err)
		return fmt.Errorf("failed to update balance: %w", err)
	}

	// Создаем запись о транзакции
	transaction := &models.Transaction{
		FromAccount: &card.AccountID,
		ToAccount:   nil, // Платеж во внешнюю систему
		Amount:      amount,
		Type:        "payment",
		Status:      "completed",
		Description: fmt.Sprintf("Card payment (Card ID: %d)", cardID),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.transactionRepo.Create(ctx, transaction); err != nil {
		s.logger.Error("Failed to create transaction record for card payment",
			"card_id", cardID,
			"amount", amount,
			"error", err)
		// Не возвращаем ошибку, так как платеж уже выполнен
	}

	s.logger.Info("Card payment processed successfully",
		"card_id", cardID,
		"account_id", card.AccountID,
		"amount", amount,
		"new_balance", newBalance)

	return nil
}
