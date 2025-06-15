package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"github.com/vterdunov/learn-bank-app/pkg/logger"
)

// contextKey тип для ключей контекста
type contextKey string

const (
	// UserIDKey ключ для ID пользователя в контексте
	UserIDKey contextKey = "userID"
	// RequestIDKey ключ для ID запроса в контексте
	RequestIDKey contextKey = "requestID"
)

// AuthMiddleware middleware для проверки JWT токенов
func AuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	log := logger.NewDefault()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Получаем заголовок Authorization
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				log.Warn("Missing Authorization header",
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.String("remote_addr", r.RemoteAddr),
				)
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			// Проверяем формат Bearer token
			if !strings.HasPrefix(authHeader, "Bearer ") {
				log.Warn("Invalid Authorization header format",
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.String("remote_addr", r.RemoteAddr),
				)
				http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
				return
			}

			// Извлекаем токен
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")

			// Парсим и валидируем токен
			claims := &jwt.RegisteredClaims{}
			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				// Проверяем алгоритм подписи
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(jwtSecret), nil
			})

			if err != nil {
				log.Warn("Invalid JWT token",
					slog.String("error", err.Error()),
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.String("remote_addr", r.RemoteAddr),
				)
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			if !token.Valid {
				log.Warn("Invalid JWT token",
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.String("remote_addr", r.RemoteAddr),
				)
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// Извлекаем userID из токена
			userIDStr := claims.Subject
			if userIDStr == "" {
				log.Warn("Missing user ID in token",
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.String("remote_addr", r.RemoteAddr),
				)
				http.Error(w, "Invalid token claims", http.StatusUnauthorized)
				return
			}

			userID, err := strconv.Atoi(userIDStr)
			if err != nil {
				log.Warn("Invalid user ID in token",
					slog.String("error", err.Error()),
					slog.String("user_id", userIDStr),
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
				)
				http.Error(w, "Invalid token claims", http.StatusUnauthorized)
				return
			}

			// Добавляем userID в контекст
			ctx := context.WithValue(r.Context(), UserIDKey, userID)

			log.Info("User authenticated",
				slog.Int("user_id", userID),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
			)

			// Передаем управление следующему handler
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserIDFromContext извлекает ID пользователя из контекста
func GetUserIDFromContext(ctx context.Context) (int, bool) {
	userID, ok := ctx.Value(UserIDKey).(int)
	return userID, ok
}

// GetRequestIDFromContext извлекает ID запроса из контекста
func GetRequestIDFromContext(ctx context.Context) (string, bool) {
	requestID, ok := ctx.Value(RequestIDKey).(string)
	return requestID, ok
}

// RequireAuth helper для проверки аутентификации в handlers
func RequireAuth(ctx context.Context) (int, error) {
	userID, ok := GetUserIDFromContext(ctx)
	if !ok {
		return 0, ErrUnauthorized
	}
	return userID, nil
}

// ErrUnauthorized ошибка неавторизованного доступа
var ErrUnauthorized = &AuthError{Message: "unauthorized access"}

// AuthError кастомный тип ошибки аутентификации
type AuthError struct {
	Message string
}

func (e *AuthError) Error() string {
	return e.Message
}
