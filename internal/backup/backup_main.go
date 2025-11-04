package backup

import (
	"sfDBTools/internal/appconfig"
	"sfDBTools/internal/applog"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/database"
)

type Service struct {
	Config          *appconfig.Config
	Log             applog.Logger
	DBInfo          *types.DBInfo
	Profile         *types.ProfileInfo
	BackupDBOptions *types.BackupDBOptions
	BackupEntry     *types.BackupEntryConfig
	Client          *database.Client // Client database aktif selama backup
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
