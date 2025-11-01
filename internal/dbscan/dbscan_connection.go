package dbscan

import (
	"context"
	"fmt"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/helper"
)

// ConnectToTargetDB membuat koneksi ke database pusat untuk menyimpan hasil.
// Logika untuk mendapatkan konfigurasi dipisahkan untuk kejelasan.
func (s *Service) ConnectToTargetDB(ctx context.Context) (*database.Client, error) {
	// Gunakan konfigurasi app database dari environment (ENV_DB_HOST, dsb)
	client, err := database.ConnectToAppDatabase()
	if err != nil {
		return nil, fmt.Errorf("gagal koneksi ke target database: %w", err)
	}
	// Verify koneksi dengan ping
	if err := client.Ping(ctx); err != nil {
		client.Close()
		return nil, fmt.Errorf("gagal verifikasi koneksi: %w", err)
	}
	// ui.PrintSuccess("Koneksi ke target database berhasil")
	return client, nil
}

// getTargetDBConfig memisahkan logika pengambilan konfigurasi dari env vars.
// Ini membuat ConnectToTargetDB lebih fokus pada tugas koneksi.
func (s *Service) getTargetDBConfig() types.ServerDBConnection {
	// Ambil dari ScanOptions jika tersedia, jika tidak dari env
	conn := s.ScanOptions.TargetDB
	if conn.Host == "" {
		conn.Host = helper.GetEnvOrDefault("SFDB_DB_HOST", "localhost")
	}
	if conn.Port == 0 {
		conn.Port = helper.GetEnvOrDefaultInt("SFDB_DB_PORT", 3306)
	}
	if conn.User == "" {
		conn.User = helper.GetEnvOrDefault("SFDB_DB_USER", "root")
	}
	if conn.Password == "" {
		conn.Password = helper.GetEnvOrDefault("SFDB_DB_PASSWORD", "")
	}
	if conn.Database == "" {
		conn.Database = helper.GetEnvOrDefault("SFDB_DB_NAME", "sfDBTools")
	}

	return types.ServerDBConnection{
		Host:     conn.Host,
		Port:     conn.Port,
		User:     conn.User,
		Password: conn.Password,
		Database: conn.Database,
	}
}
