package cmdbackup

import (
	"fmt"
	"sfDBTools/internal/backup"
	"sfDBTools/internal/types"
	defaultVal "sfDBTools/pkg/defaultval"
	"sfDBTools/pkg/flags"

	"github.com/spf13/cobra"
)

// CmdDBBackupFilter adalah perintah untuk melakukan backup database dengan filter
var CmdDBBackupFilter = &cobra.Command{
	Use:   "filter",
	Short: "Backup database dengan filter dan multi-select (single-file atau multi-file)",
	Long: `Perintah ini akan melakukan backup database dengan metode filter.
Mode single-file akan menggabungkan semua database dalam satu file (sebelumnya: combined).
Mode multi-file akan memisahkan setiap database ke file terpisah (sebelumnya: separated).

Jika tidak ada --db atau --db-file yang di-provide, akan muncul multi-select untuk memilih database.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Pastikan dependencies tersedia
		if types.Deps == nil {
			fmt.Println("✗ Dependencies tidak tersedia. Pastikan aplikasi diinisialisasi dengan benar.")
			return
		}

		// Dapatkan mode dari flag
		mode, _ := cmd.Flags().GetString("mode")

		// Map mode ke backup mode internal
		var backupMode string
		switch mode {
		case "single-file":
			backupMode = "combined"
		case "multi-file":
			backupMode = "separated"
		default:
			types.Deps.Logger.Error(fmt.Sprintf("Mode tidak valid: %s. Gunakan 'single-file' atau 'multi-file'", mode))
			return
		}

		if err := backup.ExecuteBackup(cmd, types.Deps, backupMode); err != nil {
			types.Deps.Logger.Error("db-backup filter gagal: " + err.Error())
		}
	},
}

func init() {
	defaultOpts := defaultVal.DefaultBackupOptions("combined") // Default ke combined
	flags.AddBackupFilterFlags(CmdDBBackupFilter, &defaultOpts)

	// Tambahkan flag --mode khusus untuk filter command
	CmdDBBackupFilter.Flags().String("mode", "single-file", "Mode backup: single-file (semua database dalam satu file) atau multi-file (satu file per database)")
}
