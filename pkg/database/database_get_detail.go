package database

import (
	"context"
	"database/sql"
	"fmt"
	"sfDBTools/internal/types"
)

// GetSingleDatabaseDetail mengambil detail database dari tabel database_details
// berdasarkan database_name, server_host, dan server_port
func (c *Client) GetSingleDatabaseDetail(ctx context.Context, databaseName, serverHost string, serverPort int) (*types.DatabaseDetail, error) {
	query := `
		SELECT 
			database_name,
			size_bytes,
			size_human,
			table_count,
			procedure_count,
			function_count,
			view_count,
			user_grant_count,
			collection_time,
			error_message,
			created_at,
			updated_at
		FROM database_details
		WHERE database_name = ? 
			AND server_host = ? 
			AND server_port = ? AND error_message IS NULL
		ORDER BY collection_time DESC
		LIMIT 1
	`

	var detail types.DatabaseDetail
	var errorMessage sql.NullString

	err := c.db.QueryRowContext(ctx, query, databaseName, serverHost, serverPort).Scan(
		&detail.DatabaseName,
		&detail.SizeBytes,
		&detail.SizeHuman,
		&detail.TableCount,
		&detail.ProcedureCount,
		&detail.FunctionCount,
		&detail.ViewCount,
		&detail.UserGrantCount,
		&detail.CollectionTime,
		&errorMessage,
		&detail.CreatedAt,
		&detail.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("detail database tidak ditemukan untuk database '%s'", databaseName)
		}
		return nil, fmt.Errorf("gagal mengambil detail database: %w", err)
	}

	// Handle nullable error_message
	if errorMessage.Valid {
		detail.ErrorMessage = &errorMessage.String
	}

	return &detail, nil
}
