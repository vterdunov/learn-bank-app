package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/vterdunov/learn-bank-app/internal/domain"
	"github.com/vterdunov/learn-bank-app/internal/repository"
	"github.com/vterdunov/learn-bank-app/internal/utils"
	"github.com/vterdunov/learn-bank-app/pkg/logger"
)

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
)

// authService реализует интерфейс AuthService
type authService struct {
	userRepo repository.UserRepository
	logger   *slog.Logger
}

// NewAuthService создает новый экземпляр сервиса аутентификации
func NewAuthService(userRepo repository.UserRepository, lg *slog.Logger) AuthService {
	return &authService{
		userRepo: userRepo,
		logger:   logger.WithService(lg, "auth_service"),
	}
}

// Register регистрирует нового пользователя
func (s *authService) Register(ctx context.Context, req RegisterRequest) (*domain.User, error) {
	start := time.Now()

	// Валидация входных данных
	if err := utils.ValidateEmail(req.Email); err != nil {
		logger.LogError(s.logger, err, "Invalid email format during registration", "email", req.Email)
		return nil, fmt.Errorf("invalid email format: %w", err)
	}

	if err := utils.ValidateUsername(req.Username); err != nil {
		logger.LogError(s.logger, err, "Invalid username during registration", "username", req.Username)
		return nil, fmt.Errorf("invalid username: %w", err)
	}

	if err := utils.ValidatePassword(req.Password); err != nil {
		logger.LogError(s.logger, err, "Invalid password during registration")
		return nil, fmt.Errorf("invalid password: %w", err)
	}

	// Проверка уникальности email
	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		logger.LogSecurityEvent(s.logger, "duplicate_email_registration", "medium", map[string]interface{}{
			"email": req.Email,
		})
		return nil, ErrUserAlreadyExists
	}

	// Проверка уникальности username
	existingUser, err = s.userRepo.GetByUsername(ctx, req.Username)
	if err == nil && existingUser != nil {
		logger.LogSecurityEvent(s.logger, "duplicate_username_registration", "medium", map[string]interface{}{
			"username": req.Username,
		})
		return nil, ErrUserAlreadyExists
	}

	// Хеширование пароля
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		logger.LogError(s.logger, err, "Failed to hash password during registration")
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Создание пользователя
	user := &domain.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		logger.LogError(s.logger, err, "Failed to create user in database")
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Убираем пароль из ответа
	user.PasswordHash = ""

	// Логируем успешную регистрацию
	logger.LogOperation(s.logger, "user_registration", true, time.Since(start).Milliseconds(),
		"user_id", user.ID,
		"username", user.Username,
		"email", user.Email,
	)

	logger.LogUserAction(s.logger, user.ID, "user_registered", map[string]interface{}{
		"username": user.Username,
		"email":    user.Email,
	})

	return user, nil
}

// Login выполняет аутентификацию пользователя и возвращает JWT токен
func (s *authService) Login(ctx context.Context, req LoginRequest) (string, error) {
	start := time.Now()

	// Валидация email
	if err := utils.ValidateEmail(req.Email); err != nil {
		logger.LogError(s.logger, err, "Invalid email format during login", "email", req.Email)
		return "", fmt.Errorf("invalid email format: %w", err)
	}

	// Получение пользователя по email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		logger.LogSecurityEvent(s.logger, "login_attempt_unknown_email", "medium", map[string]interface{}{
			"email": req.Email,
		})
		return "", ErrInvalidCredentials
	}

	// Проверка пароля
	if err := utils.VerifyPassword(user.PasswordHash, req.Password); err != nil {
		logger.LogSecurityEvent(s.logger, "login_attempt_invalid_password", "high", map[string]interface{}{
			"email":   req.Email,
			"user_id": user.ID,
		})
		return "", ErrInvalidCredentials
	}

	// Генерация JWT токена
	token, err := utils.GenerateJWT(strconv.Itoa(user.ID))
	if err != nil {
		logger.LogError(s.logger, err, "Failed to generate JWT token", "user_id", user.ID)
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	// Логируем успешный вход
	logger.LogOperation(s.logger, "user_login", true, time.Since(start).Milliseconds(),
		"user_id", user.ID,
		"email", user.Email,
	)

	logger.LogUserAction(s.logger, user.ID, "user_logged_in", map[string]interface{}{
		"email": user.Email,
	})

	return token, nil
}

// ValidateToken проверяет валидность JWT токена и возвращает пользователя
func (s *authService) ValidateToken(ctx context.Context, token string) (*domain.User, error) {
	start := time.Now()

	// Валидация JWT токена
	userIDStr, err := utils.ValidateJWT(token)
	if err != nil {
		logger.LogSecurityEvent(s.logger, "invalid_token_validation", "high", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, ErrInvalidToken
	}

	// Преобразование ID пользователя
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		logger.LogError(s.logger, err, "Failed to parse user ID from token", "user_id_str", userIDStr)
		return nil, ErrInvalidToken
	}

	// Получение пользователя по ID
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		logger.LogSecurityEvent(s.logger, "token_validation_user_not_found", "high", map[string]interface{}{
			"user_id": userID,
		})
		return nil, ErrInvalidToken
	}

	// Убираем пароль из ответа
	user.PasswordHash = ""

	// Логируем успешную валидацию токена
	logger.LogOperation(s.logger, "token_validation", true, time.Since(start).Milliseconds(),
		"user_id", user.ID,
	)

	return user, nil
}
