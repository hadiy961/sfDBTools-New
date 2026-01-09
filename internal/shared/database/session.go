package database

import (
	"context"
	"fmt"
)

// File: pkg/database/database_session_var.go
// Deskripsi: Helper untuk mengatur variabel sesi database (max_statement_time) secara aman
// Catatan: Gunakan fungsi ini untuk memastikan nilai sesi dikembalikan ke nilai awal.

// withVarScope adalah helper generik untuk mengatur variabel database dengan restore function.
func withVarScope(ctx context.Context, getter func(context.Context) (float64, error), setter func(context.Context, float64) error, newVal float64, errPrefix string) (restore func(context.Context) error, original float64, err error) {
	orig, err := getter(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: gagal mengambil nilai: %w", errPrefix, err)
	}
	if err := setter(ctx, newVal); err != nil {
		return nil, orig, fmt.Errorf("%s: gagal mengatur nilai: %w", errPrefix, err)
	}
	return func(rctx context.Context) error { return setter(rctx, orig) }, orig, nil
}

// WithSessionMaxStatementTime mengatur nilai SESSION max_statement_time ke newSeconds
// dan mengembalikan fungsi restore untuk mengembalikan ke nilai awal.
func WithSessionMaxStatementTime(ctx context.Context, c *Client, newSeconds float64) (restore func(context.Context) error, original float64, err error) {
	return withVarScope(ctx, c.GetMaxStatementsTime, c.SetMaxStatementsTime, newSeconds, "SESSION max_statement_time")
}
