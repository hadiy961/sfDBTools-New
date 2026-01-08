package parsing

import (
	dbscanmodel "sfdbtools/internal/app/dbscan/model"
	defaultVal "sfdbtools/internal/cli/defaults"
	"sfdbtools/internal/cli/resolver"
	appconfig "sfdbtools/internal/services/config"
	"sfdbtools/internal/shared/listx"

	"github.com/spf13/cobra"
)

// ParsingScanFilterOptions membaca flag untuk perintah `dbscan filter`.
// Hanya melakukan mapping flag ke struct ScanOptions.
// Pembacaan file dan penggabungan list dilakukan oleh caller (Service/Logic).
func ParsingScanFilterOptions(cmd *cobra.Command, cfg *appconfig.Config) (dbscanmodel.ScanOptions, error) {
	// Start dengan default untuk mode database (memuat fallback dari config/env)
	opts := defaultVal.GetDefaultScanOptions("database")

	// Profile & key (Shared Helper)
	if err := PopulateProfileFlags(cmd, &opts.ProfileInfo); err != nil {
		return dbscanmodel.ScanOptions{}, err
	}

	// Exclude system
	opts.ExcludeSystem = resolver.GetBoolFlagOrEnv(cmd, "exclude-system", "")

	// Include sources
	includeCSV := resolver.GetStringFlagOrEnv(cmd, "db", "")
	opts.IncludeList = listx.CSVToCleanList(includeCSV)

	includeFile := resolver.GetStringFlagOrEnv(cmd, "db-file", "")
	if includeFile != "" {
		opts.DatabaseList.File = includeFile
	}
	// Note: opts.DatabaseList.File might already be set by defaultVal from config.

	// Exclude sources
	excludeCSV := resolver.GetStringFlagOrEnv(cmd, "exclude-db", "")
	opts.ExcludeList = listx.CSVToCleanList(excludeCSV)

	excludeFile := resolver.GetStringFlagOrEnv(cmd, "exclude-file", "")
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
