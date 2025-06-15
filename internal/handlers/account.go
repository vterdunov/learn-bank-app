package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/vterdunov/learn-bank-app/internal/domain"
	"github.com/vterdunov/learn-bank-app/internal/service"
	"github.com/vterdunov/learn-bank-app/pkg/logger"
)

// Account Request DTOs
type CreateAccountRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	AccountType string `json:"account_type" validate:"required,oneof=savings checking"`
}

type DepositRequest struct {
	Amount float64 `json:"amount" validate:"required,gt=0"`
}

type WithdrawRequest struct {
	Amount float64 `json:"amount" validate:"required,gt=0"`
}

type TransferRequest struct {
	FromAccountID string  `json:"from_account_id" validate:"required,uuid"`
	ToAccountID   string  `json:"to_account_id" validate:"required,uuid"`
	Amount        float64 `json:"amount" validate:"required,gt=0"`
	Description   string  `json:"description,omitempty" validate:"max=255"`
}

// Account Response DTOs
type AccountResponse struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	AccountNumber string    `json:"account_number"`
	Name          string    `json:"name"`
	AccountType   string    `json:"account_type"`
	Balance       float64   `json:"balance"`
	Currency      string    `json:"currency"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type TransactionResponse struct {
	ID            string    `json:"id"`
	FromAccountID *string   `json:"from_account_id"`
	ToAccountID   *string   `json:"to_account_id"`
	Amount        float64   `json:"amount"`
	Type          string    `json:"type"`
	Status        string    `json:"status"`
	Description   string    `json:"description"`
	CreatedAt     time.Time `json:"created_at"`
}

// AccountHandler обрабатывает запросы управления счетами
type AccountHandler struct {
	accountService service.AccountService
	logger         *slog.Logger
}

func NewAccountHandler(accountService service.AccountService, logger *slog.Logger) *AccountHandler {
	return &AccountHandler{
		accountService: accountService,
		logger:         logger,
	}
}

// CreateAccount создает новый банковский счет
func (h *AccountHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	var req CreateAccountRequest

	// Валидация JSON
	if err := ValidateJSON(r, &req); err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	// Валидация полей
	if validationErr := Validate(&req); validationErr != nil {
		WriteErrorResponse(w, http.StatusBadRequest, validationErr)
		return
	}

	// Получение userID из контекста (устанавливается middleware)
	userID, err := GetUserIDFromRequest(r)
	if err != nil {
		WriteErrorResponse(w, http.StatusUnauthorized, err)
		return
	}

	// Создание счета
	serviceReq := service.CreateAccountRequest{
		Currency: "RUB", // По умолчанию RUB согласно ТЗ
	}

	account, err := h.accountService.CreateAccount(r.Context(), userID, serviceReq)
	if err != nil {
		h.logger.Error("Failed to create account", "user_id", userID, "error", err.Error())
		WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	// Логирование
	logger.LogUserAction(h.logger, userID, "account_created", map[string]interface{}{
		"account_id": account.ID,
	})

	// Успешный ответ
	response := AccountToResponse(account)
	WriteSuccessResponse(w, response)
}

// GetUserAccounts получает все счета пользователя
func (h *AccountHandler) GetUserAccounts(w http.ResponseWriter, r *http.Request) {
	// Получение userID из контекста
	userID, err := GetUserIDFromRequest(r)
	if err != nil {
		WriteErrorResponse(w, http.StatusUnauthorized, err)
		return
	}

	// Получение счетов
	accounts, err := h.accountService.GetUserAccounts(r.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get user accounts", "user_id", userID, "error", err.Error())
		WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	// Конвертация в response
	var responses []*AccountResponse
	for _, account := range accounts {
		responses = append(responses, AccountToResponse(account))
	}

	WriteSuccessResponse(w, responses)
}

// Deposit пополняет счет
func (h *AccountHandler) Deposit(w http.ResponseWriter, r *http.Request) {
	var req DepositRequest

	// Валидация JSON
	if err := ValidateJSON(r, &req); err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	// Валидация полей
	if validationErr := Validate(&req); validationErr != nil {
		WriteErrorResponse(w, http.StatusBadRequest, validationErr)
		return
	}

	// Получение account ID из URL
	accountIDStr := r.PathValue("id")
	if accountIDStr == "" {
		WriteErrorResponse(w, http.StatusBadRequest, fmt.Errorf("account ID required"))
		return
	}

	accountID, err := strconv.Atoi(accountIDStr)
	if err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, fmt.Errorf("invalid account ID"))
		return
	}

	// Получение userID из контекста
	userID, err := GetUserIDFromRequest(r)
	if err != nil {
		WriteErrorResponse(w, http.StatusUnauthorized, err)
		return
	}

	// Пополнение счета (проверка прав доступа встроена в сервис)
	if err := h.accountService.DepositMoney(r.Context(), userID, accountID, req.Amount); err != nil {
		h.logger.Error("Failed to deposit money", "account_id", accountID, "amount", req.Amount, "error", err.Error())
		WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	// Логирование
	h.logger.Info("Money deposited", "account_id", accountID, "amount", req.Amount)

	WriteSuccessResponse(w, map[string]string{"message": "Deposit successful"})
}

// Withdraw списывает средства со счета
func (h *AccountHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	var req WithdrawRequest

	// Валидация JSON
	if err := ValidateJSON(r, &req); err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	// Валидация полей
	if validationErr := Validate(&req); validationErr != nil {
		WriteErrorResponse(w, http.StatusBadRequest, validationErr)
		return
	}

	// Получение account ID из URL
	accountIDStr := r.PathValue("id")
	if accountIDStr == "" {
		WriteErrorResponse(w, http.StatusBadRequest, fmt.Errorf("account ID required"))
		return
	}

	accountID, err := strconv.Atoi(accountIDStr)
	if err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, fmt.Errorf("invalid account ID"))
		return
	}

	// Получение userID из контекста
	userID, err := GetUserIDFromRequest(r)
	if err != nil {
		WriteErrorResponse(w, http.StatusUnauthorized, err)
		return
	}

	// Списание средств (проверка прав доступа встроена в сервис)
	if err := h.accountService.WithdrawMoney(r.Context(), userID, accountID, req.Amount); err != nil {
		h.logger.Error("Failed to withdraw money", "account_id", accountID, "amount", req.Amount, "error", err.Error())

		// Определяем статус код на основе ошибки
		statusCode := http.StatusInternalServerError
		if err.Error() == "insufficient funds" {
			statusCode = http.StatusBadRequest
		}

		WriteErrorResponse(w, statusCode, err)
		return
	}

	// Логирование
	h.logger.Info("Money withdrawn", "account_id", accountID, "amount", req.Amount)

	WriteSuccessResponse(w, map[string]string{"message": "Withdrawal successful"})
}

// Transfer выполняет перевод между счетами
func (h *AccountHandler) Transfer(w http.ResponseWriter, r *http.Request) {
	var req TransferRequest

	// Валидация JSON
	if err := ValidateJSON(r, &req); err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	// Валидация полей
	if validationErr := Validate(&req); validationErr != nil {
		WriteErrorResponse(w, http.StatusBadRequest, validationErr)
		return
	}

	// Конвертация ID из строк в int
	fromAccountID, err := strconv.Atoi(req.FromAccountID)
	if err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, fmt.Errorf("invalid from_account_id"))
		return
	}

	toAccountID, err := strconv.Atoi(req.ToAccountID)
	if err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, fmt.Errorf("invalid to_account_id"))
		return
	}

	// Получение userID из контекста
	userID, err := GetUserIDFromRequest(r)
	if err != nil {
		WriteErrorResponse(w, http.StatusUnauthorized, err)
		return
	}

	// Выполнение перевода (проверка прав доступа встроена в сервис)
	if err := h.accountService.TransferMoney(r.Context(), userID, fromAccountID, toAccountID, req.Amount); err != nil {
		h.logger.Error("Failed to transfer money",
			"from_account_id", fromAccountID,
			"to_account_id", toAccountID,
			"amount", req.Amount,
			"error", err.Error())

		// Определяем статус код на основе ошибки
		statusCode := http.StatusInternalServerError
		if err.Error() == "insufficient funds" || err.Error() == "account not found" {
			statusCode = http.StatusBadRequest
		}

		WriteErrorResponse(w, statusCode, err)
		return
	}

	// Логирование
	h.logger.Info("Money transferred",
		"from_account_id", fromAccountID,
		"to_account_id", toAccountID,
		"amount", req.Amount)

	WriteSuccessResponse(w, map[string]string{"message": "Transfer successful"})
}

// Conversion functions
func AccountToResponse(account *domain.Account) *AccountResponse {
	return &AccountResponse{
		ID:            fmt.Sprintf("%d", account.ID),
		UserID:        fmt.Sprintf("%d", account.UserID),
		AccountNumber: account.Number,
		Name:          "", // Domain doesn't have Name field
		AccountType:   "", // Domain doesn't have AccountType field
		Balance:       account.Balance,
		Currency:      account.Currency,
		Status:        account.Status,
		CreatedAt:     account.CreatedAt,
		UpdatedAt:     account.UpdatedAt,
	}
}

func TransactionToResponse(transaction *domain.Transaction) *TransactionResponse {
	var fromAccountID, toAccountID *string
	if transaction.FromAccount != nil {
		fromAccountIDStr := fmt.Sprintf("%d", *transaction.FromAccount)
		fromAccountID = &fromAccountIDStr
	}
	if transaction.ToAccount != nil {
		toAccountIDStr := fmt.Sprintf("%d", *transaction.ToAccount)
		toAccountID = &toAccountIDStr
	}

	return &TransactionResponse{
		ID:            fmt.Sprintf("%d", transaction.ID),
		FromAccountID: fromAccountID,
		ToAccountID:   toAccountID,
		Amount:        transaction.Amount,
		Type:          transaction.Type,
		Status:        transaction.Status,
		Description:   transaction.Description,
		CreatedAt:     transaction.CreatedAt,
	}
}
