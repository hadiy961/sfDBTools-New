package database

import (
	"context"
	"database/sql"
	"fmt"
)

// CountDatabases menghitung jumlah database pada server
func (s *Client) CountDatabases() (int, error) {
	var count int
	row := s.db.QueryRow("SELECT COUNT(*) FROM information_schema.SCHEMATA")
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

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
	query := fmt.Sprintf("SET STATEMENT max_statement_time=0 FOR SHOW FULL TABLES FROM `%s` WHERE Table_type = 'BASE TABLE'", dbName)
	rows, err := s.DB().QueryContext(ctx, query)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		count++
	}

	if err = rows.Err(); err != nil {
		return 0, err
	}
	return count, nil
}

func (s *Client) GetProcedureCount(ctx context.Context, dbName string) (int, error) {
	// 1. Gunakan perintah SHOW PROCEDURE STATUS yang jauh lebih cepat.
	//    Filter langsung dengan klausa WHERE Db = ?.
	query := `SET STATEMENT max_statement_time=0 FOR SHOW PROCEDURE STATUS WHERE Db = ?`

	rows, err := s.DB().QueryContext(ctx, query, dbName)
	if err != nil {
		// 2. Kembalikan error agar pemanggil tahu ada masalah.
		return 0, err
	}
	defer rows.Close()

	var count int
	// Cukup hitung jumlah baris yang dikembalikan.
	for rows.Next() {
		count++
	}

	// Selalu periksa error setelah iterasi selesai.
	if err = rows.Err(); err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Client) GetFunctionCount(ctx context.Context, dbName string) (int, error) {
	// 1. Gunakan perintah SHOW FUNCTION STATUS yang jauh lebih cepat.
	query := `SET STATEMENT max_statement_time=0 FOR SHOW FUNCTION STATUS WHERE Db = ?`

	rows, err := s.DB().QueryContext(ctx, query, dbName)
	if err != nil {
		// 2. Kembalikan error yang sebenarnya agar masalah tidak tersembunyi.
		return 0, err
	}
	defer rows.Close()

	var count int
	// Hitung jumlah baris yang dikembalikan.
	for rows.Next() {
		count++
	}

	// Periksa error yang mungkin terjadi saat iterasi.
	if err = rows.Err(); err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Client) GetViewCount(ctx context.Context, dbName string) (int, error) {
	// 1. Gunakan SHOW FULL TABLES, ini jauh lebih cepat.
	//    Filter hasilnya dengan Table_type = 'VIEW'.
	query := fmt.Sprintf("SET STATEMENT max_statement_time=0 FOR SHOW FULL TABLES FROM `%s` WHERE Table_type = 'VIEW'", dbName)

	rows, err := s.DB().QueryContext(ctx, query)
	if err != nil {
		// 2. Kembalikan error agar pemanggil tahu ada masalah.
		return 0, err
	}
	defer rows.Close()

	var count int
	// Hitung jumlah baris yang cocok.
	for rows.Next() {
		count++
	}

	// Periksa error yang mungkin terjadi saat iterasi.
	if err = rows.Err(); err != nil {
		return 0, err
	}

	return count, nil
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
