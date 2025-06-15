package utils

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// Константы для криптографии
const (
	// BCryptCost стоимость хеширования bcrypt
	BCryptCost = 12
	// HMACKeySize размер ключа HMAC
	HMACKeySize = 32
	// EncryptionKeySize размер ключа шифрования
	EncryptionKeySize = 32
)

// Ошибки криптографии
var (
	ErrInvalidPassword  = errors.New("invalid password")
	ErrInvalidHMAC      = errors.New("invalid HMAC")
	ErrEncryptionFailed = errors.New("encryption failed")
	ErrDecryptionFailed = errors.New("decryption failed")
)

// HashPassword хеширует пароль с использованием bcrypt
func HashPassword(password string) (string, error) {
	if len(password) == 0 {
		return "", ErrInvalidPassword
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), BCryptCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hashedBytes), nil
}

// VerifyPassword проверяет пароль против хеша
func VerifyPassword(hashedPassword, password string) error {
	if len(password) == 0 || len(hashedPassword) == 0 {
		return ErrInvalidPassword
	}

	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrInvalidPassword
		}
		return fmt.Errorf("failed to verify password: %w", err)
	}

	return nil
}

// HashCVV хеширует CVV с использованием bcrypt
func HashCVV(cvv string) (string, error) {
	if len(cvv) < 3 || len(cvv) > 4 {
		return "", errors.New("invalid CVV length")
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(cvv), BCryptCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash CVV: %w", err)
	}

	return string(hashedBytes), nil
}

// VerifyCVV проверяет CVV против хеша
func VerifyCVV(hashedCVV, cvv string) error {
	if len(cvv) < 3 || len(cvv) > 4 || len(hashedCVV) == 0 {
		return errors.New("invalid CVV")
	}

	err := bcrypt.CompareHashAndPassword([]byte(hashedCVV), []byte(cvv))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return errors.New("invalid CVV")
		}
		return fmt.Errorf("failed to verify CVV: %w", err)
	}

	return nil
}

// GenerateRandomKey генерирует случайный ключ заданного размера
func GenerateRandomKey(size int) ([]byte, error) {
	key := make([]byte, size)
	_, err := rand.Read(key)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random key: %w", err)
	}
	return key, nil
}

// ComputeHMAC вычисляет HMAC-SHA256 для данных
func ComputeHMAC(data string, key []byte) string {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// VerifyHMAC проверяет HMAC для данных
func VerifyHMAC(data string, expectedHMAC string, key []byte) error {
	computedHMAC := ComputeHMAC(data, key)
	if !hmac.Equal([]byte(computedHMAC), []byte(expectedHMAC)) {
		return ErrInvalidHMAC
	}
	return nil
}

// EncryptedData представляет зашифрованные данные с HMAC
type EncryptedData struct {
	Data []byte `json:"data"`
	HMAC string `json:"hmac"`
}

// SimpleEncrypt простое XOR шифрование (для демонстрации, в продакшене использовать AES)
func SimpleEncrypt(plaintext string, key []byte) (*EncryptedData, error) {
	if len(key) < EncryptionKeySize {
		return nil, errors.New("key too short")
	}

	// Простое XOR шифрование
	data := make([]byte, len(plaintext))
	for i, b := range []byte(plaintext) {
		data[i] = b ^ key[i%len(key)]
	}

	// Вычисляем HMAC для проверки целостности
	hmacValue := ComputeHMAC(string(data), key)

	return &EncryptedData{
		Data: data,
		HMAC: hmacValue,
	}, nil
}

// SimpleDecrypt простая расшифровка XOR
func SimpleDecrypt(encrypted *EncryptedData, key []byte) (string, error) {
	if len(key) < EncryptionKeySize {
		return "", errors.New("key too short")
	}

	// Проверяем HMAC
	if err := VerifyHMAC(string(encrypted.Data), encrypted.HMAC, key); err != nil {
		return "", fmt.Errorf("HMAC verification failed: %w", err)
	}

	// Расшифровываем данные
	plaintext := make([]byte, len(encrypted.Data))
	for i, b := range encrypted.Data {
		plaintext[i] = b ^ key[i%len(key)]
	}

	return string(plaintext), nil
}

// EncryptCardData шифрует данные карты
func EncryptCardData(cardNumber, expiryDate string, key []byte) (*EncryptedData, *EncryptedData, error) {
	encryptedNumber, err := SimpleEncrypt(cardNumber, key)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encrypt card number: %w", err)
	}

	encryptedExpiry, err := SimpleEncrypt(expiryDate, key)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encrypt expiry date: %w", err)
	}

	return encryptedNumber, encryptedExpiry, nil
}

// DecryptCardData расшифровывает данные карты
func DecryptCardData(encryptedNumber, encryptedExpiry *EncryptedData, key []byte) (string, string, error) {
	cardNumber, err := SimpleDecrypt(encryptedNumber, key)
	if err != nil {
		return "", "", fmt.Errorf("failed to decrypt card number: %w", err)
	}

	expiryDate, err := SimpleDecrypt(encryptedExpiry, key)
	if err != nil {
		return "", "", fmt.Errorf("failed to decrypt expiry date: %w", err)
	}

	return cardNumber, expiryDate, nil
}
