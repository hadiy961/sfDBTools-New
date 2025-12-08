package cmdbackup

import (
	"fmt"
	"sfDBTools/internal/backup"
	"sfDBTools/internal/types"
	defaultVal "sfDBTools/pkg/defaultval"
	"sfDBTools/pkg/flags"

	"github.com/spf13/cobra"
)

// CmdBackupPrimary adalah perintah untuk melakukan backup database primary.
var CmdBackupPrimary = &cobra.Command{
	Use:   "primary",
	Short: "Backup database primary (tanpa secondary)",
	Long:  `Perintah ini akan melakukan backup database primary. Hanya database tanpa suffix '_secondary' yang ditampilkan.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Pastikan dependencies tersedia
		if types.Deps == nil {
			fmt.Println("✗ Dependencies tidak tersedia. Pastikan aplikasi diinisialisasi dengan benar.")
			return
		}

		if err := backup.ExecuteBackup(cmd, types.Deps, "primary"); err != nil {
			types.Deps.Logger.Error("db-backup primary gagal: " + err.Error())
		}
	},
}

func init() {
	defaultOpts := defaultVal.DefaultBackupOptions("primary")
	flags.AddBackupFlgs(CmdBackupPrimary, &defaultOpts, "primary")
}
