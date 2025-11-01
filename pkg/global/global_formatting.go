package global

import "github.com/dustin/go-humanize"

// formatFileSize mengubah ukuran file dalam bytes menjadi format yang mudah dibaca.
func FormatFileSize(bytes int64) string {
	return humanize.Bytes(uint64(bytes))
}
