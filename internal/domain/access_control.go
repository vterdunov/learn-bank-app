package domain

import (
	"context"
	"fmt"
)

// AccessControlService определяет интерфейс для контроля доступа на уровне домена
type AccessControlService interface {
	CanAccessAccount(ctx context.Context, userID, accountID int) error
	CanAccessCard(ctx context.Context, userID, cardID int) error
	CanAccessCredit(ctx context.Context, userID, creditID int) error
}

// AccessControlDomain реализует доменную логику контроля доступа
type AccessControlDomain struct {
	// Зависимости от репозиториев для проверки связей
	accountRepo AccountRepositoryInterface
	cardRepo    CardRepositoryInterface
	creditRepo  CreditRepositoryInterface
}

// NewAccessControlDomain создает новый экземпляр domain service
func NewAccessControlDomain(
	accountRepo AccountRepositoryInterface,
	cardRepo CardRepositoryInterface,
	creditRepo CreditRepositoryInterface,
) *AccessControlDomain {
	return &AccessControlDomain{
		accountRepo: accountRepo,
		cardRepo:    cardRepo,
		creditRepo:  creditRepo,
	}
}

// CanAccessAccount проверяет может ли пользователь работать со счетом
func (d *AccessControlDomain) CanAccessAccount(ctx context.Context, userID, accountID int) error {
	account, err := d.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return fmt.Errorf("account not found: %w", err)
	}

	if account.UserID != userID {
		return NewAccessDeniedError("account", accountID, userID)
	}

	return nil
}

// CanAccessCard проверяет может ли пользователь работать с картой
func (d *AccessControlDomain) CanAccessCard(ctx context.Context, userID, cardID int) error {
	card, err := d.cardRepo.GetByID(ctx, cardID)
	if err != nil {
		return fmt.Errorf("card not found: %w", err)
	}

	// Проверяем принадлежность карты через счет
	return d.CanAccessAccount(ctx, userID, card.AccountID)
}

// CanAccessCredit проверяет может ли пользователь работать с кредитом
func (d *AccessControlDomain) CanAccessCredit(ctx context.Context, userID, creditID int) error {
	credit, err := d.creditRepo.GetByID(ctx, creditID)
	if err != nil {
		return fmt.Errorf("credit not found: %w", err)
	}

	// Проверяем принадлежность кредита через счет
	return d.CanAccessAccount(ctx, userID, credit.AccountID)
}

// Интерфейсы репозиториев для domain service (только нужные методы)
type AccountRepositoryInterface interface {
	GetByID(ctx context.Context, id int) (*Account, error)
}

type CardRepositoryInterface interface {
	GetByID(ctx context.Context, id int) (*Card, error)
}

type CreditRepositoryInterface interface {
	GetByID(ctx context.Context, id int) (*Credit, error)
}

// AccessDeniedError кастомная ошибка для нарушения прав доступа
type AccessDeniedError struct {
	ResourceType string
	ResourceID   int
	UserID       int
}

func (e *AccessDeniedError) Error() string {
	return fmt.Sprintf("access denied: user %d cannot access %s %d", e.UserID, e.ResourceType, e.ResourceID)
}

func NewAccessDeniedError(resourceType string, resourceID, userID int) *AccessDeniedError {
	return &AccessDeniedError{
		ResourceType: resourceType,
		ResourceID:   resourceID,
		UserID:       userID,
	}
}

// IsAccessDeniedError проверяет является ли ошибка нарушением прав доступа
func IsAccessDeniedError(err error) bool {
	_, ok := err.(*AccessDeniedError)
	return ok
}
