package cmdcleanup

import (
	"fmt"
	"sfDBTools/internal/cleanup"
	"sfDBTools/internal/types"

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
		if types.Deps == nil || types.Deps.Config == nil || types.Deps.Logger == nil {
			fmt.Println("✗ Dependencies tidak tersedia. Pastikan aplikasi diinisialisasi dengan benar.")
			return
		}

		// Inject logger dan config ke package cleanup
		cleanup.Logger = types.Deps.Logger
		cleanup.SetConfig(types.Deps.Config)

		if err := cleanup.CleanupDryRun(); err != nil {
			types.Deps.Logger.Errorf("Cleanup dry-run gagal: %v", err)
			return
		}
	},
}

func init() {
	CmdCleanupMain.AddCommand(CmdCleanupDryRun)
}
