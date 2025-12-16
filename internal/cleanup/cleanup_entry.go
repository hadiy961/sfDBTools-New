// File : internal/cleanup/cleanup_entry.go
// Deskripsi : Entry point untuk cleanup command execution
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-16
// Last Modified : 2025-12-16

package cleanup

import (
	"fmt"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/ui"
)

// ExecuteCleanupCommand adalah entry point untuk cleanup command execution
func (s *Service) ExecuteCleanupCommand(config types.CleanupEntryConfig) error {
	// Log prefix untuk tracking
	if config.LogPrefix != "" {
		s.Log.Infof("[%s] Memulai cleanup dengan mode: %s", config.LogPrefix, config.Mode)
	}

	// Tampilkan options jika diminta
	if config.ShowOptions {
		s.displayCleanupOptions()
	}

	// Jalankan cleanup berdasarkan mode
	var err error
	switch config.Mode {
	case "run":
		err = s.cleanupCore(false, s.CleanupOptions.Pattern)
	case "dry-run":
		err = s.cleanupCore(true, s.CleanupOptions.Pattern)
	case "pattern":
		if s.CleanupOptions.Pattern == "" {
			return ErrInvalidCleanupMode
		}
		err = s.cleanupCore(false, s.CleanupOptions.Pattern)
	default:
		return ErrInvalidCleanupMode
	}

	return err
}

// displayCleanupOptions menampilkan konfigurasi cleanup yang akan dijalankan
func (s *Service) displayCleanupOptions() {
	ui.PrintSubHeader("Konfigurasi Cleanup")

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

	ui.FormatTable([]string{"Setting", "Value"}, data)
}
