package utils

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWT константы
const (
	// DefaultTokenExpiry время жизни токена по умолчанию
	DefaultTokenExpiry = 24 * time.Hour
	// RefreshTokenExpiry время жизни refresh токена
	RefreshTokenExpiry = 7 * 24 * time.Hour
)

// JWT ошибки
var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrTokenExpired     = errors.New("token expired")
	ErrTokenNotFound    = errors.New("token not found")
	ErrInvalidSignature = errors.New("invalid token signature")
	ErrInvalidClaims    = errors.New("invalid token claims")
)

// JWTClaims кастомные claims для JWT
type JWTClaims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	jwt.RegisteredClaims
}

// JWTManager управляет JWT токенами
type JWTManager struct {
	secretKey     []byte
	tokenExpiry   time.Duration
	refreshExpiry time.Duration
}

// NewJWTManager создает новый менеджер JWT
func NewJWTManager(secretKey string) *JWTManager {
	return &JWTManager{
		secretKey:     []byte(secretKey),
		tokenExpiry:   DefaultTokenExpiry,
		refreshExpiry: RefreshTokenExpiry,
	}
}

// GenerateToken генерирует JWT токен для пользователя
func (j *JWTManager) GenerateToken(userID int, username, email string) (string, error) {
	claims := &JWTClaims{
		UserID:   userID,
		Username: username,
		Email:    email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.Itoa(userID),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.tokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "learn-bank-app",
			Audience:  []string{"learn-bank-app-users"},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(j.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// GenerateRefreshToken генерирует refresh токен
func (j *JWTManager) GenerateRefreshToken(userID int) (string, error) {
	claims := &jwt.RegisteredClaims{
		Subject:   strconv.Itoa(userID),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.refreshExpiry)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
		Issuer:    "learn-bank-app",
		Audience:  []string{"learn-bank-app-refresh"},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(j.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken проверяет и парсит JWT токен
func (j *JWTManager) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Проверяем метод подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		if errors.Is(err, jwt.ErrSignatureInvalid) {
			return nil, ErrInvalidSignature
		}
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	// Дополнительная проверка claims
	if claims.UserID <= 0 {
		return nil, ErrInvalidClaims
	}

	return claims, nil
}

// ValidateRefreshToken проверяет refresh токен
func (j *JWTManager) ValidateRefreshToken(tokenString string) (int, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return 0, ErrTokenExpired
		}
		if errors.Is(err, jwt.ErrSignatureInvalid) {
			return 0, ErrInvalidSignature
		}
		return 0, fmt.Errorf("failed to parse refresh token: %w", err)
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return 0, ErrInvalidToken
	}

	userID, err := strconv.Atoi(claims.Subject)
	if err != nil || userID <= 0 {
		return 0, ErrInvalidClaims
	}

	return userID, nil
}

// ExtractUserIDFromToken извлекает ID пользователя из токена без полной валидации
func (j *JWTManager) ExtractUserIDFromToken(tokenString string) (int, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return j.secretKey, nil
	}, jwt.WithoutClaimsValidation())

	if err != nil {
		return 0, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return 0, ErrInvalidClaims
	}

	return claims.UserID, nil
}

// IsTokenExpired проверяет истек ли токен
func (j *JWTManager) IsTokenExpired(tokenString string) bool {
	_, err := j.ValidateToken(tokenString)
	return errors.Is(err, ErrTokenExpired)
}

// GetTokenRemainingTime возвращает оставшееся время жизни токена
func (j *JWTManager) GetTokenRemainingTime(tokenString string) (time.Duration, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return 0, err
	}

	if claims.ExpiresAt == nil {
		return 0, errors.New("token has no expiration time")
	}

	remaining := time.Until(claims.ExpiresAt.Time)
	if remaining < 0 {
		return 0, ErrTokenExpired
	}

	return remaining, nil
}

// RefreshAccessToken обновляет access токен используя refresh токен
func (j *JWTManager) RefreshAccessToken(refreshToken string, username, email string) (string, error) {
	userID, err := j.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", fmt.Errorf("invalid refresh token: %w", err)
	}

	return j.GenerateToken(userID, username, email)
}

// SetTokenExpiry устанавливает время жизни токена
func (j *JWTManager) SetTokenExpiry(expiry time.Duration) {
	j.tokenExpiry = expiry
}

// SetRefreshExpiry устанавливает время жизни refresh токена
func (j *JWTManager) SetRefreshExpiry(expiry time.Duration) {
	j.refreshExpiry = expiry
}
