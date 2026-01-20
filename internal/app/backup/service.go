// File : internal/app/backup/service.go
// Deskripsi : Service utama untuk backup operations dengan interface implementation
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-05
// Last Modified : 20 Januari 2026
package backup

import (
	"fmt"
	"os"
	"sfdbtools/internal/app/backup/gtid"
	"sfdbtools/internal/app/backup/model/types_backup"
	"sfdbtools/internal/app/backup/modes"
	"sfdbtools/internal/domain"
	appconfig "sfdbtools/internal/services/config"
	applog "sfdbtools/internal/services/log"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/database"
	"sfdbtools/internal/shared/errorlog"
	"sfdbtools/internal/shared/servicehelper"
	"sfdbtools/internal/ui/print"
	"sfdbtools/internal/ui/progress"
	"sync"
)

// BackupExecutionState menyimpan state yang mutable selama eksekusi backup.
// State ini dibuat per-execution untuk menghindari race condition.
type BackupExecutionState struct {
	GTIDInfo          *gtid.GTIDInfo
	CurrentBackupFile string
	BackupInProgress  bool
	ExcludedDatabases []string
	CleanupOnCancel   bool // Flag untuk cleanup partial files saat context cancelled
	CleanupLog        applog.Logger
	mu                sync.Mutex
}

// SetCurrentBackupFile mencatat file backup yang sedang dibuat (thread-safe)
func (s *BackupExecutionState) SetCurrentBackupFile(filePath string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.CurrentBackupFile = filePath
	s.BackupInProgress = true
}

// ClearCurrentBackupFile menghapus catatan file backup setelah selesai (thread-safe)
func (s *BackupExecutionState) ClearCurrentBackupFile() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.CurrentBackupFile = ""
	s.BackupInProgress = false
}

// GetCurrentBackupFile returns current backup file path (thread-safe)
func (s *BackupExecutionState) GetCurrentBackupFile() (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.CurrentBackupFile, s.BackupInProgress
}

// EnableCleanup mengaktifkan cleanup on cancel dengan logger yang diberikan.
// Dipanggil dari ExecuteBackupLoop sebelum memulai backup loop.
func (s *BackupExecutionState) EnableCleanup(log applog.Logger) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.CleanupOnCancel = true
	s.CleanupLog = log
}

// Cleanup menghapus partial backup file saat context cancelled (best-effort cleanup).
// Dipanggil dari context cancellation handler untuk cleanup incomplete backups.
func (s *BackupExecutionState) Cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.CleanupOnCancel || s.CurrentBackupFile == "" || !s.BackupInProgress {
		return
	}

	log := s.CleanupLog
	if log == nil {
		// Fallback jika log tidak diset (tidak seharusnya terjadi)
		return
	}

	// Remove partial backup file (best effort)
	if err := os.Remove(s.CurrentBackupFile); err != nil {
		// Log error tapi jangan fail (cleanup is best-effort)
		log.Debugf("Gagal cleanup partial backup file %s: %v", s.CurrentBackupFile, err)
	} else {
		log.Infof("✓ Cleanup partial backup file: %s", s.CurrentBackupFile)
	}

	// Remove metadata file jika ada (best effort)
	metaFile := s.CurrentBackupFile + ".meta.json"
	if err := os.Remove(metaFile); err != nil {
		log.Debugf("Gagal cleanup metadata file %s: %v (mungkin belum dibuat)", metaFile, err)
	} else {
		log.Infof("✓ Cleanup metadata file: %s", metaFile)
	}

	s.CurrentBackupFile = ""
	s.BackupInProgress = false
}

// Service adalah service utama untuk backup operations.
// Service bersifat stateless dan immutable setelah dibuat, sehingga thread-safe.
type Service struct {
	servicehelper.BaseService

	Config          *appconfig.Config
	Log             applog.Logger
	ErrorLog        *errorlog.ErrorLogger
	DBInfo          *domain.DBInfo
	Profile         *domain.ProfileInfo
	BackupDBOptions *types_backup.BackupDBOptions
	BackupEntry     *types_backup.BackupEntryConfig
	Client          *database.Client
}

// NewBackupService membuat instance baru Service (stateless)
func NewBackupService(logs applog.Logger, cfg *appconfig.Config, backup interface{}) *Service {
	logDir := cfg.Log.Output.File.Dir
	if logDir == "" {
		logDir = consts.DefaultLogDir
	}

	svc := &Service{
		Log:      logs,
		Config:   cfg,
		ErrorLog: errorlog.NewErrorLogger(logs, logDir, consts.FeatureBackup),
	}

	if backup != nil {
		switch v := backup.(type) {
		case *types_backup.BackupDBOptions:
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

// NewExecutionState membuat execution state baru untuk backup operation
func NewExecutionState() *BackupExecutionState {
	return &BackupExecutionState{
		ExcludedDatabases: []string{},
	}
}

// Verify interface implementation at compile time
var _ modes.BackupService = (*Service)(nil)

// =============================================================================
// Interface helpers (used by modes.BackupService)
// =============================================================================

// GetLog returns logger instance
func (s *Service) GetLog() applog.Logger { return s.Log }

// GetConfig returns config instance
func (s *Service) GetConfig() *appconfig.Config { return s.Config }

// GetOptions returns backup options
func (s *Service) GetOptions() *types_backup.BackupDBOptions { return s.BackupDBOptions }

// ToBackupResult konversi BackupLoopResult ke BackupResult
func (s *Service) ToBackupResult(loopResult types_backup.BackupLoopResult) types_backup.BackupResult {
	return types_backup.BackupResult{
		BackupInfo:          loopResult.BackupInfos,
		FailedDatabaseInfos: loopResult.FailedDBs,
		Errors:              loopResult.Errors,
	}
}

// =============================================================================
// Service shutdown
// =============================================================================

// HandleShutdown menangani graceful shutdown saat CTRL+C atau interrupt.
// State di-pass sebagai parameter untuk menghindari shared mutable state di Service.
func (s *Service) HandleShutdown(state *BackupExecutionState) {
	if state == nil {
		progress.RunWithSpinnerSuspended(func() {
			s.Log.Warn("Menerima signal interrupt, tidak ada execution state")
		})
		s.Cancel()
		return
	}

	fileToRemove, inProgress := state.GetCurrentBackupFile()
	if inProgress && fileToRemove != "" {
		// Clear state sebelum cleanup
		state.ClearCurrentBackupFile()

		progress.RunWithSpinnerSuspended(func() {
			s.Log.Warn("Proses backup dihentikan, melakukan rollback...")
			if err := os.Remove(fileToRemove); err != nil {
				s.Log.Errorf("Gagal menghapus file backup: %v", err)
				print.PrintError(fmt.Sprintf("⚠ WARNING: File backup partial mungkin masih tersisa: %s", fileToRemove))
				print.PrintError("Silakan hapus manual jika diperlukan.")
			} else {
				s.Log.Infof("File backup yang belum selesai berhasil dihapus: %s", fileToRemove)
				s.Log.Info("File backup partial berhasil dihapus")
			}
		})
	} else {
		progress.RunWithSpinnerSuspended(func() {
			s.Log.Warn("Menerima signal interrupt, tidak ada proses backup aktif")
		})
	}

	s.Cancel()
}
