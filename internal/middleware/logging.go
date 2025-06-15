package middleware

import (
	"log/slog"
	"net/http"
	"time"
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

// LoggingMiddleware логирует HTTP запросы
func LoggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Создаем wrapper для response writer
			wrapped := newResponseWriter(w)

			// Выполняем запрос
			next.ServeHTTP(wrapped, r)

			// Вычисляем время выполнения
			duration := time.Since(start)

			// Логируем информацию о запросе
			logger.Info("HTTP Request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("query", r.URL.RawQuery),
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("user_agent", r.UserAgent()),
				slog.Int("status_code", wrapped.statusCode),
				slog.Int("response_size", wrapped.written),
				slog.Duration("duration", duration),
			)
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
