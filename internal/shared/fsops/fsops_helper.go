package fsops

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sfdbtools/internal/shared/validation"
	"strings"
	"time"
)

func BuildSubdirPath(structurePattern, client string) (string, error) {
	if structurePattern == "" {
		return "", fmt.Errorf("empty structure pattern")
	}
	vars := map[string]string{
		"client": client,
	}
	return BuildSubdirPathFromPattern(structurePattern, vars, time.Now())
}

// BuildSubdirPathFromPattern menggantikan token-token dalam pola menggunakan vars dan waktu sekarang.
// Format token: {nama} atau {nama:format}. Untuk token 'date' dan 'timestamp' format mengikuti Go time.Format.
func BuildSubdirPathFromPattern(pattern string, vars map[string]string, now time.Time) (string, error) {
	if pattern == "" {
		return "", fmt.Errorf("pola struktur kosong")
	}

	// Validasi dasar sebelum penggantian token
	if err := validation.ValidateSubdirPattern(pattern, vars); err != nil {
		return "", fmt.Errorf("pola struktur tidak valid: %w", err)
	}

	// regex menangkap {name} atau {name:format}
	re := regexp.MustCompile(`\{([a-zA-Z0-9_]+)(?::([^}]+))?\}`)

	replaced := re.ReplaceAllStringFunc(pattern, func(tok string) string {
		parts := re.FindStringSubmatch(tok)
		if len(parts) < 2 {
			return ""
		}
		name := parts[1]
		format := ""
		if len(parts) >= 3 {
			format = parts[2]
		}

		switch name {
		case "date":
			if format == "" {
				format = "2006-01-02"
			}
			return now.Format(format)
		case "timestamp":
			if format == "" {
				// default format: YYYYMMDD_HHMMSS
				format = "20060102_150405"
			}
			return now.Format(format)
		case "year":
			return fmt.Sprintf("%04d", now.Year())
		case "month":
			return fmt.Sprintf("%02d", int(now.Month()))
		case "day":
			return fmt.Sprintf("%02d", now.Day())
		default:
			if v, ok := vars[name]; ok {
				return v
			}
			// jika tidak ada, kosongkan token
			return ""
		}
	})

	// sanitasi path dan cegah path-traversal
	cleaned := filepath.Clean(replaced)
	// jika hasil menjadi '..' atau memiliki ../ di awal, buang bagian tersebut
	if cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(os.PathSeparator)) {
		cleaned = strings.ReplaceAll(cleaned, "..", "")
	}
	// pastikan tidak ada absolute path (hindari memaksa root)
	if filepath.IsAbs(cleaned) {
		cleaned = strings.TrimPrefix(cleaned, string(os.PathSeparator))
	}

	return cleaned, nil
}
