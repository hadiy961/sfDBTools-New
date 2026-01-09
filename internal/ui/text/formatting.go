// File : internal/ui/text/formatting.go
// Deskripsi : Helper formatting untuk output user-facing (size, duration)
// Author : Hadiyatna Muflihun
// Tanggal : 9 Januari 2026
// Last Modified : 9 Januari 2026

package text

import (
	"fmt"
	"time"

	"github.com/dustin/go-humanize"
)

// FormatFileSize mengubah ukuran file dalam bytes menjadi format yang mudah dibaca.
func FormatFileSize(bytes int64) string {
	return humanize.Bytes(uint64(bytes))
}

// FormatDuration memformat durasi menjadi string human-readable.
func FormatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%d jam %d menit %d detik", hours, minutes, seconds)
}
