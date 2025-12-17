package cleanupcmd

import (
	"sfDBTools/internal/cleanup"
	"sfDBTools/internal/types"
	defaultVal "sfDBTools/internal/defaultval"
	"sfDBTools/internal/flags"

	"github.com/spf13/cobra"
)

// CmdCleanupRun menjalankan pembersihan sebenarnya berdasarkan retention policy
var CmdCleanupRun = &cobra.Command{
	Use:   "run",
	Short: "Jalankan pembersihan file backup lama (sesuai retensi)",
	Long: `Menjalankan pembersihan file backup lama sesuai konfigurasi retensi (backup.retention.days).
File yang lebih tua dari jumlah hari retensi akan dihapus.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := cleanup.ExecuteCleanup(cmd, types.Deps, "run"); err != nil {
			types.Deps.Logger.Error("cleanup gagal: " + err.Error())
		}
	},
}

func init() {
	// Set default values
	defaultOpts := defaultVal.DefaultCleanupOptions()

	flags.AddCleanupFlags(CmdCleanupRun, &defaultOpts)
}
