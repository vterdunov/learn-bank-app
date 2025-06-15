package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vterdunov/learn-bank-app/internal/domain"
)

// AccountRepositoryImpl реализация AccountRepository
type AccountRepositoryImpl struct {
	db *pgxpool.Pool
}

// NewAccountRepository создает новый экземпляр AccountRepository
func NewAccountRepository(db *pgxpool.Pool) AccountRepository {
	return &AccountRepositoryImpl{db: db}
}

// Create создает новый счет
func (r *AccountRepositoryImpl) Create(ctx context.Context, account *domain.Account) error {
	query := `
		INSERT INTO accounts (user_id, number, balance, currency, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`

	now := time.Now()
	account.CreatedAt = now
	account.UpdatedAt = now

	err := r.db.QueryRow(ctx, query,
		account.UserID,
		account.Number,
		account.Balance,
		account.Currency,
		account.Status,
		account.CreatedAt,
		account.UpdatedAt,
	).Scan(&account.ID)

	return err
}

// GetByID получает счет по ID
func (r *AccountRepositoryImpl) GetByID(ctx context.Context, id int) (*domain.Account, error) {
	query := `
		SELECT id, user_id, number, balance, currency, status, created_at, updated_at
		FROM accounts
		WHERE id = $1`

	account := &domain.Account{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&account.ID,
		&account.UserID,
		&account.Number,
		&account.Balance,
		&account.Currency,
		&account.Status,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("account not found")
		}
		return nil, err
	}

	return account, nil
}

// GetByUserID получает все счета пользователя
func (r *AccountRepositoryImpl) GetByUserID(ctx context.Context, userID int) ([]*domain.Account, error) {
	query := `
		SELECT id, user_id, number, balance, currency, status, created_at, updated_at
		FROM accounts
		WHERE user_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []*domain.Account
	for rows.Next() {
		account := &domain.Account{}
		err := rows.Scan(
			&account.ID,
			&account.UserID,
			&account.Number,
			&account.Balance,
			&account.Currency,
			&account.Status,
			&account.CreatedAt,
			&account.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}

// GetByNumber получает счет по номеру
func (r *AccountRepositoryImpl) GetByNumber(ctx context.Context, number string) (*domain.Account, error) {
	query := `
		SELECT id, user_id, number, balance, currency, status, created_at, updated_at
		FROM accounts
		WHERE number = $1`

	account := &domain.Account{}
	err := r.db.QueryRow(ctx, query, number).Scan(
		&account.ID,
		&account.UserID,
		&account.Number,
		&account.Balance,
		&account.Currency,
		&account.Status,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("account not found")
		}
		return nil, err
	}

	return account, nil
}

// Update обновляет данные счета
func (r *AccountRepositoryImpl) Update(ctx context.Context, account *domain.Account) error {
	query := `
		UPDATE accounts
		SET balance = $2, currency = $3, status = $4, updated_at = $5
		WHERE id = $1`

	account.UpdatedAt = time.Now()

	result, err := r.db.Exec(ctx, query,
		account.ID,
		account.Balance,
		account.Currency,
		account.Status,
		account.UpdatedAt,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("account not found")
	}

	return nil
}

// UpdateBalance обновляет баланс счета
func (r *AccountRepositoryImpl) UpdateBalance(ctx context.Context, id int, balance float64) error {
	query := `
		UPDATE accounts
		SET balance = $2, updated_at = $3
		WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id, balance, time.Now())
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("account not found")
	}

	return nil
}

// Delete удаляет счет
func (r *AccountRepositoryImpl) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM accounts WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("account not found")
	}

	return nil
}

// Transfer выполняет перевод между счетами в транзакции
func (r *AccountRepositoryImpl) Transfer(ctx context.Context, fromID, toID int, amount float64) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Проверяем баланс отправителя
	var fromBalance float64
	err = tx.QueryRow(ctx, "SELECT balance FROM accounts WHERE id = $1 FOR UPDATE", fromID).Scan(&fromBalance)
	if err != nil {
		return err
	}

	if fromBalance < amount {
		return errors.New("insufficient funds")
	}

	// Списываем с отправителя
	_, err = tx.Exec(ctx,
		"UPDATE accounts SET balance = balance - $1, updated_at = $2 WHERE id = $3",
		amount, time.Now(), fromID)
	if err != nil {
		return err
	}

	// Пополняем получателя
	_, err = tx.Exec(ctx,
		"UPDATE accounts SET balance = balance + $1, updated_at = $2 WHERE id = $3",
		amount, time.Now(), toID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// GetBalance получает баланс счета
func (r *AccountRepositoryImpl) GetBalance(ctx context.Context, id int) (float64, error) {
	query := `SELECT balance FROM accounts WHERE id = $1`

	var balance float64
	err := r.db.QueryRow(ctx, query, id).Scan(&balance)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, errors.New("account not found")
		}
		return 0, err
	}

	return balance, nil
}
