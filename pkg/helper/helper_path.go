package helper

import (
	"fmt"
	"path/filepath"
	"sfDBTools/pkg/compress"
	"sfDBTools/pkg/consts"
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
	if compressionType != compress.CompressionType(consts.CompressionTypeNone) && compressionType != "" {
		compressionExt = compress.GetFileExtension(compressionType)
	}
	// Tentukan ekstensi enkripsi
	encryptionExt := ""
	if encrypted {
		encryptionExt = consts.ExtEnc
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
		// Selalu tambahkan .sql terlebih dahulu
		result = result + consts.ExtSQL

		// Kemudian tambahkan ekstensi kompresi (jika ada)
		result = result + r.CompressionExt

		// Terakhir tambahkan ekstensi enkripsi (jika ada)
		result = result + r.EncryptionExt
	}

	return result
}

// GenerateBackupFilename menghasilkan nama file backup menggunakan fixed pattern
// Fixed pattern: {database}_{year}{month}{day}_{hour}{minute}{second}_{hostname}
// Untuk mode combined, format: combined_{hostname}_{timestamp}_{jumlah_db}
func GenerateBackupFilename(database string, mode string, hostname string, compressionType compress.CompressionType, encrypted bool, excludeData bool) (string, error) {
	return GenerateBackupFilenameWithCount(database, mode, hostname, compressionType, encrypted, 0, excludeData)
}

// GenerateBackupFilenameWithCount menghasilkan nama file backup dengan jumlah database
// Untuk mode combined/all dengan dbCount > 0, format: {prefix}_{hostname}_{timestamp}_{jumlah_db}
// prefix: "all" untuk backup all, "combined" untuk filter --mode=single-file
func GenerateBackupFilenameWithCount(database string, mode string, hostname string, compressionType compress.CompressionType, encrypted bool, dbCount int, excludeData bool) (string, error) {
	// Untuk mode separated, validasi bahwa database tidak kosong
	if (mode == "separated" || mode == "separate") && database == "" {
		return "", fmt.Errorf("database name tidak boleh kosong untuk mode separated")
	}

	// Untuk mode combined atau all, gunakan format khusus
	if mode == "combined" || mode == "all" {
		if dbCount > 0 {
			// Format: {prefix}_{hostname}_{timestamp}_{jumlah_db}
			// prefix berbeda: "all" untuk backup all, "combined" untuk filter single-file
			prefix := mode
			if excludeData {
				prefix += "_nodata"
			}
			timestamp := time.Now().Format("20060102_150405")
			database = fmt.Sprintf("%s_%s_%s_%ddb", prefix, hostname, timestamp, dbCount)
		} else if database == "" {
			database = "all_databases"
		}
	} else if excludeData {
		// Untuk mode single/separated/primary/secondary, tambahkan suffix _nodata
		// jika exclude-data aktif
		database += "_nodata"
	}

	replacer, err := NewPathPatternReplacer(database, hostname, compressionType, encrypted, true)
	if err != nil {
		return "", fmt.Errorf("gagal membuat pattern replacer: %w", err)
	}

	// Untuk combined/all dengan custom format, gunakan pattern sederhana
	var filename string
	if (mode == "combined" || mode == "all") && dbCount > 0 {
		// Sudah include hostname dan timestamp di database name
		// Ekstensi kompresi dan enkripsi akan ditambahkan otomatis oleh replacer
		filename = replacer.ReplacePattern("{database}")
	} else {
		// Gunakan fixed pattern
		filename = replacer.ReplacePattern(consts.FixedBackupPattern)
	}

	// Validasi hasil tidak boleh kosong atau hanya ekstensi
	if filename == "" || filename == consts.ExtSQL || strings.HasPrefix(filename, ".") {
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
	replacer, err := NewPathPatternReplacer("", hostname, compress.CompressionType(consts.CompressionTypeNone), false, false)
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
func GenerateFullBackupPath(baseDir string, structurePattern string, filenamePattern string, database string, mode string, hostname string, compressionType compress.CompressionType, encrypted bool, excludeData bool) (string, error) {
	// Generate directory
	dir, err := GenerateBackupDirectory(baseDir, structurePattern, hostname)
	if err != nil {
		return "", err
	}

	// Generate filename - filenamePattern diabaikan, gunakan fixed pattern
	filename, err := GenerateBackupFilename(database, mode, hostname, compressionType, encrypted, excludeData)
	if err != nil {
		return "", err
	}

	// Gabungkan
	fullPath := filepath.Join(dir, filename)
	return fullPath, nil
}
