// File : internal/backup/helper/compression_helper.go
// Deskripsi : Helper functions untuk menangani compression operations
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-05
// Last Modified : 2025-12-05

package helper

import (
	"sfDBTools/internal/types/types_backup"
	"sfDBTools/pkg/compress"
)

// ConvertCompressionType mengkonversi string compression type menjadi compress.CompressionType
// Jika compression tidak enabled atau tipe invalid, mengembalikan CompressionNone
func ConvertCompressionType(enabled bool, compressionType string) compress.CompressionType {
	if !enabled {
		return compress.CompressionNone
	}
	return compress.CompressionType(compressionType)
}

// NewCompressionSettings membuat instance CompressionSettings dari konfigurasi
func NewCompressionSettings(enabled bool, compressionType string, level int) types_backup.CompressionSettings {
	return types_backup.CompressionSettings{
		Type:    ConvertCompressionType(enabled, compressionType),
		Enabled: enabled,
		Level:   level,
	}
}
