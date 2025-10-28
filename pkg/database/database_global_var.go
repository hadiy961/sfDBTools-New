package database

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
)

// GetMaxStatementsTime adalah helper internal untuk mengambil nilai @@max_statement_time.
func (c *Client) GetMaxStatementsTime(ctx context.Context) (float64, error) {
	var name string
	var raw sql.NullString
	if err := c.db.QueryRowContext(ctx, "SHOW GLOBAL VARIABLES LIKE 'max_statement_time'").Scan(&name, &raw); err != nil {
		// Jika tidak ada baris, MariaDB biasanya mengembalikan nilai default,
		// jadi error ini jarang terjadi kecuali ada masalah koneksi.
		return 0, fmt.Errorf("query max_statement_time gagal: %w", err)
	}

	if !raw.Valid || raw.String == "" {
		// Ini berarti nilainya tidak di-set atau 0
		return 0, nil
	}

	val, err := strconv.ParseFloat(raw.String, 64)
	if err != nil {
		return 0, fmt.Errorf("parsing max_statement_time '%s' gagal: %w", raw.String, err)
	}
	return val, nil
}

// SetMaxStatementsTime mengatur nilai max_statements_time untuk sesi saat ini.
func (c *Client) SetMaxStatementsTime(ctx context.Context, seconds float64) error {
	_, err := c.db.ExecContext(ctx, "SET GLOBAL max_statement_time = ?", seconds)
	if err != nil {
		return fmt.Errorf("gagal set GLOBAL max_statement_time: %w", err)
	}
	return nil
}
