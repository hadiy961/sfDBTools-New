package cmdbackup

import (
	"github.com/spf13/cobra"
)

// CmdDBBackupCombined adalah perintah untuk melakukan backup database secara combined
var CmdDBBackupSeparated = &cobra.Command{
	Use:   "separated",
	Short: "Backup database secara combined",
	Long:  `Perintah ini akan melakukan backup database dengan metode combined.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Implementasi logika backup combined
	},
}
