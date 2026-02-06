package dbuser

import (
	"github.com/spf13/cobra"
)

// AddExportFlags mendaftarkan flags untuk `db-user export`.
func AddExportFlags(cmd *cobra.Command) {
	if cmd == nil {
		return
	}

	// Profile (source)
	flags := cmd.Flags()
	flags.StringP("profile", "p", "", "Path ke file profil database source terenkripsi (ENV: SFDB_SOURCE_PROFILE)")
	flags.StringP("profile-key", "k", "", "Encryption key untuk decrypt file profil database source (ENV: SFDB_SOURCE_PROFILE_KEY)")

	// Filters
	flags.StringArray("user", nil, "Export hanya untuk account tertentu, format: user@host (repeatable)")
	flags.StringArray("db", nil, "Filter grants yang relevan untuk database tertentu (repeatable)")
	flags.String("db-file", "", "File berisi daftar database (satu per baris) untuk filter grants")
	flags.String("client-code", "", "Ambil daftar database berdasarkan client code (pattern primary NBC) lalu filter grants untuk DB tersebut")

	// Behavior
	flags.Bool("exclude-system-users", true, "Skip system users (mysql.sys, mysql.session, dll)")
	flags.Bool("include-create-user", true, "Best-effort include CREATE USER agar bisa restore di server kosong")

	// Output
	flags.StringP("out", "o", "", "Output file path untuk hasil export (.sql)")
	flags.String("out-perm", "0600", "File permission output (octal string), default 0600")
}

// AddApplyFlags mendaftarkan flags untuk `db-user apply`.
func AddApplyFlags(cmd *cobra.Command) {
	if cmd == nil {
		return
	}

	flags := cmd.Flags()
	// Profile (target)
	flags.StringP("profile", "p", "", "Path ke file profil database target terenkripsi (ENV: SFDB_TARGET_PROFILE)")
	flags.StringP("profile-key", "k", "", "Encryption key untuk decrypt file profil database target (ENV: SFDB_TARGET_PROFILE_KEY)")

	flags.StringP("file", "f", "", "Path file SQL user+grants yang akan di-apply")
	flags.Bool("force", true, "Jalankan mysql client dengan -f (best-effort, lanjut meski ada SQL error)")
}
