package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// PostgresConfig holds Postgres-specific connection parameters.
type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string // e.g., "disable", "require"
}

// DSN generates a string Data Source Name (DSN) for Postgres.
func (c *PostgresConfig) DSN() string {
	if c.SSLMode == "" {
		c.SSLMode = "disable"
	}
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode)
}

// NewPostgresClient creates a new Client for Postgres.
func NewPostgresClient(ctx context.Context, cfg PostgresConfig, timeout time.Duration) (*Client, error) {
	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres connection: %w", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(1 * time.Hour)

	pingCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	return &Client{db: db, driver: "postgres"}, nil
}
