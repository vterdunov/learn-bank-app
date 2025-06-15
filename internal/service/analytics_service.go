package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/vterdunov/learn-bank-app/internal/repository"
	"github.com/vterdunov/learn-bank-app/pkg/logger"
)

// analyticsService реализует интерфейс AnalyticsService
type analyticsService struct {
	accountRepo     repository.AccountRepository
	transactionRepo repository.TransactionRepository
	creditRepo      repository.CreditRepository
	logger          *slog.Logger
}

// NewAnalyticsService создает новый экземпляр сервиса аналитики
func NewAnalyticsService(
	accountRepo repository.AccountRepository,
	transactionRepo repository.TransactionRepository,
	creditRepo repository.CreditRepository,
) AnalyticsService {
	return &analyticsService{
		accountRepo:     accountRepo,
		transactionRepo: transactionRepo,
		creditRepo:      creditRepo,
		logger:          logger.NewDefault(),
	}
}

// GetMonthlyStatistics возвращает статистику доходов/расходов за месяц
func (s *analyticsService) GetMonthlyStatistics(ctx context.Context, userID int, month time.Time) (*MonthlyStats, error) {
	s.logger.Info("Getting monthly statistics",
		slog.Int("user_id", userID),
		slog.String("month", month.Format("2006-01")),
	)

	// Простая заглушка - возвращаем моковые данные
	stats := &MonthlyStats{
		Income:   50000.0, // Доходы за месяц
		Expenses: 35000.0, // Расходы за месяц
		Balance:  15000.0, // Остаток
	}

	s.logger.Info("Monthly statistics calculated",
		slog.Int("user_id", userID),
		slog.Float64("income", stats.Income),
		slog.Float64("expenses", stats.Expenses),
	)

	return stats, nil
}

// GetCreditLoad возвращает информацию о кредитной нагрузке пользователя
func (s *analyticsService) GetCreditLoad(ctx context.Context, userID int) (*CreditLoad, error) {
	s.logger.Info("Getting credit load",
		slog.Int("user_id", userID),
	)

	// Простая заглушка - возвращаем моковые данные
	load := &CreditLoad{
		TotalDebt:       120000.0, // Общая задолженность
		MonthlyPayments: 8500.0,   // Ежемесячные платежи
		CreditRatio:     0.35,     // Коэффициент кредитной нагрузки (35%)
	}

	s.logger.Info("Credit load calculated",
		slog.Int("user_id", userID),
		slog.Float64("total_debt", load.TotalDebt),
		slog.Float64("monthly_payments", load.MonthlyPayments),
		slog.Float64("credit_ratio", load.CreditRatio),
	)

	return load, nil
}

// PredictBalance возвращает прогноз баланса счета на N дней
func (s *analyticsService) PredictBalance(ctx context.Context, accountID int, days int) (*BalancePrediction, error) {
	s.logger.Info("Predicting balance",
		slog.Int("account_id", accountID),
		slog.Int("days", days),
	)

	// Простая заглушка - возвращаем моковые данные
	prediction := &BalancePrediction{
		CurrentBalance:   25000.0,                        // Текущий баланс
		PredictedBalance: 22000.0,                        // Прогнозируемый баланс
		PredictionDate:   time.Now().AddDate(0, 0, days), // Дата прогноза
	}

	s.logger.Info("Balance prediction calculated",
		slog.Int("account_id", accountID),
		slog.Float64("current_balance", prediction.CurrentBalance),
		slog.Float64("predicted_balance", prediction.PredictedBalance),
		slog.Time("prediction_date", prediction.PredictionDate),
	)

	return prediction, nil
}
