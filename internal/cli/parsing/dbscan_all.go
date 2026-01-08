package parsing

import (
	dbscanmodel "sfdbtools/internal/app/dbscan/model"
	defaultVal "sfdbtools/internal/cli/defaults"
	resolver "sfdbtools/internal/cli/resolver"

	"github.com/spf13/cobra"
)

// ParsingScanAllOptions membaca flag untuk perintah `dbscan all` dengan
// nama flag yang konsisten dengan perintah `dbscan filter`.
// Mengembalikan ScanOptions yang siap dipakai service.
func ParsingScanAllOptions(cmd *cobra.Command) (dbscanmodel.ScanOptions, error) {
	// Mulai dari default untuk mode all
	opts := defaultVal.GetDefaultScanOptions("all")

	// Profile & key (Shared Helper)
	if err := PopulateProfileFlags(cmd, &opts.ProfileInfo); err != nil {
		return dbscanmodel.ScanOptions{}, err
	}

	// Options lain yang diminta
	opts.ExcludeSystem = resolver.GetBoolFlagOrEnv(cmd, "exclude-system", "")
	opts.Background = resolver.GetBoolFlagOrEnv(cmd, "background", "")
	opts.ShowOptions = resolver.GetBoolFlagOrEnv(cmd, "show-options", "")

	return opts, nil
}
