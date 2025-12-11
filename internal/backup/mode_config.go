// File : internal/backup/mode_config.go
// Deskripsi : Centralized configuration for backup modes
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-11
// Last Modified : 2025-12-11

package backup

import (
	"fmt"
	"sfDBTools/internal/types/types_backup"
)

// GetExecutionConfig returns the execution configuration for a given backup mode
func GetExecutionConfig(mode string) (types_backup.ExecutionConfig, error) {
	// Centralized map configuration
	modeConfigs := map[string]types_backup.ExecutionConfig{
		"single": {
			Mode:        "single",
			HeaderTitle: "Database Backup - Single",
			LogPrefix:   "[Backup Single]",
			SuccessMsg:  "Proses backup database single selesai.",
		},
		"separated": {
			Mode:        "separated",
			HeaderTitle: "Database Backup - Separated",
			LogPrefix:   "[Backup Separated]",
			SuccessMsg:  "Proses backup database separated selesai.",
		},
		"combined": {
			Mode:        "combined",
			HeaderTitle: "Database Backup - Combined",
			LogPrefix:   "[Backup Combined]",
			SuccessMsg:  "Proses backup database combined selesai.",
		},
		"primary": {
			Mode:        "primary",
			HeaderTitle: "Database Backup - Primary",
			LogPrefix:   "[Backup Primary]",
			SuccessMsg:  "Proses backup database primary selesai.",
		},
		"secondary": {
			Mode:        "secondary",
			HeaderTitle: "Database Backup - Secondary",
			LogPrefix:   "[Backup Secondary]",
			SuccessMsg:  "Proses backup database secondary selesai.",
		},
	}

	config, exists := modeConfigs[mode]
	if !exists {
		return types_backup.ExecutionConfig{}, fmt.Errorf("mode backup tidak dikenali: %s", mode)
	}

	return config, nil
}
