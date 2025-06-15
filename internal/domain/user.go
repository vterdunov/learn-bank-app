package domain

import (
	"errors"
	"time"
)

// User представляет пользователя системы
type User struct {
	ID           int       `json:"id" db:"id"`
	Username     string    `json:"username" db:"username"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// RegisterRequest представляет запрос на регистрацию
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest представляет запрос на авторизацию
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse представляет ответ при успешной авторизации
type AuthResponse struct {
	User  *User  `json:"user"`
	Token string `json:"token"`
}

// Validation errors
var (
	ErrEmptyUsername = errors.New("username cannot be empty")
	ErrEmptyEmail    = errors.New("email cannot be empty")
	ErrEmptyPassword = errors.New("password cannot be empty")
)

// Validate валидирует пользователя
func (u *User) Validate() error {
	if u.Username == "" {
		return ErrEmptyUsername
	}
	if u.Email == "" {
		return ErrEmptyEmail
	}
	return nil
}

// Validate валидирует запрос на регистрацию
func (r *RegisterRequest) Validate() error {
	if r.Username == "" {
		return ErrEmptyUsername
	}
	if r.Email == "" {
		return ErrEmptyEmail
	}
	if r.Password == "" {
		return ErrEmptyPassword
	}
	return nil
}

// Validate валидирует запрос на авторизацию
func (r *LoginRequest) Validate() error {
	if r.Email == "" {
		return ErrEmptyEmail
	}
	if r.Password == "" {
		return ErrEmptyPassword
	}
	return nil
}
