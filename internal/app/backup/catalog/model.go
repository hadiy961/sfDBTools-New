package catalog

import "time"

// CatalogEntry merepresentasikan satu record backup di catalog
type CatalogEntry struct {
	ID              string    `json:"id"`
	BackupFile      string    `json:"backup_file"`
	MetadataFile    string    `json:"metadata_file"`
	DatabaseNames   []string  `json:"database_names"`
	Hostname        string    `json:"hostname"`
	BackupType      string    `json:"backup_type"`
	BackupMode      string    `json:"backup_mode"`
	BackupStatus    string    `json:"backup_status"`
	BackupTime      time.Time `json:"backup_time"`
	FileSizeBytes   int64     `json:"file_size_bytes"`
	FileSizeHuman   string    `json:"file_size_human"`
	Compressed      bool      `json:"compressed"`
	CompressionType string    `json:"compression_type,omitempty"`
	Encrypted       bool      `json:"encrypted"`
	Ticket          string    `json:"ticket,omitempty"`
	GTIDInfo        string    `json:"gtid_info,omitempty"`
	ChecksumHash    string    `json:"checksum_hash,omitempty"`
	ProfileUsed     string    `json:"profile_used,omitempty"`
	RegisteredAt    time.Time `json:"registered_at"`
}

// Catalog adalah root structure dari catalog file
type Catalog struct {
	Version   string         `json:"version"`
	UpdatedAt time.Time      `json:"updated_at"`
	Entries   []CatalogEntry `json:"entries"`
}

// QueryOptions menentukan parameter pencarian pada catalog
type QueryOptions struct {
	Database string
	Since    string
	Status   string
	Hostname string
	Limit    int
}

// DatabaseCoverage menyimpan metrik agregasi untuk satu database
type DatabaseCoverage struct {
	DatabaseName string
	LastBackup   time.Time
	BackupCount  int
	TotalSize    int64
}

// StorageTrendPoint menyimpan tren storage pada satu waktu tertentu
type StorageTrendPoint struct {
	Date        string
	SizeBytes   int64
	SizeHuman   string
	BackupCount int
}

// ReportSummary menyimpan hasil laporan agregasi catalog
type ReportSummary struct {
	Period           string
	TotalBackups     int
	TotalSizeBytes   int64
	TotalSizeHuman   string
	SuccessCount     int
	FailedCount      int
	SuccessRate      float64
	DatabaseCoverage []DatabaseCoverage
	StorageTrend     []StorageTrendPoint
}
