// File : internal/cleanup/cleanup_backup.go
// Deskripsi : Fungsi cleanup yang dipanggil dari backup package
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-08
// Last Modified : 2025-12-08

package cleanup

import (
	"fmt"
	"sfDBTools/internal/appconfig"
	"sfDBTools/internal/applog"
	"sfDBTools/pkg/fsops"
	"time"
)

// CleanupOldBackupsFromBackup menjalankan cleanup old backup files setelah backup selesai
// Dipanggil dari backup service setelah proses backup berhasil
func CleanupOldBackupsFromBackup(config *appconfig.Config, logger applog.Logger) error {
	retentionDays := config.Backup.Cleanup.Days
	if retentionDays <= 0 {
		logger.Info("Retention days tidak valid, melewati cleanup")
		return nil
	}

	baseDir := config.Backup.Output.BaseDirectory
	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)

	logger.Infof("Melakukan cleanup backup files lebih dari %d hari (sebelum %s)",
		retentionDays, cutoffTime.Format(timeFormat))

	// Scan files yang perlu dihapus - reuse scanFilesWithLogger
	filesToDelete, err := scanFilesWithLogger(baseDir, cutoffTime, "", logger)
	if err != nil {
		return fmt.Errorf("gagal scan old backup files: %w", err)
	}

	if len(filesToDelete) == 0 {
		logger.Info("Tidak ada old backup files yang perlu dihapus")
		return nil
	}

	// Hapus files - reuse performDeletionWithLogger
	performDeletionWithLogger(filesToDelete, logger)

	return nil
}

// CleanupFailedBackup menghapus file backup yang gagal
// Dipanggil dari backup error handler ketika backup gagal
func CleanupFailedBackup(filePath string, logger applog.Logger) {
	if fsops.FileExists(filePath) {
		logger.Infof("Menghapus file backup yang gagal: %s", filePath)
		if err := fsops.RemoveFile(filePath); err != nil {
			logger.Warnf("Gagal menghapus file backup yang gagal: %v", err)
		}
	}
}

// CleanupPartialBackup menghapus file backup yang belum selesai saat interrupt
// Dipanggil dari backup service saat graceful shutdown
func CleanupPartialBackup(filePath string, logger applog.Logger) error {
	if err := fsops.RemoveFile(filePath); err != nil {
		return fmt.Errorf("gagal menghapus file backup: %w", err)
	}
	logger.Infof("File backup yang belum selesai berhasil dihapus: %s", filePath)
	return nil
}
