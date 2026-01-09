package database

import (
	"context"
	"database/sql"
	"fmt"
)

// Database metric collection functions

// GetDatabaseSize menghitung total ukuran data + indeks untuk sebuah database.
// Ini adalah alternatif yang jauh lebih cepat daripada query ke information_schema.
func (s *Client) GetDatabaseSize(ctx context.Context, dbName string) (int64, error) {
	// Use information_schema aggregation to reliably compute total size.
	// This is portable across MySQL/MariaDB versions and returns a single row.
	var total sql.NullInt64
	query := `SET STATEMENT max_statement_time=0 FOR SELECT IFNULL(SUM(DATA_LENGTH + INDEX_LENGTH), 0) FROM information_schema.TABLES WHERE TABLE_SCHEMA = ?`
	err := s.DB().QueryRowContext(ctx, query, dbName).Scan(&total)
	if err != nil {
		return 0, err
	}
	return total.Int64, nil
}

func (s *Client) GetTableCount(ctx context.Context, dbName string) (int, error) {
	return s.countRows(ctx, fmt.Sprintf("SET STATEMENT max_statement_time=0 FOR SHOW FULL TABLES FROM `%s` WHERE Table_type = 'BASE TABLE'", dbName))
}

// countRows adalah helper untuk menghitung jumlah baris dari query.
func (s *Client) countRows(ctx context.Context, query string, args ...interface{}) (int, error) {
	rows, err := s.DB().QueryContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
	}
	return count, rows.Err()
}

func (s *Client) GetProcedureCount(ctx context.Context, dbName string) (int, error) {
	return s.countRows(ctx, `SET STATEMENT max_statement_time=0 FOR SHOW PROCEDURE STATUS WHERE Db = ?`, dbName)
}

func (s *Client) GetFunctionCount(ctx context.Context, dbName string) (int, error) {
	return s.countRows(ctx, `SET STATEMENT max_statement_time=0 FOR SHOW FUNCTION STATUS WHERE Db = ?`, dbName)
}

func (s *Client) GetViewCount(ctx context.Context, dbName string) (int, error) {
	return s.countRows(ctx, fmt.Sprintf("SET STATEMENT max_statement_time=0 FOR SHOW FULL TABLES FROM `%s` WHERE Table_type = 'VIEW'", dbName))
}

func (s *Client) GetUserGrantCount(ctx context.Context, dbName string) (int, error) {
	// 1. Kueri langsung ke tabel mysql.db yang jauh lebih cepat.
	//    Kita hitung kombinasi unik dari User dan Host.
	query := `
		SET STATEMENT max_statement_time=0 FOR
		SELECT COUNT(DISTINCT CONCAT(User, '@', Host)) 
		FROM mysql.db 
		WHERE Db = ?`

	var count int
	// Karena query COUNT(*) mengembalikan satu baris, QueryRowContext cocok di sini.
	err := s.DB().QueryRowContext(ctx, query, dbName).Scan(&count)
	if err != nil {
		// 2. Kembalikan error yang sebenarnya, jangan disembunyikan.
		return 0, err
	}
	return count, nil
}
