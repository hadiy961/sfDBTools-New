// File : cmd/dbcopy/main.go
// Deskripsi : Root command untuk db-copy (copy database via backup+restore)
// Author : Hadiyatna Muflihun
// Tanggal : 26 Januari 2026
// Last Modified : 26 Januari 2026
package dbcopycmd

import (
	"github.com/spf13/cobra"
)

// CmdDBCopyMain adalah perintah induk (parent command) untuk semua perintah 'db-copy'.
var CmdDBCopyMain = &cobra.Command{
	Use:     "db-copy",
	Aliases: []string{"copy"},
	Short:   "Copy database antar server/profile (automation-first)",
	Long: `Copy database dengan cara streaming backup → restore.

Fokus utama:
	  - Automation-first (non-interaktif dengan --skip-confirm / --quiet)
  - Aman (opsional pre-backup target sebelum overwrite)
  - Hemat RAM (streaming pipeline)`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

func init() {
	// Global flags (persistent) untuk semua subcommand
	CmdDBCopyMain.PersistentFlags().String("source-profile", "", "Profile database source (ENV: SFDB_SOURCE_PROFILE)")
	CmdDBCopyMain.PersistentFlags().String("source-profile-key", "", "Kunci enkripsi profile source (ENV: SFDB_SOURCE_PROFILE_KEY)")
	CmdDBCopyMain.PersistentFlags().String("target-profile", "", "Profile database target (ENV: SFDB_TARGET_PROFILE). Kosong = sama dengan source")
	CmdDBCopyMain.PersistentFlags().String("target-profile-key", "", "Kunci enkripsi profile target (ENV: SFDB_TARGET_PROFILE_KEY)")

	CmdDBCopyMain.PersistentFlags().StringP("ticket", "t", "", "Ticket number untuk audit (wajib)")

	CmdDBCopyMain.PersistentFlags().Bool("skip-confirm", false, "Non-interaktif dan skip semua prompt/konfirmasi (wajib untuk automation)")
	CmdDBCopyMain.PersistentFlags().Bool("continue-on-error", false, "Lanjutkan meskipun langkah non-kritis gagal")
	CmdDBCopyMain.PersistentFlags().Bool("dry-run", false, "Validasi + print rencana eksekusi tanpa menjalankan backup/restore")
	CmdDBCopyMain.PersistentFlags().BoolP("exclude-data", "x", false, "Mengecualikan data dari pencadangan (schema only)")

	CmdDBCopyMain.PersistentFlags().Bool("include-dmart", true, "Ikut copy companion database (_dmart) jika ada")
	CmdDBCopyMain.PersistentFlags().String("workdir", "", "Direktori kerja untuk file dump sementara (default: temp dir)")

	// Subcommands
	CmdDBCopyMain.AddCommand(CmdCopyP2S)
	CmdDBCopyMain.AddCommand(CmdCopyP2P)
	CmdDBCopyMain.AddCommand(CmdCopyS2S)
}
