package database

import (
	"context"
	"time"
)

// SaveDatabaseDetail menyimpan detail database ke tabel database_details menggunakan stored procedure
func (s *Client) SaveDatabaseDetail(ctx context.Context, detail DatabaseDetailInfo, serverHost string, serverPort int) error {
	// Parse collection time
	collectionTime, err := time.Parse("2006-01-02 15:04:05", detail.CollectionTime)
	if err != nil {
		collectionTime = time.Now()
	}

	// Prepare error message (NULL if no error)
	var errorMsg *string
	if detail.Error != "" {
		errorMsg = &detail.Error
	}

	// Call stored procedure
	query := `CALL sp_insert_database_detail(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = s.DB().ExecContext(ctx, query,
		detail.DatabaseName,
		serverHost,
		serverPort,
		detail.SizeBytes,
		detail.SizeHuman,
		detail.TableCount,
		detail.ProcedureCount,
		detail.FunctionCount,
		detail.ViewCount,
		detail.UserGrantCount,
		collectionTime,
		errorMsg,
		0,
	)

	return err
}
