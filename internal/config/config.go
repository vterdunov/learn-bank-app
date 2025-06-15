package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	SMTP     SMTPConfig
	CBR      CBRConfig
}

type ServerConfig struct {
	Port string
	Host string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	Database string
	Username string
	Password string
	SSLMode  string
}

type JWTConfig struct {
	Secret string
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
}

type CBRConfig struct {
	ServiceURL string
	BankMargin float64
}

func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Host: getEnvString("SERVER_HOST", "localhost"),
			Port: getEnvString("SERVER_PORT", "8080"),
		},
		Database: DatabaseConfig{
			Host:     getEnvString("DB_HOST", "localhost"),
			Port:     getEnvString("DB_PORT", "5432"),
			Database: getEnvString("DB_NAME", "bankapp"),
			Username: getEnvString("DB_USER", "postgres"),
			Password: getEnvString("DB_PASSWORD", "postgres"),
			SSLMode:  getEnvString("DB_SSL_MODE", "disable"),
		},
		JWT: JWTConfig{
			Secret: getEnvString("JWT_SECRET", ""),
		},
		SMTP: SMTPConfig{
			Host:     getEnvString("SMTP_HOST", ""),
			Port:     getEnvInt("SMTP_PORT", 587),
			Username: getEnvString("SMTP_USERNAME", ""),
			Password: getEnvString("SMTP_PASSWORD", ""),
		},
		CBR: CBRConfig{
			ServiceURL: getEnvString("CBR_SERVICE_URL", "https://www.cbr.ru/DailyInfoWebServ/DailyInfo.asmx"),
			BankMargin: getEnvFloat("CBR_BANK_MARGIN", 5.0),
		},
	}

	if cfg.JWT.Secret == "" {
		return nil, fmt.Errorf("JWT_SECRET environment variable is required")
	}

	return cfg, nil
}

func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}
