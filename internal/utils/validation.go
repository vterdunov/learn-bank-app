package utils

import (
	"errors"
	"regexp"
	"strings"
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
)

// emailRegex регулярное выражение для валидации email
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// usernameRegex регулярное выражение для валидации username
var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

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
