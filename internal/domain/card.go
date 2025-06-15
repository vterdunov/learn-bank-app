package domain

import (
	"errors"
	"time"
)

// Card представляет банковскую карту
type Card struct {
	ID            int       `json:"id" db:"id"`
	AccountID     int       `json:"account_id" db:"account_id"`
	EncryptedData string    `json:"-" db:"encrypted_data"`
	HMAC          string    `json:"-" db:"hmac"`
	CVVHash       string    `json:"-" db:"cvv_hash"`
	ExpiryDate    time.Time `json:"expiry_date" db:"expiry_date"`
	Status        string    `json:"status" db:"status"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// CardData представляет расшифрованные данные карты
type CardData struct {
	Number     string    `json:"number"`
	ExpiryDate time.Time `json:"expiry_date"`
}

// CreateCardRequest представляет запрос на создание карты
type CreateCardRequest struct {
	AccountID int `json:"account_id"`
}

// PaymentRequest представляет запрос на оплату картой
type PaymentRequest struct {
	CardID int     `json:"card_id"`
	Amount float64 `json:"amount"`
	CVV    string  `json:"cvv"`
}

// CardStatus определяет возможные статусы карты
const (
	CardStatusActive  = "active"
	CardStatusBlocked = "blocked"
	CardStatusExpired = "expired"
)

// CardType определяет тип карты LearnBank
const (
	CardTypeLearnBank = "LEARNBANK"
)

// Validation errors
var (
	ErrInvalidCardStatus    = errors.New("invalid card status")
	ErrInvalidPaymentCVV    = errors.New("invalid CVV")
	ErrInvalidPaymentAmount = errors.New("invalid payment amount")
	ErrCardExpired          = errors.New("card is expired")
)

// Validate валидирует карту
func (c *Card) Validate() error {
	if c.Status != CardStatusActive && c.Status != CardStatusBlocked && c.Status != CardStatusExpired {
		return ErrInvalidCardStatus
	}
	if c.ExpiryDate.Before(time.Now()) {
		return ErrCardExpired
	}
	return nil
}

// Validate валидирует запрос на создание карты
func (r *CreateCardRequest) Validate() error {
	if r.AccountID <= 0 {
		return errors.New("invalid account ID")
	}
	return nil
}

// Validate валидирует запрос на оплату
func (r *PaymentRequest) Validate() error {
	if r.CardID <= 0 {
		return errors.New("invalid card ID")
	}
	if r.Amount <= 0 {
		return ErrInvalidPaymentAmount
	}
	if r.Amount > 1000000000 {
		return ErrInvalidPaymentAmount
	}
	if len(r.CVV) != 3 {
		return ErrInvalidPaymentCVV
	}
	return nil
}
