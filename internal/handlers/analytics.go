package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/vterdunov/learn-bank-app/internal/service"
)

// Analytics Request DTOs
type MonthlyStatsRequest struct {
	Month int `json:"month" validate:"required,min=1,max=12"`
	Year  int `json:"year" validate:"required,min=2020,max=2030"`
}

type BalancePredictionRequest struct {
	AccountID string `json:"account_id" validate:"required,uuid"`
	Days      int    `json:"days" validate:"required,min=1,max=365"`
}

// Analytics Response DTOs
type MonthlyStatsResponse struct {
	Income   float64 `json:"income"`
	Expenses float64 `json:"expenses"`
	Balance  float64 `json:"balance"`
	Month    int     `json:"month"`
	Year     int     `json:"year"`
}

type CreditLoadResponse struct {
	TotalDebt       float64 `json:"total_debt"`
	MonthlyPayments float64 `json:"monthly_payments"`
	CreditRatio     float64 `json:"credit_ratio"`
}

type BalancePredictionResponse struct {
	CurrentBalance    float64   `json:"current_balance"`
	PredictedBalance  float64   `json:"predicted_balance"`
	PredictionDate    time.Time `json:"prediction_date"`
	ScheduledPayments float64   `json:"scheduled_payments"`
}

// AnalyticsHandler обрабатывает запросы аналитики
type AnalyticsHandler struct {
	analyticsService service.AnalyticsService
	logger           *slog.Logger
}

func NewAnalyticsHandler(analyticsService service.AnalyticsService, logger *slog.Logger) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: analyticsService,
		logger:           logger,
	}
}

// GetMonthlyStats получает месячную статистику
func (h *AnalyticsHandler) GetMonthlyStats(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value("userID")
	if userIDStr == nil {
		WriteErrorResponse(w, http.StatusUnauthorized, fmt.Errorf("user not authenticated"))
		return
	}

	userID, err := strconv.Atoi(userIDStr.(string))
	if err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, fmt.Errorf("invalid user ID"))
		return
	}

	// Получаем текущий месяц по умолчанию
	now := time.Now()

	stats, err := h.analyticsService.GetMonthlyStatistics(context.Background(), userID, now)
	if err != nil {
		h.logger.Error("Failed to get monthly stats", "user_id", userID, "error", err.Error())
		WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	response := &MonthlyStatsResponse{
		Income:   stats.Income,
		Expenses: stats.Expenses,
		Balance:  stats.Balance,
		Month:    int(now.Month()),
		Year:     now.Year(),
	}

	WriteSuccessResponse(w, response)
}

// GetCreditLoad получает кредитную нагрузку пользователя
func (h *AnalyticsHandler) GetCreditLoad(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value("userID")
	if userIDStr == nil {
		WriteErrorResponse(w, http.StatusUnauthorized, fmt.Errorf("user not authenticated"))
		return
	}

	userID, err := strconv.Atoi(userIDStr.(string))
	if err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, fmt.Errorf("invalid user ID"))
		return
	}

	creditLoad, err := h.analyticsService.GetCreditLoad(context.Background(), userID)
	if err != nil {
		h.logger.Error("Failed to get credit load", "user_id", userID, "error", err.Error())
		WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	response := &CreditLoadResponse{
		TotalDebt:       creditLoad.TotalDebt,
		MonthlyPayments: creditLoad.MonthlyPayments,
		CreditRatio:     creditLoad.CreditRatio,
	}

	WriteSuccessResponse(w, response)
}

// PredictBalance прогнозирует баланс счета
func (h *AnalyticsHandler) PredictBalance(w http.ResponseWriter, r *http.Request) {
	var req BalancePredictionRequest

	if err := ValidateJSON(r, &req); err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	if validationErr := Validate(&req); validationErr != nil {
		WriteErrorResponse(w, http.StatusBadRequest, validationErr)
		return
	}

	accountID, err := strconv.Atoi(req.AccountID)
	if err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, fmt.Errorf("invalid account_id"))
		return
	}

	prediction, err := h.analyticsService.PredictBalance(context.Background(), accountID, req.Days)
	if err != nil {
		h.logger.Error("Failed to predict balance", "account_id", accountID, "days", req.Days, "error", err.Error())
		WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	response := &BalancePredictionResponse{
		CurrentBalance:   prediction.CurrentBalance,
		PredictedBalance: prediction.PredictedBalance,
		PredictionDate:   prediction.PredictionDate,
	}

	WriteSuccessResponse(w, response)
}
