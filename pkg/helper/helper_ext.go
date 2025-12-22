package helper

import (
	"path/filepath"
	"sfDBTools/pkg/compress"
	"sfDBTools/pkg/consts"
	"strings"
)

var (
	// backupExtensions mendefinisikan ekstensi file yang dianggap sebagai file backup.
	backupExtensions = []string{consts.ExtSQL, consts.ExtGzip, consts.ExtZstd, consts.ExtXz, consts.ExtZlib, consts.ExtEnc}

	// allBackupExtensions adalah list lengkap dari semua ekstensi backup yang perlu di-strip
	// untuk ekstraksi nama database dari filename
	allBackupExtensions = []string{consts.ExtEnc, consts.ExtGzip, consts.ExtZstd, consts.ExtXz, consts.ExtZlib, consts.ExtSQL}
)

// helper.TrimProfileSuffix menghapus suffix .cnf.enc dari nama jika ada.
func TrimProfileSuffix(name string) string {
	return strings.TrimSuffix(strings.TrimSuffix(name, consts.ExtEnc), consts.ExtCnf)
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

// ExtractFileExtensions mengekstrak ekstensi dari filename dan mengembalikan nama tanpa ekstensi + list ekstensi
// Contoh: "mydb.sql.gz.enc" -> ("mydb", [".sql", ".gz", ".enc"])
func ExtractFileExtensions(filename string) (string, []string) {
	nameWithoutExt := filename
	extensions := []string{}

	// Remove .enc
	if strings.HasSuffix(strings.ToLower(nameWithoutExt), consts.ExtEnc) {
		nameWithoutExt = strings.TrimSuffix(nameWithoutExt, consts.ExtEnc)
		extensions = append([]string{consts.ExtEnc}, extensions...)
	}

	// Remove compression extension
	for _, ext := range compress.SupportedCompressionExtensions() {
		if strings.HasSuffix(strings.ToLower(nameWithoutExt), ext) {
			nameWithoutExt = nameWithoutExt[:len(nameWithoutExt)-len(ext)]
			extensions = append([]string{ext}, extensions...)
			break
		}
	}

	// Remove .sql
	if strings.HasSuffix(strings.ToLower(nameWithoutExt), consts.ExtSQL) {
		nameWithoutExt = strings.TrimSuffix(nameWithoutExt, consts.ExtSQL)
		extensions = append([]string{consts.ExtSQL}, extensions...)
	}

	return nameWithoutExt, extensions
}
