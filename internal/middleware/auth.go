package middleware

import (
	"net/http"
	"strings"

	"github.com/vterdunov/learn-bank-app/internal/utils"
)

// AuthMiddleware проверяет JWT токен и добавляет данные пользователя в контекст
func AuthMiddleware(jwtManager *utils.JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Получаем Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			// Проверяем формат Bearer token
			if !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
			}

			// Извлекаем токен
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == "" {
				http.Error(w, "Token not provided", http.StatusUnauthorized)
				return
			}

			// Валидируем JWT токен
			claims, err := jwtManager.ValidateToken(tokenString)
			if err != nil {
				switch err {
				case utils.ErrTokenExpired:
					http.Error(w, "Token expired", http.StatusUnauthorized)
				case utils.ErrInvalidSignature:
					http.Error(w, "Invalid token signature", http.StatusUnauthorized)
				case utils.ErrInvalidToken, utils.ErrInvalidClaims:
					http.Error(w, "Invalid token", http.StatusUnauthorized)
				default:
					http.Error(w, "Token validation failed", http.StatusUnauthorized)
				}
				return
			}

			// Добавляем данные пользователя в контекст
			ctx := utils.SetUserDataInContext(r.Context(), claims.UserID, claims.Username, claims.Email)

			// Передаем запрос дальше с обновленным контекстом
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuthMiddleware проверяет JWT токен если он предоставлен, но не требует его
func OptionalAuthMiddleware(jwtManager *utils.JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")

			// Если заголовок отсутствует, продолжаем без авторизации
			if authHeader == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Если заголовок есть, валидируем токен
			if strings.HasPrefix(authHeader, "Bearer ") {
				tokenString := strings.TrimPrefix(authHeader, "Bearer ")
				if tokenString != "" {
					claims, err := jwtManager.ValidateToken(tokenString)
					if err == nil {
						// Добавляем данные пользователя в контекст только при успешной валидации
						ctx := utils.SetUserDataInContext(r.Context(), claims.UserID, claims.Username, claims.Email)
						r = r.WithContext(ctx)
					}
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireOwnershipMiddleware проверяет что пользователь является владельцем ресурса
func RequireOwnershipMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Проверяем что пользователь авторизован
			if err := utils.RequireAuth(r.Context()); err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Middleware только проверяет авторизацию, конкретная проверка владения
			// должна выполняться в handlers для каждого ресурса отдельно
			next.ServeHTTP(w, r)
		})
	}
}

// AdminMiddleware проверяет права администратора (заготовка для будущего)
func AdminMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Проверяем авторизацию
			if err := utils.RequireAuth(r.Context()); err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// TODO: Добавить проверку роли администратора когда будет система ролей
			// Пока что пропускаем всех авторизованных пользователей

			next.ServeHTTP(w, r)
		})
	}
}
