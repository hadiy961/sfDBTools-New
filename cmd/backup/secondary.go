package backupcmd

import (
	"fmt"
	"sfDBTools/internal/backup"
	"sfDBTools/internal/types"
	defaultVal "sfDBTools/internal/defaultval"
	"sfDBTools/internal/flags"

	"github.com/spf13/cobra"
)

// CmdBackupSecondary adalah perintah untuk melakukan backup database secondary.
var CmdBackupSecondary = &cobra.Command{
	Use:   "secondary",
	Short: "Backup database secondary (dengan _secondary suffix)",
	Long:  `Perintah ini akan melakukan backup database secondary. Hanya database dengan suffix '_secondary' yang ditampilkan.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Pastikan dependencies tersedia
		if types.Deps == nil {
			fmt.Println("✗ Dependencies tidak tersedia. Pastikan aplikasi diinisialisasi dengan benar.")
			return
		}

		if err := backup.ExecuteBackup(cmd, types.Deps, "secondary"); err != nil {
			// Error has been logged by ExecuteBackup
			return
		}
	},
}

func init() {
	defaultOpts := defaultVal.DefaultBackupOptions("secondary")
	flags.AddBackupFlgs(CmdBackupSecondary, &defaultOpts, "secondary")
}
