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

// WithGlobalMaxStatementTime mengatur nilai GLOBAL max_statement_time ke newSeconds
// dan mengembalikan fungsi restore untuk mengembalikan ke nilai awal.
//
// PENTING: GLOBAL scope berlaku untuk SEMUA koneksi baru ke server, termasuk:
// - mysqldump command
// - mysql CLI restore command
// - Koneksi baru dari aplikasi lain
//
// Ini berbeda dengan SESSION scope yang hanya berlaku untuk satu koneksi.
//
// Pemanggil harus:
// 1. Memiliki privilege SUPER atau SYSTEM_VARIABLES_ADMIN di database
// 2. Memanggil restore function (disarankan gunakan context.Background() agar tidak terpengaruh pembatalan)
// 3. Menangani error jika privilege tidak ada
func WithGlobalMaxStatementTime(ctx context.Context, c *Client, newSeconds float64) (restore func(context.Context) error, original float64, err error) {
	orig, err := c.GetGlobalMaxStatementTime(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal mengambil nilai GLOBAL max_statement_time: %w", err)
	}
	if err := c.SetGlobalMaxStatementTime(ctx, newSeconds); err != nil {
		return nil, orig, fmt.Errorf("gagal mengatur GLOBAL max_statement_time (periksa privilege SUPER): %w", err)
	}
	restore = func(rctx context.Context) error {
		return c.SetGlobalMaxStatementTime(rctx, orig)
	}
	return restore, orig, nil
}
