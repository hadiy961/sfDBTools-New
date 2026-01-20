package gtid

import (
	"context"
	"database/sql"
	"fmt"
	"sfdbtools/internal/app/backup/model"

	"sfdbtools/internal/shared/database"
)

// GTIDInfo menyimpan informasi GTID dari database server.
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
		return nil, fmt.Errorf("GetGTIDInfo: %w (enable binary logging)", model.ErrBinlogNotEnabled)
	}

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("gagal mendapatkan kolom: %w", err)
	}

	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	if err := rows.Scan(valuePtrs...); err != nil {
		return nil, fmt.Errorf("gagal scan hasil: %w", err)
	}

	var binlogFile string
	var binlogPos int64

	if values[0] != nil {
		binlogFile = string(values[0].([]byte))
	}
	if values[1] != nil {
		switch v := values[1].(type) {
		case int64:
			binlogPos = v
		case uint64:
			binlogPos = int64(v)
		default:
			return nil, fmt.Errorf("unexpected type for position: %T", values[1])
		}
	}

	return &GTIDInfo{MasterLogFile: binlogFile, MasterLogPos: binlogPos}, nil
}

// GetBinlogGTIDPos mengambil GTID position dari binlog file dan position.
// Fungsi ini hanya bekerja di MariaDB yang memiliki fungsi BINLOG_GTID_POS.
func GetBinlogGTIDPos(ctx context.Context, client *database.Client, binlogFile string, binlogPos int64) (string, error) {
	query := fmt.Sprintf("SELECT BINLOG_GTID_POS('%s', %d)", binlogFile, binlogPos)

	var gtidPos sql.NullString
	if err := client.DB().QueryRowContext(ctx, query).Scan(&gtidPos); err != nil {
		return "", fmt.Errorf("gagal mendapatkan BINLOG_GTID_POS: %w", err)
	}

	if !gtidPos.Valid {
		return "", fmt.Errorf("GetGTIDPos: %w", model.ErrGTIDPosNull)
	}

	return gtidPos.String, nil
}

// GetFullGTIDInfo mengambil informasi GTID lengkap termasuk GTID position.
// Fungsi ini adalah kombinasi dari GetMasterStatus dan GetBinlogGTIDPos.
func GetFullGTIDInfo(ctx context.Context, client *database.Client) (*GTIDInfo, error) {
	gtidInfo, err := GetMasterStatus(ctx, client)
	if err != nil {
		return nil, err
	}

	gtidPos, err := GetBinlogGTIDPos(ctx, client, gtidInfo.MasterLogFile, gtidInfo.MasterLogPos)
	if err != nil {
		// Jika gagal mendapatkan GTID position, kembalikan info tanpa GTID.
		return gtidInfo, nil
	}

	gtidInfo.GTIDBinlog = gtidPos
	return gtidInfo, nil
}
