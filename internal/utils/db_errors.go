package utils

import (
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Database error types
var (
	// ErrDuplicateKey возвращается при нарушении уникальности
	ErrDuplicateKey = errors.New("duplicate key violation")
	// ErrForeignKeyViolation возвращается при нарушении внешнего ключа
	ErrForeignKeyViolation = errors.New("foreign key violation")
	// ErrNotNull возвращается при попытке вставить NULL в NOT NULL поле
	ErrNotNull = errors.New("not null violation")
	// ErrCheckViolation возвращается при нарушении CHECK constraint
	ErrCheckViolation = errors.New("check constraint violation")
	// ErrDataTypeMismatch возвращается при несоответствии типов данных
	ErrDataTypeMismatch = errors.New("data type mismatch")
	// ErrConnectionFailed возвращается при ошибке подключения к БД
	ErrConnectionFailed = errors.New("database connection failed")
	// ErrInvalidInput возвращается при некорректных входных данных
	ErrInvalidInput = errors.New("invalid input data")
	// ErrRecordNotFound возвращается когда запись не найдена
	ErrRecordNotFound = errors.New("record not found")
	// ErrTransactionFailed возвращается при ошибке транзакции
	ErrTransactionFailed = errors.New("transaction failed")
)

// PostgreSQL error codes
const (
	// Constraint violations
	PgErrUniqueViolation     = "23505"
	PgErrForeignKeyViolation = "23503"
	PgErrNotNullViolation    = "23502"
	PgErrCheckViolation      = "23514"

	// Data type errors
	PgErrInvalidTextRepresentation = "22P02"
	PgErrNumericValueOutOfRange    = "22003"
	PgErrDatetimeFieldOverflow     = "22008"

	// Connection errors
	PgErrConnectionException      = "08000"
	PgErrConnectionFailure        = "08006"
	PgErrSQLClientUnableToConnect = "08001"
)

// DBError представляет ошибку базы данных с дополнительной информацией
type DBError struct {
	Code       string
	Message    string
	Detail     string
	Table      string
	Column     string
	Constraint string
	Original   error
}

func (e *DBError) Error() string {
	if e.Detail != "" {
		return e.Message + ": " + e.Detail
	}
	return e.Message
}

func (e *DBError) Unwrap() error {
	return e.Original
}

// ParseDBError анализирует ошибку PostgreSQL и возвращает типизированную ошибку
func ParseDBError(err error) error {
	if err == nil {
		return nil
	}

	// Проверяем на ErrNoRows
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrRecordNotFound
	}

	// Проверяем на pgconn.PgError
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		dbErr := &DBError{
			Code:       pgErr.Code,
			Message:    pgErr.Message,
			Detail:     pgErr.Detail,
			Table:      pgErr.TableName,
			Column:     pgErr.ColumnName,
			Constraint: pgErr.ConstraintName,
			Original:   err,
		}

		switch pgErr.Code {
		case PgErrUniqueViolation:
			return &UniqueViolationError{
				DBError:    *dbErr,
				Field:      extractFieldFromConstraint(pgErr.ConstraintName),
				Constraint: pgErr.ConstraintName,
			}
		case PgErrForeignKeyViolation:
			return &ForeignKeyViolationError{
				DBError:          *dbErr,
				ReferencedTable:  pgErr.TableName,
				ReferencedColumn: pgErr.ColumnName,
			}
		case PgErrNotNullViolation:
			return &NotNullViolationError{
				DBError: *dbErr,
				Column:  pgErr.ColumnName,
			}
		case PgErrCheckViolation:
			return &CheckViolationError{
				DBError:    *dbErr,
				Constraint: pgErr.ConstraintName,
			}
		case PgErrInvalidTextRepresentation,
			PgErrNumericValueOutOfRange,
			PgErrDatetimeFieldOverflow:
			return &DataTypeMismatchError{
				DBError: *dbErr,
				Column:  pgErr.ColumnName,
			}
		case PgErrConnectionException,
			PgErrConnectionFailure,
			PgErrSQLClientUnableToConnect:
			return &ConnectionError{
				DBError: *dbErr,
			}
		default:
			return dbErr
		}
	}

	// Проверяем на строковые ошибки
	errStr := strings.ToLower(err.Error())

	if strings.Contains(errStr, "connection") {
		return ErrConnectionFailed
	}

	if strings.Contains(errStr, "duplicate") || strings.Contains(errStr, "unique") {
		return ErrDuplicateKey
	}

	if strings.Contains(errStr, "foreign key") {
		return ErrForeignKeyViolation
	}

	return err
}

