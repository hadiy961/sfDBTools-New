// File : internal/backup/helpers/gtid.go
// Deskripsi : Fungsi untuk mengambil informasi GTID dari MariaDB/MySQL (Backup-specific)
// Author : Hadiyatna Muflihun
// Tanggal : 22 Desember 2024
// Last Modified : 22 Desember 2024
// Moved from: pkg/database/database_gtid.go

package helpers

import (
	"context"
	"database/sql"
	"fmt"
	"sfDBTools/pkg/database"
)

// GTIDInfo menyimpan informasi GTID dari database server
type GTIDInfo struct {
	MasterLogFile string // Nama file binlog master
	MasterLogPos  int64  // Posisi di dalam binlog
	GTIDBinlog    string // GTID position dalam format MariaDB
}

// GetMasterStatus mengambil informasi MASTER STATUS dari database server.
// Fungsi ini mengembalikan informasi binlog file dan position.
func GetMasterStatus(ctx context.Context, client *database.Client) (*GTIDInfo, error) {
	query := "SHOW MASTER STATUS"

	rows, err := client.DB().QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("gagal menjalankan SHOW MASTER STATUS: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("SHOW MASTER STATUS tidak mengembalikan hasil. Pastikan binary logging diaktifkan")
	}

	// Get column names to determine structure
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("gagal mendapatkan kolom: %w", err)
	}

	// Prepare slice for scanning based on column count
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	err = rows.Scan(valuePtrs...)
	if err != nil {
		return nil, fmt.Errorf("gagal scan hasil: %w", err)
	}

	// First two columns are always File and Position
	var binlogFile string
	var binlogPos int64

	if values[0] != nil {
		binlogFile = string(values[0].([]byte))
	}
	if values[1] != nil {
		// Handle both int64 and uint64 types (depends on MySQL/MariaDB version)
		switch v := values[1].(type) {
		case int64:
			binlogPos = v
		case uint64:
			binlogPos = int64(v)
		default:
			return nil, fmt.Errorf("unexpected type for position: %T", values[1])
		}
	}

	gtidInfo := &GTIDInfo{
		MasterLogFile: binlogFile,
		MasterLogPos:  binlogPos,
	}

	return gtidInfo, nil
}

// GetBinlogGTIDPos mengambil GTID position dari binlog file dan position.
// Fungsi ini hanya bekerja di MariaDB yang memiliki fungsi BINLOG_GTID_POS.
func GetBinlogGTIDPos(ctx context.Context, client *database.Client, binlogFile string, binlogPos int64) (string, error) {
	query := fmt.Sprintf("SELECT BINLOG_GTID_POS('%s', %d)", binlogFile, binlogPos)

	var gtidPos sql.NullString
	err := client.DB().QueryRowContext(ctx, query).Scan(&gtidPos)

	if err != nil {
		return "", fmt.Errorf("gagal mendapatkan BINLOG_GTID_POS: %w", err)
	}

	if !gtidPos.Valid {
		return "", fmt.Errorf("BINLOG_GTID_POS mengembalikan NULL")
	}

	return gtidPos.String, nil
}

// GetFullGTIDInfo mengambil informasi GTID lengkap termasuk GTID position.
// Fungsi ini adalah kombinasi dari GetMasterStatus dan GetBinlogGTIDPos.
func GetFullGTIDInfo(ctx context.Context, client *database.Client) (*GTIDInfo, error) {
	// Dapatkan master status terlebih dahulu
	gtidInfo, err := GetMasterStatus(ctx, client)
	if err != nil {
		return nil, err
	}

	// Dapatkan GTID position dari binlog
	gtidPos, err := GetBinlogGTIDPos(ctx, client, gtidInfo.MasterLogFile, gtidInfo.MasterLogPos)
	if err != nil {
		// Jika gagal mendapatkan GTID position, kembalikan info tanpa GTID
		// (ini bisa terjadi di MySQL yang tidak support BINLOG_GTID_POS)
		return gtidInfo, nil
	}

	gtidInfo.GTIDBinlog = gtidPos
	return gtidInfo, nil
}
