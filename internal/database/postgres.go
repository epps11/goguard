package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

// Config holds database configuration
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// DB wraps the sql.DB connection
type DB struct {
	*sql.DB
}

// NewFromEnv creates a new database connection from environment variables
func NewFromEnv() (*DB, error) {
	cfg := Config{
		Host:     getEnv("GOGUARD_DB_HOST", "localhost"),
		Port:     getEnv("GOGUARD_DB_PORT", "5432"),
		User:     getEnv("GOGUARD_DB_USER", "goguard"),
		Password: getEnv("GOGUARD_DB_PASSWORD", "goguard_secret"),
		DBName:   getEnv("GOGUARD_DB_NAME", "goguard"),
		SSLMode:  getEnv("GOGUARD_DB_SSLMODE", "disable"),
	}
	return New(cfg)
}

// New creates a new database connection
func New(cfg Config) (*DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection with retry
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for i := 0; i < 5; i++ {
		if err = db.PingContext(ctx); err == nil {
			log.Info().Msg("Connected to PostgreSQL database")
			return &DB{db}, nil
		}
		log.Warn().Err(err).Int("attempt", i+1).Msg("Failed to connect to database, retrying...")
		time.Sleep(2 * time.Second)
	}

	return nil, fmt.Errorf("failed to connect to database after 5 attempts: %w", err)
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
