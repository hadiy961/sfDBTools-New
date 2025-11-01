package cmdcleanup

import (
	"fmt"
	"sfDBTools/internal/cleanup"
	"sfDBTools/internal/types"

	"github.com/spf13/cobra"
)

// CmdCleanupRun menjalankan pembersihan sebenarnya berdasarkan retention policy
var CmdCleanupRun = &cobra.Command{
	Use:   "run",
	Short: "Jalankan pembersihan file backup lama (sesuai retensi)",
	Long: `Menjalankan pembersihan file backup lama sesuai konfigurasi retensi (backup.retention.days).
File yang lebih tua dari jumlah hari retensi akan dihapus.`,
	Run: func(cmd *cobra.Command, args []string) {
		if types.Deps == nil || types.Deps.Config == nil || types.Deps.Logger == nil {
			fmt.Println("✗ Dependencies tidak tersedia. Pastikan aplikasi diinisialisasi dengan benar.")
			return
		}

		// Inject logger dan config ke package cleanup
		cleanup.Logger = types.Deps.Logger
		cleanup.SetConfig(types.Deps.Config)

		if err := cleanup.CleanupOldBackups(); err != nil {
			types.Deps.Logger.Errorf("Cleanup gagal: %v", err)
			return
		}
	},
}

func init() {
	CmdCleanupMain.AddCommand(CmdCleanupRun)
}
