// File : internal/backup/metadata/user.go
// Deskripsi : Helper path untuk user grants (legacy export logic dihapus; gunakan internal/app/backup/grants)
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-05
// Last Modified : 12 Februari 2026
package metadata

import (
	"path/filepath"
	backupfile "sfdbtools/internal/app/backup/helpers/file"
)

// GenerateUserFilePath menghasilkan path file untuk user grants berdasarkan backup file path
// Contoh: /backup/db_20250101.sql.gz -> /backup/db_20250101_users.sql
func GenerateUserFilePath(backupFilePath string) string {
	dir := filepath.Dir(backupFilePath)
	base := filepath.Base(backupFilePath)
	return filepath.Join(dir, backupfile.GenerateGrantsFilename(base))
}
