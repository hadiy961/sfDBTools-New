package parsing

import (
	"sfDBTools/internal/types/types_backup"
	defaultVal "sfDBTools/internal/defaultval"
	"sfDBTools/pkg/helper"

	"github.com/spf13/cobra"
)

// ParsingBackupOptions melakukan parsing opsi untuk backup combined
func ParsingBackupOptions(cmd *cobra.Command, mode string) (types_backup.BackupDBOptions, error) {
	// Mulai dari default untuk mode combined
	opts := defaultVal.DefaultBackupOptions(mode)

	// Deteksi apakah ini command filter (untuk multi-select logic)
	// Command filter memiliki Use="filter", sedangkan all memiliki Use="all"
	isFilterCommand := cmd.Use == "filter"

	// Profile & key (Shared Helper)
	PopulateProfileFlags(cmd, &opts.Profile)

	// Filters (Shared Helper)
	PopulateFilterFlags(cmd, &opts.Filter)

	// Set flag untuk command filter agar bisa tampilkan multi-select jika tidak ada include/exclude
	// Ini digunakan di setup.go untuk menentukan apakah perlu multi-select atau tidak
	if isFilterCommand {
		opts.Filter.IsFilterCommand = true
	}

	// Compression (Shared Helper)
	PopulateCompressionFlags(cmd, &opts.Compression)

	// Encryption (Shared Helper)
	PopulateEncryptionFlags(cmd, &opts.Encryption)

	// Capture GTID (hanya untuk combined)
	if mode == "combined" {
		opts.CaptureGTID = helper.GetBoolFlagOrEnv(cmd, "capture-gtid", "")
	} else {
		opts.CaptureGTID = false
	}

	// Exclude User
	opts.ExcludeUser = helper.GetBoolFlagOrEnv(cmd, "exclude-user", "")

	// Dry Run
	opts.DryRun = helper.GetBoolFlagOrEnv(cmd, "dry-run", "")

	// Output Directory
	if v := helper.GetStringFlagOrEnv(cmd, "output-dir", ""); v != "" {
		opts.OutputDir = v
	}
	opts.Force = helper.GetBoolFlagOrEnv(cmd, "force", "")

	// Mode-specific options
	if mode == "single" {
		if v := helper.GetStringFlagOrEnv(cmd, "database", ""); v != "" {
			opts.DBName = v
		}
		if v := helper.GetStringFlagOrEnv(cmd, "filename", ""); v != "" {
			opts.File.Filename = v
		}
		opts.Filter.ExcludeData = helper.GetBoolFlagOrEnv(cmd, "exclude-data", "")
		opts.IncludeDmart = helper.GetBoolFlagOrEnv(cmd, "include-dmart", "")
		opts.IncludeTemp = helper.GetBoolFlagOrEnv(cmd, "include-temp", "")
		opts.IncludeArchive = helper.GetBoolFlagOrEnv(cmd, "include-archive", "")
	} else if mode == "primary" {
		// Mode primary sama seperti single, hanya tanpa --database flag
		if v := helper.GetStringFlagOrEnv(cmd, "filename", ""); v != "" {
			opts.File.Filename = v
		}
		opts.Filter.ExcludeData = helper.GetBoolFlagOrEnv(cmd, "exclude-data", "")
		opts.IncludeDmart = helper.GetBoolFlagOrEnv(cmd, "include-dmart", "")
		opts.IncludeTemp = helper.GetBoolFlagOrEnv(cmd, "include-temp", "")
		opts.IncludeArchive = helper.GetBoolFlagOrEnv(cmd, "include-archive", "")
		if v := helper.GetStringFlagOrEnv(cmd, "client-code", ""); v != "" {
			opts.ClientCode = v
		}
	} else if mode == "secondary" {
		// Mode secondary sama seperti primary, hanya untuk database dengan suffix _secondary
		if v := helper.GetStringFlagOrEnv(cmd, "filename", ""); v != "" {
			opts.File.Filename = v
		}
		opts.Filter.ExcludeData = helper.GetBoolFlagOrEnv(cmd, "exclude-data", "")
		opts.IncludeDmart = helper.GetBoolFlagOrEnv(cmd, "include-dmart", "")
		opts.IncludeTemp = helper.GetBoolFlagOrEnv(cmd, "include-temp", "")
		opts.IncludeArchive = helper.GetBoolFlagOrEnv(cmd, "include-archive", "")
		if v := helper.GetStringFlagOrEnv(cmd, "client-code", ""); v != "" {
			opts.ClientCode = v
		}
		if v := helper.GetStringFlagOrEnv(cmd, "instance", ""); v != "" {
			opts.Instance = v
		}
	}

	// Ticket (wajib untuk semua mode)
	if v := helper.GetStringFlagOrEnv(cmd, "ticket", ""); v != "" {
		opts.Ticket = v
	}

	// Mode
	opts.Mode = mode

	return opts, nil
}
