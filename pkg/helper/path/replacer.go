package path

import (
	"strings"
	"time"

	"sfdbtools/pkg/compress"
	"sfdbtools/pkg/consts"
)

// PathPatternReplacer menyimpan nilai-nilai untuk menggantikan pattern dalam path/filename.
type PathPatternReplacer struct {
	Database       string
	Timestamp      time.Time
	Hostname       string
	Year           string
	Month          string
	Day            string
	Hour           string
	Minute         string
	Second         string
	CompressionExt string
	EncryptionExt  string
	IsFilename     bool
}

// NewPathPatternReplacer membuat instance baru PathPatternReplacer dengan timestamp saat ini.
func NewPathPatternReplacer(database string, hostname string, compressionType compress.CompressionType, encrypted bool, isFilename bool) (*PathPatternReplacer, error) {
	if hostname == "" {
		hostname = "unknown"
	}

	now := time.Now()

	compressionExt := ""
	if compressionType != compress.CompressionType(consts.CompressionTypeNone) && compressionType != "" {
		compressionExt = compress.GetFileExtension(compressionType)
	}

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

// ReplacePattern mengganti semua pattern dalam string dengan nilai yang sesuai.
func (r *PathPatternReplacer) ReplacePattern(pattern string, excludeHostname ...bool) string {
	result := pattern
	skipHostname := len(excludeHostname) > 0 && excludeHostname[0]

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

	if !skipHostname {
		replacements["{hostname}"] = r.Hostname
	}

	for pattern, value := range replacements {
		result = strings.ReplaceAll(result, pattern, value)
	}

	if r.IsFilename {
		result = result + consts.ExtSQL
		result = result + r.CompressionExt
		result = result + r.EncryptionExt
	}

	return result
}
