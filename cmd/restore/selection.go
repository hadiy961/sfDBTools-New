// File : cmd/cmd_restore/cmd_restore_selection.go
// Deskripsi : Command untuk restore selection berbasis CSV
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-19

package restorecmd

import (
	"fmt"
	"sfDBTools/internal/flags"
	"sfDBTools/internal/restore"
	"sfDBTools/internal/types"

	"github.com/spf13/cobra"
)

// CmdRestoreSelection adalah command untuk restore multiple database berdasarkan CSV
var CmdRestoreSelection = &cobra.Command{
	Use:   "selection",
	Short: "Restore banyak database dari CSV (file,db,enc,grants)",
	Run: func(cmd *cobra.Command, args []string) {
		if types.Deps == nil {
			fmt.Println("✗ Dependencies tidak tersedia. Pastikan aplikasi diinisialisasi dengan benar.")
			return
		}
		if err := restore.ExecuteRestoreSelectionCommand(cmd, types.Deps); err != nil {
			types.Deps.Logger.Error("restore selection gagal: " + err.Error())
		}
	},
}

func init() {
	flags.AddRestoreSelectionAllFlags(CmdRestoreSelection)
}