// UniqueViolationError специфичная ошибка нарушения уникальности
type UniqueViolationError struct {
	DBError
	Field      string
	Constraint string
}

func (e *UniqueViolationError) Error() string {
	if e.Field != "" {
		return e.Field + " already exists"
	}
	return "unique constraint violation: " + e.Constraint
}

// ForeignKeyViolationError специфичная ошибка нарушения внешнего ключа
type ForeignKeyViolationError struct {
	DBError
	ReferencedTable  string
	ReferencedColumn string
}

func (e *ForeignKeyViolationError) Error() string {
	return "foreign key constraint violation: referenced record does not exist"
}

// NotNullViolationError специфичная ошибка NULL в NOT NULL поле
type NotNullViolationError struct {
	DBError
	Column string
}

func (e *NotNullViolationError) Error() string {
	return "field " + e.Column + " cannot be null"
}

// CheckViolationError специфичная ошибка нарушения CHECK constraint
type CheckViolationError struct {
	DBError
	Constraint string
}

func (e *CheckViolationError) Error() string {
	return "check constraint violation: " + e.Constraint
}

// DataTypeMismatchError специфичная ошибка несоответствия типов данных
type DataTypeMismatchError struct {
	DBError
	Column string
}

func (e *DataTypeMismatchError) Error() string {
	return "invalid data type for field " + e.Column
}

// ConnectionError специфичная ошибка подключения к БД
type ConnectionError struct {
	DBError
}

func (e *ConnectionError) Error() string {
	return "database connection error: " + e.Message
}

// extractFieldFromConstraint извлекает название поля из имени constraint
func extractFieldFromConstraint(constraintName string) string {
	// Примеры: users_email_key -> email, users_username_key -> username
	parts := strings.Split(constraintName, "_")
	if len(parts) >= 3 && parts[len(parts)-1] == "key" {
		return parts[len(parts)-2]
	}

	// Если не удалось извлечь, возвращаем весь constraint
	return constraintName
}

// IsUniqueViolation проверяет является ли ошибка нарушением уникальности
func IsUniqueViolation(err error) bool {
	var uniqueErr *UniqueViolationError
	return errors.As(err, &uniqueErr) || errors.Is(err, ErrDuplicateKey)
}

// IsForeignKeyViolation проверяет является ли ошибка нарушением внешнего ключа
func IsForeignKeyViolation(err error) bool {
	var fkErr *ForeignKeyViolationError
	return errors.As(err, &fkErr) || errors.Is(err, ErrForeignKeyViolation)
}

// IsNotNullViolation проверяет является ли ошибка нарушением NOT NULL
func IsNotNullViolation(err error) bool {
	var notNullErr *NotNullViolationError
	return errors.As(err, &notNullErr) || errors.Is(err, ErrNotNull)
}

// IsConnectionError проверяет является ли ошибка проблемой подключения
func IsConnectionError(err error) bool {
	var connErr *ConnectionError
	return errors.As(err, &connErr) || errors.Is(err, ErrConnectionFailed)
}

// IsRecordNotFound проверяет является ли ошибка "запись не найдена"
func IsRecordNotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows) || errors.Is(err, ErrRecordNotFound)
}

// WrapDBError оборачивает ошибку БД с дополнительным контекстом
func WrapDBError(err error, operation string) error {
	if err == nil {
		return nil
	}

	parsedErr := ParseDBError(err)

	// Добавляем контекст операции
	switch e := parsedErr.(type) {
	case *UniqueViolationError:
		e.Message = operation + ": " + e.Message
		return e
	case *ForeignKeyViolationError:
		e.Message = operation + ": " + e.Message
		return e
	case *DBError:
		e.Message = operation + ": " + e.Message
		return e
	default:
		return errors.New(operation + ": " + parsedErr.Error())
	}
}
