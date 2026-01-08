// File : cmd/backup/filter.go
// Deskripsi : Command untuk backup database dengan filter
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-30
// Last Modified : 2026-01-05
package backupcmd

import (
	"fmt"
	defaultVal "sfdbtools/internal/cli/defaults"
	"sfdbtools/internal/cli/flags"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/validation"
	"sfdbtools/internal/ui/prompt"

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
	sfdbtools db-backup filter --db "db_satu" --db "db_dua" --mode multi-file

  # 3. Backup database spesifik digabung jadi satu file (Single-file)
	sfdbtools db-backup filter --db "db_utama" --db "db_pendukung" --mode single-file`,
	Run: func(cmd *cobra.Command, args []string) {
		runBackupCommand(cmd, func() (string, error) {
			return getBackupMode(cmd)
		})
	},
}

func init() {
	defaultOpts := defaultVal.DefaultBackupOptions(consts.ModeCombined)
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

		selected, _, err := prompt.SelectOne("Pilih mode backup:", modeOptions, 0)
		if err != nil {
			return "", fmt.Errorf("mode non-interaktif: set --mode secara eksplisit (single-file/multi-file): %w", validation.HandleInputError(err))
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
		return consts.ModeCombined, nil
	case "multi-file":
		return consts.ModeSeparated, nil
	default:
		return "", fmt.Errorf("mode tidak valid: %s. Gunakan 'single-file' atau 'multi-file'", mode)
	}
}
