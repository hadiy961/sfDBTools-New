package types

import "sfdbtools/internal/domain"

// ScanEntryConfig untuk konfigurasi scan entry point
type ScanEntryConfig struct {
	HeaderTitle string
	ShowOptions bool
	SuccessMsg  string
	LogPrefix   string
	Mode        string // "all" atau "database"
}

// ScanResult berisi hasil scanning
type ScanResult struct {
	TotalDatabases int
	SuccessCount   int
	FailedCount    int
	Duration       string
	Errors         []string
}

// ScanOptions berisi opsi untuk database scan
type ScanOptions struct {
	// Database Configuration
	ProfileInfo domain.ProfileInfo
	LocalScan   bool
	// Encryption
	Encryption struct {
		Key string
	}

	// Database Selection
	DatabaseList struct {
		File      string
		Databases []string
		UseFile   bool
	}

	// Source Database untuk mode single
	SourceDatabase string

	// Filter Options
	ExcludeSystem bool
	IncludeList   []string
	ExcludeList   []string
	ExcludeFile   string // Path ke file berisi blacklist database

	// Output Options
	DisplayResults bool

	// Internal use only
	Mode        string // "all" atau "selection" atau "single"
	ShowOptions bool   // Tampilkan opsi scanning sebelum eksekusi
}

// SystemDatabases adalah canonical list dari database sistem MySQL/MariaDB
// Menggunakan map untuk O(1) lookup performance
var SystemDatabases = map[string]struct{}{
	"information_schema": {},
	"mysql":              {},
	"performance_schema": {},
	"sys":                {},
	"innodb":             {},
}

// FailedDatabaseScanInfo berisi informasi database yang gagal
type FailedDatabaseScanInfo struct {
	DatabaseName   string
	ErrorMessage   string
	CollectionTime string
	ServerHost     string
	ServerPort     int
}

// DatabaseDetailInfo berisi informasi detail database
type DatabaseDetailInfo struct {
	DatabaseName   string `json:"database_name"`
	SizeBytes      int64  `json:"size_bytes"`
	SizeHuman      string `json:"size_human"`
	TableCount     int    `json:"table_count"`
	ProcedureCount int    `json:"procedure_count"`
	FunctionCount  int    `json:"function_count"`
	ViewCount      int    `json:"view_count"`
	UserGrantCount int    `json:"user_grant_count"`
	CollectionTime string `json:"collection_time"`
	Error          string `json:"error,omitempty"` // jika ada error saat collect
}
