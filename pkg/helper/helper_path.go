package helper

import (
	"fmt"
	"path/filepath"
	"sfDBTools/pkg/compress"
	"strings"
	"time"
)

// Fixed pattern yang digunakan untuk filename backup
const FixedBackupPattern = "{database}_{year}{month}{day}_{hour}{minute}{second}_{hostname}"

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
// Parameters:
// - pattern: string yang berisi pattern yang ingin di-replace
// - excludeHostname: jika true, {hostname} tidak akan di-replace (gunakan untuk directory pattern)
func (r *PathPatternReplacer) ReplacePattern(pattern string, excludeHostname ...bool) string {
	result := pattern
	skipHostname := len(excludeHostname) > 0 && excludeHostname[0]

	// Daftar replacements
	replacements := map[string]string{
		"{database}":  r.Database,
		"{timestamp}": r.Timestamp.Format("20060102_150405"),
		"{year}":      r.Year,
		"{month}":     r.Month,
		"{day}":       r.Day,
		"{hour}":      r.Hour,
		"{minute}":    r.Minute,
		"{second}":    r.Second,
	}

	// Hanya replace {hostname} jika tidak di-exclude (untuk filename saja)
	if !skipHostname {
		replacements["{hostname}"] = r.Hostname
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

// GenerateBackupFilename menghasilkan nama file backup menggunakan fixed pattern
// Fixed pattern: {database}_{year}{month}{day}_{hour}{minute}{second}_{hostname}
// Untuk mode combined, format: combined_{hostname}_{timestamp}_{jumlah_db}
func GenerateBackupFilename(database string, mode string, hostname string, compressionType compress.CompressionType, encrypted bool) (string, error) {
	return GenerateBackupFilenameWithCount(database, mode, hostname, compressionType, encrypted, 0)
}

// GenerateBackupFilenameWithCount menghasilkan nama file backup dengan jumlah database
// Untuk mode combined dengan dbCount > 0, format: combined_{hostname}_{timestamp}_{jumlah_db}
func GenerateBackupFilenameWithCount(database string, mode string, hostname string, compressionType compress.CompressionType, encrypted bool, dbCount int) (string, error) {
	// Untuk mode separated, validasi bahwa database tidak kosong
	if (mode == "separated" || mode == "separate") && database == "" {
		return "", fmt.Errorf("database name tidak boleh kosong untuk mode separated")
	}

	// Untuk mode combined, gunakan format khusus
	if mode == "combined" {
		if dbCount > 0 {
			// Format: combined_{hostname}_{timestamp}_{jumlah_db}
			timestamp := time.Now().Format("20060102_150405")
			database = fmt.Sprintf("combined_%s_%s_%ddb", hostname, timestamp, dbCount)
		} else if database == "" {
			database = "all_databases"
		}
	}

	replacer, err := NewPathPatternReplacer(database, hostname, compressionType, encrypted, true)
	if err != nil {
		return "", fmt.Errorf("gagal membuat pattern replacer: %w", err)
	}

	// Untuk combined dengan custom format, gunakan pattern sederhana
	var filename string
	if mode == "combined" && dbCount > 0 {
		// Sudah include hostname dan timestamp di database name
		// Ekstensi kompresi dan enkripsi akan ditambahkan otomatis oleh replacer
		filename = replacer.ReplacePattern("{database}")
	} else {
		// Gunakan fixed pattern
		filename = replacer.ReplacePattern(FixedBackupPattern)
	}

	// Validasi hasil tidak boleh kosong atau hanya ekstensi
	if filename == "" || filename == ".sql" || strings.HasPrefix(filename, ".") {
		return "", fmt.Errorf("hasil generate filename tidak valid: %s (database: %s)", filename, database)
	}

	return filename, nil
}

// GenerateBackupDirectory menghasilkan path direktori backup berdasarkan base directory dan structure pattern
// Hostname akan di-exclude dari directory path (hanya untuk filename)
//
// Examples:
//
//	GenerateBackupDirectory("/media/ArchiveDB", "{year}{month}{day}/", "dbserver1")
//	returns: /media/ArchiveDB/20251205/
func GenerateBackupDirectory(baseDir string, structurePattern string, hostname string) (string, error) {
	// Gunakan database kosong karena directory tidak butuh nama database
	// Untuk directory tidak perlu ekstensi kompresi/enkripsi
	replacer, err := NewPathPatternReplacer("", hostname, compress.CompressionNone, false, false)
	if err != nil {
		return "", fmt.Errorf("gagal membuat pattern replacer: %w", err)
	}

	// Replace pattern di structure - exclude hostname dari directory path
	subPath := replacer.ReplacePattern(structurePattern, true)

	// Gabungkan base directory dengan subpath
	fullPath := filepath.Join(baseDir, subPath)

	return fullPath, nil
}

// GenerateFullBackupPath menghasilkan full path untuk file backup (directory + filename)
// Note: filenamePattern parameter diabaikan karena menggunakan fixed pattern
func GenerateFullBackupPath(baseDir string, structurePattern string, filenamePattern string, database string, mode string, hostname string, compressionType compress.CompressionType, encrypted bool) (string, error) {
	// Generate directory
	dir, err := GenerateBackupDirectory(baseDir, structurePattern, hostname)
	if err != nil {
		return "", err
	}

	// Generate filename - filenamePattern diabaikan, gunakan fixed pattern
	filename, err := GenerateBackupFilename(database, mode, hostname, compressionType, encrypted)
	if err != nil {
		return "", err
	}

	// Gabungkan
	fullPath := filepath.Join(dir, filename)
	return fullPath, nil
}
