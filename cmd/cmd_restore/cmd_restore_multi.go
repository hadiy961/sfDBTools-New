// File : cmd/cmd_restore/cmd_restore_multi.go
// Deskripsi : Command untuk restore multiple databases (placeholder)
// Author : Hadiyatna Muflihun
// Tanggal : 2025-11-05
// Last Modified : 2025-11-05

package cmdrestore

import (
	"fmt"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/flags"

	"github.com/spf13/cobra"
)

// CmdRestoreMulti adalah command untuk restore multiple databases (placeholder)
var CmdRestoreMulti = &cobra.Command{
	Use:   "multi",
	Short: "Restore multiple databases dari multiple backup files (coming soon)",
	Long: `Perintah ini akan melakukan restore multiple databases dari multiple backup files.

FITUR INI MASIH DALAM PENGEMBANGAN

Fitur yang akan tersedia:
- Support multiple source files
- Support wildcard pattern untuk file selection
- Support parallel restore untuk performa
- Support selective database restore dari combined backup

Contoh penggunaan di masa depan:
  sfdbtools restore multi --source "backup/*.sql.gz.enc"
  sfdbtools restore multi --source-list files.txt
  sfdbtools restore multi --pattern "dbname_*" --parallel 4`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("❌ Fitur restore multi belum diimplementasikan")
		fmt.Println("📋 Fitur ini akan ditambahkan di versi berikutnya")
		fmt.Println()
		fmt.Println("Untuk saat ini, gunakan:")
		fmt.Println("  - 'restore single' untuk restore satu database dari backup terpisah")
		fmt.Println("  - 'restore all' untuk restore semua database dari combined backup")
	},
}

func init() {
	// Placeholder options untuk restore multi
	defaultOpts := types.RestoreOptions{
		Mode:           "multi",
		VerifyChecksum: true,
		Force:          false,
		DryRun:         false,
		ShowOptions:    false,
	}
	flags.AddRestoreMultiFlags(CmdRestoreMulti, &defaultOpts)
}
