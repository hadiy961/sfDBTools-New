package parsing

import (
	defaultVal "sfDBTools/internal/cli/defaults"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/helper"

	"github.com/spf13/cobra"
)

// ParsingCleanupOptions membaca flag untuk perintah cleanup
// Mengembalikan CleanupOptions yang siap dipakai service.
func ParsingCleanupOptions(cmd *cobra.Command) (types.CleanupOptions, error) {
	// Mulai dari default
	opts := defaultVal.DefaultCleanupOptions()

	// Days - retention days untuk cleanup
	if v := helper.GetIntFlagOrEnv(cmd, "days", ""); v != 0 {
		opts.Days = v
	}

	// Pattern - glob pattern untuk filter files
	if v := helper.GetStringFlagOrEnv(cmd, "pattern", ""); v != "" {
		opts.Pattern = v
	}

	// Background mode
	opts.Background = helper.GetBoolFlagOrEnv(cmd, "background", "")

	// Dry-run mode
	opts.DryRun = helper.GetBoolFlagOrEnv(cmd, "dry-run", "")

	return opts, nil
}
