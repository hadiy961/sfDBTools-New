// File : internal/cleanup/executor.go
// Deskripsi : Core execution logic untuk scanning dan deletion
// Author : Hadiyatna Muflihun
// Tanggal : 16 Desember 2025
// Last Modified : 17 Desember 2025

package cleanup

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"sfDBTools/internal/appconfig"
	"sfDBTools/internal/applog"
	"sfDBTools/internal/types"
	"sfDBTools/internal/types/types_backup"
	"sfDBTools/pkg/fsops"
	"sfDBTools/pkg/global"
	"sfDBTools/pkg/helper"

	"github.com/bmatcuk/doublestar/v4"
)

const (
	// timeFormat mendefinisikan format timestamp standar untuk logging.
	timeFormat = "2006-01-02 15:04:05"
)

// cleanupCore adalah fungsi inti terpadu untuk semua logika pembersihan.
func (s *Service) cleanupCore(dryRun bool, pattern string) error {
	mode := "Menjalankan"
	if dryRun {
		mode = "Menjalankan DRY-RUN"
	}
	
	if pattern != "" {
		s.Log.Infof("%s proses cleanup untuk pattern: %s", mode, pattern)
	} else {
		s.Log.Infof("%s proses cleanup backup...", mode)
	}

	retentionDays := s.Config.Backup.Cleanup.Days
	if retentionDays <= 0 {
		s.Log.Info("Retention days tidak valid, melewati proses")
		return nil
	}

	s.Log.Info("Path backup base directory:", s.Config.Backup.Output.BaseDirectory)
	s.Log.Infof("Cleanup policy: hapus file backup lebih dari %d hari", retentionDays)
	
	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)
	s.Log.Infof("Cutoff time: %s", cutoffTime.Format(timeFormat))

	// Pindai file
	filesToDelete, err := s.scanFiles(s.Config.Backup.Output.BaseDirectory, cutoffTime, pattern)
	if err != nil {
		return fmt.Errorf("gagal memindai file backup: %w", err)
	}

	if len(filesToDelete) == 0 {
		s.Log.Info("Tidak ada file backup lama yang perlu dihapus")
		return nil
	}

	if dryRun {
		s.logDryRunSummary(filesToDelete)
	} else {
		s.performDeletion(filesToDelete)
	}

	return nil
}

// scanFiles memindai file berdasarkan kriteria retensi dan pattern.
func (s *Service) scanFiles(baseDir string, cutoff time.Time, pattern string) ([]types_backup.BackupFileInfo, error) {
	if pattern == "" {
		pattern = "**/*"
	}

	// Gunakan doublestar untuk recursive globbing
	paths, err := doublestar.Glob(os.DirFS(baseDir), pattern)
	if err != nil {
		return nil, fmt.Errorf("gagal memproses pattern glob %s: %w", pattern, err)
	}

	var filesToDelete []types_backup.BackupFileInfo
	for _, path := range paths {
		fullPath := filepath.Join(baseDir, path)

		info, err := os.Stat(fullPath)
		if err != nil {
			s.Log.Errorf("Gagal mendapatkan info file %s: %v", fullPath, err)
			continue
		}

		// Validasi tipe file dan kriteria
		if info.IsDir() || (pattern == "**/*" && !helper.IsBackupFile(fullPath)) {
			continue
		}

		if info.ModTime().Before(cutoff) {
			filesToDelete = append(filesToDelete, types_backup.BackupFileInfo{
				Path:    fullPath,
				ModTime: info.ModTime(),
				Size:    info.Size(),
			})
		}
	}

	// Sort berdasarkan waktu modifikasi (terlama dulu)
	sort.Slice(filesToDelete, func(i, j int) bool {
		return filesToDelete[i].ModTime.Before(filesToDelete[j].ModTime)
	})

	return filesToDelete, nil
}

// performDeletion menghapus file yang ada dalam daftar.
func (s *Service) performDeletion(files []types_backup.BackupFileInfo) {
	s.Log.Infof("Ditemukan %d file backup lama yang akan dihapus", len(files))

	var deletedCount int
	var totalFreedSize int64

	for _, file := range files {
		if err := os.Remove(file.Path); err != nil {
			s.Log.Errorf("Gagal menghapus file %s: %v", file.Path, err)
			continue
		}
		deletedCount++
		totalFreedSize += file.Size
		s.Log.Infof("Dihapus: %s (size: %s)", file.Path, global.FormatFileSize(file.Size))
	}

	s.Log.Infof("Cleanup selesai: %d file dihapus, total %s ruang dibebaskan.",
		deletedCount, global.FormatFileSize(totalFreedSize))
}

// =============================================================================
// Public Helper Functions (Used by other modules like Backup)
// =============================================================================

// CleanupOldBackupsFromBackup menjalankan cleanup old backup files (untuk integrasi dengan backup module).
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

	opts := types.CleanupOptions{
		Enabled: true,
		Days:    retentionDays,
		Pattern: "",
	}
	svc := NewCleanupService(config, logger, opts)

	filesToDelete, err := svc.scanFiles(baseDir, cutoffTime, "")
	if err != nil {
		return fmt.Errorf("gagal scan old backup files: %w", err)
	}

	if len(filesToDelete) == 0 {
		logger.Info("Tidak ada old backup files yang perlu dihapus")
		return nil
	}

	svc.performDeletion(filesToDelete)
	return nil
}

// CleanupFailedBackup menghapus file backup yang gagal.
func CleanupFailedBackup(filePath string, logger applog.Logger) {
	if fsops.FileExists(filePath) {
		logger.Infof("Menghapus file backup yang gagal: %s", filePath)
		if err := fsops.RemoveFile(filePath); err != nil {
			logger.Warnf("Gagal menghapus file backup yang gagal: %v", err)
		}
	}
}

// CleanupPartialBackup menghapus file backup yang belum selesai (graceful shutdown).
func CleanupPartialBackup(filePath string, logger applog.Logger) error {
	if err := fsops.RemoveFile(filePath); err != nil {
		return fmt.Errorf("gagal menghapus file backup: %w", err)
	}
	logger.Infof("File backup yang belum selesai berhasil dihapus: %s", filePath)
	return nil
}
