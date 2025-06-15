package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vterdunov/learn-bank-app/internal/config"
	"github.com/vterdunov/learn-bank-app/internal/database"
	"github.com/vterdunov/learn-bank-app/internal/domain"
	"github.com/vterdunov/learn-bank-app/internal/repository"
	"github.com/vterdunov/learn-bank-app/internal/router"
	"github.com/vterdunov/learn-bank-app/internal/service"
	"github.com/vterdunov/learn-bank-app/internal/utils"
)

func main() {
	// Настройка логирования
	lg := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(lg)

	// Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Инициализация JWT
	utils.InitJWT(cfg.JWT.Secret)

	// Создание контекста с отменой
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Подключение к базе данных
	dbCfg := database.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		Database: cfg.Database.Database,
		Username: cfg.Database.Username,
		Password: cfg.Database.Password,
		SSLMode:  cfg.Database.SSLMode,
	}

	db, err := database.NewConnection(ctx, dbCfg)
	if err != nil {
		slog.Error("Failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()

	// Применение миграций
	if err := db.ApplyMigrations(ctx); err != nil {
		slog.Error("Failed to apply migrations", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Инициализация репозиториев
	userRepo := repository.NewUserRepository(db.Pool)
	accountRepo := repository.NewAccountRepository(db.Pool)
	cardRepo := repository.NewCardRepository(db.Pool)
	transactionRepo := repository.NewTransactionRepository(db.Pool)
	creditRepo := repository.NewCreditRepository(db.Pool)
	paymentScheduleRepo := repository.NewPaymentScheduleRepository(db.Pool)

	// Инициализация внешних сервисов
	cbrService := service.NewCBRService(cfg, lg)

	// Инициализация access control
	accessControl := domain.NewAccessControlDomain(accountRepo, cardRepo, creditRepo)

	// Инициализация основных сервисов
	authService := service.NewAuthService(userRepo, lg)
	accountService := service.NewAccountService(accountRepo, transactionRepo, accessControl, lg)
	cardService := service.NewCardService(cardRepo, accountRepo, transactionRepo, lg)
	creditService := service.NewCreditService(creditRepo, paymentScheduleRepo, accountRepo, transactionRepo, cbrService, lg)
	analyticsService := service.NewAnalyticsService(accountRepo, transactionRepo, creditRepo)

	// Инициализация email сервиса для шедулера
	emailService := service.NewEmailService(cfg, lg)

	// Инициализация шедулера
	scheduler := service.NewSchedulerService(cfg, creditRepo, paymentScheduleRepo, accountRepo, transactionRepo, userRepo, emailService, lg)

	// Инициализация роутера со всеми сервисами
	routerConfig := router.Config{
		Logger:    lg,
		JWTSecret: cfg.JWT.Secret,
		Services: &router.Services{
			Auth:      authService,
			Account:   accountService,
			Card:      cardService,
			Credit:    creditService,
			Analytics: analyticsService,
			CBR:       cbrService,
		},
	}

	appRouter := router.New(routerConfig)

	// Запуск шедулера
	if err := scheduler.Start(ctx); err != nil {
		slog.Error("Failed to start scheduler", slog.String("error", err.Error()))
		os.Exit(1)
	}
	slog.Info("Scheduler started")

	// Настройка HTTP сервера
	server := &http.Server{
		Addr:         cfg.Server.Host + ":" + cfg.Server.Port,
		Handler:      appRouter.Handler(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		slog.Info("Shutting down server and scheduler...")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		// Останавливаем шедулер
		scheduler.Stop()
		slog.Info("Scheduler stopped")

		// Останавливаем HTTP сервер
		if err := server.Shutdown(shutdownCtx); err != nil {
			slog.Error("Server shutdown error", slog.String("error", err.Error()))
		}

		cancel()
	}()

	// Запуск сервера
	slog.Info("Starting server",
		slog.String("address", server.Addr),
		slog.String("database", cfg.Database.Database),
	)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("Server failed to start", slog.String("error", err.Error()))
		os.Exit(1)
	}

	slog.Info("Server stopped")
}
