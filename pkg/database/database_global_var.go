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
	// Gunakan SESSION agar perubahan hanya berlaku pada koneksi ini
	if err := c.db.QueryRowContext(ctx, "SHOW SESSION VARIABLES LIKE 'max_statement_time'").Scan(&name, &raw); err != nil {
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

// SetMaxStatementsTime mengatur nilai max_statements_time untuk sesi saat ini (SESSION scope).
func (c *Client) SetMaxStatementsTime(ctx context.Context, seconds float64) error {
	_, err := c.db.ExecContext(ctx, "SET SESSION max_statement_time = ?", seconds)
	if err != nil {
		return fmt.Errorf("gagal set SESSION max_statement_time: %w", err)
	}
	return nil
}

// GetGlobalMaxStatementTime mengambil nilai GLOBAL max_statement_time
// GLOBAL scope berlaku untuk SEMUA koneksi baru ke server
func (c *Client) GetGlobalMaxStatementTime(ctx context.Context) (float64, error) {
	var name string
	var raw sql.NullString
	// Gunakan GLOBAL untuk mendapatkan nilai global (berlaku untuk semua koneksi baru)
	if err := c.db.QueryRowContext(ctx, "SHOW GLOBAL VARIABLES LIKE 'max_statement_time'").Scan(&name, &raw); err != nil {
		return 0, fmt.Errorf("query GLOBAL max_statement_time gagal: %w", err)
	}

	if !raw.Valid || raw.String == "" {
		return 0, nil
	}

	val, err := strconv.ParseFloat(raw.String, 64)
	if err != nil {
		return 0, fmt.Errorf("parsing GLOBAL max_statement_time '%s' gagal: %w", raw.String, err)
	}
	return val, nil
}

// SetGlobalMaxStatementTime mengatur nilai GLOBAL max_statement_time
// GLOBAL scope berlaku untuk SEMUA koneksi baru ke server (termasuk mysqldump, mysql CLI, dsb)
// Catatan: Memerlukan privilege SUPER atau SYSTEM_VARIABLES_ADMIN
func (c *Client) SetGlobalMaxStatementTime(ctx context.Context, seconds float64) error {
	_, err := c.db.ExecContext(ctx, "SET GLOBAL max_statement_time = ?", seconds)
	if err != nil {
		return fmt.Errorf("gagal set GLOBAL max_statement_time (periksa privilege SUPER): %w", err)
	}
	return nil
}

// GetServerHostname mendapatkan hostname dari MySQL/MariaDB server menggunakan query SELECT @@hostname
func (c *Client) GetServerHostname(ctx context.Context) (string, error) {
	var hostname string
	if err := c.db.QueryRowContext(ctx, "SELECT @@hostname").Scan(&hostname); err != nil {
		return "", fmt.Errorf("gagal mendapatkan server hostname: %w", err)
	}
	return hostname, nil
}

// GetServerInfo mendapatkan hostname dan port dari MySQL/MariaDB server
func (c *Client) GetServerInfo(ctx context.Context) (hostname string, port int, err error) {
	if err := c.db.QueryRowContext(ctx, "SELECT @@hostname, @@port").Scan(&hostname, &port); err != nil {
		return "", 0, fmt.Errorf("gagal mendapatkan server info: %w", err)
	}
	return hostname, port, nil
}

// GetMaxAllowedPacket mendapatkan nilai max_allowed_packet dalam bytes
func (c *Client) GetMaxAllowedPacket(ctx context.Context) (int64, error) {
	var value int64
	if err := c.db.QueryRowContext(ctx, "SELECT @@max_allowed_packet").Scan(&value); err != nil {
		return 0, fmt.Errorf("gagal mendapatkan max_allowed_packet: %w", err)
	}
	return value, nil
}
