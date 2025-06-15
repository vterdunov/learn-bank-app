package database

import (
	"context"
	"embed"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/stdlib"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

func (db *DB) ApplyMigrations(ctx context.Context) error {
	slog.Info("Starting database migrations")

	// Получаем стандартное database/sql соединение из pgxpool
	sqlDB := stdlib.OpenDBFromPool(db.Pool)
	defer func() {
		if err := sqlDB.Close(); err != nil {
			slog.Error("Failed to close SQL DB", slog.String("error", err.Error()))
		}
	}()

	// Создаем database driver для postgres
	databaseDriver, err := postgres.WithInstance(sqlDB, &postgres.Config{
		MigrationsTable:       "schema_migrations",
		MigrationsTableQuoted: false,
		MultiStatementEnabled: false,
		DatabaseName:          "",
		SchemaName:            "",
		StatementTimeout:      0,
		MultiStatementMaxSize: 0,
	})
	if err != nil {
		return fmt.Errorf("failed to create database driver: %w", err)
	}

	// Создаем источник миграций из встроенных файлов
	sourceDriver, err := iofs.New(migrationFiles, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create source driver: %w", err)
	}

	// Создаем migrate instance с database и source
	m, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", databaseDriver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer func() {
		if sourceErr, dbErr := m.Close(); sourceErr != nil || dbErr != nil {
			slog.Error("Failed to close migrate instance",
				slog.String("source_error", fmt.Sprintf("%v", sourceErr)),
				slog.String("db_error", fmt.Sprintf("%v", dbErr)),
			)
		}
	}()

	// Применяем миграции
	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			slog.Info("No new migrations to apply")
			return nil
		}
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	slog.Info("Database migrations completed successfully")
	return nil
}
