package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

// Config конфигурация логгера
type Config struct {
	Level  string // debug, info, warn, error
	Format string // json, text
	Output string // stdout, stderr, file path
}

// New создает новый логгер с заданной конфигурацией
func New(cfg Config) *slog.Logger {
	var level slog.Level
	switch strings.ToLower(cfg.Level) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	var output io.Writer
	switch cfg.Output {
	case "stdout":
		output = os.Stdout
	case "stderr":
		output = os.Stderr
	default:
		// Если указан путь к файлу, пытаемся его открыть
		if cfg.Output != "" {
			file, err := os.OpenFile(cfg.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
			if err == nil {
				output = file
			} else {
				output = os.Stdout
			}
		} else {
			output = os.Stdout
		}
	}

	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
	}

	switch strings.ToLower(cfg.Format) {
	case "json":
		handler = slog.NewJSONHandler(output, opts)
	case "text":
		handler = slog.NewTextHandler(output, opts)
	default:
		handler = slog.NewTextHandler(output, opts)
	}

	return slog.New(handler)
}

// NewDefault создает логгер с настройками по умолчанию
func NewDefault() *slog.Logger {
	return New(Config{
		Level:  "info",
		Format: "text",
		Output: "stdout",
	})
}

// WithFields добавляет поля к логгеру
func WithFields(logger *slog.Logger, fields map[string]interface{}) *slog.Logger {
	args := make([]interface{}, 0, len(fields)*2)
	for key, value := range fields {
		args = append(args, key, value)
	}
	return logger.With(args...)
}

// WithService добавляет поле service к логгеру
func WithService(logger *slog.Logger, service string) *slog.Logger {
	return logger.With("service", service)
}

// WithUserID добавляет поле user_id к логгеру
func WithUserID(logger *slog.Logger, userID int) *slog.Logger {
	return logger.With("user_id", userID)
}

// WithRequestID добавляет поле request_id к логгеру
func WithRequestID(logger *slog.Logger, requestID string) *slog.Logger {
	return logger.With("request_id", requestID)
}

// LogError логирует ошибку с контекстом
func LogError(logger *slog.Logger, err error, message string, fields ...interface{}) {
	args := append([]interface{}{"error", err.Error()}, fields...)
	logger.Error(message, args...)
}

// LogOperation логирует операцию с метриками
func LogOperation(logger *slog.Logger, operation string, success bool, duration int64, fields ...interface{}) {
	args := append([]interface{}{
		"operation", operation,
		"success", success,
		"duration_ms", duration,
	}, fields...)

	if success {
		logger.Info("Operation completed", args...)
	} else {
		logger.Error("Operation failed", args...)
	}
}

// LogUserAction логирует действие пользователя
func LogUserAction(logger *slog.Logger, userID int, action string, details map[string]interface{}) {
	args := []interface{}{
		"user_id", userID,
		"action", action,
	}

	for key, value := range details {
		args = append(args, key, value)
	}

	logger.Info("User action", args...)
}

// LogSecurityEvent логирует событие безопасности
func LogSecurityEvent(logger *slog.Logger, event string, severity string, details map[string]interface{}) {
	args := []interface{}{
		"security_event", event,
		"severity", severity,
	}

	for key, value := range details {
		args = append(args, key, value)
	}

	switch severity {
	case "critical", "high":
		logger.Error("Security event", args...)
	case "medium":
		logger.Warn("Security event", args...)
	default:
		logger.Info("Security event", args...)
	}
}
