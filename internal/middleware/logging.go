package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/vterdunov/learn-bank-app/pkg/logger"
)

// ResponseWriter wrapper для захвата статус кода
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.written += n
	return n, err
}

// LoggingMiddleware логирует все HTTP запросы
func LoggingMiddleware(lg *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Создаем wrapped response writer для отслеживания статуса
			wrapped := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Выполняем запрос
			next.ServeHTTP(wrapped, r)

			// Вычисляем время выполнения
			duration := time.Since(start)

			// Логируем запрос
			logger.LogOperation(
				lg,
				"http_request",
				wrapped.statusCode < 400,
				duration.Milliseconds(),
				"method", r.Method,
				"uri", r.RequestURI,
				"remote_addr", r.RemoteAddr,
				"user_agent", r.UserAgent(),
				"status_code", wrapped.statusCode,
			)

			// Логируем ошибки отдельно
			if wrapped.statusCode >= 400 {
				severity := "medium"
				if wrapped.statusCode >= 500 {
					severity = "high"
				}

				logger.LogSecurityEvent(
					lg,
					"http_error",
					severity,
					map[string]interface{}{
						"method":      r.Method,
						"uri":         r.RequestURI,
						"status_code": wrapped.statusCode,
						"remote_addr": r.RemoteAddr,
						"user_agent":  r.UserAgent(),
					},
				)
			}
		})
	}
}

// ErrorLoggingMiddleware логирует ошибки с дополнительным контекстом
func ErrorLoggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wrapped := newResponseWriter(w)

			// Добавляем контекст запроса
			ctx := r.Context()

			defer func() {
				// Логируем ошибки (статус коды >= 400)
				if wrapped.statusCode >= 400 {
					logger.Error("HTTP Error",
						slog.String("method", r.Method),
						slog.String("path", r.URL.Path),
						slog.String("remote_addr", r.RemoteAddr),
						slog.Int("status_code", wrapped.statusCode),
						slog.Any("context", ctx),
					)
				}
			}()

			next.ServeHTTP(wrapped, r)
		})
	}
}
