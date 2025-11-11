package helper

import (
	"path/filepath"
	"strings"
)

var (
	// backupExtensions mendefinisikan ekstensi file yang dianggap sebagai file backup.
	backupExtensions = []string{".sql", ".gz", ".zst", ".lz4", ".enc"}

	// allBackupExtensions adalah list lengkap dari semua ekstensi backup yang perlu di-strip
	// untuk ekstraksi nama database dari filename
	allBackupExtensions = []string{".enc", ".gz", ".zst", ".xz", ".zlib", ".lz4", ".sql"}
)

// helper.TrimProfileSuffix menghapus suffix .cnf.enc dari nama jika ada.
func TrimProfileSuffix(name string) string {
	return strings.TrimSuffix(strings.TrimSuffix(name, ".enc"), ".cnf")
}

// IsBackupFile memeriksa apakah sebuah file dianggap sebagai file backup berdasarkan ekstensinya.
func IsBackupFile(filename string) bool {
	lowerFilename := strings.ToLower(filename)
	for _, ext := range backupExtensions {
		if strings.HasSuffix(lowerFilename, ext) {
			return true
		}
	}
	return false
}

// StripAllBackupExtensions menghilangkan semua ekstensi backup dari filename
// Mengembalikan base filename tanpa ekstensi backup
// Contoh: "mydb_20251110.sql.gz.enc" -> "mydb_20251110"
func StripAllBackupExtensions(filename string) string {
	base := filepath.Base(filename)

	// Loop untuk menghilangkan semua ekstensi yang match
	// Menggunakan loop karena file bisa punya multiple extensions (misal: .sql.gz.enc)
	changed := true
	for changed {
		changed = false
		lower := strings.ToLower(base)
		for _, ext := range allBackupExtensions {
			if strings.HasSuffix(lower, ext) {
				base = base[:len(base)-len(ext)]
				changed = true
				break
			}
		}
	}

	return base
}
