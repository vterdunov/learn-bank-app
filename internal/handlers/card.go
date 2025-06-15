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

// CreateCardRequest структура запроса для создания карты
type CreateCardRequest struct {
	AccountID    string `json:"account_id" validate:"required"`
	CardType     string `json:"card_type" validate:"required,oneof=debit credit"`
	PinCode      string `json:"pin_code" validate:"required,len=4,numeric"`
	DailyLimit   int    `json:"daily_limit" validate:"required,min=1000,max=1000000"`
	MonthlyLimit int    `json:"monthly_limit" validate:"required,min=5000,max=10000000"`
}

// CardPaymentRequest структура запроса для оплаты картой
type CardPaymentRequest struct {
	Amount      float64 `json:"amount" validate:"required,gt=0"`
	MerchantID  string  `json:"merchant_id" validate:"required"`
	Description string  `json:"description,omitempty" validate:"max=255"`
	CVV         string  `json:"cvv" validate:"required,len=3,numeric"`
}

// CardResponse структура ответа с информацией о карте
type CardResponse struct {
	ID           string    `json:"id"`
	AccountID    string    `json:"account_id"`
	MaskedNumber string    `json:"masked_number"`
	CardType     string    `json:"card_type"`
	ExpiryMonth  int       `json:"expiry_month"`
	ExpiryYear   int       `json:"expiry_year"`
	Status       string    `json:"status"`
	DailyLimit   int       `json:"daily_limit"`
	MonthlyLimit int       `json:"monthly_limit"`
	DailySpent   float64   `json:"daily_spent"`
	MonthlySpent float64   `json:"monthly_spent"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CardHandler обрабатывает запросы управления картами
type CardHandler struct {
	cardService service.CardService
	logger      *slog.Logger
}

func NewCardHandler(cardService service.CardService, logger *slog.Logger) *CardHandler {
	return &CardHandler{
		cardService: cardService,
		logger:      logger,
	}
}

// CreateCard создает новую карту для счета
func (h *CardHandler) CreateCard(w http.ResponseWriter, r *http.Request) {
	var req CreateCardRequest

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

	// Конвертация account ID
	accountID, err := strconv.Atoi(req.AccountID)
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

	// Создание карты
	card, err := h.cardService.CreateCard(r.Context(), userID, accountID)
	if err != nil {
		h.logger.Error("Failed to create card", "account_id", accountID, "user_id", userID, "error", err.Error())
		WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	// Логирование
	h.logger.Info("Card created", "card_id", card.ID, "account_id", accountID)

	WriteSuccessResponse(w, CardToResponse(card))
}

// GetAccountCards получает все карты для указанного счета
func (h *CardHandler) GetAccountCards(w http.ResponseWriter, r *http.Request) {
	// Получение account ID из URL
	accountIDStr := r.PathValue("accountId")
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

	// Получение карт
	cards, err := h.cardService.GetAccountCards(r.Context(), userID, accountID)
	if err != nil {
		h.logger.Error("Failed to get account cards", "account_id", accountID, "user_id", userID, "error", err.Error())
		WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	// Конвертация в response формат
	var cardResponses []*CardResponse
	for _, card := range cards {
		cardResponses = append(cardResponses, CardToResponse(card))
	}

	WriteSuccessResponse(w, map[string]interface{}{
		"cards": cardResponses,
		"count": len(cardResponses),
	})
}

// CardPayment выполняет оплату картой
func (h *CardHandler) CardPayment(w http.ResponseWriter, r *http.Request) {
	var req CardPaymentRequest

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

	// Получение card ID из URL
	cardIDStr := r.PathValue("id")
	if cardIDStr == "" {
		WriteErrorResponse(w, http.StatusBadRequest, fmt.Errorf("card ID required"))
		return
	}

	cardID, err := strconv.Atoi(cardIDStr)
	if err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, fmt.Errorf("invalid card ID"))
		return
	}

	// Получение userID из контекста
	userID, err := GetUserIDFromRequest(r)
	if err != nil {
		WriteErrorResponse(w, http.StatusUnauthorized, err)
		return
	}

	// Выполнение платежа
	if err := h.cardService.ProcessPayment(r.Context(), userID, cardID, req.Amount); err != nil {
		h.logger.Error("Failed to process card payment",
			"card_id", cardID,
			"user_id", userID,
			"amount", req.Amount,
			"error", err.Error())

		// Определяем статус код на основе ошибки
		statusCode := http.StatusInternalServerError
		if err.Error() == "insufficient funds" || err.Error() == "invalid CVV" || err.Error() == "card expired" {
			statusCode = http.StatusBadRequest
		}

		WriteErrorResponse(w, statusCode, err)
		return
	}

	// Логирование
	h.logger.Info("Card payment processed",
		"card_id", cardID,
		"amount", req.Amount,
		"merchant_id", req.MerchantID)

	WriteSuccessResponse(w, map[string]string{"message": "Payment successful"})
}

// Conversion functions
func CardToResponse(card *domain.Card) *CardResponse {
	return &CardResponse{
		ID:           fmt.Sprintf("%d", card.ID),
		AccountID:    fmt.Sprintf("%d", card.AccountID),
		MaskedNumber: "****-****-****-XXXX", // Маскированный номер
		CardType:     domain.CardTypeLearnBank,
		ExpiryMonth:  int(card.ExpiryDate.Month()),
		ExpiryYear:   card.ExpiryDate.Year(),
		Status:       card.Status,
		DailyLimit:   100000,  // Пример лимита
		MonthlyLimit: 1000000, // Пример лимита
		DailySpent:   0.0,     // Заглушка
		MonthlySpent: 0.0,     // Заглушка
		CreatedAt:    card.CreatedAt,
		UpdatedAt:    card.UpdatedAt,
	}
}
