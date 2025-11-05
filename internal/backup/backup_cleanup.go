package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/global"
	"sfDBTools/pkg/helper"
	"time"
)

// cleanupOldBackups menjalankan cleanup old backup files sebelum backup dimulai
func (s *Service) cleanupOldBackups() error {
	retentionDays := s.Config.Backup.Cleanup.Days
	if retentionDays <= 0 {
		s.Log.Info("Retention days tidak valid, melewati cleanup")
		return nil
	}

	baseDir := s.Config.Backup.Output.BaseDirectory
	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)

	s.Log.Infof("Melakukan cleanup backup files lebih dari %d hari (sebelum %s)",
		retentionDays, cutoffTime.Format("2006-01-02 15:04:05"))

	// Scan files yang perlu dihapus
	filesToDelete, err := s.scanOldBackupFiles(baseDir, cutoffTime)
	if err != nil {
		return fmt.Errorf("gagal scan old backup files: %w", err)
	}

	if len(filesToDelete) == 0 {
		s.Log.Info("Tidak ada old backup files yang perlu dihapus")
		return nil
	}

	// Hapus files
	var deletedCount int
	var totalFreedSize int64

	for _, file := range filesToDelete {
		if err := os.Remove(file.Path); err != nil {
			s.Log.Warnf("Gagal menghapus file %s: %v", file.Path, err)
			continue
		}
		deletedCount++
		totalFreedSize += file.Size
		s.Log.Infof("Dihapus: %s (size: %s)", file.Path, global.FormatFileSize(file.Size))
	}

	s.Log.Infof("Cleanup selesai: %d file dihapus, total %s ruang dibebaskan",
		deletedCount, global.FormatFileSize(totalFreedSize))

	return nil
}

// scanOldBackupFiles mencari backup files yang lebih lama dari cutoff time
func (s *Service) scanOldBackupFiles(baseDir string, cutoff time.Time) ([]types.BackupFileInfo, error) {
	var filesToDelete []types.BackupFileInfo

	err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			s.Log.Warnf("Error mengakses path %s: %v", path, err)
			return nil // Continue walking
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if it's a backup file and older than cutoff
		if helper.IsBackupFile(path) && info.ModTime().Before(cutoff) {
			filesToDelete = append(filesToDelete, types.BackupFileInfo{
				Path:    path,
				ModTime: info.ModTime(),
				Size:    info.Size(),
			})
		}

		return nil
	})

	return filesToDelete, err
}
