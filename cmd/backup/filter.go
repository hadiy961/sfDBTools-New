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

// CmdBackupFilter adalah perintah untuk melakukan backup database dengan filter
var CmdBackupFilter = &cobra.Command{
	Use:   "filter",
	Short: "Backup banyak database pilihan (Selective/Bulk Backup)",
	Long: `Melakukan backup untuk beberapa database sekaligus yang dipilih berdasarkan nama atau pola.

Command ini memiliki dua mode output utama:
  1. Single-File (Combined): Menggabungkan semua database terpilih menjadi SATU file dump (.sql).
  2. Multi-File (Separated): Membuat file dump TERPISAH untuk setiap database.

Jika tidak ada database yang ditentukan lewat flag, mode interaktif (multi-select) akan muncul.`,
	Example: `  # 1. Mode Interaktif (Pilih database dari list)
  sfdbtools db-backup filter

  # 2. Backup database spesifik ke file terpisah (Multi-file)
  sfdbtools db-backup filter --db "db_satu,db_dua" --mode multi-file

  # 3. Backup database spesifik digabung jadi satu file (Single-file)
  sfdbtools db-backup filter --db "db_utama,db_pendukung" --mode single-file

  # 4. Backup menggunakan pola nama (misal: semua yg berawalan 'shop_')
  sfdbtools db-backup filter --pattern "shop_*" --mode multi-file`,
	Run: func(cmd *cobra.Command, args []string) {
		// Pastikan dependencies tersedia
		if types.Deps == nil {
			fmt.Println("✗ Dependencies tidak tersedia. Pastikan aplikasi diinisialisasi dengan benar.")
			return
		}

		backupMode, err := getBackupMode(cmd)
		if err != nil {
			types.Deps.Logger.Error(err.Error())
			return
		}

		if err := backup.ExecuteBackup(cmd, types.Deps, backupMode); err != nil {
			types.Deps.Logger.Error("db-backup filter gagal: " + err.Error())
		}
	},
}

func init() {
	defaultOpts := defaultVal.DefaultBackupOptions("combined") // Default ke combined
	flags.AddBackupFilterFlags(CmdBackupFilter, &defaultOpts)

	// Tambahkan flag --mode khusus untuk filter command
	CmdBackupFilter.Flags().String("mode", "single-file", "Mode backup: single-file (semua database dalam satu file) atau multi-file (satu file per database)")
}

func getBackupMode(cmd *cobra.Command) (string, error) {
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
			return "", fmt.Errorf("pemilihan mode dibatalkan: %w", err)
		}

		// Parse pilihan untuk mendapatkan mode
		if selected == modeOptions[0] {
			mode = "single-file"
		} else {
			mode = "multi-file"
		}
	}

	// Map mode ke backup mode internal
	switch mode {
	case "single-file":
		return "combined", nil
	case "multi-file":
		return "separated", nil
	default:
		return "", fmt.Errorf("mode tidak valid: %s. Gunakan 'single-file' atau 'multi-file'", mode)
	}
}
