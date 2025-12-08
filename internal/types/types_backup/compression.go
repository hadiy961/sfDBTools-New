// File : internal/types/types_backup/compression.go
// Deskripsi : Compression settings struct
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-05
// Last Modified : 2025-12-05

package types_backup

import "sfDBTools/pkg/compress"

// CompressionSettings menyimpan konfigurasi compression
type CompressionSettings struct {
	Type    compress.CompressionType
	Enabled bool
	Level   int
}
