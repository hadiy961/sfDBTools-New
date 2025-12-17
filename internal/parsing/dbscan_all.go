package parsing

import (
	"sfDBTools/internal/types"
	"sfDBTools/pkg/consts"
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

	// Profile & key (konsisten dengan filter)
	if v := helper.GetStringFlagOrEnv(cmd, "profile", consts.ENV_SOURCE_PROFILE); v != "" {
		opts.ProfileInfo.Path = v
	}
	if v := helper.GetStringFlagOrEnv(cmd, "profile-key", consts.ENV_SOURCE_PROFILE_KEY); v != "" {
		opts.Encryption.Key = v
	}

	// Options lain yang diminta
	opts.ExcludeSystem = helper.GetBoolFlagOrEnv(cmd, "exclude-system", "")
	opts.Background = helper.GetBoolFlagOrEnv(cmd, "background", "")
	opts.ShowOptions = helper.GetBoolFlagOrEnv(cmd, "show-options", "")

	return opts, nil
}
