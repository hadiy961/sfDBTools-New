package helpers

import (
	"context"
	"time"

	"sfDBTools/internal/types"
	"sfDBTools/pkg/database"
)

// SaveDatabaseDetail menyimpan detail database ke tabel database_details menggunakan stored procedure.
func SaveDatabaseDetail(ctx context.Context, client *database.Client, detail types.DatabaseDetailInfo, serverHost string, serverPort int) error {
	collectionTime, err := time.Parse("2006-01-02 15:04:05", detail.CollectionTime)
	if err != nil {
		collectionTime = time.Now()
	}

	var errorMsg *string
	if detail.Error != "" {
		errorMsg = &detail.Error
	}

	query := `CALL sp_insert_database_detail(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = client.DB().ExecContext(ctx, query,
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
