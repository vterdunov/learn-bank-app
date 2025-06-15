package utils

import (
	"context"
	"errors"
)

// Ошибки авторизации
var (
	ErrUnauthorized     = errors.New("unauthorized")
	ErrForbidden        = errors.New("forbidden")
	ErrUserIDNotFound   = errors.New("user ID not found in context")
	ErrInvalidUserID    = errors.New("invalid user ID")
	ErrAccessDenied     = errors.New("access denied")
	ErrResourceNotOwned = errors.New("resource not owned by user")
)

// ContextKey тип для ключей контекста
type ContextKey string

const (
	// UserIDKey ключ для user ID в контексте
	UserIDKey ContextKey = "userID"
	// UsernameKey ключ для username в контексте
	UsernameKey ContextKey = "username"
	// EmailKey ключ для email в контексте
	EmailKey ContextKey = "email"
)

// GetUserIDFromContext извлекает ID пользователя из контекста
func GetUserIDFromContext(ctx context.Context) (int, error) {
	userID, ok := ctx.Value(UserIDKey).(int)
	if !ok {
		return 0, ErrUserIDNotFound
	}

	if userID <= 0 {
		return 0, ErrInvalidUserID
	}

	return userID, nil
}

// GetUsernameFromContext извлекает username пользователя из контекста
func GetUsernameFromContext(ctx context.Context) (string, error) {
	username, ok := ctx.Value(UsernameKey).(string)
	if !ok {
		return "", errors.New("username not found in context")
	}

	return username, nil
}

// GetEmailFromContext извлекает email пользователя из контекста
func GetEmailFromContext(ctx context.Context) (string, error) {
	email, ok := ctx.Value(EmailKey).(string)
	if !ok {
		return "", errors.New("email not found in context")
	}

	return email, nil
}

// SetUserIDInContext добавляет ID пользователя в контекст
func SetUserIDInContext(ctx context.Context, userID int) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// SetUsernameInContext добавляет username пользователя в контекст
func SetUsernameInContext(ctx context.Context, username string) context.Context {
	return context.WithValue(ctx, UsernameKey, username)
}

// SetEmailInContext добавляет email пользователя в контекст
func SetEmailInContext(ctx context.Context, email string) context.Context {
	return context.WithValue(ctx, EmailKey, email)
}

// SetUserDataInContext добавляет все данные пользователя в контекст
func SetUserDataInContext(ctx context.Context, userID int, username, email string) context.Context {
	ctx = SetUserIDInContext(ctx, userID)
	ctx = SetUsernameInContext(ctx, username)
	ctx = SetEmailInContext(ctx, email)
	return ctx
}

// CheckAccountOwnership проверяет является ли пользователь владельцем счета
func CheckAccountOwnership(ctx context.Context, accountUserID int) error {
	currentUserID, err := GetUserIDFromContext(ctx)
	if err != nil {
		return err
	}

	if currentUserID != accountUserID {
		return ErrResourceNotOwned
	}

	return nil
}

// CheckCardOwnership проверяет является ли пользователь владельцем карты через счет
func CheckCardOwnership(ctx context.Context, cardAccountUserID int) error {
	currentUserID, err := GetUserIDFromContext(ctx)
	if err != nil {
		return err
	}

	if currentUserID != cardAccountUserID {
		return ErrResourceNotOwned
	}

	return nil
}

// CheckCreditOwnership проверяет является ли пользователь владельцем кредита
func CheckCreditOwnership(ctx context.Context, creditUserID int) error {
	currentUserID, err := GetUserIDFromContext(ctx)
	if err != nil {
		return err
	}

	if currentUserID != creditUserID {
		return ErrResourceNotOwned
	}

	return nil
}

// CheckTransactionOwnership проверяет участвует ли пользователь в транзакции
func CheckTransactionOwnership(ctx context.Context, fromAccountUserID, toAccountUserID int) error {
	currentUserID, err := GetUserIDFromContext(ctx)
	if err != nil {
		return err
	}

	// Пользователь должен быть владельцем либо отправляющего, либо получающего счета
	if currentUserID != fromAccountUserID && currentUserID != toAccountUserID {
		return ErrResourceNotOwned
	}

	return nil
}

// CheckResourceAccess общая функция для проверки доступа к ресурсу
func CheckResourceAccess(ctx context.Context, resourceOwnerID int) error {
	currentUserID, err := GetUserIDFromContext(ctx)
	if err != nil {
		return err
	}

	if currentUserID != resourceOwnerID {
		return ErrAccessDenied
	}

	return nil
}

// IsAuthorized проверяет что пользователь авторизован (есть в контексте)
func IsAuthorized(ctx context.Context) bool {
	_, err := GetUserIDFromContext(ctx)
	return err == nil
}

// RequireAuth проверяет авторизацию и возвращает ошибку если не авторизован
func RequireAuth(ctx context.Context) error {
	if !IsAuthorized(ctx) {
		return ErrUnauthorized
	}
	return nil
}

// RequireOwnership проверяет что пользователь является владельцем ресурса
func RequireOwnership(ctx context.Context, resourceOwnerID int) error {
	if err := RequireAuth(ctx); err != nil {
		return err
	}

	return CheckResourceAccess(ctx, resourceOwnerID)
}

// ValidateUserAccess комплексная проверка доступа пользователя
func ValidateUserAccess(ctx context.Context, requiredUserID int) error {
	// Проверяем что пользователь авторизован
	currentUserID, err := GetUserIDFromContext(ctx)
	if err != nil {
		return ErrUnauthorized
	}

	// Проверяем что это тот же пользователь
	if currentUserID != requiredUserID {
		return ErrForbidden
	}

	return nil
}

// CanAccessAccount проверяет может ли пользователь получить доступ к счету
func CanAccessAccount(ctx context.Context, accountUserID int) bool {
	err := CheckAccountOwnership(ctx, accountUserID)
	return err == nil
}

// CanAccessCard проверяет может ли пользователь получить доступ к карте
func CanAccessCard(ctx context.Context, cardAccountUserID int) bool {
	err := CheckCardOwnership(ctx, cardAccountUserID)
	return err == nil
}

// CanAccessCredit проверяет может ли пользователь получить доступ к кредиту
func CanAccessCredit(ctx context.Context, creditUserID int) bool {
	err := CheckCreditOwnership(ctx, creditUserID)
	return err == nil
}
