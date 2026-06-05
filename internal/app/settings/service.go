package settings

import (
	appconfig "sfdbtools/internal/services/config"
	appdeps "sfdbtools/internal/cli/deps"
	"sfdbtools/internal/shared/database"
	"sfdbtools/internal/ui/print"
	"sfdbtools/internal/ui/prompt"
	"strings"
)

type SettingItem struct {
	Key      string
	Value    string
	Category string
}

type Service struct {
	config *appconfig.Config
	deps   *appdeps.Dependencies
}

func NewService(deps *appdeps.Dependencies) *Service {
	return &Service{
		config: deps.Config,
		deps:   deps,
	}
}

func (s *Service) RunMenu() error {
	for {
		print.PrintAppHeader("Settings Management")
		
		options := []string{
			"1. View Configuration (Lihat Semua)",
			"2. Edit Settings (Ubah Parameter)",
			"3. Manage Database Profiles (Koneksi DB)",
			"4. Cloud Sync & Hub (Sinkronisasi)",
			"5. Health & Diagnostics (Cek Kesehatan)",
			"6. View Audit Logs (Log Aktivitas)",
			"7. Systemd & Maintenance (Lainnya)",
			"0. Back to Main Menu",
		}

		selected, _, err := prompt.SelectOne("Select Action:", options, 0)
		if err != nil || strings.Contains(selected, "Back") {
			return nil
		}

		idx := selected[0:1]
		switch idx {
		case "1":
			s.ViewAllSettings()
		case "2":
			s.EditSettingsMenu()
		case "3":
			s.ManageProfilesMenu()
		case "4":
			db, _ := database.GetSQLite()
			s.CloudMenu(db)
		case "5":
			s.RunConnectivityDiagnostics()
		case "6":
			db, _ := database.GetSQLite()
			s.ViewAuditLogs(db)
		case "7":
			s.MaintenanceMenu()
		}
		
		if idx != "0" {
			prompt.WaitForEnter()
		}
	}
}
