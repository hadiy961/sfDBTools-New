// File : internal/types/types_backup/metadata.go
// Deskripsi : Metadata config structs (tanpa method)
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-05
// Last Modified : 2025-12-05

package types_backup

import (
	"sfDBTools/internal/applog"
	"time"
)

// MetadataConfig menyimpan parameter untuk generate metadata
type MetadataConfig struct {
	BackupFile      string
	BackupType      string // "combined", "separated"
	DatabaseNames   []string
	Hostname        string
	FileSize        int64
	Compressed      bool
	CompressionType string
	Encrypted       bool
	BackupStatus    string
	Warnings        []string
	StderrOutput    string
	Duration        time.Duration
	StartTime       time.Time
	EndTime         time.Time
	Logger          applog.Logger
}
