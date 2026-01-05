// File : internal/cleanup/executor.go
// Deskripsi : Core execution logic untuk scanning dan deletion
// Author : Hadiyatna Muflihun
// Tanggal : 16 Desember 2025
// Last Modified : 30 Desember 2025

package cleanup

import (
	"fmt"
	"os"
	"path/filepath"
	"sfDBTools/internal/app/backup/model/types_backup"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/global"
	"sfDBTools/pkg/helper"
	"sort"
	"time"

	"github.com/bmatcuk/doublestar/v4"
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
	s.Log.Infof("Cutoff time: %s", cutoffTime.Format(consts.CleanupTimeFormat))

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
