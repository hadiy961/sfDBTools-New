// File : internal/backup/mode_config.go
// Deskripsi : Centralized configuration for backup modes
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-11
// Last Modified : 2025-12-11

package backup

import (
	"fmt"
	"sfdbtools/internal/app/backup/model/types_backup"
	"sfdbtools/pkg/consts"
)

// GetExecutionConfig returns the execution configuration for a given backup mode
func GetExecutionConfig(mode string) (types_backup.ExecutionConfig, error) {
	// Centralized map configuration
	modeConfigs := map[string]types_backup.ExecutionConfig{
		consts.ModeSingle: {
			Mode:        consts.ModeSingle,
			HeaderTitle: "Database Backup - Single",
			LogPrefix:   "[Backup Single]",
			SuccessMsg:  "Proses backup database single selesai.",
		},
		consts.ModeSeparated: {
			Mode:        consts.ModeSeparated,
			HeaderTitle: "Database Backup - Separated",
			LogPrefix:   "[Backup Separated]",
			SuccessMsg:  "Proses backup database separated selesai.",
		},
		consts.ModeCombined: {
			Mode:        consts.ModeCombined,
			HeaderTitle: "Database Backup - Filter (Single File)",
			LogPrefix:   "[Backup Filter Single-File]",
			SuccessMsg:  "Proses backup database filter (single file) selesai.",
		},
		consts.ModeAll: {
			Mode:        consts.ModeAll,
			HeaderTitle: "Database Backup - All (Exclude Filters)",
			LogPrefix:   "[Backup All]",
			SuccessMsg:  "Proses backup all databases selesai.",
		},
		consts.ModePrimary: {
			Mode:        consts.ModePrimary,
			HeaderTitle: "Database Backup - Primary",
			LogPrefix:   "[Backup Primary]",
			SuccessMsg:  "Proses backup database primary selesai.",
		},
		consts.ModeSecondary: {
			Mode:        consts.ModeSecondary,
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
