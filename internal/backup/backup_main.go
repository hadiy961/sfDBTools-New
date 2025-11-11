package backup

import (
	"os"
	"sfDBTools/internal/appconfig"
	"sfDBTools/internal/applog"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/errorlog"
	"sfDBTools/pkg/servicehelper"
)

type Service struct {
	servicehelper.BaseService // Embed base service untuk graceful shutdown functionality

	Config          *appconfig.Config
	Log             applog.Logger
	ErrorLog        *errorlog.ErrorLogger
	DBInfo          *types.DBInfo
	Profile         *types.ProfileInfo
	BackupDBOptions *types.BackupDBOptions
	BackupEntry     *types.BackupEntryConfig
	Client          *database.Client // Client database aktif selama backup

	// Backup-specific state
	currentBackupFile string
	backupInProgress  bool
}

func NewBackupService(logs applog.Logger, cfg *appconfig.Config, backup interface{}) *Service {
	logDir := cfg.Log.Output.File.Dir
	if logDir == "" {
		logDir = "/var/log/sfDBTools"
	}

	svc := &Service{
		Log:      logs,
		Config:   cfg, // Perbaikan: set field Config agar tidak nil
		ErrorLog: errorlog.NewErrorLogger(logs, logDir, "backup"),
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
			s.Log.Warn("Proses backup dihentikan, melakukan rollback...")
			shouldRemoveFile = true
			fileToRemove = s.currentBackupFile
			s.currentBackupFile = ""
			s.backupInProgress = false
		}
	})

	// Remove file outside of lock to avoid holding lock during I/O
	if shouldRemoveFile {
		if _, err := os.Stat(fileToRemove); err == nil {
			if removeErr := os.Remove(fileToRemove); removeErr == nil {
				s.Log.Infof("File backup yang belum selesai berhasil dihapus: %s", fileToRemove)
			} else {
				s.Log.Errorf("Gagal menghapus file backup: %v", removeErr)
			}
		}
	}

	s.Cancel() // Panggil cancel dari BaseService
}
