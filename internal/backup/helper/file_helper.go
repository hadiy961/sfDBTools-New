// File : internal/backup/helper/file_helper.go
// Deskripsi : Helper functions untuk file operations di backup
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-05
// Last Modified : 2025-12-05

package helper

import (
	"path/filepath"
	"strings"
)

// GenerateUserFilePath menghasilkan path file untuk user grants berdasarkan backup file path
// Contoh: /backup/db_20250101.sql.gz -> /backup/db_20250101_users.cnf
func GenerateUserFilePath(backupFilePath string) string {
	dir := filepath.Dir(backupFilePath)
	base := filepath.Base(backupFilePath)

	// Remove extension
	ext := filepath.Ext(base)
	nameWithoutExt := strings.TrimSuffix(base, ext)

	// Handle double extensions like .sql.gz
	nameWithoutExt = strings.TrimSuffix(nameWithoutExt, ".sql")

	userFileName := nameWithoutExt + "_users.cnf"
	return filepath.Join(dir, userFileName)
}

// GenerateGTIDFilePath menghasilkan path file untuk GTID berdasarkan backup file path
// Contoh: /backup/db_20250101.sql.gz -> /backup/db_20250101_gtid.info
func GenerateGTIDFilePath(backupFilePath string) string {
	dir := filepath.Dir(backupFilePath)
	base := filepath.Base(backupFilePath)

	// Remove extension
	ext := filepath.Ext(base)
	nameWithoutExt := strings.TrimSuffix(base, ext)

	// Handle double extensions like .sql.gz
	nameWithoutExt = strings.TrimSuffix(nameWithoutExt, ".sql")

	gtidFileName := nameWithoutExt + "_gtid.info"
	return filepath.Join(dir, gtidFileName)
}

// GenerateBackupMetadataFilePath menghasilkan path file untuk metadata berdasarkan backup file path
// Contoh: /backup/db_20250101.sql.gz -> /backup/db_20250101.metadata.json
func GenerateBackupMetadataFilePath(backupFilePath string) string {
	dir := filepath.Dir(backupFilePath)
	base := filepath.Base(backupFilePath)

	// Remove extension
	ext := filepath.Ext(base)
	nameWithoutExt := strings.TrimSuffix(base, ext)

	// Handle double extensions like .sql.gz
	nameWithoutExt = strings.TrimSuffix(nameWithoutExt, ".sql")

	metadataFileName := nameWithoutExt + ".metadata.json"
	return filepath.Join(dir, metadataFileName)
}
