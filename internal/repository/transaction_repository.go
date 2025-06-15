package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vterdunov/learn-bank-app/internal/domain"
)

// TransactionRepositoryImpl реализация TransactionRepository
type TransactionRepositoryImpl struct {
	db *pgxpool.Pool
}

// NewTransactionRepository создает новый экземпляр TransactionRepository
func NewTransactionRepository(db *pgxpool.Pool) TransactionRepository {
	return &TransactionRepositoryImpl{db: db}
}

// Create создает новую транзакцию
func (r *TransactionRepositoryImpl) Create(ctx context.Context, transaction *domain.Transaction) error {
	query := `
		INSERT INTO transactions (from_account, to_account, amount, type, status, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`

	now := time.Now()
	transaction.CreatedAt = now
	transaction.UpdatedAt = now

	err := r.db.QueryRow(ctx, query,
		transaction.FromAccount,
		transaction.ToAccount,
		transaction.Amount,
		transaction.Type,
		transaction.Status,
		transaction.Description,
		transaction.CreatedAt,
		transaction.UpdatedAt,
	).Scan(&transaction.ID)

	return err
}

// GetByID получает транзакцию по ID
func (r *TransactionRepositoryImpl) GetByID(ctx context.Context, id int) (*domain.Transaction, error) {
	query := `
		SELECT id, from_account, to_account, amount, type, status, description, created_at, updated_at
		FROM transactions
		WHERE id = $1`

	transaction := &domain.Transaction{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&transaction.ID,
		&transaction.FromAccount,
		&transaction.ToAccount,
		&transaction.Amount,
		&transaction.Type,
		&transaction.Status,
		&transaction.Description,
		&transaction.CreatedAt,
		&transaction.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("transaction not found")
		}
		return nil, err
	}

	return transaction, nil
}

// GetByAccountID получает транзакции по ID счета с пагинацией
func (r *TransactionRepositoryImpl) GetByAccountID(ctx context.Context, accountID int, limit, offset int) ([]*domain.Transaction, error) {
	query := `
		SELECT id, from_account, to_account, amount, type, status, description, created_at, updated_at
		FROM transactions
		WHERE from_account = $1 OR to_account = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, query, accountID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*domain.Transaction
	for rows.Next() {
		transaction := &domain.Transaction{}
		err := rows.Scan(
			&transaction.ID,
			&transaction.FromAccount,
			&transaction.ToAccount,
			&transaction.Amount,
			&transaction.Type,
			&transaction.Status,
			&transaction.Description,
			&transaction.CreatedAt,
			&transaction.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

// GetByUserID получает транзакции пользователя с пагинацией
func (r *TransactionRepositoryImpl) GetByUserID(ctx context.Context, userID int, limit, offset int) ([]*domain.Transaction, error) {
	query := `
		SELECT t.id, t.from_account, t.to_account, t.amount, t.type, t.status, t.description, t.created_at, t.updated_at
		FROM transactions t
		LEFT JOIN accounts a1 ON t.from_account = a1.id
		LEFT JOIN accounts a2 ON t.to_account = a2.id
		WHERE a1.user_id = $1 OR a2.user_id = $1
		ORDER BY t.created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*domain.Transaction
	for rows.Next() {
		transaction := &domain.Transaction{}
		err := rows.Scan(
			&transaction.ID,
			&transaction.FromAccount,
			&transaction.ToAccount,
			&transaction.Amount,
			&transaction.Type,
			&transaction.Status,
			&transaction.Description,
			&transaction.CreatedAt,
			&transaction.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

// Update обновляет транзакцию
func (r *TransactionRepositoryImpl) Update(ctx context.Context, transaction *domain.Transaction) error {
	query := `
		UPDATE transactions
		SET amount = $2, type = $3, status = $4, description = $5, updated_at = $6
		WHERE id = $1`

	transaction.UpdatedAt = time.Now()

	result, err := r.db.Exec(ctx, query,
		transaction.ID,
		transaction.Amount,
		transaction.Type,
		transaction.Status,
		transaction.Description,
		transaction.UpdatedAt,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("transaction not found")
	}

	return nil
}

// Delete удаляет транзакцию
func (r *TransactionRepositoryImpl) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM transactions WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("transaction not found")
	}

	return nil
}

// GetTransactionsByDateRange получает транзакции за период
func (r *TransactionRepositoryImpl) GetTransactionsByDateRange(ctx context.Context, accountID int, startDate, endDate time.Time) ([]*domain.Transaction, error) {
	query := `
		SELECT id, from_account, to_account, amount, type, status, description, created_at, updated_at
		FROM transactions
		WHERE (from_account = $1 OR to_account = $1)
		  AND created_at >= $2 AND created_at <= $3
		ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query, accountID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*domain.Transaction
	for rows.Next() {
		transaction := &domain.Transaction{}
		err := rows.Scan(
			&transaction.ID,
			&transaction.FromAccount,
			&transaction.ToAccount,
			&transaction.Amount,
			&transaction.Type,
			&transaction.Status,
			&transaction.Description,
			&transaction.CreatedAt,
			&transaction.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

// GetMonthlyStatistics получает месячную статистику транзакций
func (r *TransactionRepositoryImpl) GetMonthlyStatistics(ctx context.Context, userID int, year int, month int) (*domain.MonthlyStatistics, error) {
	query := `
		SELECT
			COALESCE(SUM(CASE
				WHEN t.type IN ('deposit', 'transfer') AND a2.user_id = $1
				THEN t.amount
				ELSE 0
			END), 0) as income,
			COALESCE(SUM(CASE
				WHEN t.type IN ('withdraw', 'transfer', 'payment') AND a1.user_id = $1
				THEN t.amount
				ELSE 0
			END), 0) as expenses
		FROM transactions t
		LEFT JOIN accounts a1 ON t.from_account = a1.id
		LEFT JOIN accounts a2 ON t.to_account = a2.id
		WHERE (a1.user_id = $1 OR a2.user_id = $1)
		  AND EXTRACT(YEAR FROM t.created_at) = $2
		  AND EXTRACT(MONTH FROM t.created_at) = $3
		  AND t.status = 'completed'`

	stats := &domain.MonthlyStatistics{
		Year:  year,
		Month: month,
	}

	err := r.db.QueryRow(ctx, query, userID, year, month).Scan(&stats.Income, &stats.Expenses)
	if err != nil {
		return nil, err
	}

	stats.Balance = stats.Income - stats.Expenses

	return stats, nil
}
