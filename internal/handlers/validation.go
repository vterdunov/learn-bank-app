package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/vterdunov/learn-bank-app/internal/utils"
)

var (
	ErrInvalidJSON      = errors.New("invalid JSON format")
	ErrValidationFailed = errors.New("validation failed")
)

// FieldError представляет ошибку валидации поля
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationError представляет ошибки валидации
type ValidationError struct {
	Errors []FieldError `json:"errors"`
}

func (v ValidationError) Error() string {
	if len(v.Errors) == 0 {
		return "validation failed"
	}
	return fmt.Sprintf("validation failed: %s", v.Errors[0].Message)
}

// ValidateJSON парсит JSON и валидирует структуру
func ValidateJSON(r *http.Request, dst interface{}) error {
	if r.Body == nil {
		return ErrInvalidJSON
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}

	return nil
}

// Validate валидирует DTO и возвращает ошибки
func Validate(dto interface{}) *ValidationError {
	var errors []FieldError

	switch v := dto.(type) {
	case *RegisterRequest:
		errors = validateRegisterRequest(v)
	case *LoginRequest:
		errors = validateLoginRequest(v)
	case *CreateAccountRequest:
		errors = validateCreateAccountRequest(v)
	case *DepositRequest:
		errors = validateAmountRequest(v.Amount, "amount")
	case *WithdrawRequest:
		errors = validateAmountRequest(v.Amount, "amount")
	case *TransferRequest:
		errors = validateTransferRequest(v)
	case *CreateCardRequest:
		errors = validateCreateCardRequest(v)
	case *CardPaymentRequest:
		errors = validateCardPaymentRequest(v)
	case *CreateCreditRequest:
		errors = validateCreateCreditRequest(v)
	case *MonthlyStatsRequest:
		errors = validateMonthlyStatsRequest(v)
	case *BalancePredictionRequest:
		errors = validateBalancePredictionRequest(v)
	}

	if len(errors) > 0 {
		return &ValidationError{Errors: errors}
	}

	return nil
}

func validateRegisterRequest(req *RegisterRequest) []FieldError {
	var errors []FieldError

	if req.Username == "" {
		errors = append(errors, FieldError{
			Field:   "username",
			Message: "username is required",
		})
	} else if len(req.Username) < 3 || len(req.Username) > 50 {
		errors = append(errors, FieldError{
			Field:   "username",
			Message: "username must be between 3 and 50 characters",
		})
	}

	if req.Email == "" {
		errors = append(errors, FieldError{
			Field:   "email",
			Message: "email is required",
		})
	} else if utils.ValidateEmail(req.Email) != nil {
		errors = append(errors, FieldError{
			Field:   "email",
			Message: "invalid email format",
		})
	}

	if req.Password == "" {
		errors = append(errors, FieldError{
			Field:   "password",
			Message: "password is required",
		})
	} else if len(req.Password) < 8 {
		errors = append(errors, FieldError{
			Field:   "password",
			Message: "password must be at least 8 characters",
		})
	}

	return errors
}

func validateLoginRequest(req *LoginRequest) []FieldError {
	var errors []FieldError

	if req.Email == "" {
		errors = append(errors, FieldError{
			Field:   "email",
			Message: "email is required",
		})
	} else if utils.ValidateEmail(req.Email) != nil {
		errors = append(errors, FieldError{
			Field:   "email",
			Message: "invalid email format",
		})
	}

	if req.Password == "" {
		errors = append(errors, FieldError{
			Field:   "password",
			Message: "password is required",
		})
	}

	return errors
}

func validateCreateAccountRequest(req *CreateAccountRequest) []FieldError {
	var errors []FieldError

	if req.Name == "" {
		errors = append(errors, FieldError{
			Field:   "name",
			Message: "account name is required",
		})
	} else if len(req.Name) > 100 {
		errors = append(errors, FieldError{
			Field:   "name",
			Message: "account name must not exceed 100 characters",
		})
	}

	if req.AccountType == "" {
		errors = append(errors, FieldError{
			Field:   "account_type",
			Message: "account type is required",
		})
	} else if req.AccountType != "savings" && req.AccountType != "checking" {
		errors = append(errors, FieldError{
			Field:   "account_type",
			Message: "account type must be 'savings' or 'checking'",
		})
	}

	return errors
}

func validateAmountRequest(amount float64, fieldName string) []FieldError {
	var errors []FieldError

	if amount <= 0 {
		errors = append(errors, FieldError{
			Field:   fieldName,
			Message: fieldName + " must be positive",
		})
	}

	return errors
}

func validateTransferRequest(req *TransferRequest) []FieldError {
	var errors []FieldError

	if req.FromAccountID == "" {
		errors = append(errors, FieldError{
			Field:   "from_account_id",
			Message: "from_account_id is required",
		})
	}

	if req.ToAccountID == "" {
		errors = append(errors, FieldError{
			Field:   "to_account_id",
			Message: "to_account_id is required",
		})
	}

	if req.Amount <= 0 {
		errors = append(errors, FieldError{
			Field:   "amount",
			Message: "amount must be positive",
		})
	}

	if len(req.Description) > 255 {
		errors = append(errors, FieldError{
			Field:   "description",
			Message: "description must not exceed 255 characters",
		})
	}

	return errors
}

