// File : cmd/cmd_restore/cmd_restore_custom.go
// Deskripsi : Command untuk restore custom dari SFCola account detail
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-24

package restorecmd

import (
	"fmt"
	"sfDBTools/internal/app/restore"
	appdeps "sfDBTools/internal/cli/deps"
	"sfDBTools/internal/cli/flags"

	"github.com/spf13/cobra"
)

// CmdRestoreCustom adalah command untuk restore custom (SFCola account detail)
var CmdRestoreCustom = &cobra.Command{
	Use:   "custom",
	Short: "Restore custom (paste account detail → provision DB+users → restore DB & DMART)",
	Run: func(cmd *cobra.Command, args []string) {
		if appdeps.Deps == nil {
			fmt.Println("✗ Dependencies tidak tersedia. Pastikan aplikasi diinisialisasi dengan benar.")
			return
		}

		if err := restore.ExecuteRestoreCustomCommand(cmd, appdeps.Deps); err != nil {
			appdeps.Deps.Logger.Error("restore custom gagal: " + err.Error())
		}
	},
}

func init() {
	flags.AddRestoreCustomFlags(CmdRestoreCustom)
}
