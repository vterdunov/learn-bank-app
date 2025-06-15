package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/vterdunov/learn-bank-app/internal/domain"
	"github.com/vterdunov/learn-bank-app/internal/service"
)

// Card Request DTOs
type CreateCardRequest struct {
	AccountID string `json:"account_id" validate:"required,uuid"`
	CardType  string `json:"card_type" validate:"required,oneof=debit credit"`
}

type CardPaymentRequest struct {
	Amount      float64 `json:"amount" validate:"required,gt=0"`
	Merchant    string  `json:"merchant" validate:"required,min=1,max=100"`
	CVV         string  `json:"cvv" validate:"required,len=3,numeric"`
	Description string  `json:"description,omitempty" validate:"max=255"`
}

// Card Response DTOs
type CardResponse struct {
	ID         string    `json:"id"`
	AccountID  string    `json:"account_id"`
	Number     string    `json:"number"` // Masked/encrypted
	CVV        string    `json:"cvv"`    // Hashed
	ExpiryDate time.Time `json:"expiry_date"`
	Status     string    `json:"status"`
	CardType   string    `json:"card_type"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
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

// CreateCard создает новую карту
func (h *CardHandler) CreateCard(w http.ResponseWriter, r *http.Request) {
	var req CreateCardRequest

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

	card, err := h.cardService.CreateCard(context.Background(), accountID)
	if err != nil {
		h.logger.Error("Failed to create card", "account_id", accountID, "error", err.Error())
		WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	response := CardToResponse(card, "****-****-****-XXXX")
	WriteSuccessResponse(w, response)
}

// GetAccountCards получает карты по номеру счета
func (h *CardHandler) GetAccountCards(w http.ResponseWriter, r *http.Request) {
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

	cards, err := h.cardService.GetAccountCards(context.Background(), accountID)
	if err != nil {
		h.logger.Error("Failed to get account cards", "account_id", accountID, "error", err.Error())
		WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	var responses []*CardResponse
	for _, card := range cards {
		responses = append(responses, CardToResponse(card, "****-****-****-XXXX"))
	}

	WriteSuccessResponse(w, responses)
}

// ProcessPayment обрабатывает оплату картой
func (h *CardHandler) ProcessPayment(w http.ResponseWriter, r *http.Request) {
	var req CardPaymentRequest

	if err := ValidateJSON(r, &req); err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	if validationErr := Validate(&req); validationErr != nil {
		WriteErrorResponse(w, http.StatusBadRequest, validationErr)
		return
	}

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

	if err := h.cardService.ProcessPayment(context.Background(), cardID, req.Amount); err != nil {
		h.logger.Error("Failed to process payment", "card_id", cardID, "amount", req.Amount, "error", err.Error())
		WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	h.logger.Info("Payment processed", "card_id", cardID, "amount", req.Amount, "merchant", req.Merchant)
	WriteSuccessResponse(w, map[string]string{"message": "Payment successful"})
}

// Conversion functions
func CardToResponse(card *domain.Card, number string) *CardResponse {
	return &CardResponse{
		ID:         fmt.Sprintf("%d", card.ID),
		AccountID:  fmt.Sprintf("%d", card.AccountID),
		Number:     number,
		CVV:        "***",
		ExpiryDate: card.ExpiryDate,
		Status:     card.Status,
		CardType:   domain.CardTypeLearnBank,
		CreatedAt:  card.CreatedAt,
		UpdatedAt:  card.UpdatedAt,
	}
}
