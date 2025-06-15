package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vterdunov/learn-bank-app/internal/domain"
)

// CardRepositoryImpl реализация CardRepository
type CardRepositoryImpl struct {
	db *pgxpool.Pool
}

// NewCardRepository создает новый экземпляр CardRepository
func NewCardRepository(db *pgxpool.Pool) CardRepository {
	return &CardRepositoryImpl{db: db}
}

// Create создает новую карту
func (r *CardRepositoryImpl) Create(ctx context.Context, card *domain.Card) error {
	query := `
		INSERT INTO cards (account_id, encrypted_data, hmac, cvv_hash, expiry_date, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`

	now := time.Now()
	card.CreatedAt = now
	card.UpdatedAt = now

	err := r.db.QueryRow(ctx, query,
		card.AccountID,
		card.EncryptedData,
		card.HMAC,
		card.CVVHash,
		card.ExpiryDate,
		card.Status,
		card.CreatedAt,
		card.UpdatedAt,
	).Scan(&card.ID)

	return err
}

// GetByID получает карту по ID
func (r *CardRepositoryImpl) GetByID(ctx context.Context, id int) (*domain.Card, error) {
	query := `
		SELECT id, account_id, encrypted_data, hmac, cvv_hash, expiry_date, status, created_at, updated_at
		FROM cards
		WHERE id = $1`

	card := &domain.Card{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&card.ID,
		&card.AccountID,
		&card.EncryptedData,
		&card.HMAC,
		&card.CVVHash,
		&card.ExpiryDate,
		&card.Status,
		&card.CreatedAt,
		&card.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("card not found")
		}
		return nil, err
	}

	return card, nil
}

// GetByAccountID получает все карты счета
func (r *CardRepositoryImpl) GetByAccountID(ctx context.Context, accountID int) ([]*domain.Card, error) {
	query := `
		SELECT id, account_id, encrypted_data, hmac, cvv_hash, expiry_date, status, created_at, updated_at
		FROM cards
		WHERE account_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []*domain.Card
	for rows.Next() {
		card := &domain.Card{}
		err := rows.Scan(
			&card.ID,
			&card.AccountID,
			&card.EncryptedData,
			&card.HMAC,
			&card.CVVHash,
			&card.ExpiryDate,
			&card.Status,
			&card.CreatedAt,
			&card.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}

	return cards, nil
}

// Update обновляет данные карты
func (r *CardRepositoryImpl) Update(ctx context.Context, card *domain.Card) error {
	query := `
		UPDATE cards
		SET encrypted_data = $2, hmac = $3, cvv_hash = $4, expiry_date = $5, status = $6, updated_at = $7
		WHERE id = $1`

	card.UpdatedAt = time.Now()

	result, err := r.db.Exec(ctx, query,
		card.ID,
		card.EncryptedData,
		card.HMAC,
		card.CVVHash,
		card.ExpiryDate,
		card.Status,
		card.UpdatedAt,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("card not found")
	}

	return nil
}

// Delete удаляет карту
func (r *CardRepositoryImpl) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM cards WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("card not found")
	}

	return nil
}

// UpdateStatus обновляет статус карты
func (r *CardRepositoryImpl) UpdateStatus(ctx context.Context, id int, status string) error {
	query := `
		UPDATE cards
		SET status = $2, updated_at = $3
		WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id, status, time.Now())
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("card not found")
	}

	return nil
}

// GetActiveCardsByAccount получает активные карты счета
func (r *CardRepositoryImpl) GetActiveCardsByAccount(ctx context.Context, accountID int) ([]*domain.Card, error) {
	query := `
		SELECT id, account_id, encrypted_data, hmac, cvv_hash, expiry_date, status, created_at, updated_at
		FROM cards
		WHERE account_id = $1 AND status = 'active' AND expiry_date > NOW()
		ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []*domain.Card
	for rows.Next() {
		card := &domain.Card{}
		err := rows.Scan(
			&card.ID,
			&card.AccountID,
			&card.EncryptedData,
			&card.HMAC,
			&card.CVVHash,
			&card.ExpiryDate,
			&card.Status,
			&card.CreatedAt,
			&card.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}

	return cards, nil
}
