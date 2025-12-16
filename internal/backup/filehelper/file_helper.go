// File : internal/backup/filehelper/file_helper.go
// Deskripsi : Helper functions untuk file operations di backup
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-05
// Last Modified : 2025-12-05

package filehelper

import (
	"path/filepath"
	"strings"
)

// GenerateUserFilePath menghasilkan path file untuk user grants berdasarkan backup file path
// Contoh: /backup/db_20250101.sql.gz -> /backup/db_20250101_users.sql
func GenerateUserFilePath(backupFilePath string) string {
	dir := filepath.Dir(backupFilePath)
	base := filepath.Base(backupFilePath)

	// Remove all extensions (.enc, .gz, .sql, etc)
	nameWithoutExt := base
	// Remove .enc if exists
	nameWithoutExt = strings.TrimSuffix(nameWithoutExt, ".enc")
	// Remove compression extensions
	for _, ext := range []string{".gz", ".zst", ".xz", ".zlib"} {
		nameWithoutExt = strings.TrimSuffix(nameWithoutExt, ext)
	}
	// Remove .sql
	nameWithoutExt = strings.TrimSuffix(nameWithoutExt, ".sql")

	userFileName := nameWithoutExt + "_users.sql"
	return filepath.Join(dir, userFileName)
}
