package cmdbackup

import (
	"fmt"
	"sfDBTools/internal/backup"
	"sfDBTools/internal/types"
	defaultVal "sfDBTools/pkg/defaultval"
	"sfDBTools/pkg/flags"

	"github.com/spf13/cobra"
)

// CmdDBBackupCombined adalah perintah untuk melakukan backup database secara combined
var CmdDBBackupCombined = &cobra.Command{
	Use:   "combined",
	Short: "Backup database secara combined",
	Long:  `Perintah ini akan melakukan backup database dengan metode combined.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Pastikan dependencies tersedia
		if types.Deps == nil {
			fmt.Println("✗ Dependencies tidak tersedia. Pastikan aplikasi diinisialisasi dengan benar.")
			return
		}

		if err := backup.ExecuteBackup(cmd, types.Deps, "combined"); err != nil {
			types.Deps.Logger.Error("db-backup combined gagal: " + err.Error())
		}
	},
}

func init() {
	defaultOpts := defaultVal.DefaultBackupOptions("combined") // Tambahkan flags untuk backup combined
	flags.AddBackupFlags(CmdDBBackupCombined, &defaultOpts)
}
