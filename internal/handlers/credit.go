package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/vterdunov/learn-bank-app/internal/domain"
	"github.com/vterdunov/learn-bank-app/internal/service"
)

// Credit Request DTOs
type CreateCreditRequest struct {
	AccountID   string  `json:"account_id" validate:"required,uuid"`
	Amount      float64 `json:"amount" validate:"required,gt=0"`
	TermMonths  int     `json:"term_months" validate:"required,min=1,max=360"`
	Description string  `json:"description,omitempty" validate:"max=255"`
}

// Credit Response DTOs
type CreditResponse struct {
	ID             string    `json:"id"`
	AccountID      string    `json:"account_id"`
	Amount         float64   `json:"amount"`
	InterestRate   float64   `json:"interest_rate"`
	TermMonths     int       `json:"term_months"`
	MonthlyPayment float64   `json:"monthly_payment"`
	RemainingDebt  float64   `json:"remaining_debt"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type PaymentScheduleResponse struct {
	ID                string     `json:"id"`
	CreditID          string     `json:"credit_id"`
	PaymentNumber     int        `json:"payment_number"`
	PaymentDate       time.Time  `json:"payment_date"`
	PaymentAmount     float64    `json:"payment_amount"`
	PrincipalAmount   float64    `json:"principal_amount"`
	InterestAmount    float64    `json:"interest_amount"`
	RemainingBalance  float64    `json:"remaining_balance"`
	Status            string     `json:"status"`
	ActualPaymentDate *time.Time `json:"actual_payment_date"`
	CreatedAt         time.Time  `json:"created_at"`
}

// CreditHandler обрабатывает запросы кредитования
type CreditHandler struct {
	creditService service.CreditService
	logger        *slog.Logger
}

func NewCreditHandler(creditService service.CreditService, logger *slog.Logger) *CreditHandler {
	return &CreditHandler{
		creditService: creditService,
		logger:        logger,
	}
}

// CreateCredit создает новый кредит
func (h *CreditHandler) CreateCredit(w http.ResponseWriter, r *http.Request) {
	var req CreateCreditRequest

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

	// Получение userID из контекста
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

	// Создание запроса для сервиса (используя правильную структуру из domain)
	creditReq := domain.CreateCreditRequest{
		AccountID:    accountID,
		Amount:       req.Amount,
		TermMonths:   req.TermMonths,
		InterestRate: 15.0, // Базовая ставка, может быть получена из ЦБ РФ
	}

	credit, err := h.creditService.CreateCredit(r.Context(), userID, creditReq)
	if err != nil {
		h.logger.Error("Failed to create credit",
			"account_id", accountID,
			"user_id", userID,
			"amount", req.Amount,
			"error", err.Error())
		WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	h.logger.Info("Credit created", "credit_id", credit.ID, "account_id", accountID, "amount", req.Amount)

	response := CreditToResponse(credit)
	WriteSuccessResponse(w, response)
}

// GetCreditSchedule получает график платежей по кредиту
func (h *CreditHandler) GetCreditSchedule(w http.ResponseWriter, r *http.Request) {
	creditIDStr := r.PathValue("id")
	if creditIDStr == "" {
		WriteErrorResponse(w, http.StatusBadRequest, fmt.Errorf("credit ID required"))
		return
	}

	creditID, err := strconv.Atoi(creditIDStr)
	if err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, fmt.Errorf("invalid credit ID"))
		return
	}

	// Получение userID из контекста
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

	schedule, err := h.creditService.GetCreditSchedule(r.Context(), userID, creditID)
	if err != nil {
		h.logger.Error("Failed to get credit schedule", "credit_id", creditID, "user_id", userID, "error", err.Error())
		WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	var responses []*PaymentScheduleResponse
	for _, payment := range schedule {
		responses = append(responses, PaymentScheduleToResponse(payment))
	}

	WriteSuccessResponse(w, responses)
}

// Conversion functions
func CreditToResponse(credit *domain.Credit) *CreditResponse {
	return &CreditResponse{
		ID:             fmt.Sprintf("%d", credit.ID),
		AccountID:      fmt.Sprintf("%d", credit.AccountID),
		Amount:         credit.Amount,
		InterestRate:   credit.InterestRate,
		TermMonths:     credit.TermMonths,
		MonthlyPayment: credit.MonthlyPayment,
		RemainingDebt:  credit.RemainingDebt,
		Status:         credit.Status,
		CreatedAt:      credit.CreatedAt,
		UpdatedAt:      credit.UpdatedAt,
	}
}

func PaymentScheduleToResponse(payment *domain.PaymentSchedule) *PaymentScheduleResponse {
	return &PaymentScheduleResponse{
		ID:                fmt.Sprintf("%d", payment.ID),
		CreditID:          fmt.Sprintf("%d", payment.CreditID),
		PaymentNumber:     payment.PaymentNumber,
		PaymentDate:       payment.DueDate,
		PaymentAmount:     payment.PaymentAmount,
		PrincipalAmount:   payment.PrincipalAmount,
		InterestAmount:    payment.InterestAmount,
		RemainingBalance:  payment.RemainingBalance,
		Status:            payment.Status,
		ActualPaymentDate: payment.PaidDate,
		CreatedAt:         payment.CreatedAt,
	}
}
