// File : internal/cleanup/cleanup_core.go
// Deskripsi : Core cleanup logic - scan dan delete files
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-16
// Last Modified : 2025-12-16

package cleanup

import (
	"fmt"
	"os"
	"path/filepath"
	"sfDBTools/internal/types/types_backup"
	"sfDBTools/pkg/global"
	"sfDBTools/pkg/helper"
	"sort"
	"time"

	"github.com/bmatcuk/doublestar/v4"
)

const (
	// timeFormat mendefinisikan format timestamp standar untuk logging.
	timeFormat = "2006-01-02 15:04:05"
)

// cleanupCore adalah fungsi inti terpadu untuk semua logika pembersihan.
func (s *Service) cleanupCore(dryRun bool, pattern string) error {
	// Tentukan mode operasi untuk logging
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

	// Pindai file berdasarkan mode (dengan atau tanpa pattern)
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

// scanFiles memilih metode pemindaian file (menyeluruh atau berdasarkan pola).
func (s *Service) scanFiles(baseDir string, cutoff time.Time, pattern string) ([]types_backup.BackupFileInfo, error) {
	// Jika tidak ada pattern, kita buat pattern default untuk mencari semua file secara rekursif.
	// Tanda '**/*' berarti "semua file di semua sub-direktori".
	if pattern == "" {
		pattern = "**/*"
	}

	// Satu panggilan untuk menemukan semua file yang cocok, di mana pun lokasinya!
	paths, err := doublestar.Glob(os.DirFS(baseDir), pattern)
	if err != nil {
		return nil, fmt.Errorf("gagal memproses pattern glob %s: %w", pattern, err)
	}

	var filesToDelete []types_backup.BackupFileInfo
	for _, path := range paths {
		// Karena Glob mengembalikan path relatif, kita gabungkan lagi dengan baseDir
		fullPath := filepath.Join(baseDir, path)

		info, err := os.Stat(fullPath)
		if err != nil {
			s.Log.Errorf("Gagal mendapatkan info file %s: %v", fullPath, err)
			continue
		}

		// Lewati direktori dan file yang tidak sesuai kriteria
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

	sort.Slice(filesToDelete, func(i, j int) bool {
		return filesToDelete[i].ModTime.Before(filesToDelete[j].ModTime)
	})

	return filesToDelete, nil
}

// performDeletion menghapus file-file yang ada dalam daftar.
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

// logDryRunSummary mencatat ringkasan file yang akan dihapus dalam mode dry-run.
func (s *Service) logDryRunSummary(files []types_backup.BackupFileInfo) {
	s.Log.Infof("DRY-RUN: Ditemukan %d file backup yang AKAN dihapus:", len(files))

	var totalSize int64
	for i, file := range files {
		totalSize += file.Size
		s.Log.Infof("  [%d] %s (modified: %s, size: %s)",
			i+1,
			file.Path,
			file.ModTime.Format(timeFormat),
			global.FormatFileSize(file.Size))
	}

	s.Log.Infof("DRY-RUN: Total %d file dengan ukuran %s akan dibebaskan.",
		len(files), global.FormatFileSize(totalSize))
	s.Log.Info("DRY-RUN: Untuk menjalankan cleanup sebenarnya, jalankan tanpa flag --dry-run.")
}
