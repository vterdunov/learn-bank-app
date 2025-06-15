package models

import "time"

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
