package global

import (
	"fmt"
	"time"

	"github.com/dustin/go-humanize"
)

// formatFileSize mengubah ukuran file dalam bytes menjadi format yang mudah dibaca.
func FormatFileSize(bytes int64) string {
	return humanize.Bytes(uint64(bytes))
}

// formatDuration memformat durasi menjadi string human-readable
func FormatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%d jam %d menit %d detik", hours, minutes, seconds)
}
