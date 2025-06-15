package utils

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

var (
	// ErrInvalidEmail возвращается при некорректном email
	ErrInvalidEmail = errors.New("invalid email format")
	// ErrInvalidUsername возвращается при некорректном username
	ErrInvalidUsername = errors.New("invalid username format")
	// ErrWeakPassword возвращается при слабом пароле
	ErrWeakPassword = errors.New("password is too weak")
	// ErrEmptyField возвращается при пустом обязательном поле
	ErrEmptyField = errors.New("field cannot be empty")
	// ErrEmailExists возвращается когда email уже существует
	ErrEmailExists = errors.New("email already exists")
	// ErrUsernameExists возвращается когда username уже существует
	ErrUsernameExists = errors.New("username already exists")
	// ErrUserNotFound возвращается когда пользователь не найден
	ErrUserNotFound = errors.New("user not found")
	// ErrInvalidAccountNumber возвращается при некорректном номере счета
	ErrInvalidAccountNumber = errors.New("invalid account number")
	// ErrInvalidCardNumber возвращается при некорректном номере карты
	ErrInvalidCardNumber = errors.New("invalid card number")
	// ErrInvalidDate возвращается при некорректной дате
	ErrInvalidDate = errors.New("invalid date")
	// ErrInvalidStatus возвращается при некорректном статусе
	ErrInvalidStatus = errors.New("invalid status")
	// ErrInvalidCVV возвращается при некорректном CVV
	ErrInvalidCVV = errors.New("invalid CVV")
)

// emailRegex регулярное выражение для валидации email
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// usernameRegex регулярное выражение для валидации username
var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// accountNumberRegex регулярное выражение для номера счета (20 цифр)
var accountNumberRegex = regexp.MustCompile(`^\d{20}$`)

// ValidateEmail проверяет корректность email адреса
func ValidateEmail(email string) error {
	if strings.TrimSpace(email) == "" {
		return ErrEmptyField
	}

	if !emailRegex.MatchString(email) {
		return ErrInvalidEmail
	}

	return nil
}

// ValidateUsername проверяет корректность имени пользователя
func ValidateUsername(username string) error {
	if strings.TrimSpace(username) == "" {
		return ErrEmptyField
	}

	// Проверка длины (от 3 до 30 символов)
	if len(username) < 3 || len(username) > 30 {
		return ErrInvalidUsername
	}

	// Проверка допустимых символов (буквы, цифры, _, -)
	if !usernameRegex.MatchString(username) {
		return ErrInvalidUsername
	}

	return nil
}