func validateCreateCardRequest(req *CreateCardRequest) []FieldError {
	var errors []FieldError

	if req.AccountID == "" {
		errors = append(errors, FieldError{
			Field:   "account_id",
			Message: "account_id is required",
		})
	}

	if req.CardType == "" {
		errors = append(errors, FieldError{
			Field:   "card_type",
			Message: "card_type is required",
		})
	} else if req.CardType != "debit" && req.CardType != "credit" {
		errors = append(errors, FieldError{
			Field:   "card_type",
			Message: "card_type must be 'debit' or 'credit'",
		})
	}

	return errors
}

func validateCardPaymentRequest(req *CardPaymentRequest) []FieldError {
	var errors []FieldError

	if req.Amount <= 0 {
		errors = append(errors, FieldError{
			Field:   "amount",
			Message: "amount must be positive",
		})
	}

	if req.CVV == "" {
		errors = append(errors, FieldError{
			Field:   "cvv",
			Message: "CVV is required",
		})
	} else if len(req.CVV) != 3 {
		errors = append(errors, FieldError{
			Field:   "cvv",
			Message: "CVV must be exactly 3 digits",
		})
	} else if !isNumeric(req.CVV) {
		errors = append(errors, FieldError{
			Field:   "cvv",
			Message: "CVV must contain only numbers",
		})
	}

	if req.Merchant == "" {
		errors = append(errors, FieldError{
			Field:   "merchant",
			Message: "merchant is required",
		})
	} else if len(req.Merchant) > 100 {
		errors = append(errors, FieldError{
			Field:   "merchant",
			Message: "merchant name must not exceed 100 characters",
		})
	}

	return errors
}

func validateCreateCreditRequest(req *CreateCreditRequest) []FieldError {
	var errors []FieldError

	if req.AccountID == "" {
		errors = append(errors, FieldError{
			Field:   "account_id",
			Message: "account_id is required",
		})
	}

	if req.Amount <= 0 {
		errors = append(errors, FieldError{
			Field:   "amount",
			Message: "amount must be positive",
		})
	} else if req.Amount > 5000000 {
		errors = append(errors, FieldError{
			Field:   "amount",
			Message: "amount must not exceed 5,000,000",
		})
	}

	if req.TermMonths <= 0 {
		errors = append(errors, FieldError{
			Field:   "term_months",
			Message: "term_months must be positive",
		})
	} else if req.TermMonths > 360 {
		errors = append(errors, FieldError{
			Field:   "term_months",
			Message: "term_months must not exceed 360 (30 years)",
		})
	}

	if len(req.Description) > 255 {
		errors = append(errors, FieldError{
			Field:   "description",
			Message: "description must not exceed 255 characters",
		})
	}

	return errors
}

func validateMonthlyStatsRequest(req *MonthlyStatsRequest) []FieldError {
	var errors []FieldError

	if req.Year < 2020 || req.Year > 2030 {
		errors = append(errors, FieldError{
			Field:   "year",
			Message: "year must be between 2020 and 2030",
		})
	}

	if req.Month < 1 || req.Month > 12 {
		errors = append(errors, FieldError{
			Field:   "month",
			Message: "month must be between 1 and 12",
		})
	}

	return errors
}

func validateBalancePredictionRequest(req *BalancePredictionRequest) []FieldError {
	var errors []FieldError

	if req.AccountID == "" {
		errors = append(errors, FieldError{
			Field:   "account_id",
			Message: "account_id is required",
		})
	}

	if req.Days <= 0 {
		errors = append(errors, FieldError{
			Field:   "days",
			Message: "days must be positive",
		})
	} else if req.Days > 365 {
		errors = append(errors, FieldError{
			Field:   "days",
			Message: "days must not exceed 365",
		})
	}

	return errors
}

// Helper functions
func isNumeric(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

// WriteErrorResponse записывает ошибку в HTTP ответ
func WriteErrorResponse(w http.ResponseWriter, statusCode int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResponse := map[string]interface{}{
		"success": false,
		"error": map[string]string{
			"code":    strings.ToLower(strings.ReplaceAll(err.Error(), " ", "_")),
			"message": err.Error(),
		},
	}

	if validationErr, ok := err.(*ValidationError); ok {
		errorResponse["error"] = map[string]string{
			"code":    "validation_error",
			"message": validationErr.Error(),
		}
	}

	_ = json.NewEncoder(w).Encode(errorResponse)
}

// WriteSuccessResponse записывает успешный ответ
func WriteSuccessResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"success": true,
		"data":    data,
	}

	_ = json.NewEncoder(w).Encode(response)
}
