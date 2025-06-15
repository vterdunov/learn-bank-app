package service

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/vterdunov/learn-bank-app/internal/domain"
	"github.com/vterdunov/learn-bank-app/internal/utils"
)

// MockUserRepository для тестирования
type MockUserRepository struct {
	users       map[string]*domain.User
	usersByID   map[int]*domain.User
	createError error
	getError    error
	nextID      int
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users:     make(map[string]*domain.User),
		usersByID: make(map[int]*domain.User),
		nextID:    1,
	}
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	if m.createError != nil {
		return m.createError
	}

	user.ID = m.nextID
	m.nextID++
	m.users[user.Email] = user
	m.usersByID[user.ID] = user
	return nil
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	if m.getError != nil {
		return nil, m.getError
	}

	user, exists := m.users[email]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	if m.getError != nil {
		return nil, m.getError
	}

	for _, user := range m.users {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, errors.New("user not found")
}

func (m *MockUserRepository) GetByID(ctx context.Context, id int) (*domain.User, error) {
	if m.getError != nil {
		return nil, m.getError
	}

	user, exists := m.usersByID[id]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	if m.createError != nil {
		return m.createError
	}

	m.users[user.Email] = user
	m.usersByID[user.ID] = user
	return nil
}

func (m *MockUserRepository) Delete(ctx context.Context, id int) error {
	if m.createError != nil {
		return m.createError
	}

	user, exists := m.usersByID[id]
	if !exists {
		return errors.New("user not found")
	}

	delete(m.users, user.Email)
	delete(m.usersByID, id)
	return nil
}

func (m *MockUserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	if m.getError != nil {
		return false, m.getError
	}

	_, exists := m.users[email]
	return exists, nil
}

func (m *MockUserRepository) UsernameExists(ctx context.Context, username string) (bool, error) {
	if m.getError != nil {
		return false, m.getError
	}

	for _, user := range m.users {
		if user.Username == username {
			return true, nil
		}
	}
	return false, nil
}

func setupAuthService() (*authService, *MockUserRepository) {
	// Инициализируем JWT для тестов
	utils.InitJWT("test-secret-key-for-testing")

	mockRepo := NewMockUserRepository()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewAuthService(mockRepo, logger).(*authService)
	return service, mockRepo
}

