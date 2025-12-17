package backupcmd

import (
	"fmt"
	"sfDBTools/internal/backup"
	"sfDBTools/internal/types"
	defaultVal "sfDBTools/internal/defaultval"
	"sfDBTools/internal/flags"

	"github.com/AlecAivazis/survey/v2"
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

		// Jika mode tidak di-provide (masih default "single-file"), tanyakan interaktif
		if !cmd.Flags().Changed("mode") {
			modeOptions := []string{
				"single-file (gabungkan semua database dalam satu file)",
				"multi-file (pisahkan setiap database ke file terpisah)",
			}

			var selected string
			prompt := &survey.Select{
				Message: "Pilih mode backup:",
				Options: modeOptions,
				Default: modeOptions[0], // Default single-file
			}

			if err := survey.AskOne(prompt, &selected); err != nil {
				types.Deps.Logger.Error("Pemilihan mode dibatalkan: " + err.Error())
				return
			}

			// Parse pilihan untuk mendapatkan mode
			if selected == modeOptions[0] {
				mode = "single-file"
			} else {
				mode = "multi-file"
			}
		}

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
