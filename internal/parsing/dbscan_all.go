package parsing

import (
	"sfDBTools/internal/types"
	defaultVal "sfDBTools/internal/defaultval"
	"sfDBTools/pkg/helper"

	"github.com/spf13/cobra"
)

// ParsingScanAllOptions membaca flag untuk perintah `dbscan all` dengan
// nama flag yang konsisten dengan perintah `dbscan filter`.
// Mengembalikan ScanOptions yang siap dipakai service.
func ParsingScanAllOptions(cmd *cobra.Command) (types.ScanOptions, error) {
	// Mulai dari default untuk mode all
	opts := defaultVal.GetDefaultScanOptions("all")

	// Profile & key (Shared Helper)
	PopulateProfileFlags(cmd, &opts.ProfileInfo)

	// Options lain yang diminta
	opts.ExcludeSystem = helper.GetBoolFlagOrEnv(cmd, "exclude-system", "")
	opts.Background = helper.GetBoolFlagOrEnv(cmd, "background", "")
	opts.ShowOptions = helper.GetBoolFlagOrEnv(cmd, "show-options", "")

	return opts, nil
}