// ValidatePassword проверяет сложность пароля
func ValidatePassword(password string) error {
	if strings.TrimSpace(password) == "" {
		return ErrEmptyField
	}

	// Минимальная длина 8 символов
	if len(password) < 8 {
		return ErrWeakPassword
	}

	var (
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	// Проверка наличия разных типов символов
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	// Пароль должен содержать минимум 3 из 4 типов символов
	typesCount := 0
	if hasUpper {
		typesCount++
	}
	if hasLower {
		typesCount++
	}
	if hasNumber {
		typesCount++
	}
	if hasSpecial {
		typesCount++
	}

	if typesCount < 3 {
		return ErrWeakPassword
	}

	return nil
}

// ValidateAmount проверяет корректность суммы
func ValidateAmount(amount float64) error {
	if amount <= 0 {
		return errors.New("amount must be positive")
	}

	if amount > 1000000000 { // Лимит в 1 млрд
		return errors.New("amount too large")
	}

	return nil
}

// ValidateCurrency проверяет поддерживаемые валюты
func ValidateCurrency(currency string) error {
	if strings.TrimSpace(currency) == "" {
		return ErrEmptyField
	}

	// Поддерживается только RUB согласно ТЗ
	if strings.ToUpper(currency) != "RUB" {
		return errors.New("only RUB currency is supported")
	}

	return nil
}

// ValidateAccountNumber проверяет номер банковского счета
func ValidateAccountNumber(accountNumber string) error {
	if strings.TrimSpace(accountNumber) == "" {
		return ErrEmptyField
	}

	// Номер счета должен состоять из 20 цифр
	if !accountNumberRegex.MatchString(accountNumber) {
		return ErrInvalidAccountNumber
	}

	return nil
}

// ValidateCardNumber проверяет номер карты LearnBank с помощью алгоритма Луна
func ValidateCardNumber(cardNumber string) error {
	if strings.TrimSpace(cardNumber) == "" {
		return ErrEmptyField
	}

	// Удаляем пробелы и дефисы
	cardNumber = strings.ReplaceAll(strings.ReplaceAll(cardNumber, " ", ""), "-", "")

	// Проверяем что состоит только из цифр
	if _, err := strconv.Atoi(cardNumber); err != nil {
		return ErrInvalidCardNumber
	}

	// Длина должна быть 16 цифр для карт LearnBank
	if len(cardNumber) != 16 {
		return ErrInvalidCardNumber
	}

	// Проверяем что это карта LearnBank (префикс 7777)
	if len(cardNumber) >= 4 {
		prefix := cardNumber[:4]
		if prefix != "7777" {
			return errors.New("only LearnBank cards are supported")
		}
	}

	// Проверка алгоритмом Луна
	if !isValidLuhn(cardNumber) {
		return ErrInvalidCardNumber
	}

	return nil
}

// isValidLuhn проверяет номер карты алгоритмом Луна
func isValidLuhn(cardNumber string) bool {
	var sum int
	alternate := false

	// Обходим цифры справа налево
	for i := len(cardNumber) - 1; i >= 0; i-- {
		digit := int(cardNumber[i] - '0')

		if alternate {
			digit *= 2
			if digit > 9 {
				digit = digit%10 + digit/10
			}
		}

		sum += digit
		alternate = !alternate
	}

	return sum%10 == 0
}

// ValidateCVV проверяет CVV код карты
func ValidateCVV(cvv string) error {
	if strings.TrimSpace(cvv) == "" {
		return ErrEmptyField
	}

	// CVV должен состоять из 3 или 4 цифр
	if len(cvv) < 3 || len(cvv) > 4 {
		return ErrInvalidCVV
	}

	// Проверяем что состоит только из цифр
	if _, err := strconv.Atoi(cvv); err != nil {
		return ErrInvalidCVV
	}

	return nil
}

// ValidateExpiryDate проверяет срок действия карты
func ValidateExpiryDate(expiryDate time.Time) error {
	now := time.Now()

	// Срок действия не должен быть в прошлом
	if expiryDate.Before(now) {
		return ErrInvalidDate
	}

	// Срок действия не должен быть более чем через 10 лет
	maxDate := now.AddDate(10, 0, 0)
	if expiryDate.After(maxDate) {
		return ErrInvalidDate
	}

	return nil
}

// ValidateDueDate проверяет дату платежа
func ValidateDueDate(dueDate time.Time) error {
	now := time.Now()

	// Дата платежа должна быть в будущем
	if dueDate.Before(now.AddDate(0, 0, -1)) { // Разрешаем вчерашнюю дату
		return ErrInvalidDate
	}

	// Дата платежа не должна быть более чем через 50 лет
	maxDate := now.AddDate(50, 0, 0)
	if dueDate.After(maxDate) {
		return ErrInvalidDate
	}

	return nil
}

// ValidateAccountStatus проверяет статус счета
func ValidateAccountStatus(status string) error {
	validStatuses := []string{"active", "blocked", "closed"}

	for _, validStatus := range validStatuses {
		if status == validStatus {
			return nil
		}
	}

	return ErrInvalidStatus
}

// ValidateCardStatus проверяет статус карты
func ValidateCardStatus(status string) error {
	validStatuses := []string{"active", "blocked", "expired"}

	for _, validStatus := range validStatuses {
		if status == validStatus {
			return nil
		}
	}

	return ErrInvalidStatus
}

// ValidateCardType проверяет тип карты LearnBank
func ValidateCardType(cardType string) error {
	if cardType == "LEARNBANK" {
		return nil
	}
	return errors.New("invalid card type")
}

// ValidateTransactionType проверяет тип транзакции
func ValidateTransactionType(transactionType string) error {
	validTypes := []string{"deposit", "withdraw", "transfer", "payment", "credit"}

	for _, validType := range validTypes {
		if transactionType == validType {
			return nil
		}
	}

	return ErrInvalidStatus
}

// ValidateTransactionStatus проверяет статус транзакции
func ValidateTransactionStatus(status string) error {
	validStatuses := []string{"pending", "completed", "failed", "cancelled"}

	for _, validStatus := range validStatuses {
		if status == validStatus {
			return nil
		}
	}

	return ErrInvalidStatus
}

// ValidateCreditStatus проверяет статус кредита
func ValidateCreditStatus(status string) error {
	validStatuses := []string{"active", "paid_off", "overdue", "cancelled"}

	for _, validStatus := range validStatuses {
		if status == validStatus {
			return nil
		}
	}

	return ErrInvalidStatus
}

// ValidatePaymentStatus проверяет статус платежа
func ValidatePaymentStatus(status string) error {
	validStatuses := []string{"pending", "paid", "overdue", "cancelled"}

	for _, validStatus := range validStatuses {
		if status == validStatus {
			return nil
		}
	}

	return ErrInvalidStatus
}

// ValidateTermMonths проверяет срок кредита в месяцах
func ValidateTermMonths(termMonths int) error {
	if termMonths <= 0 {
		return errors.New("term months must be positive")
	}

	// Максимальный срок кредита 30 лет (360 месяцев)
	if termMonths > 360 {
		return errors.New("term months too large")
	}

	return nil
}

// ValidateInterestRate проверяет процентную ставку
func ValidateInterestRate(rate float64) error {
	if rate < 0 {
		return errors.New("interest rate cannot be negative")
	}

	// Максимальная ставка 100%
	if rate > 100 {
		return errors.New("interest rate too high")
	}

	return nil
}

// IsUniqueConstraintError проверяет является ли ошибка нарушением уникальности
func IsUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "unique") ||
		strings.Contains(errStr, "duplicate") ||
		strings.Contains(errStr, "already exists")
}

// ParseUniqueConstraintError определяет какое поле нарушает уникальность
func ParseUniqueConstraintError(err error, email, username string) error {
	if err == nil {
		return nil
	}

	errStr := strings.ToLower(err.Error())

	// Проверяем какое поле нарушает уникальность
	if strings.Contains(errStr, "email") || strings.Contains(errStr, email) {
		return ErrEmailExists
	}

	if strings.Contains(errStr, "username") || strings.Contains(errStr, username) {
		return ErrUsernameExists
	}

	// Общая ошибка уникальности если не удалось определить поле
	return errors.New("unique constraint violation")
}

// ValidateRegistrationData проверяет все поля при регистрации
func ValidateRegistrationData(email, username, password string) error {
	if err := ValidateEmail(email); err != nil {
		return err
	}

	if err := ValidateUsername(username); err != nil {
		return err
	}

	if err := ValidatePassword(password); err != nil {
		return err
	}

	return nil
}

// ValidateLoginData проверяет поля при входе
func ValidateLoginData(email, password string) error {
	if err := ValidateEmail(email); err != nil {
		return err
	}

	if strings.TrimSpace(password) == "" {
		return ErrEmptyField
	}

	return nil
}
