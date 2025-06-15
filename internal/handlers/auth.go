package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/vterdunov/learn-bank-app/internal/domain"
	"github.com/vterdunov/learn-bank-app/internal/service"
	"github.com/vterdunov/learn-bank-app/pkg/logger"
)

// Auth Request DTOs
type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// Auth Response DTOs
type AuthResponse struct {
	Token     string    `json:"token"`
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	ExpiresAt time.Time `json:"expires_at"`
}

// AuthHandler обрабатывает запросы аутентификации
type AuthHandler struct {
	authService service.AuthService
	logger      *slog.Logger
}

func NewAuthHandler(authService service.AuthService, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
	}
}

// Register обрабатывает регистрацию нового пользователя
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest

	// Валидация JSON
	if err := ValidateJSON(r, &req); err != nil {
		logger.LogSecurityEvent(h.logger, "invalid_registration_request", "medium", map[string]interface{}{
			"error": err.Error(),
			"ip":    r.RemoteAddr,
		})
		WriteErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	// Валидация полей
	if validationErr := Validate(&req); validationErr != nil {
		logger.LogSecurityEvent(h.logger, "registration_validation_failed", "medium", map[string]interface{}{
			"email": req.Email,
			"ip":    r.RemoteAddr,
		})
		WriteErrorResponse(w, http.StatusBadRequest, validationErr)
		return
	}

	// Создание пользователя
	serviceReq := service.RegisterRequest{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	}

	user, err := h.authService.Register(context.Background(), serviceReq)
	if err != nil {
		logger.LogSecurityEvent(h.logger, "registration_failed", "high", map[string]interface{}{
			"email": req.Email,
			"error": err.Error(),
			"ip":    r.RemoteAddr,
		})

		// Определяем статус код на основе ошибки
		statusCode := http.StatusInternalServerError
		if err.Error() == "email already exists" || err.Error() == "username already exists" {
			statusCode = http.StatusConflict
		}

		WriteErrorResponse(w, statusCode, err)
		return
	}

	// Логирование успешной регистрации
	logger.LogUserAction(h.logger, user.ID, "user_registered", map[string]interface{}{
		"email": user.Email,
		"ip":    r.RemoteAddr,
	})

	// Успешный ответ - без JWT токена, так как в интерфейсе нет метода GenerateToken
	response := AuthResponse{
		Token:     "", // TODO: Implement JWT generation in auth service
		UserID:    strconv.Itoa(user.ID),
		Email:     user.Email,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	WriteSuccessResponse(w, response)
}

// Login обрабатывает вход пользователя в систему
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest

	// Валидация JSON
	if err := ValidateJSON(r, &req); err != nil {
		logger.LogSecurityEvent(h.logger, "invalid_login_request", "medium", map[string]interface{}{
			"error": err.Error(),
			"ip":    r.RemoteAddr,
		})
		WriteErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	// Валидация полей
	if validationErr := Validate(&req); validationErr != nil {
		logger.LogSecurityEvent(h.logger, "login_validation_failed", "medium", map[string]interface{}{
			"email": req.Email,
			"ip":    r.RemoteAddr,
		})
		WriteErrorResponse(w, http.StatusBadRequest, validationErr)
		return
	}

	// Аутентификация пользователя
	serviceReq := service.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	}

	token, err := h.authService.Login(context.Background(), serviceReq)
	if err != nil {
		logger.LogSecurityEvent(h.logger, "login_failed", "high", map[string]interface{}{
			"email": req.Email,
			"error": err.Error(),
			"ip":    r.RemoteAddr,
		})
		WriteErrorResponse(w, http.StatusUnauthorized, err)
		return
	}

	// Логирование успешного входа
	h.logger.Info("User logged in", "email", req.Email, "ip", r.RemoteAddr)

	// Успешный ответ
	response := AuthResponse{
		Token:     token,
		UserID:    "", // TODO: Extract user ID from token or change service interface
		Email:     req.Email,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	WriteSuccessResponse(w, response)
}

// Conversion functions
func (r *RegisterRequest) ToDomain() *domain.User {
	return &domain.User{
		Username: r.Username,
		Email:    r.Email,
	}
}
