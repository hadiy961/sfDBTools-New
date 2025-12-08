package cmdbackup

import (
	"fmt"
	"sfDBTools/internal/backup"
	"sfDBTools/internal/types"
	defaultVal "sfDBTools/pkg/defaultval"
	"sfDBTools/pkg/flags"

	"github.com/spf13/cobra"
)

// CmdDBBackupSeparated adalah perintah untuk melakukan backup database secara separated
var CmdDBBackupSeparated = &cobra.Command{
	Use:   "separated",
	Short: "Backup database secara separated",
	Long:  `Perintah ini akan melakukan backup database dengan metode separated (satu file per database).`,
	Run: func(cmd *cobra.Command, args []string) {
		// Pastikan dependencies tersedia
		if types.Deps == nil {
			fmt.Println("✗ Dependencies tidak tersedia. Pastikan aplikasi diinisialisasi dengan benar.")
			return
		}

		if err := backup.ExecuteBackup(cmd, types.Deps, "separated"); err != nil {
			types.Deps.Logger.Error("db-backup separated gagal: " + err.Error())
		}
	},
}

func init() {
	defaultOpts := defaultVal.DefaultBackupOptions("separated")
	flags.AddBackupFlags(CmdDBBackupSeparated, &defaultOpts)
}
