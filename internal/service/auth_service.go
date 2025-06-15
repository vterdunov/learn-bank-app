package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/vterdunov/learn-bank-app/internal/models"
	"github.com/vterdunov/learn-bank-app/internal/repository"
	"github.com/vterdunov/learn-bank-app/internal/utils"
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
func NewAuthService(userRepo repository.UserRepository, logger *slog.Logger) AuthService {
	return &authService{
		userRepo: userRepo,
		logger:   logger,
	}
}

// Register регистрирует нового пользователя
func (s *authService) Register(ctx context.Context, req RegisterRequest) (*models.User, error) {
	// Валидация входных данных
	if err := utils.ValidateEmail(req.Email); err != nil {
		s.logger.Warn("Invalid email format", "email", req.Email, "error", err)
		return nil, fmt.Errorf("invalid email format: %w", err)
	}

	if err := utils.ValidateUsername(req.Username); err != nil {
		s.logger.Warn("Invalid username", "username", req.Username, "error", err)
		return nil, fmt.Errorf("invalid username: %w", err)
	}

	if err := utils.ValidatePassword(req.Password); err != nil {
		s.logger.Warn("Invalid password", "error", err)
		return nil, fmt.Errorf("invalid password: %w", err)
	}

	// Проверка уникальности email
	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		s.logger.Warn("User with email already exists", "email", req.Email)
		return nil, ErrUserAlreadyExists
	}

	// Проверка уникальности username
	existingUser, err = s.userRepo.GetByUsername(ctx, req.Username)
	if err == nil && existingUser != nil {
		s.logger.Warn("User with username already exists", "username", req.Username)
		return nil, ErrUserAlreadyExists
	}

	// Хеширование пароля
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		s.logger.Error("Failed to hash password", "error", err)
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Создание пользователя
	user := &models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		s.logger.Error("Failed to create user", "error", err)
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Убираем пароль из ответа
	user.PasswordHash = ""

	s.logger.Info("User registered successfully", "user_id", user.ID, "username", user.Username, "email", user.Email)
	return user, nil
}

// Login выполняет аутентификацию пользователя и возвращает JWT токен
func (s *authService) Login(ctx context.Context, req LoginRequest) (string, error) {
	// Валидация email
	if err := utils.ValidateEmail(req.Email); err != nil {
		s.logger.Warn("Invalid email format on login", "email", req.Email, "error", err)
		return "", fmt.Errorf("invalid email format: %w", err)
	}

	// Получение пользователя по email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		s.logger.Warn("User not found", "email", req.Email, "error", err)
		return "", ErrInvalidCredentials
	}

	// Проверка пароля
	if err := utils.VerifyPassword(req.Password, user.PasswordHash); err != nil {
		s.logger.Warn("Invalid password", "email", req.Email, "error", err)
		return "", ErrInvalidCredentials
	}

	// Генерация JWT токена
	token, err := utils.GenerateJWT(strconv.Itoa(user.ID))
	if err != nil {
		s.logger.Error("Failed to generate JWT token", "user_id", user.ID, "error", err)
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	s.logger.Info("User logged in successfully", "user_id", user.ID, "email", user.Email)
	return token, nil
}

// ValidateToken проверяет валидность JWT токена и возвращает пользователя
func (s *authService) ValidateToken(ctx context.Context, token string) (*models.User, error) {
	// Валидация JWT токена
	userIDStr, err := utils.ValidateJWT(token)
	if err != nil {
		s.logger.Warn("Invalid JWT token", "error", err)
		return nil, ErrInvalidToken
	}

	// Преобразование ID пользователя
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		s.logger.Error("Failed to parse user ID from token", "user_id_str", userIDStr, "error", err)
		return nil, ErrInvalidToken
	}

	// Получение пользователя по ID
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Warn("User not found for token", "user_id", userID, "error", err)
		return nil, ErrInvalidToken
	}

	// Убираем пароль из ответа
	user.PasswordHash = ""

	return user, nil
}
