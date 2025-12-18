// File : internal/backup/service.go
// Deskripsi : Service utama untuk backup operations dengan interface implementation
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-05
// Last Modified : 2025-12-05

package backup

import (
	"context"
	"sfDBTools/internal/appconfig"
	"sfDBTools/internal/applog"
	"sfDBTools/internal/backup/metadata"
	"sfDBTools/internal/backup/modes"
	"sfDBTools/internal/cleanup"
	"sfDBTools/internal/types"
	"sfDBTools/internal/types/types_backup"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/errorlog"
	"sfDBTools/pkg/servicehelper"
	"sfDBTools/pkg/ui"
)

// Service adalah service utama untuk backup operations
type Service struct {
	servicehelper.BaseService

	Config          *appconfig.Config
	Log             applog.Logger
	ErrorLog        *errorlog.ErrorLogger
	DBInfo          *types.DBInfo
	Profile         *types.ProfileInfo
	BackupDBOptions *types_backup.BackupDBOptions
	BackupEntry     *types_backup.BackupEntryConfig
	Client          *database.Client

	// Backup-specific state
	currentBackupFile string
	backupInProgress  bool
	gtidInfo          *database.GTIDInfo
	excludedDatabases []string // List database yang dikecualikan (untuk mode 'all')
}

// NewBackupService membuat instance baru Service
func NewBackupService(logs applog.Logger, cfg *appconfig.Config, backup interface{}) *Service {
	logDir := cfg.Log.Output.File.Dir
	if logDir == "" {
		logDir = "/var/log/sfDBTools"
	}

	svc := &Service{
		Log:               logs,
		Config:            cfg,
		ErrorLog:          errorlog.NewErrorLogger(logs, logDir, "backup"),
		excludedDatabases: []string{}, // Initialize dengan empty slice, bukan nil
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

// =============================================================================
// GTID and User Grants Helpers
// =============================================================================

// CaptureAndSaveGTID mengambil dan menyimpan GTID info jika diperlukan
func (s *Service) CaptureAndSaveGTID(ctx context.Context, backupFilePath string) error {
	if !s.BackupDBOptions.CaptureGTID {
		return nil
	}

	s.Log.Info("Mengambil informasi GTID sebelum backup...")
	gtidInfo, err := s.Client.GetFullGTIDInfo(ctx)
	if err != nil {
		s.Log.Warnf("Gagal mendapatkan GTID: %v", err)
		return nil
	}

	s.Log.Infof("GTID berhasil diambil: File=%s, Pos=%d", gtidInfo.MasterLogFile, gtidInfo.MasterLogPos)

	// Simpan GTID info ke service untuk dimasukkan ke metadata nanti
	s.gtidInfo = gtidInfo

	return nil
}

// GetTotalDatabaseCount mengambil total database dari server
func (s *Service) GetTotalDatabaseCount(ctx context.Context, dbFiltered []string) int {
	allDatabases, err := s.Client.GetDatabaseList(ctx)
	totalDBFound := len(allDatabases)
	if err != nil {
		s.Log.Warnf("Gagal mendapatkan total database: %v, menggunakan fallback", err)
		totalDBFound = len(dbFiltered)
	}
	return totalDBFound
}

// ExportUserGrantsIfNeeded export user grants jika diperlukan
// Delegates to metadata.ExportUserGrantsIfNeededWithLogging dengan BackupDBOptions.ExcludeUser
func (s *Service) ExportUserGrantsIfNeeded(ctx context.Context, referenceBackupFile string, databases []string) string {
	// Skip export user grants jika dry-run mode
	if s.BackupDBOptions.DryRun {
		s.Log.Info("[DRY-RUN] Skip export user grants (dry-run mode aktif)")
		return ""
	}
	return metadata.ExportUserGrantsIfNeededWithLogging(ctx, s.Client, s.Log, referenceBackupFile, s.BackupDBOptions.ExcludeUser, databases)
}

// UpdateMetadataUserGrantsPath update metadata dengan actual user grants path
func (s *Service) UpdateMetadataUserGrantsPath(backupFilePath string, userGrantsPath string) {
	if err := metadata.UpdateMetadataUserGrantsFile(backupFilePath, userGrantsPath, s.Log); err != nil {
		s.Log.Warnf("Gagal update metadata user grants path: %v", err)
	}
}

// =============================================================================
// Interface Implementation - modes.BackupService
// =============================================================================

// GetLog returns logger instance
func (s *Service) GetLog() applog.Logger { return s.Log }

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

// Verify interface implementation at compile time
var _ modes.BackupService = (*Service)(nil)
