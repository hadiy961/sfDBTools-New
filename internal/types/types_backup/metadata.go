// File : internal/types/types_backup/metadata.go
// Deskripsi : Metadata config structs (tanpa method)
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-05
// Last Modified : 2026-01-05
package types_backup

import (
	"sfDBTools/internal/services/log"
	"time"
)

// MetadataConfig menyimpan parameter untuk generate metadata
type MetadataConfig struct {
	BackupFile        string
	BackupType        string // "combined", "separated", "all"
	DatabaseNames     []string
	ExcludedDatabases []string // List database yang dikecualikan (untuk mode 'all')
	Hostname          string
	FileSize          int64
	Compressed        bool
	CompressionType   string
	Encrypted         bool
	BackupStatus      string
	Warnings          []string
	StderrOutput      string
	Duration          time.Duration
	StartTime         time.Time
	EndTime           time.Time
	GTIDInfo          string
	Logger            applog.Logger
	// Replication information
	ReplicationUser     string
	ReplicationPassword string
	SourceHost          string
	SourcePort          int
	// Additional files
	UserGrantsFile string
	// Version information
	MysqldumpVersion string
	MariaDBVersion   string
	// Ticket information
	Ticket string
	// Backup options
	ExcludeData bool // Jika true, hanya backup struktur database (tanpa data)
}
