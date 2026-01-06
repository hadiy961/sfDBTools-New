// File : internal/cleanup/display.go
// Deskripsi : Display functions untuk cleanup results dan options
// Author : Hadiyatna Muflihun
// Tanggal : 16 Desember 2025
// Last Modified : 17 Desember 2025

package cleanup

import (
	"fmt"
	"sfdbtools/internal/app/backup/model/types_backup"
	"sfdbtools/internal/ui/print"
	"sfdbtools/internal/ui/table"
	"sfdbtools/pkg/consts"
	"sfdbtools/pkg/global"
)

// displayCleanupOptions menampilkan konfigurasi cleanup yang akan dijalankan
func (s *Service) displayCleanupOptions() {
	print.PrintSubHeader("Konfigurasi Cleanup")

	data := [][]string{
		{"Base Directory", s.Config.Backup.Output.BaseDirectory},
		{"Retention Days", fmt.Sprintf("%d", s.Config.Backup.Cleanup.Days)},
	}

	if s.CleanupOptions.Pattern != "" {
		data = append(data, []string{"Pattern", s.CleanupOptions.Pattern})
	}

	if s.Config.Backup.Cleanup.Schedule != "" {
		data = append(data, []string{"Schedule", s.Config.Backup.Cleanup.Schedule})
	}

	table.Render([]string{"Setting", "Value"}, data)
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
			file.ModTime.Format(consts.CleanupTimeFormat),
			global.FormatFileSize(file.Size))
	}

	s.Log.Infof("DRY-RUN: Total %d file dengan ukuran %s akan dibebaskan.",
		len(files), global.FormatFileSize(totalSize))
	s.Log.Info("DRY-RUN: Untuk menjalankan cleanup sebenarnya, jalankan tanpa flag --dry-run.")
}
