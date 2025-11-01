package cmdcleanup

import (
	"fmt"
	"sfDBTools/internal/cleanup"
	"sfDBTools/internal/types"

	"github.com/spf13/cobra"
)

var pattern string

// CmdCleanupPattern menjalankan pembersihan berdasarkan pola glob tertentu
var CmdCleanupPattern = &cobra.Command{
	Use:   "pattern",
	Short: "Hapus file backup yang cocok dengan pola tertentu",
	Long: `Membersihkan file backup lama yang cocok dengan pola glob tertentu (contoh: "**/*.sql.gz").
Hanya file yang lebih tua dari kebijakan retensi yang akan dipertimbangkan.`,
	Run: func(cmd *cobra.Command, args []string) {
		if types.Deps == nil || types.Deps.Config == nil || types.Deps.Logger == nil {
			fmt.Println("✗ Dependencies tidak tersedia. Pastikan aplikasi diinisialisasi dengan benar.")
			return
		}

		if pattern == "" {
			fmt.Println("✗ Harap tentukan --pattern. Lihat 'sfdbtools cleanup pattern --help'.")
			return
		}

		// Inject logger dan config ke package cleanup
		cleanup.Logger = types.Deps.Logger
		cleanup.SetConfig(types.Deps.Config)

		if err := cleanup.CleanupByPattern(pattern); err != nil {
			types.Deps.Logger.Errorf("Cleanup by pattern gagal: %v", err)
			return
		}
	},
}

func init() {
	CmdCleanupPattern.Flags().StringVarP(&pattern, "pattern", "p", "", "Pola glob untuk memilih file (contoh: '**/*.sql.gz')")
	CmdCleanupMain.AddCommand(CmdCleanupPattern)
}
