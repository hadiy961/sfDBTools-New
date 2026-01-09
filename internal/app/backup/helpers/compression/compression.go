// File : internal/app/backup/helpers/compression/compression.go
// Deskripsi : Helper untuk backup compression settings
// Author : Hadiyatna Muflihun
// Tanggal : 23 Desember 2025
// Last Modified : 9 Januari 2026

package compression

import (
	"sfdbtools/internal/app/backup/model/types_backup"
	"sfdbtools/internal/shared/compress"
)

// BuildCompressionSettings membuat CompressionSettings dari BackupDBOptions.
// Ditempatkan di domain backup (bukan shared) karena hanya dipakai fitur backup.
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
