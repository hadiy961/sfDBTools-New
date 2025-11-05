package backup

import (
	"context"
	"os"
	"sfDBTools/internal/appconfig"
	"sfDBTools/internal/applog"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/database"
	"sync"
)

type Service struct {
	Config          *appconfig.Config
	Log             applog.Logger
	DBInfo          *types.DBInfo
	Profile         *types.ProfileInfo
	BackupDBOptions *types.BackupDBOptions
	BackupEntry     *types.BackupEntryConfig
	Client          *database.Client // Client database aktif selama backup

	// Untuk graceful shutdown
	cancelFunc        context.CancelFunc
	currentBackupFile string
	backupInProgress  bool
	mu                sync.Mutex // Melindungi akses ke currentBackupFile dan backupInProgress
}

func NewBackupService(logs applog.Logger, cfg *appconfig.Config, backup interface{}) *Service {
	svc := &Service{
		Log:    logs,
		Config: cfg, // Perbaikan: set field Config agar tidak nil
	}

	if backup != nil {
		switch v := backup.(type) {
		case *types.BackupDBOptions:
			svc.BackupDBOptions = v
			svc.Profile = &v.Profile
			svc.BackupEntry = &v.Entry
			svc.DBInfo = &v.Profile.DBInfo
		default:
			logs.Warn("Tipe backup tidak dikenali dalam Service")
		}
	} else {
		logs.Warn("Tipe backup tidak dikenali dalam Service")
	}

	return svc
}

// SetCancelFunc menyimpan cancel function untuk graceful shutdown
func (s *Service) SetCancelFunc(cancel context.CancelFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cancelFunc = cancel
}

// SetCurrentBackupFile mencatat file backup yang sedang dibuat
func (s *Service) SetCurrentBackupFile(filePath string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.currentBackupFile = filePath
	s.backupInProgress = true
}

// ClearCurrentBackupFile menghapus catatan file backup setelah selesai
func (s *Service) ClearCurrentBackupFile() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.currentBackupFile = ""
	s.backupInProgress = false
}

// HandleShutdown menangani graceful shutdown saat CTRL+C atau interrupt
func (s *Service) HandleShutdown() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.backupInProgress && s.currentBackupFile != "" {
		s.Log.Warn("Proses backup dihentikan, melakukan rollback...")

		// Hapus file backup yang belum selesai
		if _, err := os.Stat(s.currentBackupFile); err == nil {
			if removeErr := os.Remove(s.currentBackupFile); removeErr == nil {
				s.Log.Infof("File backup yang belum selesai berhasil dihapus: %s", s.currentBackupFile)
			} else {
				s.Log.Errorf("Gagal menghapus file backup: %v", removeErr)
			}
		}

		s.currentBackupFile = ""
		s.backupInProgress = false
	}

	// Cancel context jika ada
	if s.cancelFunc != nil {
		s.cancelFunc()
	}
}
