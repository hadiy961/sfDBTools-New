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

// GenerateUserGrantsFilePath menghasilkan path file untuk user grants berdasarkan backup file path
// Contoh: /backup/db_20250101.sql.gz -> /backup/db_20250101_grants.sql
func GenerateUserGrantsFilePath(backupFilePath string) string {
	dir := filepath.Dir(backupFilePath)
	base := filepath.Base(backupFilePath)
	return filepath.Join(dir, backupfile.GenerateGrantsFilename(base))
}

// GenerateUserDefinitionFilePath menghasilkan path file untuk user definitions (CREATE USER)
// Contoh: /backup/db_20250101.sql.gz -> /backup/db_20250101_users.sql
func GenerateUserDefinitionFilePath(backupFilePath string) string {
	dir := filepath.Dir(backupFilePath)
	base := filepath.Base(backupFilePath)
	// Helper file utils mungkin belum support ini, kita hardcode suffix consistency dulu
	// Pattern: <original>_users.sql
	name := backupfile.GenerateGrantsFilename(base) // ini outputnya _grants.sql atau _users.sql tergantung implementasi lama
	// Karena kita mau paksa _users.sql, kita replace manual
	if len(name) > 11 && name[len(name)-11:] == "_grants.sql" {
		name = name[:len(name)-11] + "_users.sql"
	}
	// Fallback logic lama generate suffix _users.sql (cek backupfile helper)
	return filepath.Join(dir, name)
}
