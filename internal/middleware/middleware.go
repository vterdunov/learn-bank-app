package middleware

import (
	"log/slog"
	"net/http"
)

// Chain helper для объединения middleware
func Chain(middlewares ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}

// RecoveryMiddleware восстанавливается после паники
func RecoveryMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("Panic recovered",
						slog.Any("error", err),
						slog.String("method", r.Method),
						slog.String("path", r.URL.Path),
						slog.String("remote_addr", r.RemoteAddr),
					)

					// Возвращаем 500 ошибку
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
