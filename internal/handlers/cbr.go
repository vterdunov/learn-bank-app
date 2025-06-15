package handlers

import (
	"log/slog"
	"net/http"

	"github.com/vterdunov/learn-bank-app/internal/service"
)

// CBR Response DTOs
type CBRRateResponse struct {
	Rate   float64 `json:"rate"`
	Source string  `json:"source"`
}

// CBRHandler обрабатывает запросы к ЦБ РФ
type CBRHandler struct {
	cbrService service.CBRService
	logger     *slog.Logger
}

func NewCBRHandler(cbrService service.CBRService, logger *slog.Logger) *CBRHandler {
	return &CBRHandler{
		cbrService: cbrService,
		logger:     logger,
	}
}

// GetCBRRate получает ключевую ставку ЦБ РФ
func (h *CBRHandler) GetCBRRate(w http.ResponseWriter, r *http.Request) {
	rate, err := h.cbrService.GetKeyRate(r.Context())
	if err != nil {
		h.logger.Error("Failed to get CBR rate", "error", err.Error())
		WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	response := &CBRRateResponse{
		Rate:   rate,
		Source: "Central Bank of Russia",
	}

	h.logger.Info("CBR rate retrieved successfully", "rate", rate)
	WriteSuccessResponse(w, response)
}
