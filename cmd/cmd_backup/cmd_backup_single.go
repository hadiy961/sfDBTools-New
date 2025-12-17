package cmdbackup

import (
	"fmt"
	"sfDBTools/internal/backup"
	"sfDBTools/internal/types"
	defaultVal "sfDBTools/internal/defaultval"
	"sfDBTools/internal/flags"

	"github.com/spf13/cobra"
)

// CmdBackupSingle adalah perintah untuk melakukan backup satu database.
var CmdBackupSingle = &cobra.Command{
	Use:   "single",
	Short: "Backup satu database",
	Long:  `Perintah ini akan melakukan backup satu database.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Pastikan dependencies tersedia
		if types.Deps == nil {
			fmt.Println("✗ Dependencies tidak tersedia. Pastikan aplikasi diinisialisasi dengan benar.")
			return
		}

		if err := backup.ExecuteBackup(cmd, types.Deps, "single"); err != nil {
			types.Deps.Logger.Error("db-backup single gagal: " + err.Error())
		}
	},
}

func init() {
	defaultOpts := defaultVal.DefaultBackupOptions("single")
	flags.AddBackupFlgs(CmdBackupSingle, &defaultOpts, "single")
}