func TestAuthService_Register(t *testing.T) {
	service, mockRepo := setupAuthService()
	ctx := context.Background()

	t.Run("successful registration", func(t *testing.T) {
		req := RegisterRequest{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "SecurePass123!",
		}

		user, err := service.Register(ctx, req)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if user.Username != req.Username {
			t.Errorf("Expected username %s, got %s", req.Username, user.Username)
		}

		if user.Email != req.Email {
			t.Errorf("Expected email %s, got %s", req.Email, user.Email)
		}

		if user.PasswordHash != "" {
			t.Error("Expected password hash to be empty in response")
		}

		// Проверяем что пользователь сохранен в репозитории
		savedUser, err := mockRepo.GetByEmail(ctx, req.Email)
		if err != nil {
			t.Fatalf("Expected user to be saved, got error: %v", err)
		}

		if savedUser.PasswordHash == "" {
			t.Error("Expected password hash to be saved in repository")
		} else {
			// Проверяем что пароль корректно хешируется
			err = utils.VerifyPassword(req.Password, savedUser.PasswordHash)
			if err != nil {
				t.Errorf("Expected password to be correctly hashed, got error: %v", err)
			}
		}
	})

	t.Run("invalid email", func(t *testing.T) {
		req := RegisterRequest{
			Username: "testuser2",
			Email:    "invalid-email",
			Password: "SecurePass123!",
		}

		_, err := service.Register(ctx, req)
		if err == nil {
			t.Error("Expected error for invalid email")
		}
	})

	t.Run("invalid username", func(t *testing.T) {
		req := RegisterRequest{
			Username: "u", // слишком короткий
			Email:    "test2@example.com",
			Password: "SecurePass123!",
		}

		_, err := service.Register(ctx, req)
		if err == nil {
			t.Error("Expected error for invalid username")
		}
	})

	t.Run("invalid password", func(t *testing.T) {
		req := RegisterRequest{
			Username: "testuser3",
			Email:    "test3@example.com",
			Password: "weak", // слишком слабый пароль
		}

		_, err := service.Register(ctx, req)
		if err == nil {
			t.Error("Expected error for invalid password")
		}
	})

	t.Run("duplicate email", func(t *testing.T) {
		// Сначала создаем пользователя
		req1 := RegisterRequest{
			Username: "testuser4",
			Email:    "duplicate@example.com",
			Password: "SecurePass123!",
		}
		_, err := service.Register(ctx, req1)
		if err != nil {
			t.Fatalf("Failed to create first user: %v", err)
		}

		// Пытаемся создать с тем же email
		req2 := RegisterRequest{
			Username: "testuser5",
			Email:    "duplicate@example.com",
			Password: "SecurePass123!",
		}
		_, err = service.Register(ctx, req2)
		if err != ErrUserAlreadyExists {
			t.Errorf("Expected ErrUserAlreadyExists, got %v", err)
		}
	})

	t.Run("duplicate username", func(t *testing.T) {
		// Сначала создаем пользователя
		req1 := RegisterRequest{
			Username: "duplicateuser",
			Email:    "test6@example.com",
			Password: "SecurePass123!",
		}
		_, err := service.Register(ctx, req1)
		if err != nil {
			t.Fatalf("Failed to create first user: %v", err)
		}

		// Пытаемся создать с тем же username
		req2 := RegisterRequest{
			Username: "duplicateuser",
			Email:    "test7@example.com",
			Password: "SecurePass123!",
		}
		_, err = service.Register(ctx, req2)
		if err != ErrUserAlreadyExists {
			t.Errorf("Expected ErrUserAlreadyExists, got %v", err)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.createError = errors.New("database error")

		req := RegisterRequest{
			Username: "testuser8",
			Email:    "test8@example.com",
			Password: "SecurePass123!",
		}

		_, err := service.Register(ctx, req)
		if err == nil {
			t.Error("Expected error from repository")
		}

		mockRepo.createError = nil // сбрасываем ошибку
	})
}

func TestAuthService_Login(t *testing.T) {
	service, mockRepo := setupAuthService()
	ctx := context.Background()

	// Создаем пользователя для тестов через регистрацию
	regReq := RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "SecurePass123!",
	}

	// Регистрируем пользователя чтобы он попал в mockRepo
	_, err := service.Register(ctx, regReq)
	if err != nil {
		t.Fatalf("Failed to register test user: %v", err)
	}

	t.Run("successful login", func(t *testing.T) {
		req := LoginRequest{
			Email:    "test@example.com",
			Password: "SecurePass123!",
		}

		token, err := service.Login(ctx, req)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if token == "" {
			t.Error("Expected token to be returned")
		}

		// Проверяем что токен валидный
		_, err = utils.ValidateJWT(token)
		if err != nil {
			t.Errorf("Expected valid token, got error: %v", err)
		}
	})

	t.Run("invalid email format", func(t *testing.T) {
		req := LoginRequest{
			Email:    "invalid-email",
			Password: "SecurePass123!",
		}

		_, err := service.Login(ctx, req)
		if err == nil {
			t.Error("Expected error for invalid email format")
		}
	})

	t.Run("user not found", func(t *testing.T) {
		req := LoginRequest{
			Email:    "notfound@example.com",
			Password: "SecurePass123!",
		}

		_, err := service.Login(ctx, req)
		if err != ErrInvalidCredentials {
			t.Errorf("Expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("invalid password", func(t *testing.T) {
		req := LoginRequest{
			Email:    "test@example.com",
			Password: "WrongPassword",
		}

		_, err := service.Login(ctx, req)
		if err != ErrInvalidCredentials {
			t.Errorf("Expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.getError = errors.New("database error")

		req := LoginRequest{
			Email:    "test@example.com",
			Password: "SecurePass123!",
		}

		_, err := service.Login(ctx, req)
		if err != ErrInvalidCredentials {
			t.Errorf("Expected ErrInvalidCredentials, got %v", err)
		}

		mockRepo.getError = nil // сбрасываем ошибку
	})
}

func TestAuthService_ValidateToken(t *testing.T) {
	service, mockRepo := setupAuthService()
	ctx := context.Background()

	// Создаем пользователя для тестов
	testUser := &domain.User{
		ID:           1,
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	mockRepo.usersByID[testUser.ID] = testUser

	t.Run("valid token", func(t *testing.T) {
		// Генерируем валидный токен
		token, err := utils.GenerateJWT("1")
		if err != nil {
			t.Fatalf("Failed to generate token: %v", err)
		}

		user, err := service.ValidateToken(ctx, token)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if user.ID != testUser.ID {
			t.Errorf("Expected user ID %d, got %d", testUser.ID, user.ID)
		}

		if user.PasswordHash != "" {
			t.Error("Expected password hash to be empty in response")
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		_, err := service.ValidateToken(ctx, "invalid-token")
		if err != ErrInvalidToken {
			t.Errorf("Expected ErrInvalidToken, got %v", err)
		}
	})

	t.Run("expired token", func(t *testing.T) {
		// Создаем токен с истекшим сроком
		// Это сложно протестировать без изменения времени системы
		// Поэтому пропускаем этот тест
		t.Skip("Testing expired token requires time manipulation")
	})

	t.Run("user not found", func(t *testing.T) {
		// Генерируем токен для несуществующего пользователя
		token, err := utils.GenerateJWT("999")
		if err != nil {
			t.Fatalf("Failed to generate token: %v", err)
		}

		_, err = service.ValidateToken(ctx, token)
		if err != ErrInvalidToken {
			t.Errorf("Expected ErrInvalidToken, got %v", err)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.getError = errors.New("database error")

		token, err := utils.GenerateJWT("1")
		if err != nil {
			t.Fatalf("Failed to generate token: %v", err)
		}

		_, err = service.ValidateToken(ctx, token)
		if err != ErrInvalidToken {
			t.Errorf("Expected ErrInvalidToken, got %v", err)
		}

		mockRepo.getError = nil // сбрасываем ошибку
	})
}
