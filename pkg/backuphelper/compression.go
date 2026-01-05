// File : pkg/backuphelper/compression.go
// Deskripsi : Shared helper untuk backup compression settings
// Author : Hadiyatna Muflihun
// Tanggal : 23 Desember 2025
// Last Modified : 23 Desember 2025

package backuphelper

import (
	"sfDBTools/internal/app/backup/model/types_backup"
	"sfDBTools/pkg/compress"
)

// BuildCompressionSettings membuat CompressionSettings dari BackupDBOptions
// Shared function untuk menghindari duplikasi di service dan selection packages
func BuildCompressionSettings(opts *types_backup.BackupDBOptions) types_backup.CompressionSettings {
	compressionType := opts.Compression.Type
	if !opts.Compression.Enabled {
		compressionType = ""
	}
	return types_backup.CompressionSettings{
		Type:    compress.CompressionType(compressionType),
		Enabled: opts.Compression.Enabled,
		Level:   opts.Compression.Level,
	}
}
