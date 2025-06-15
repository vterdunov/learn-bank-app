package service

import (
	"context"
	"time"

	"github.com/vterdunov/learn-bank-app/internal/domain"
)

// AuthService определяет интерфейс сервиса аутентификации
type AuthService interface {
	Register(ctx context.Context, req RegisterRequest) (*domain.User, error)
	Login(ctx context.Context, req LoginRequest) (string, error)
	ValidateToken(ctx context.Context, token string) (*domain.User, error)
}

// AccountService определяет интерфейс сервиса управления счетами
type AccountService interface {
	CreateAccount(ctx context.Context, userID int, req CreateAccountRequest) (*domain.Account, error)
	GetUserAccounts(ctx context.Context, userID int) ([]*domain.Account, error)
	DepositMoney(ctx context.Context, accountID int, amount float64) error
	WithdrawMoney(ctx context.Context, accountID int, amount float64) error
	TransferMoney(ctx context.Context, fromAccountID, toAccountID int, amount float64) error
}

// CardService определяет интерфейс сервиса управления картами
type CardService interface {
	CreateCard(ctx context.Context, accountID int) (*domain.Card, error)
	GetAccountCards(ctx context.Context, accountID int) ([]*domain.Card, error)
	DecryptCardData(ctx context.Context, card *domain.Card) (*CardData, error)
	ProcessPayment(ctx context.Context, cardID int, amount float64) error
}

// CreditService определяет интерфейс сервиса кредитования
type CreditService interface {
	CreateCredit(ctx context.Context, req domain.CreateCreditRequest) (*domain.Credit, error)
	GetCreditSchedule(ctx context.Context, creditID int) ([]*domain.PaymentSchedule, error)
	CalculateAnnuityPayment(principal, rate float64, months int) float64
	ProcessOverduePayments(ctx context.Context) error
}

// AnalyticsService определяет интерфейс сервиса аналитики
type AnalyticsService interface {
	GetMonthlyStatistics(ctx context.Context, userID int, month time.Time) (*MonthlyStats, error)
	GetCreditLoad(ctx context.Context, userID int) (*CreditLoad, error)
	PredictBalance(ctx context.Context, accountID int, days int) (*BalancePrediction, error)
}

// EmailService определяет интерфейс сервиса отправки email
type EmailService interface {
	SendPaymentNotification(userEmail string, amount float64) error
	SendCreditNotification(userEmail string, credit *domain.Credit) error
	SendOverdueNotification(userEmail string, payment *domain.PaymentSchedule) error
}

// CBRService определяет интерфейс сервиса интеграции с ЦБ РФ
type CBRService interface {
	GetKeyRate(ctx context.Context) (float64, error)
}

// SchedulerService определяет интерфейс сервиса шедулера
type SchedulerService interface {
	Start(ctx context.Context) error
	Stop()
	ProcessOverduePayments(ctx context.Context) error
}

// DTO структуры для запросов и ответов

// RegisterRequest структура запроса регистрации
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest структура запроса авторизации
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// CreateAccountRequest структура запроса создания счета
type CreateAccountRequest struct {
	Currency string `json:"currency"`
}

// CardData структура расшифрованных данных карты
type CardData struct {
	Number     string    `json:"number"`
	ExpiryDate time.Time `json:"expiry_date"`
}

// MonthlyStats структура месячной статистики
type MonthlyStats struct {
	Income   float64 `json:"income"`
	Expenses float64 `json:"expenses"`
	Balance  float64 `json:"balance"`
}

// CreditLoad структура кредитной нагрузки
type CreditLoad struct {
	TotalDebt       float64 `json:"total_debt"`
	MonthlyPayments float64 `json:"monthly_payments"`
	CreditRatio     float64 `json:"credit_ratio"`
}

// BalancePrediction структура прогноза баланса
type BalancePrediction struct {
	CurrentBalance   float64   `json:"current_balance"`
	PredictedBalance float64   `json:"predicted_balance"`
	PredictionDate   time.Time `json:"prediction_date"`
}
