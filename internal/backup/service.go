// File : internal/backup/service.go
// Deskripsi : Service utama untuk backup operations dengan interface implementation
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-05
// Last Modified : 2025-12-05

package backup

import (
	"sfDBTools/internal/appconfig"
	"sfDBTools/internal/applog"
	"sfDBTools/internal/backup/modes"
	"sfDBTools/internal/types"
	"sfDBTools/internal/types/types_backup"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/errorlog"
	"sfDBTools/pkg/servicehelper"
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
		logDir = consts.DefaultLogDir
	}

	svc := &Service{
		Log:               logs,
		Config:            cfg,
		ErrorLog:          errorlog.NewErrorLogger(logs, logDir, consts.FeatureBackup),
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

// Verify interface implementation at compile time
var _ modes.BackupService = (*Service)(nil)
