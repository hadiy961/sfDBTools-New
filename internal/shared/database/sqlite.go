// File : internal/shared/database/sqlite.go
// Deskripsi : Driver & Manager untuk SQLite Lokal (ModernC - No CGO)
// Author : Antigravity
// Tanggal : 30 April 2026

package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	_ "modernc.org/sqlite"
)

var (
	sqliteInstance *sql.DB
	sqliteOnce     sync.Once
	sqliteErr      error
)

// GetSQLite returns the global SQLite connection instance.
func GetSQLite() (*sql.DB, error) {
	if sqliteInstance == nil {
		return nil, fmt.Errorf("sqlite database not initialized, call InitSQLite first")
	}
	return sqliteInstance, nil
}

// InitSQLite menginisialisasi koneksi database SQLite dan menjalankan migrasi dasar.
func InitSQLite(dbPath string) error {
	var err error
	sqliteOnce.Do(func() {
		// Pastikan direktori ada
		dbDir := filepath.Dir(dbPath)
		if _, err = os.Stat(dbDir); os.IsNotExist(err) {
			if err = os.MkdirAll(dbDir, 0755); err != nil {
				sqliteErr = fmt.Errorf("failed to create database directory: %v", err)
				return
			}
		}

		// Buka koneksi
		sqliteInstance, sqliteErr = sql.Open("sqlite", dbPath)
		if sqliteErr != nil {
			return
		}

		// Test koneksi
		if sqliteErr = sqliteInstance.Ping(); sqliteErr != nil {
			return
		}

		// Jalankan migrasi dasar (Auto-DDL)
		if sqliteErr = runBasicMigrations(sqliteInstance); sqliteErr != nil {
			return
		}

		// Pastikan semua key setting ada
		_ = EnsureDefaultSettings()
	})

	return sqliteErr
}

// EnsureDefaultSettings memastikan semua key setting yang dibutuhkan ada di database.
func EnsureDefaultSettings() error {
	db, err := GetSQLite()
	if err != nil {
		return err
	}

	defaults := map[string][]string{
		"sync_enabled":           {"false", "cloud"},
		"sync_auto":              {"true", "cloud"},
		"sync_mode":              {"two-way", "cloud"},
		"sync_interval":          {"15", "cloud"},
		"sync_jobs_enabled":      {"true", "cloud"},
		"sync_type":              {"postgres", "cloud"},
		"sync_host":              {"", "cloud"},
		"sync_port":              {"5432", "cloud"},
		"sync_user":              {"", "cloud"},
		"sync_password":          {"", "cloud"},
		"sync_database":          {"sfdbtools_sync", "cloud"},
		"auto_update_enabled":    {"true", "cloud"},
		"check_internet_startup": {"true", "cloud"},
		"heartbeat_enabled":      {"true", "cloud"},
		"sync_key":               {"", "cloud"},
		"client_alias":           {"", "cloud"},
		"group_tag":              {"Production", "cloud"},
	}

	for k, v := range defaults {
		_, err = db.Exec("INSERT OR IGNORE INTO app_settings (key, value, category) VALUES (?, ?, ?)", k, v[0], v[1])
		if err != nil {
			return fmt.Errorf("failed to ensure setting %s: %w", k, err)
		}
	}

	return nil
}

// IsInitialized mengecek apakah database sudah ada dan tabel utama sudah terbuat.
func IsInitialized(dbPath string) bool {
	if dbPath == "" {
		return false
	}
	// 1. Cek apakah file ada
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return false
	}

	// 2. Cek apakah bisa dibuka dan ada tabel app_settings
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return false
	}
	defer db.Close()

	var name string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='app_settings'").Scan(&name)
	return err == nil
}

// runBasicMigrations menjalankan DDL awal untuk struktur database.
func runBasicMigrations(db *sql.DB) error {
	queries := []string{
		// Tabel untuk setting aplikasi (KV Store)
		`CREATE TABLE IF NOT EXISTS app_settings (
			key TEXT PRIMARY KEY,
			value TEXT,
			category TEXT,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,

		// Tabel untuk profiles database (MariaDB connection)
		`CREATE TABLE IF NOT EXISTS profiles (
			name TEXT PRIMARY KEY,
			encrypted_data TEXT,
			account_code TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,

		// Tabel untuk catalog backup
		`CREATE TABLE IF NOT EXISTS backup_catalog (
			id TEXT PRIMARY KEY,
			backup_file TEXT,
			backup_type TEXT,
			status TEXT,
			size_bytes INTEGER,
			registered_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,

		// Tabel untuk jadwal backup (pindahan dari YAML jobs)
		`CREATE TABLE IF NOT EXISTS backup_jobs (
			name TEXT PRIMARY KEY,
			enabled INTEGER DEFAULT 1,
			schedule TEXT,
			mode TEXT,
			output_mode TEXT,
			include_file TEXT,
			profile_name TEXT,
			ticket TEXT,
			output_dir TEXT,
			retention_days INTEGER,
			last_run DATETIME,
			last_status TEXT
		);`,

		// Tabel untuk riwayat sinkronisasi
		`CREATE TABLE IF NOT EXISTS sync_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			direction TEXT,
			status TEXT,
			changes_count INTEGER,
			error_msg TEXT
		);`,

		// Tabel untuk log audit
		`CREATE TABLE IF NOT EXISTS audit_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			event_type TEXT,
			details TEXT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute migration: %v (query: %s)", err, query)
		}
	}

	// Migrasi tambahan (Update skema)
	// Tambah kolom is_locked ke app_settings jika belum ada
	_, _ = db.Exec("ALTER TABLE app_settings ADD COLUMN is_locked INTEGER DEFAULT 0")
	_, _ = db.Exec("ALTER TABLE profiles ADD COLUMN updated_at DATETIME DEFAULT CURRENT_TIMESTAMP")
	_, _ = db.Exec("ALTER TABLE backup_jobs ADD COLUMN updated_at DATETIME DEFAULT CURRENT_TIMESTAMP")
	_, _ = db.Exec("ALTER TABLE audit_logs ADD COLUMN is_synced INTEGER DEFAULT 0")

	return nil
}

// GetSetting mengambil nilai setting dari tabel app_settings berdasarkan key.
func GetSetting(key string) string {
	db, err := GetSQLite()
	if err != nil {
		return ""
	}
	var value string
	err = db.QueryRow("SELECT value FROM app_settings WHERE key = ?", key).Scan(&value)
	if err != nil {
		return ""
	}
	return value
}

// GetSettingBool mengambil nilai setting sebagai boolean.
func GetSettingBool(key string) bool {
	val := GetSetting(key)
	return val == "true" || val == "1" || val == "yes"
}
