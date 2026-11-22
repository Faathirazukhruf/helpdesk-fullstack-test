package config

import (
	"fmt"
	"os"
)

// Config menyimpan seluruh konfigurasi aplikasi dari environment variable.
type Config struct {
	Port       string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
	JWTSecret  string
}

// Load membaca environment variables dan mengisi nilai default bila tidak diset.
func Load() *Config {
	return &Config{
		Port:       getEnv("PORT", "8080"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "helpdesk_user"),
		DBPassword: getEnv("DB_PASSWORD", "helpdesk_secret"),
		DBName:     getEnv("DB_NAME", "helpdesk_db"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
		JWTSecret:  getEnv("JWT_SECRET", "helpdesk-secret-key-change-in-production"),
	}
}

// DSN menghasilkan PostgreSQL Data Source Name untuk koneksi database/sql.
func (c *Config) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
