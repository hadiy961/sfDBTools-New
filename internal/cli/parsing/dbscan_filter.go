package parsing

import (
	dbscanmodel "sfdbtools/internal/app/dbscan/model"
	defaultVal "sfdbtools/internal/cli/defaults"
	appconfig "sfdbtools/internal/services/config"
	"sfdbtools/pkg/helper"

	"github.com/spf13/cobra"
)

// ParsingScanFilterOptions membaca flag untuk perintah `dbscan filter`.
// Hanya melakukan mapping flag ke struct ScanOptions.
// Pembacaan file dan penggabungan list dilakukan oleh caller (Service/Logic).
func ParsingScanFilterOptions(cmd *cobra.Command, cfg *appconfig.Config) (dbscanmodel.ScanOptions, error) {
	// Start dengan default untuk mode database (memuat fallback dari config/env)
	opts := defaultVal.GetDefaultScanOptions("database")

	// Profile & key (Shared Helper)
	PopulateProfileFlags(cmd, &opts.ProfileInfo)

	// Exclude system
	opts.ExcludeSystem = helper.GetBoolFlagOrEnv(cmd, "exclude-system", "")

	// Include sources
	includeCSV := helper.GetStringFlagOrEnv(cmd, "db", "")
	opts.IncludeList = helper.CSVToCleanList(includeCSV)

	includeFile := helper.GetStringFlagOrEnv(cmd, "db-file", "")
	if includeFile != "" {
		opts.DatabaseList.File = includeFile
	}
	// Note: opts.DatabaseList.File might already be set by defaultVal from config.

	// Exclude sources
	excludeCSV := helper.GetStringFlagOrEnv(cmd, "exclude-db", "")
	opts.ExcludeList = helper.CSVToCleanList(excludeCSV)

	excludeFile := helper.GetStringFlagOrEnv(cmd, "exclude-file", "")
	if excludeFile != "" {
		opts.ExcludeFile = excludeFile
	} else {
		// Fallback ke konfigurasi bila tidak ada flag exclude-file
		// Kita masukkan ke ExcludeList agar nanti di-merge
		if len(cfg.Backup.Exclude.Databases) > 0 {
			opts.ExcludeList = append(opts.ExcludeList, cfg.Backup.Exclude.Databases...)
		}
	}

	return opts, nil
}
