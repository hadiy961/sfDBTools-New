package copycmd

import (
	"github.com/spf13/cobra"
)

// CmdCopyMain adalah perintah induk (parent command) untuk semua perintah 'copy'.
var CmdCopyMain = &cobra.Command{
	Use:     "copy",
	Aliases: []string{"clone", "cp"},
	Short:   "Salin database atau tabel dalam satu server",
	Long: `Utilitas untuk menyalin database atau tabel secara instan di dalam server yang sama.
Mendukung mode Direct Stream (Piping) untuk kecepatan maksimal dan mode Disk-based untuk keamanan.`,
	Example: `  # Salin database secara interaktif
  sfdbtools copy db my_database

  # Salin database dengan nama target spesifik
  sfdbtools copy db source_db target_db --profile myprof

  # Salin tabel spesifik
  sfdbtools copy table mydb.users mydb.users_backup --profile myprof`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	// Tambahkan sub-command
	CmdCopyMain.AddCommand(CmdCopyDB)
	CmdCopyMain.AddCommand(CmdCopyTable)
}
