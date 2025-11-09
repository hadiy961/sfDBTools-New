// File : cmd/cmd_restore/cmd_restore_main.go
// Deskripsi : Main command untuk restore operations
// Author : Hadiyatna Muflihun
// Tanggal : 2025-11-05
// Last Modified : 2025-11-05

package cmdrestore

import (
	"github.com/spf13/cobra"
)

// CmdRestore adalah root command untuk restore operations
var CmdRestore = &cobra.Command{
	Use:   "restore",
	Short: "Restore database dari backup files",
	Long: `Restore database dari backup files dengan dekripsi, dekompresi, dan verifikasi checksum otomatis.

Available Commands:
  single  - Restore database dari satu file backup terpisah
  all     - Restore semua database dari file combined backup
  multi   - Restore multiple databases dari multiple files (coming soon)

Environment Variables:
  SFDB_TARGET_PROFILE          - Default target profile path
  SFDB_TARGET_PROFILE_KEY      - Default target profile encryption key
  SFDB_BACKUP_ENCRYPTION_KEY   - Default backup encryption key

Examples:
  # Restore single database
  sfdbtools restore single --source backup/mydb_20251105.sql.gz.enc --target-db mydb_restored

  # Restore all databases from combined backup
  sfdbtools restore all --source backup/combined_20251105.sql.gz.enc

  # Dry run untuk simulasi
  sfdbtools restore single --source backup/mydb.sql.gz.enc --dry-run --show-options`,
}

func init() {
	// Register subcommands
	CmdRestore.AddCommand(CmdRestoreSingle)
	CmdRestore.AddCommand(CmdRestoreAll)
	CmdRestore.AddCommand(CmdRestoreMulti)
}
