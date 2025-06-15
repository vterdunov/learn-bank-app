package utils

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"time"
)

// Константы для карт
const (
	// LearnBankPrefix префикс для карт LearnBank
	LearnBankPrefix = "7777"
	// CardNumberLength длина номера карты
	CardNumberLength = 16
	// CVVLength длина CVV
	CVVLength = 3
)

// Тип карты LearnBank
const (
	CardTypeLearnBank = "LEARNBANK"
	CardTypeUnknown   = "UNKNOWN"
)

// Ошибки карт
var (
	ErrCardExpired = errors.New("card expired")
)

// GenerateCardNumber генерирует номер карты по алгоритму Луна
func GenerateCardNumber() (string, error) {
	prefix := LearnBankPrefix

	// Генерируем первые 15 цифр
	cardNumber := prefix
	for len(cardNumber) < CardNumberLength-1 {
		digit, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", fmt.Errorf("failed to generate random digit: %w", err)
		}
		cardNumber += digit.String()
	}

	// Вычисляем контрольную цифру по алгоритму Луна
	checkDigit := calculateLuhnCheckDigit(cardNumber)
	cardNumber += strconv.Itoa(checkDigit)

	return cardNumber, nil
}

// calculateLuhnCheckDigit вычисляет контрольную цифру по алгоритму Луна
func calculateLuhnCheckDigit(cardNumber string) int {
	sum := 0
	alternate := true

	// Проходим справа налево (исключая последнюю позицию для контрольной цифры)
	for i := len(cardNumber) - 1; i >= 0; i-- {
		digit, _ := strconv.Atoi(string(cardNumber[i]))

		if alternate {
			digit *= 2
			if digit > 9 {
				digit = digit/10 + digit%10
			}
		}

		sum += digit
		alternate = !alternate
	}

	return (10 - (sum % 10)) % 10
}

// GetCardType определяет тип карты по номеру
func GetCardType(cardNumber string) string {
	if len(cardNumber) < 4 {
		return CardTypeUnknown
	}

	prefix := cardNumber[:4]
	if prefix == LearnBankPrefix {
		return CardTypeLearnBank
	}

	return CardTypeUnknown
}

// GenerateCVV генерирует случайный CVV
func GenerateCVV() (string, error) {
	cvv := ""
	for i := 0; i < CVVLength; i++ {
		digit, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", fmt.Errorf("failed to generate CVV digit: %w", err)
		}
		cvv += digit.String()
	}
	return cvv, nil
}

// GenerateExpiryDate генерирует дату истечения карты (через 3 года)
func GenerateExpiryDate() string {
	expiry := time.Now().AddDate(3, 0, 0)
	return expiry.Format("01/06") // MM/YY
}

// ValidateExpiryDateCard проверяет дату истечения карты (отдельная функция от validation.go)
func ValidateExpiryDateCard(expiryDate string) error {
	if len(expiryDate) != 5 || expiryDate[2] != '/' {
		return errors.New("invalid expiry date format")
	}

	monthStr := expiryDate[:2]
	yearStr := expiryDate[3:]

	month, err := strconv.Atoi(monthStr)
	if err != nil || month < 1 || month > 12 {
		return errors.New("invalid expiry date month")
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		return errors.New("invalid expiry date year")
	}

	// Преобразуем двузначный год в четырехзначный
	if year < 50 {
		year += 2000
	} else if year < 100 {
		year += 1900
	}

	// Проверяем что карта не истекла
	expiry := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	lastDayOfMonth := expiry.AddDate(0, 1, -1)

	if time.Now().After(lastDayOfMonth) {
		return ErrCardExpired
	}

	return nil
}

// MaskCardNumber маскирует номер карты для отображения
func MaskCardNumber(cardNumber string) string {
	if len(cardNumber) != CardNumberLength {
		return "****-****-****-****"
	}

	return cardNumber[:4] + "-****-****-" + cardNumber[12:]
}

// FormatCardNumber форматирует номер карты с дефисами
func FormatCardNumber(cardNumber string) string {
	if len(cardNumber) != CardNumberLength {
		return cardNumber
	}

	return cardNumber[:4] + "-" + cardNumber[4:8] + "-" + cardNumber[8:12] + "-" + cardNumber[12:]
}

// IsCardExpiringSoon проверяет истекает ли карта в ближайшие N дней
func IsCardExpiringSoon(expiryDate string, days int) (bool, error) {
	if err := ValidateExpiryDateCard(expiryDate); err != nil {
		return false, err
	}

	monthStr := expiryDate[:2]
	yearStr := expiryDate[3:]

	month, _ := strconv.Atoi(monthStr)
	year, _ := strconv.Atoi(yearStr)

	if year < 50 {
		year += 2000
	} else if year < 100 {
		year += 1900
	}

	expiry := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	lastDayOfMonth := expiry.AddDate(0, 1, -1)

	warningDate := time.Now().AddDate(0, 0, days)

	return lastDayOfMonth.Before(warningDate), nil
}

// GetCardTypeDisplayName возвращает отображаемое название типа карты
func GetCardTypeDisplayName(cardType string) string {
	if cardType == CardTypeLearnBank {
		return "LearnBank"
	}
	return "Неизвестная карта"
}

// GetCardLimits возвращает лимиты для карты LearnBank
func GetCardLimits() (dailyLimit, monthlyLimit float64) {
	return 100000.0, 500000.0 // 100k в день, 500k в месяц
}

// GetCardFee возвращает комиссию за обслуживание карты (годовая)
func GetCardFee() float64 {
	return 1000.0 // 1000 руб в год
}

// GetCashbackRate возвращает процент кэшбэка для карты
func GetCashbackRate() float64 {
	return 1.0 // 1%
}

// IsValidCardType проверяет является ли тип карты валидным
func IsValidCardType(cardType string) bool {
	return cardType == CardTypeLearnBank
}
