package handlers

import (
	"fmt"
	"net/http"

	"github.com/vterdunov/learn-bank-app/internal/middleware"
)

// GetUserIDFromRequest извлекает userID из контекста запроса
func GetUserIDFromRequest(r *http.Request) (int, error) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		return 0, fmt.Errorf("user not authenticated")
	}
	return userID, nil
}

// WriteAccessDeniedResponse возвращает 403 Forbidden
func WriteAccessDeniedResponse(w http.ResponseWriter, err error) {
	WriteErrorResponse(w, http.StatusForbidden, err)
}
