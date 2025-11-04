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
}

// NewPathPatternReplacer membuat instance baru PathPatternReplacer dengan timestamp saat ini
func NewPathPatternReplacer(database string, hostname string, compressionType compress.CompressionType, encrypted bool) (*PathPatternReplacer, error) {
	// Gunakan hostname yang diberikan (dari database server)
	if hostname == "" {
		hostname = "unknown"
	}

	now := time.Now()

	// Tentukan ekstensi kompresi
	compressionExt := ".sql"
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
	for pattern, value := range replacements {
		result = strings.ReplaceAll(result, pattern, value)
	}

	// Tambahkan ekstensi kompresi dan enkripsi di akhir
	result = result + r.CompressionExt + r.EncryptionExt

	return result
}

// GenerateBackupFilename menghasilkan nama file backup berdasarkan pattern
// Untuk mode combined, database akan diganti dengan "all_databases"
func GenerateBackupFilename(pattern string, database string, mode string, hostname string, compressionType compress.CompressionType, encrypted bool) (string, error) {
	// Untuk mode combined, gunakan nama khusus
	if mode == "combined" && database == "" {
		database = "all_databases"
	}

	replacer, err := NewPathPatternReplacer(database, hostname, compressionType, encrypted)
	if err != nil {
		return "", fmt.Errorf("gagal membuat pattern replacer: %w", err)
	}

	filename := replacer.ReplacePattern(pattern)
	return filename, nil
}

// GenerateBackupDirectory menghasilkan path direktori backup berdasarkan base directory dan structure pattern
func GenerateBackupDirectory(baseDir string, structurePattern string, hostname string) (string, error) {
	// Gunakan database kosong karena directory tidak butuh nama database
	// Untuk directory tidak perlu ekstensi kompresi/enkripsi
	replacer, err := NewPathPatternReplacer("", hostname, compress.CompressionNone, false)
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
