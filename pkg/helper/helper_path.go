package helper

import (
	"fmt"
	"path/filepath"
	"sfDBTools/pkg/compress"
	"strings"
	"time"
)

// PathPatternReplacer menyimpan nilai-nilai untuk menggantikan pattern dalam path/filename
type PathPatternReplacer struct {
	Database       string
	Timestamp      time.Time
	Hostname       string // Hostname dari database server
	Year           string
	Month          string
	Day            string
	Hour           string
	Minute         string
	Second         string
	CompressionExt string // Ekstensi kompresi (.gz, .zst, etc)
	EncryptionExt  string // Ekstensi enkripsi (.enc)
	IsFilename     bool   // true jika untuk filename (butuh ekstensi), false jika untuk directory path
}

// NewPathPatternReplacer membuat instance baru PathPatternReplacer dengan timestamp saat ini
func NewPathPatternReplacer(database string, hostname string, compressionType compress.CompressionType, encrypted bool, isFilename bool) (*PathPatternReplacer, error) {
	// Gunakan hostname yang diberikan (dari database server)
	if hostname == "" {
		hostname = "unknown"
	}

	now := time.Now()

	// Tentukan ekstensi kompresi
	compressionExt := ""
	if compressionType != compress.CompressionNone && compressionType != "" {
		compressionExt = compress.GetFileExtension(compressionType)
	}
	// Tentukan ekstensi enkripsi
	encryptionExt := ""
	if encrypted {
		encryptionExt = ".enc"
	}

	return &PathPatternReplacer{
		Database:       database,
		Timestamp:      now,
		Hostname:       hostname,
		Year:           now.Format("2006"),
		Month:          now.Format("01"),
		Day:            now.Format("02"),
		Hour:           now.Format("15"),
		Minute:         now.Format("04"),
		Second:         now.Format("05"),
		CompressionExt: compressionExt,
		EncryptionExt:  encryptionExt,
		IsFilename:     isFilename,
	}, nil
}

// ReplacePattern mengganti semua pattern dalam string dengan nilai yang sesuai
func (r *PathPatternReplacer) ReplacePattern(pattern string) string {
	result := pattern

	// Daftar replacements
	replacements := map[string]string{
		"{database}":  r.Database,
		"{timestamp}": r.Timestamp.Format("20060102_150405"),
		"{hostname}":  r.Hostname,
		"{year}":      r.Year,
		"{month}":     r.Month,
		"{day}":       r.Day,
		"{hour}":      r.Hour,
		"{minute}":    r.Minute,
		"{second}":    r.Second,
	}

	// Replace semua pattern
	// PENTING: Jika database kosong, {database} akan di-replace dengan empty string
	// yang akan menghasilkan filename tidak valid (misal: _.gz.enc)
	for pattern, value := range replacements {
		result = strings.ReplaceAll(result, pattern, value)
	}

	// Tambahkan ekstensi hanya jika ini untuk filename (bukan directory path)
	if r.IsFilename {
		// Tambahkan ekstensi .sql jika tidak ada kompresi
		// (file mentah SQL harus punya ekstensi .sql)
		if r.CompressionExt == "" {
			result = result + ".sql"
		}

		// Tambahkan ekstensi kompresi dan enkripsi di akhir
		result = result + r.CompressionExt + r.EncryptionExt
	}

	return result
}

// GenerateBackupFilename menghasilkan nama file backup berdasarkan pattern
// Untuk mode combined, database akan diganti dengan "all_databases"
func GenerateBackupFilename(pattern string, database string, mode string, hostname string, compressionType compress.CompressionType, encrypted bool) (string, error) {
	// Validasi pattern tidak boleh kosong
	if pattern == "" {
		return "", fmt.Errorf("pattern filename tidak boleh kosong")
	}

	// Untuk mode separated, validasi bahwa pattern mengandung {database}
	if (mode == "separated" || mode == "separate") && database != "" {
		if !strings.Contains(pattern, "{database}") {
			return "", fmt.Errorf("pattern filename untuk mode separated harus mengandung {database} placeholder")
		}
	}

	// Untuk mode combined, gunakan nama khusus
	if mode == "combined" && database == "" {
		database = "all_databases"
	}

	replacer, err := NewPathPatternReplacer(database, hostname, compressionType, encrypted, true)
	if err != nil {
		return "", fmt.Errorf("gagal membuat pattern replacer: %w", err)
	}

	filename := replacer.ReplacePattern(pattern)

	// Validasi hasil tidak boleh kosong atau hanya ekstensi
	if filename == "" || filename == ".sql" || strings.HasPrefix(filename, ".") {
		return "", fmt.Errorf("hasil generate filename tidak valid: %s (pattern: %s, database: %s)", filename, pattern, database)
	}

	return filename, nil
}

// GenerateBackupDirectory menghasilkan path direktori backup berdasarkan base directory dan structure pattern
func GenerateBackupDirectory(baseDir string, structurePattern string, hostname string) (string, error) {
	// Gunakan database kosong karena directory tidak butuh nama database
	// Untuk directory tidak perlu ekstensi kompresi/enkripsi
	replacer, err := NewPathPatternReplacer("", hostname, compress.CompressionNone, false, false)
	if err != nil {
		return "", fmt.Errorf("gagal membuat pattern replacer: %w", err)
	}

	// Replace pattern di structure
	subPath := replacer.ReplacePattern(structurePattern)

	// Gabungkan base directory dengan subpath
	fullPath := filepath.Join(baseDir, subPath)

	return fullPath, nil
}

// GenerateFullBackupPath menghasilkan full path untuk file backup (directory + filename)
func GenerateFullBackupPath(baseDir string, structurePattern string, filenamePattern string, database string, mode string, hostname string, compressionType compress.CompressionType, encrypted bool) (string, error) {
	// Generate directory
	dir, err := GenerateBackupDirectory(baseDir, structurePattern, hostname)
	if err != nil {
		return "", err
	}

	// Generate filename
	filename, err := GenerateBackupFilename(filenamePattern, database, mode, hostname, compressionType, encrypted)
	if err != nil {
		return "", err
	}

	// Gabungkan
	fullPath := filepath.Join(dir, filename)
	return fullPath, nil
}
