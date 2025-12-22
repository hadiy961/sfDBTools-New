package backup

import (
	"sfDBTools/internal/cleanup"
	"sfDBTools/pkg/ui"
)

// SetCurrentBackupFile mencatat file backup yang sedang dibuat
func (s *Service) SetCurrentBackupFile(filePath string) {
	s.WithLock(func() {
		s.currentBackupFile = filePath
		s.backupInProgress = true
	})
}

// ClearCurrentBackupFile menghapus catatan file backup setelah selesai
func (s *Service) ClearCurrentBackupFile() {
	s.WithLock(func() {
		s.currentBackupFile = ""
		s.backupInProgress = false
	})
}

// HandleShutdown menangani graceful shutdown saat CTRL+C atau interrupt
func (s *Service) HandleShutdown() {
	var shouldRemoveFile bool
	var fileToRemove string

	s.WithLock(func() {
		if s.backupInProgress && s.currentBackupFile != "" {
			shouldRemoveFile = true
			fileToRemove = s.currentBackupFile
			s.currentBackupFile = ""
			s.backupInProgress = false
		}
	})

	if shouldRemoveFile {
		ui.RunWithSpinnerSuspended(func() {
			s.Log.Warn("Proses backup dihentikan, melakukan rollback...")
			if err := cleanup.CleanupPartialBackup(fileToRemove, s.Log); err != nil {
				s.Log.Errorf("Gagal menghapus file backup: %v", err)
			}
		})
	} else {
		ui.RunWithSpinnerSuspended(func() {
			s.Log.Warn("Menerima signal interrupt, tidak ada proses backup aktif")
		})
	}

	s.Cancel()
}
