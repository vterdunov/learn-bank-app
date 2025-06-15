package router

import (
	"log/slog"
	"net/http"

	"github.com/vterdunov/learn-bank-app/internal/handlers"
	"github.com/vterdunov/learn-bank-app/internal/middleware"
	"github.com/vterdunov/learn-bank-app/internal/service"
)

// Router содержит все маршруты приложения
type Router struct {
	mux       *http.ServeMux
	logger    *slog.Logger
	handlers  *Handlers
	jwtSecret string
}

// Handlers содержит все обработчики
type Handlers struct {
	Auth      *handlers.AuthHandler
	Account   *handlers.AccountHandler
	Card      *handlers.CardHandler
	Credit    *handlers.CreditHandler
	Analytics *handlers.AnalyticsHandler
}

// Config содержит конфигурацию для роутера
type Config struct {
	Logger    *slog.Logger
	Services  *Services
	JWTSecret string
}

// Services содержит все сервисы
type Services struct {
	Auth      service.AuthService
	Account   service.AccountService
	Card      service.CardService
	Credit    service.CreditService
	Analytics service.AnalyticsService
}

// New создает новый роутер
func New(config Config) *Router {
	// Создаем все обработчики
	h := &Handlers{
		Auth:      handlers.NewAuthHandler(config.Services.Auth, config.Logger),
		Account:   handlers.NewAccountHandler(config.Services.Account, config.Logger),
		Card:      handlers.NewCardHandler(config.Services.Card, config.Logger),
		Credit:    handlers.NewCreditHandler(config.Services.Credit, config.Logger),
		Analytics: handlers.NewAnalyticsHandler(config.Services.Analytics, config.Logger),
	}

	router := &Router{
		mux:       http.NewServeMux(),
		logger:    config.Logger,
		handlers:  h,
		jwtSecret: config.JWTSecret,
	}

	router.setupRoutes()
	return router
}

// setupRoutes настраивает все маршруты
func (r *Router) setupRoutes() {
	// Middleware для всех запросов
	commonMiddleware := middleware.Chain(
		middleware.LoggingMiddleware(),
		middleware.RequestIDMiddleware(),
	)

	// Public routes (без аутентификации)
	r.mux.Handle("POST /api/v1/auth/register", commonMiddleware(http.HandlerFunc(r.handlers.Auth.Register)))
	r.mux.Handle("POST /api/v1/auth/login", commonMiddleware(http.HandlerFunc(r.handlers.Auth.Login)))

	// Protected routes (с аутентификацией)
	authMiddleware := middleware.Chain(
		middleware.LoggingMiddleware(),
		middleware.RequestIDMiddleware(),
		middleware.AuthMiddleware(r.jwtSecret),
	)

	// Account endpoints
	r.mux.Handle("POST /api/v1/accounts", authMiddleware(http.HandlerFunc(r.handlers.Account.CreateAccount)))
	r.mux.Handle("GET /api/v1/accounts", authMiddleware(http.HandlerFunc(r.handlers.Account.GetUserAccounts)))
	r.mux.Handle("POST /api/v1/accounts/{id}/deposit", authMiddleware(http.HandlerFunc(r.handlers.Account.Deposit)))
	r.mux.Handle("POST /api/v1/accounts/{id}/withdraw", authMiddleware(http.HandlerFunc(r.handlers.Account.Withdraw)))
	r.mux.Handle("POST /api/v1/transfer", authMiddleware(http.HandlerFunc(r.handlers.Account.Transfer)))

	// Card endpoints
	r.mux.Handle("POST /api/v1/cards", authMiddleware(http.HandlerFunc(r.handlers.Card.CreateCard)))
	r.mux.Handle("GET /api/v1/accounts/{accountId}/cards", authMiddleware(http.HandlerFunc(r.handlers.Card.GetAccountCards)))
	r.mux.Handle("POST /api/v1/cards/{id}/payment", authMiddleware(http.HandlerFunc(r.handlers.Card.CardPayment)))

	// Credit endpoints
	r.mux.Handle("POST /api/v1/credits", authMiddleware(http.HandlerFunc(r.handlers.Credit.CreateCredit)))
	r.mux.Handle("GET /api/v1/credits/{id}/schedule", authMiddleware(http.HandlerFunc(r.handlers.Credit.GetCreditSchedule)))

	// Analytics endpoints
	r.mux.Handle("GET /api/v1/analytics/monthly", authMiddleware(http.HandlerFunc(r.handlers.Analytics.GetMonthlyStats)))
	r.mux.Handle("GET /api/v1/analytics/credit-load", authMiddleware(http.HandlerFunc(r.handlers.Analytics.GetCreditLoad)))
	r.mux.Handle("POST /api/v1/analytics/balance-prediction", authMiddleware(http.HandlerFunc(r.handlers.Analytics.PredictBalance)))

	// Health check endpoint
	r.mux.Handle("GET /health", commonMiddleware(http.HandlerFunc(r.healthCheck)))
}

// healthCheck простая проверка здоровья сервиса
func (r *Router) healthCheck(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok","service":"learn-bank-app"}`))
}

// Handler возвращает основной HTTP handler
func (r *Router) Handler() http.Handler {
	return r.mux
}
