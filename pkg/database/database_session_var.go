package database

import (
	"context"
	"fmt"
)

// File: pkg/database/database_session_var.go
// Deskripsi: Helper untuk mengatur variabel sesi database (max_statement_time) secara aman
// Catatan: Gunakan fungsi ini untuk memastikan nilai sesi dikembalikan ke nilai awal.

// WithSessionMaxStatementTime mengatur nilai SESSION max_statement_time ke newSeconds
// dan mengembalikan fungsi restore untuk mengembalikan ke nilai awal. Pemanggil
// bertanggung jawab untuk memanggil fungsi restore (disarankan gunakan context.Background()
// agar tidak terpengaruh pembatalan).
func WithSessionMaxStatementTime(ctx context.Context, c *Client, newSeconds float64) (restore func(context.Context) error, original float64, err error) {
	orig, err := c.GetMaxStatementsTime(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal mengambil nilai max_statement_time sesi: %w", err)
	}
	if err := c.SetMaxStatementsTime(ctx, newSeconds); err != nil {
		return nil, orig, fmt.Errorf("gagal mengatur max_statement_time sesi: %w", err)
	}
	restore = func(rctx context.Context) error {
		return c.SetMaxStatementsTime(rctx, orig)
	}
	return restore, orig, nil
}
