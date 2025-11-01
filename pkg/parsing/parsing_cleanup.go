package parsing

import (
	"sfDBTools/internal/types"
	"sfDBTools/pkg/consts"
	defaultVal "sfDBTools/pkg/defaultval"
	"sfDBTools/pkg/helper"

	"github.com/spf13/cobra"
)

// ParsingScanAllOptions membaca flag untuk perintah `dbscan all` dengan
// nama flag yang konsisten dengan perintah `dbscan filter`.
// Mengembalikan ScanOptions yang siap dipakai service.
func ParsingCleanupOptions(cmd *cobra.Command) (types.CleanupOptions, error) {
	// Mulai dari default untuk mode all
	opts := defaultVal.DefaultCleanupOptions()

	// Profile & key (konsisten dengan filter)
	if v := helper.GetIntFlagOrEnv(cmd, "days", ""); v != 0 {
		opts.Days = v
	}

	// Options lain yang diminta
	opts.Background = helper.GetBoolFlagOrEnv(cmd, "background", consts.ENV_DAEMON_MODE)

	return opts, nil
}
