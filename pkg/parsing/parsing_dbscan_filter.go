package parsing

import (
	"fmt"
	"sfDBTools/internal/appconfig"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/consts"
	defaultVal "sfDBTools/pkg/defaultval"
	"sfDBTools/pkg/fsops"
	"sfDBTools/pkg/helper"

	"github.com/spf13/cobra"
)

// ParsingScanFilterOptions membaca flag untuk perintah `dbscan filter` lalu
// menyusun types.ScanOptions lengkap (termasuk include-minus-exclude, fallback config).
func ParsingScanFilterOptions(cmd *cobra.Command, cfg *appconfig.Config) (types.ScanOptions, error) {
	// Start dengan default untuk mode database (memuat fallback dari config/env)
	opts := defaultVal.GetDefaultScanOptions("database")

	// Profile & key
	profile := helper.GetStringFlagOrEnv(cmd, "profile", consts.ENV_SOURCE_PROFILE)
	if profile != "" {
		opts.ProfileInfo.Path = profile
	}
	key := helper.GetStringFlagOrEnv(cmd, "profile-key", consts.ENV_SOURCE_PROFILE_KEY)
	if key != "" {
		opts.Encryption.Key = key
	}

	// Exclude system
	opts.ExcludeSystem = helper.GetBoolFlagOrEnv(cmd, "exclude-system", "")

	// Include sources
	includeCSV := helper.GetStringFlagOrEnv(cmd, "db", "")
	includeFromFlags := helper.CSVToCleanList(includeCSV)
	includeFile := helper.GetStringFlagOrEnv(cmd, "db-file", "")
	if includeFile == "" && opts.DatabaseList.File != "" {
		includeFile = opts.DatabaseList.File // fallback dari default config
	}
	includeFromFile := []string{}
	if includeFile != "" {
		lines, err := fsops.ReadLinesFromFile(includeFile)
		if err != nil {
			return opts, fmt.Errorf("gagal membaca db-file %s: %w", includeFile, err)
		}
		includeFromFile = helper.ListTrimNonEmpty(lines)
	}

	// Exclude sources
	excludeCSV := helper.GetStringFlagOrEnv(cmd, "exclude-db", "")
	excludeFromFlags := helper.CSVToCleanList(excludeCSV)
	excludeFile := helper.GetStringFlagOrEnv(cmd, "exclude-db-file", "")
	excludeFromFile := []string{}
	if excludeFile != "" {
		lines, err := fsops.ReadLinesFromFile(excludeFile)
		if err != nil {
			return opts, fmt.Errorf("gagal membaca exclude-db-file %s: %w", excludeFile, err)
		}
		excludeFromFile = helper.ListTrimNonEmpty(lines)
	} else {
		// Fallback ke konfigurasi bila tersedia
		if len(cfg.Backup.Exclude.Databases) > 0 {
			excludeFromFile = helper.ListTrimNonEmpty(cfg.Backup.Exclude.Databases)
		}
	}

	// Merge & validate
	includeUnion := helper.ListUnique(append(includeFromFlags, includeFromFile...))
	excludeUnion := helper.ListUnique(append(excludeFromFlags, excludeFromFile...))

	if len(includeUnion) == 0 && len(excludeUnion) == 0 {
		return opts, fmt.Errorf("minimal salah satu flag harus digunakan: gunakan --db/--db-file untuk include atau --exclude-db/--exclude-db-file untuk exclude")
	}

	if len(includeUnion) > 0 && len(excludeUnion) > 0 {
		includeUnion = helper.ListSubtract(includeUnion, excludeUnion)
		excludeUnion = []string{}
	}

	opts.IncludeList = includeUnion
	opts.ExcludeList = excludeUnion
	opts.DatabaseList.File = "" // sudah diproses lebih awal

	return opts, nil
}
