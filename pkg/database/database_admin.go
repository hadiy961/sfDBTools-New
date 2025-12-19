package database

import (
	"context"
	"database/sql"
	"fmt"
	"sfDBTools/pkg/ui"
)

// CheckDatabaseExists mengecek apakah database sudah ada
func (c *Client) CheckDatabaseExists(ctx context.Context, dbName string) (bool, error) {
	// Use retry-capable query to avoid transient invalid connection errors
	query := "SELECT SCHEMA_NAME FROM information_schema.SCHEMATA WHERE SCHEMA_NAME = ? LIMIT 1"
	rows, err := c.QueryContextWithRetry(ctx, query, dbName)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("gagal mengecek database: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		var schemaName string
		if scanErr := rows.Scan(&schemaName); scanErr != nil {
			return false, fmt.Errorf("gagal membaca hasil cek database: %w", scanErr)
		}
		return schemaName != "", nil
	}

	return false, nil
}

// DropDatabase menghapus database
func (c *Client) DropDatabase(ctx context.Context, dbName string) error {
	spin := ui.NewSpinnerWithElapsed(fmt.Sprintf("Drop database %s", dbName))
	spin.Start()
	defer spin.Stop()

	query := fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", dbName)
	_, err := c.ExecContextWithRetry(ctx, query)
	if err != nil {
		return fmt.Errorf("gagal drop database: %w", err)
	}

	return nil
}

// CreateDatabaseIfNotExists membuat database jika belum ada
func (c *Client) CreateDatabaseIfNotExists(ctx context.Context, dbName string) error {
	// Check manually first to avoid creating if exists (though IF NOT EXISTS handles it, explicit check is sometimes safer for logging/logic)
	// But IF NOT EXISTS is standard. The original code checked first.
	exists, err := c.CheckDatabaseExists(ctx, dbName)
	if err != nil {
		return err
	}

	if !exists {
		createQuery := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", dbName)
		if _, err := c.ExecContextWithRetry(ctx, createQuery); err != nil {
			return fmt.Errorf("gagal membuat database: %w", err)
		}
	}

	return nil
}

// GetNonSystemDatabases mendapatkan list database dari server mengecualikan system database
func (c *Client) GetNonSystemDatabases(ctx context.Context) ([]string, error) {
	query := "SHOW DATABASES"
	rows, err := c.QueryContextWithRetry(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("gagal query databases: %w", err)
	}
	defer rows.Close()

	var databases []string

	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err != nil {
			continue
		}

		if !IsSystemDatabase(dbName) {
			databases = append(databases, dbName)
		}
	}

	return databases, nil
}
