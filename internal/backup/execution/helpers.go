// File : internal/backup/execution/helpers.go
// Deskripsi : Shared utility functions untuk backup execution
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-30
// Last Modified :  2026-01-05
package execution

import (
	"fmt"
	"strings"

	"sfDBTools/internal/services/log"
	"sfDBTools/internal/backup/gtid"
	"sfDBTools/internal/backup/metadata"
	"sfDBTools/internal/types/types_backup"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/fsops"
)

// ExtractMysqldumpVersion mengambil versi mysqldump dari stderr output.
// Mencari baris yang mengandung "mysqldump" dan "Ver".
func ExtractMysqldumpVersion(stderrOutput string) string {
	if stderrOutput == "" {
		return ""
	}

	for _, line := range strings.Split(stderrOutput, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "mysqldump") && strings.Contains(line, "Ver") {
			return line
		}
	}

	return ""
}

// formatBackupDisplayName menghasilkan nama display untuk backup info.
// Untuk multi-DB backup, menampilkan jumlah database.
// Untuk single-DB backup, menampilkan nama database.
func formatBackupDisplayName(cfg types_backup.BackupExecutionConfig) string {
	if cfg.IsMultiDB {
		return fmt.Sprintf("Combined backup (%d databases)", len(cfg.DBList))
	}
	return cfg.DBName
}

// cleanupFailedBackup menghapus file backup yang gagal.
// Dipanggil saat backup error untuk cleanup.
func cleanupFailedBackup(filePath string, logger applog.Logger) {
	if fsops.FileExists(filePath) {
		logger.Infof("Menghapus file backup yang gagal: %s", filePath)
		if err := fsops.RemoveFile(filePath); err != nil {
			logger.Warnf("Gagal menghapus file backup yang gagal: %v", err)
		}
	}
}

// formatGTIDString menghasilkan string representasi dari GTID info.
// Returns empty string jika gtidInfo nil.
func formatGTIDString(gtidInfo *gtid.GTIDInfo) string {
	if gtidInfo == nil {
		return ""
	}

	if gtidInfo.GTIDBinlog != "" {
		return gtidInfo.GTIDBinlog
	}

	return fmt.Sprintf("File=%s, Pos=%d", gtidInfo.MasterLogFile, gtidInfo.MasterLogPos)
}

// determineUserGrantsPath menentukan path untuk user grants file.
// Returns empty string jika user grants di-exclude.
func determineUserGrantsPath(excludeUser bool, outputPath string) string {
	if excludeUser {
		return ""
	}
	return metadata.GenerateUserFilePath(outputPath)
}

// getExcludedDatabases mengembalikan list excluded databases untuk backup type tertentu.
// Hanya backup mode "all" yang memiliki excluded databases.
func getExcludedDatabases(backupType string, excludedDatabases []string) []string {
	if backupType == consts.ModeAll {
		return excludedDatabases
	}
	return []string{}
}

// determineBackupStatus menentukan status backup berdasarkan write result.
// Logs appropriate messages sesuai dengan hasil backup.
func determineBackupStatus(
	writeResult *types_backup.BackupWriteResult,
	cfg types_backup.BackupExecutionConfig,
	logger applog.Logger,
) string {
	if writeResult.StderrOutput != "" {
		if !cfg.IsMultiDB {
			logger.Warningf("Database %s backup dengan warning: %s", cfg.DBName, writeResult.StderrOutput)
		}
		return consts.BackupStatusSuccessWithWarnings
	}

	if !cfg.IsMultiDB {
		logger.Infof("✓ Database %s berhasil di-backup", cfg.DBName)
	}

	return consts.BackupStatusSuccess
}
