package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
)

// Config menyimpan parameter koneksi yang tidak akan berubah (immutable).
// Struct ini hanya bertanggung jawab untuk menyimpan data konfigurasi.
type Config struct {
	Host                 string
	Port                 int
	User                 string
	Password             string
	AllowNativePasswords bool
	ParseTime            bool
	Loc                  *time.Location
	Database             string        // Optional, bisa kosong
	ReadTimeout          time.Duration // Read timeout untuk long-running queries
	WriteTimeout         time.Duration // Write timeout untuk large data transfers
}

type Client struct {
	db *sql.DB
}

// DSN menghasilkan string Data Source Name (DSN) dari konfigurasi.
func (c *Config) DSN() string {
	// Default ke time.Local jika c.Loc tidak di-set
	loc := time.Local
	if c.Loc != nil {
		loc = c.Loc
	}

	cfg := mysql.Config{
		User:                 c.User,
		Passwd:               c.Password,
		Net:                  "tcp",
		Addr:                 fmt.Sprintf("%s:%d", c.Host, c.Port),
		DBName:               c.Database, // Set database name
		AllowNativePasswords: c.AllowNativePasswords,
		ParseTime:            c.ParseTime,
		Loc:                  loc,
		AllowOldPasswords:    true, // Aktifkan untuk kompatibilitas MariaDB lama
		// Aktifkan ini untuk kompatibilitas dengan beberapa konfigurasi MariaDB
		AllowCleartextPasswords: false,
		// Timeout settings untuk long-running operations (restore/backup)
		ReadTimeout:  c.ReadTimeout,  // 0 = unlimited
		WriteTimeout: c.WriteTimeout, // 0 = unlimited
	}

	return cfg.FormatDSN()
}

// NewClient membuat instance Client baru, membuka koneksi pool, dan melakukan ping.
// Ini adalah satu-satunya tempat di mana sql.Open dipanggil.
func NewClient(ctx context.Context, cfg Config, timeout time.Duration, maxOpenConns, maxIdleConns int, connMaxLifetime time.Duration) (*Client, error) {
	db, err := sql.Open("mysql", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("gagal membuka koneksi sql: %w", err)
	}

	// Atur parameter connection pool
	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(connMaxLifetime)
	// Set idle timeout to 0 (unlimited) untuk avoid premature connection closure
	// pada long-running operations seperti restore/backup
	db.SetConnMaxIdleTime(0)

	// Gunakan context dengan timeout untuk ping awal
	pingCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		// Pastikan db ditutup jika ping gagal
		_ = db.Close()

		return nil, fmt.Errorf("gagal melakukan ping ke database: %w", err)
	}

	return &Client{db: db}, nil
}

// Helper functions removed as we now use strings.Contains from standard library

// Close menutup connection pool. Wajib dipanggil saat aplikasi selesai.
func (c *Client) Close() error {
	if c == nil || c.db == nil {
		return nil
	}
	return c.db.Close()
}

// Ping memeriksa apakah koneksi ke database masih hidup menggunakan pool yang ada.
func (c *Client) Ping(ctx context.Context) error {
	return c.db.PingContext(ctx)
}

// DB mengembalikan instance *sql.DB jika diperlukan akses langsung.
func (c *Client) DB() *sql.DB {
	return c.db
}

// GetVersion mendapatkan versi server database sebagai string.
func (c *Client) GetVersion(ctx context.Context) (string, error) {
	var version string
	if err := c.db.QueryRowContext(ctx, "SELECT VERSION()").Scan(&version); err != nil {
		return "", fmt.Errorf("gagal mendapatkan versi database: %w", err)
	}
	return version, nil
}

// DatabaseExists memeriksa apakah database dengan nama tertentu ada
func (c *Client) DatabaseExists(ctx context.Context, dbName string) (bool, error) {
	query := "SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = ?"
	var name string
	err := c.db.QueryRowContext(ctx, query, dbName).Scan(&name)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}
	return name != "", nil
}

// CreateDatabase membuat database baru
func (c *Client) CreateDatabase(ctx context.Context, dbName string) error {
	// Sanitize database name (hanya alphanumeric dan underscore)
	// Untuk keamanan, gunakan parameterized query tidak bisa untuk CREATE DATABASE
	// Kita validasi nama database terlebih dahulu
	if !isValidDatabaseName(dbName) {
		return fmt.Errorf("invalid database name: %s", dbName)
	}

	query := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", dbName)
	_, err := c.db.ExecContext(ctx, query)
	return err
}

// DropDatabase menghapus database
func (c *Client) DropDatabase(ctx context.Context, dbName string) error {
	// Sanitize database name (hanya alphanumeric dan underscore)
	// Untuk keamanan, gunakan parameterized query tidak bisa untuk DROP DATABASE
	// Kita validasi nama database terlebih dahulu
	if !isValidDatabaseName(dbName) {
		return fmt.Errorf("invalid database name: %s", dbName)
	}

	query := fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", dbName)
	_, err := c.db.ExecContext(ctx, query)
	return err
}

// isValidDatabaseName memvalidasi nama database (alphanumeric dan underscore only)
func isValidDatabaseName(name string) bool {
	if name == "" || len(name) > 64 {
		return false
	}
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_') {
			return false
		}
	}
	return true
}
