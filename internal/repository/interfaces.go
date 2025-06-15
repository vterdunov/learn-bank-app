package repository

import (
	"context"
	"time"

	"github.com/vterdunov/learn-bank-app/internal/domain"
)

// UserRepository интерфейс для работы с пользователями
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id int) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id int) error
	EmailExists(ctx context.Context, email string) (bool, error)
	UsernameExists(ctx context.Context, username string) (bool, error)
}

// AccountRepository интерфейс для работы со счетами
type AccountRepository interface {
	Create(ctx context.Context, account *domain.Account) error
	GetByID(ctx context.Context, id int) (*domain.Account, error)
	GetByUserID(ctx context.Context, userID int) ([]*domain.Account, error)
	GetByNumber(ctx context.Context, number string) (*domain.Account, error)
	Update(ctx context.Context, account *domain.Account) error
	UpdateBalance(ctx context.Context, id int, balance float64) error
	Delete(ctx context.Context, id int) error
	Transfer(ctx context.Context, fromID, toID int, amount float64) error
	GetBalance(ctx context.Context, id int) (float64, error)
}

// CardRepository интерфейс для работы с картами
type CardRepository interface {
	Create(ctx context.Context, card *domain.Card) error
	GetByID(ctx context.Context, id int) (*domain.Card, error)
	GetByAccountID(ctx context.Context, accountID int) ([]*domain.Card, error)
	Update(ctx context.Context, card *domain.Card) error
	Delete(ctx context.Context, id int) error
	UpdateStatus(ctx context.Context, id int, status string) error
	GetActiveCardsByAccount(ctx context.Context, accountID int) ([]*domain.Card, error)
}

// TransactionRepository интерфейс для работы с транзакциями
type TransactionRepository interface {
	Create(ctx context.Context, transaction *domain.Transaction) error
	GetByID(ctx context.Context, id int) (*domain.Transaction, error)
	GetByAccountID(ctx context.Context, accountID int, limit, offset int) ([]*domain.Transaction, error)
	GetByUserID(ctx context.Context, userID int, limit, offset int) ([]*domain.Transaction, error)
	Update(ctx context.Context, transaction *domain.Transaction) error
	Delete(ctx context.Context, id int) error
	GetTransactionsByDateRange(ctx context.Context, accountID int, startDate, endDate time.Time) ([]*domain.Transaction, error)
	GetMonthlyStatistics(ctx context.Context, userID int, year int, month int) (*domain.MonthlyStatistics, error)
}

// CreditRepository интерфейс для работы с кредитами
type CreditRepository interface {
	Create(ctx context.Context, credit *domain.Credit) error
	GetByID(ctx context.Context, id int) (*domain.Credit, error)
	GetByUserID(ctx context.Context, userID int) ([]*domain.Credit, error)
	GetByAccountID(ctx context.Context, accountID int) ([]*domain.Credit, error)
	Update(ctx context.Context, credit *domain.Credit) error
	Delete(ctx context.Context, id int) error
	UpdateRemainingDebt(ctx context.Context, id int, remainingDebt float64) error
	GetActiveCredits(ctx context.Context) ([]*domain.Credit, error)
	GetCreditAnalytics(ctx context.Context, userID int) (*domain.CreditAnalytics, error)
}

// PaymentScheduleRepository интерфейс для работы с графиком платежей
type PaymentScheduleRepository interface {
	Create(ctx context.Context, payment *domain.PaymentSchedule) error
	CreateBatch(ctx context.Context, payments []*domain.PaymentSchedule) error
	GetByID(ctx context.Context, id int) (*domain.PaymentSchedule, error)
	GetByCreditID(ctx context.Context, creditID int) ([]*domain.PaymentSchedule, error)
	Update(ctx context.Context, payment *domain.PaymentSchedule) error
	Delete(ctx context.Context, id int) error
	GetOverduePayments(ctx context.Context) ([]*domain.PaymentSchedule, error)
	GetUpcomingPayments(ctx context.Context, days int) ([]*domain.PaymentSchedule, error)
	MarkAsPaid(ctx context.Context, id int, paidDate time.Time) error
	AddPenalty(ctx context.Context, id int, penaltyAmount float64) error
}

// Repositories структура содержащая все репозитории
type Repositories struct {
	User            UserRepository
	Account         AccountRepository
	Card            CardRepository
	Transaction     TransactionRepository
	Credit          CreditRepository
	PaymentSchedule PaymentScheduleRepository
}
