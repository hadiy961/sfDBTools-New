package cleanupcmd

import (
	"sfDBTools/internal/cleanup"
	appdeps "sfDBTools/internal/deps"

	"github.com/spf13/cobra"
)

// CmdCleanupDryRun menampilkan preview file yang akan dihapus tanpa menghapus apapun
var CmdCleanupDryRun = &cobra.Command{
	Use:     "dry-run",
	Aliases: []string{"preview"},
	Short:   "Tampilkan pratinjau file yang akan dihapus (tanpa menghapus)",
	Long: `Menampilkan daftar file backup yang lebih tua dari kebijakan retensi dan AKAN dihapus jika pembersihan dijalankan.
Tidak ada file yang akan dihapus pada mode ini.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := cleanup.ExecuteCleanup(cmd, appdeps.Deps, "dry-run"); err != nil {
			appdeps.Deps.Logger.Error("cleanup dry-run gagal: " + err.Error())
		}
	},
}

func init() {
	CmdCleanupMain.AddCommand(CmdCleanupDryRun)
}
