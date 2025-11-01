package cleanup

import (
	"fmt"
	"os"
	"path/filepath"
	"sfDBTools/internal/appconfig"
	"sfDBTools/internal/applog"
	"sfDBTools/internal/types"
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

var (
	// Logger adalah instance logger untuk package cleanup
	Logger applog.Logger

	// cfg adalah instance konfigurasi untuk package cleanup
	cfg *appconfig.Config
)

// SetConfig mengatur konfigurasi untuk package cleanup.
func SetConfig(conf *appconfig.Config) {
	cfg = conf
}

// CleanupOldBackups menjalankan proses penghapusan semua backup lama di direktori.
func CleanupOldBackups() error {
	return cleanupCore(false, "") // dryRun=false, tanpa pattern
}

// CleanupDryRun menampilkan preview semua backup lama yang akan dihapu
func CleanupDryRun() error {
	return cleanupCore(true, "") // dryRun=true, tanpa pattern
}

// CleanupByPattern menjalankan proses penghapusan backup lama yang cocok dengan pattern.
func CleanupByPattern(pattern string) error {
	if pattern == "" {
		return fmt.Errorf("pattern tidak boleh kosong untuk CleanupByPattern")
	}
	return cleanupCore(false, pattern) // dryRun=false, dengan pattern
}

// cleanupCore adalah fungsi inti terpadu untuk semua logika pembersihan.
func cleanupCore(dryRun bool, pattern string) error {
	// Tentukan mode operasi untuk logging
	mode := "Menjalankan"
	if dryRun {
		mode = "Menjalankan DRY-RUN"
	}
	if pattern != "" {
		Logger.Infof("%s proses cleanup untuk pattern: %s", mode, pattern)
	} else {
		Logger.Infof("%s proses cleanup backup...", mode)
	}

	retentionDays := cfg.Backup.Cleanup.Days
	if retentionDays <= 0 {
		Logger.Info("Retention days tidak valid, melewati proses")
		return nil
	}

	Logger.Info("Path backup base directory:", cfg.Backup.Output.BaseDirectory)
	Logger.Infof("Cleanup policy: hapus file backup lebih dari %d hari", retentionDays)
	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)
	Logger.Infof("Cutoff time: %s", cutoffTime.Format(timeFormat))

	// Pindai file berdasarkan mode (dengan atau tanpa pattern)
	filesToDelete, err := scanFiles(cfg.Backup.Output.BaseDirectory, cutoffTime, pattern)
	if err != nil {
		return fmt.Errorf("gagal memindai file backup: %w", err)
	}

	if len(filesToDelete) == 0 {
		Logger.Info("Tidak ada file backup lama yang perlu dihapus")
		return nil
	}

	if dryRun {
		logDryRunSummary(filesToDelete)
	} else {
		performDeletion(filesToDelete)
	}

	return nil
}

// scanFiles memilih metode pemindaian file (menyeluruh atau berdasarkan pola).
func scanFiles(baseDir string, cutoff time.Time, pattern string) ([]types.BackupFileInfo, error) {
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

	var filesToDelete []types.BackupFileInfo
	for _, path := range paths {
		// Karena Glob mengembalikan path relatif, kita gabungkan lagi dengan baseDir
		fullPath := filepath.Join(baseDir, path)

		info, err := os.Stat(fullPath)
		if err != nil {
			Logger.Errorf("Gagal mendapatkan info file %s: %v", fullPath, err)
			continue
		}

		// Lewati direktori dan file yang tidak sesuai kriteria
		if info.IsDir() || (pattern == "**/*" && !helper.IsBackupFile(fullPath)) {
			continue
		}

		if info.ModTime().Before(cutoff) {
			filesToDelete = append(filesToDelete, types.BackupFileInfo{
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
func performDeletion(files []types.BackupFileInfo) {
	Logger.Infof("Ditemukan %d file backup lama yang akan dihapus", len(files))

	var deletedCount int
	var totalFreedSize int64

	for _, file := range files {
		if err := os.Remove(file.Path); err != nil {
			Logger.Errorf("Gagal menghapus file %s: %v", file.Path, err)
			continue
		}
		deletedCount++
		totalFreedSize += file.Size
		Logger.Infof("Dihapus: %s (size: %s)", file.Path, global.FormatFileSize(file.Size))
	}

	Logger.Infof("Cleanup selesai: %d file dihapus, total %s ruang dibebaskan.",
		deletedCount, global.FormatFileSize(totalFreedSize))
}

// logDryRunSummary mencatat ringkasan file yang akan dihapus dalam mode dry-run.
func logDryRunSummary(files []types.BackupFileInfo) {
	Logger.Infof("DRY-RUN: Ditemukan %d file backup yang AKAN dihapus:", len(files))

	var totalSize int64
	for i, file := range files {
		totalSize += file.Size
		Logger.Infof("  [%d] %s (modified: %s, size: %s)",
			i+1,
			file.Path,
			file.ModTime.Format(timeFormat),
			global.FormatFileSize(file.Size))
	}

	Logger.Infof("DRY-RUN: Total %d file dengan ukuran %s akan dibebaskan.",
		len(files), global.FormatFileSize(totalSize))
	Logger.Info("DRY-RUN: Untuk menjalankan cleanup sebenarnya, jalankan tanpa flag --dry-run.")
}
