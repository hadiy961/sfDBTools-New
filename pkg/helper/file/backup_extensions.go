package file

import (
	"path/filepath"
	"strings"

	"sfDBTools/pkg/compress"
	"sfDBTools/pkg/consts"
)

var (
	backupExtensions    = []string{consts.ExtSQL, consts.ExtGzip, consts.ExtZstd, consts.ExtXz, consts.ExtZlib, consts.ExtEnc}
	allBackupExtensions = []string{consts.ExtEnc, consts.ExtGzip, consts.ExtZstd, consts.ExtXz, consts.ExtZlib, consts.ExtSQL}
)

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

// StripAllBackupExtensions menghilangkan semua ekstensi backup dari filename.
func StripAllBackupExtensions(filename string) string {
	base := filepath.Base(filename)

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

// ExtractFileExtensions mengekstrak ekstensi dari filename dan mengembalikan nama tanpa ekstensi + list ekstensi.
func ExtractFileExtensions(filename string) (string, []string) {
	nameWithoutExt := filename
	extensions := []string{}

	if strings.HasSuffix(strings.ToLower(nameWithoutExt), consts.ExtEnc) {
		nameWithoutExt = strings.TrimSuffix(nameWithoutExt, consts.ExtEnc)
		extensions = append([]string{consts.ExtEnc}, extensions...)
	}

	for _, ext := range compress.SupportedCompressionExtensions() {
		if strings.HasSuffix(strings.ToLower(nameWithoutExt), ext) {
			nameWithoutExt = nameWithoutExt[:len(nameWithoutExt)-len(ext)]
			extensions = append([]string{ext}, extensions...)
			break
		}
	}

	if strings.HasSuffix(strings.ToLower(nameWithoutExt), consts.ExtSQL) {
		nameWithoutExt = strings.TrimSuffix(nameWithoutExt, consts.ExtSQL)
		extensions = append([]string{consts.ExtSQL}, extensions...)
	}

	return nameWithoutExt, extensions
}

func IsEncryptedFile(path string) bool {
	return strings.HasSuffix(strings.ToLower(path), consts.ExtEnc)
}

// ValidBackupFileExtensionsForSelection returns the list of file extensions used by interactive file pickers.
func ValidBackupFileExtensionsForSelection() []string {
	valid := []string{consts.ExtSQL, consts.ExtSQL + consts.ExtEnc, consts.ExtEnc}
	for _, cext := range compress.SupportedCompressionExtensions() {
		valid = append(valid, consts.ExtSQL+cext)
		valid = append(valid, consts.ExtSQL+cext+consts.ExtEnc)
	}
	return valid
}
