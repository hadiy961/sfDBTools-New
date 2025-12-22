package backup

import (
	"sfDBTools/internal/applog"
	"sfDBTools/internal/types/types_backup"
)

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
