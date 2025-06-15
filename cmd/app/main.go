package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vterdunov/learn-bank-app/internal/config"
	"github.com/vterdunov/learn-bank-app/internal/database"
	"github.com/vterdunov/learn-bank-app/internal/utils"
)

func main() {
	// Настройка логирования
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

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

	// Настройка HTTP сервера
	mux := http.NewServeMux()

	// Простой health check эндпоинт
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"healthy"}`)
	})

	server := &http.Server{
		Addr:         cfg.Server.Host + ":" + cfg.Server.Port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		slog.Info("Shutting down server...")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

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
