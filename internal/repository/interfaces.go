package repository

import (
	"context"
	"time"

	"github.com/vterdunov/learn-bank-app/internal/models"
)

// UserRepository интерфейс для работы с пользователями
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id int) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id int) error
	EmailExists(ctx context.Context, email string) (bool, error)
	UsernameExists(ctx context.Context, username string) (bool, error)
}

// AccountRepository интерфейс для работы со счетами
type AccountRepository interface {
	Create(ctx context.Context, account *models.Account) error
	GetByID(ctx context.Context, id int) (*models.Account, error)
	GetByUserID(ctx context.Context, userID int) ([]*models.Account, error)
	GetByNumber(ctx context.Context, number string) (*models.Account, error)
	Update(ctx context.Context, account *models.Account) error
	UpdateBalance(ctx context.Context, id int, balance float64) error
	Delete(ctx context.Context, id int) error
	Transfer(ctx context.Context, fromID, toID int, amount float64) error
	GetBalance(ctx context.Context, id int) (float64, error)
}

// CardRepository интерфейс для работы с картами
type CardRepository interface {
	Create(ctx context.Context, card *models.Card) error
	GetByID(ctx context.Context, id int) (*models.Card, error)
	GetByAccountID(ctx context.Context, accountID int) ([]*models.Card, error)
	Update(ctx context.Context, card *models.Card) error
	Delete(ctx context.Context, id int) error
	UpdateStatus(ctx context.Context, id int, status string) error
	GetActiveCardsByAccount(ctx context.Context, accountID int) ([]*models.Card, error)
}

// TransactionRepository интерфейс для работы с транзакциями
type TransactionRepository interface {
	Create(ctx context.Context, transaction *models.Transaction) error
	GetByID(ctx context.Context, id int) (*models.Transaction, error)
	GetByAccountID(ctx context.Context, accountID int, limit, offset int) ([]*models.Transaction, error)
	GetByUserID(ctx context.Context, userID int, limit, offset int) ([]*models.Transaction, error)
	Update(ctx context.Context, transaction *models.Transaction) error
	Delete(ctx context.Context, id int) error
	GetTransactionsByDateRange(ctx context.Context, accountID int, startDate, endDate time.Time) ([]*models.Transaction, error)
	GetMonthlyStatistics(ctx context.Context, userID int, year int, month int) (*models.MonthlyStatistics, error)
}

// CreditRepository интерфейс для работы с кредитами
type CreditRepository interface {
	Create(ctx context.Context, credit *models.Credit) error
	GetByID(ctx context.Context, id int) (*models.Credit, error)
	GetByUserID(ctx context.Context, userID int) ([]*models.Credit, error)
	GetByAccountID(ctx context.Context, accountID int) ([]*models.Credit, error)
	Update(ctx context.Context, credit *models.Credit) error
	Delete(ctx context.Context, id int) error
	UpdateRemainingDebt(ctx context.Context, id int, remainingDebt float64) error
	GetActiveCredits(ctx context.Context) ([]*models.Credit, error)
	GetCreditAnalytics(ctx context.Context, userID int) (*models.CreditAnalytics, error)
}

// PaymentScheduleRepository интерфейс для работы с графиком платежей
type PaymentScheduleRepository interface {
	Create(ctx context.Context, payment *models.PaymentSchedule) error
	CreateBatch(ctx context.Context, payments []*models.PaymentSchedule) error
	GetByID(ctx context.Context, id int) (*models.PaymentSchedule, error)
	GetByCreditID(ctx context.Context, creditID int) ([]*models.PaymentSchedule, error)
	Update(ctx context.Context, payment *models.PaymentSchedule) error
	Delete(ctx context.Context, id int) error
	GetOverduePayments(ctx context.Context) ([]*models.PaymentSchedule, error)
	GetUpcomingPayments(ctx context.Context, days int) ([]*models.PaymentSchedule, error)
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
