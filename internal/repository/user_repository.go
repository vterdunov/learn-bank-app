package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vterdunov/learn-bank-app/internal/domain"
	"github.com/vterdunov/learn-bank-app/internal/utils"
)

// UserRepositoryImpl реализация UserRepository
type UserRepositoryImpl struct {
	db *pgxpool.Pool
}

// NewUserRepository создает новый экземпляр UserRepository
func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &UserRepositoryImpl{db: db}
}

// Create создает нового пользователя
func (r *UserRepositoryImpl) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (username, email, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	err := r.db.QueryRow(ctx, query,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID)

	if err != nil {
		return utils.WrapDBError(err, "create user")
	}

	return nil
}

// GetByID получает пользователя по ID
func (r *UserRepositoryImpl) GetByID(ctx context.Context, id int) (*domain.User, error) {
	query := `
		SELECT id, username, email, password_hash, created_at, updated_at
		FROM users
		WHERE id = $1`

	user := &domain.User{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if utils.IsRecordNotFound(err) {
			return nil, utils.ErrUserNotFound
		}
		return nil, utils.WrapDBError(err, "get user by id")
	}

	return user, nil
}

// GetByEmail получает пользователя по email
func (r *UserRepositoryImpl) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, username, email, password_hash, created_at, updated_at
		FROM users
		WHERE email = $1`

	user := &domain.User{}
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if utils.IsRecordNotFound(err) {
			return nil, utils.ErrUserNotFound
		}
		return nil, utils.WrapDBError(err, "get user by email")
	}

	return user, nil
}

// GetByUsername получает пользователя по username
func (r *UserRepositoryImpl) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `
		SELECT id, username, email, password_hash, created_at, updated_at
		FROM users
		WHERE username = $1`

	user := &domain.User{}
	err := r.db.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if utils.IsRecordNotFound(err) {
			return nil, utils.ErrUserNotFound
		}
		return nil, utils.WrapDBError(err, "get user by username")
	}

	return user, nil
}

// Update обновляет данные пользователя
func (r *UserRepositoryImpl) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users
		SET username = $2, email = $3, password_hash = $4, updated_at = $5
		WHERE id = $1`

	user.UpdatedAt = time.Now()

	result, err := r.db.Exec(ctx, query,
		user.ID,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.UpdatedAt,
	)

	if err != nil {
		return utils.WrapDBError(err, "update user")
	}

	if result.RowsAffected() == 0 {
		return utils.ErrUserNotFound
	}

	return nil
}

// Delete удаляет пользователя
func (r *UserRepositoryImpl) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return utils.WrapDBError(err, "delete user")
	}

	if result.RowsAffected() == 0 {
		return utils.ErrUserNotFound
	}

	return nil
}

// EmailExists проверяет существование email
func (r *UserRepositoryImpl) EmailExists(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

	var exists bool
	err := r.db.QueryRow(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, utils.WrapDBError(err, "check email exists")
	}

	return exists, nil
}

// UsernameExists проверяет существование username
func (r *UserRepositoryImpl) UsernameExists(ctx context.Context, username string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`

	var exists bool
	err := r.db.QueryRow(ctx, query, username).Scan(&exists)
	if err != nil {
		return false, utils.WrapDBError(err, "check username exists")
	}

	return exists, nil
}
