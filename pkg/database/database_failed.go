// File : pkg/database/database_failed.go
// Deskripsi : Fungsi untuk mendapatkan list database yang gagal di-scan
// Author : Hadiyatna Muflihun
// Tanggal : 16 Oktober 2025
// Last Modified : 16 Oktober 2025

package database

import (
	"context"
	"database/sql"
	"fmt"
)

// FailedDatabaseInfo berisi informasi database yang gagal
type FailedDatabaseInfo struct {
	DatabaseName   string
	ErrorMessage   string
	CollectionTime string
	ServerHost     string
	ServerPort     int
}

// GetFailedDatabases mengambil list database yang gagal di-scan (error_message IS NOT NULL)
// dari table database_details
func GetFailedDatabases(ctx context.Context, client *Client) ([]FailedDatabaseInfo, error) {
	query := `
		SELECT 
			database_name,
			COALESCE(error_message, '') as error_message,
			COALESCE(collection_time, NOW()) as collection_time,
			COALESCE(server_host, '') as server_host,
			COALESCE(server_port, 0) as server_port
		FROM database_details
		WHERE error_message IS NOT NULL
		ORDER BY collection_time DESC
	`

	rows, err := client.DB().QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("gagal query failed databases: %w", err)
	}
	defer rows.Close()

	var failedDatabases []FailedDatabaseInfo
	for rows.Next() {
		var info FailedDatabaseInfo
		var collectionTime sql.NullString

		err := rows.Scan(
			&info.DatabaseName,
			&info.ErrorMessage,
			&collectionTime,
			&info.ServerHost,
			&info.ServerPort,
		)
		if err != nil {
			return nil, fmt.Errorf("gagal scan row: %w", err)
		}

		if collectionTime.Valid {
			info.CollectionTime = collectionTime.String
		}

		failedDatabases = append(failedDatabases, info)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error saat iterasi rows: %w", err)
	}

	return failedDatabases, nil
}

// GetFailedDatabaseNames mengambil hanya nama database yang gagal
func GetFailedDatabaseNames(ctx context.Context, client *Client) ([]string, error) {
	failedDatabases, err := GetFailedDatabases(ctx, client)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(failedDatabases))
	for _, db := range failedDatabases {
		names = append(names, db.DatabaseName)
	}

	return names, nil
}

// GetFailedDatabaseCount menghitung jumlah database yang gagal
func GetFailedDatabaseCount(ctx context.Context, client *Client) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM database_details
		WHERE error_message IS NOT NULL
	`

	var count int
	err := client.DB().QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("gagal query failed database count: %w", err)
	}

	return count, nil
}
