// File : cmd/cmd_restore/cmd_restore_selection.go
// Deskripsi : Command untuk restore selection berbasis CSV
// Author : Hadiyatna Muflihun
// Tanggal : 19 Desember 2025

package restorecmd

import (
	"fmt"
	"sfdbtools/internal/app/restore"
	appdeps "sfdbtools/internal/cli/deps"
	"sfdbtools/internal/cli/flags"

	"github.com/spf13/cobra"
)

// CmdRestoreSelection adalah command untuk restore multiple database berdasarkan CSV
var CmdRestoreSelection = &cobra.Command{
	Use:   "selection",
	Short: "Restore banyak database dari CSV (file,db,enc,grants)",
	Run: func(cmd *cobra.Command, args []string) {
		if appdeps.Deps == nil {
			fmt.Println("✗ Dependencies tidak tersedia. Pastikan aplikasi diinisialisasi dengan benar.")
			return
		}
		if err := restore.ExecuteRestoreSelectionCommand(cmd, appdeps.Deps); err != nil {
			appdeps.Deps.Logger.Error("restore selection gagal: " + err.Error())
		}
	},
}

func init() {
	flags.AddRestoreSelectionAllFlags(CmdRestoreSelection)
}
