// File : internal/flags/restore.go
// Deskripsi : Helper functions untuk menambahkan flags restore commands
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-19
// Last Modified : 2025-12-19

package flags

import (
	"github.com/spf13/cobra"
)

// AddRestoreCommonFlags menambahkan flags yang umum digunakan di semua restore commands
// Flags: --profile, --profile-key, --encryption-key, --ticket
func AddRestoreCommonFlags(cmd *cobra.Command) {
	// Profile flags (target database)
	cmd.Flags().StringP("profile", "p", "", "Profile database target untuk restore (ENV: SFDB_TARGET_PROFILE)")
	cmd.Flags().StringP("profile-key", "k", "", "Kunci enkripsi profile database target (ENV: SFDB_TARGET_PROFILE_KEY)")

	// Encryption flag untuk decrypt file backup
	cmd.Flags().StringP("encryption-key", "K", "", "Kunci enkripsi untuk decrypt file backup (ENV: SFDB_BACKUP_ENCRYPTION_KEY)")

	// Ticket (wajib untuk audit)
	cmd.Flags().StringP("ticket", "t", "", "Ticket number untuk restore request (wajib)")
	cmd.MarkFlagRequired("ticket")
}

// AddRestoreFileFlags menambahkan flags untuk file input
// Flags: --file
func AddRestoreFileFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("file", "f", "", "Lokasi file backup yang akan di-restore")
	cmd.MarkFlagRequired("file")
}

// AddRestoreTargetFlags menambahkan flags untuk target database
// Flags: --target-db, --drop-target, --skip-backup, --backup-dir
func AddRestoreTargetFlags(cmd *cobra.Command, requireTarget bool) {
	cmd.Flags().StringP("target-db", "d", "", "Database target untuk restore")
	if requireTarget {
		cmd.MarkFlagRequired("target-db")
	}

	cmd.Flags().Bool("drop-target", true, "Drop target database sebelum restore")
	cmd.Flags().Bool("skip-backup", false, "Skip backup database target sebelum restore")
	cmd.Flags().StringP("backup-dir", "b", "", "Direktori output untuk backup pre-restore (default: dari config)")
}

// AddRestoreGrantsFlags menambahkan flags untuk user grants
// Flags: --grants-file, --skip-grants
func AddRestoreGrantsFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("grants-file", "g", "", "Lokasi file user grants untuk di-restore (optional)")
	cmd.Flags().Bool("skip-grants", false, "Skip restore user grants (tidak restore grants sama sekali)")
}

// AddRestoreDryRunFlag menambahkan flag untuk dry-run mode
// Flags: --dry-run
func AddRestoreDryRunFlag(cmd *cobra.Command) {
	cmd.Flags().Bool("dry-run", false, "Dry-run mode: validasi file tanpa restore")
}

// AddRestorePrimaryFlags menambahkan flags khusus untuk restore primary
// Flags: --dmart-file, --dmart-include, --dmart-detect, --skip-confirm
func AddRestorePrimaryFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("dmart-file", "c", "", "Lokasi file backup dmart (_dmart) - optional, auto-detect jika kosong")
	cmd.Flags().Bool("dmart-include", true, "Include restore companion database (_dmart)")
	cmd.Flags().Bool("dmart-detect", true, "Auto-detect file companion database (_dmart)")
	cmd.Flags().Bool("skip-confirm", false, "Skip konfirmasi jika database belum ada")

	// Backward compatibility (deprecated flags)
	cmd.Flags().String("companion-file", "", "DEPRECATED: gunakan --dmart-file")
	_ = cmd.Flags().MarkDeprecated("companion-file", "gunakan --dmart-file")
	_ = cmd.Flags().MarkHidden("companion-file")

	cmd.Flags().Bool("include-dmart", true, "DEPRECATED: gunakan --dmart-include")
	_ = cmd.Flags().MarkDeprecated("include-dmart", "gunakan --dmart-include")
	_ = cmd.Flags().MarkHidden("include-dmart")

	cmd.Flags().Bool("auto-detect-dmart", true, "DEPRECATED: gunakan --dmart-detect")
	_ = cmd.Flags().MarkDeprecated("auto-detect-dmart", "gunakan --dmart-detect")
	_ = cmd.Flags().MarkHidden("auto-detect-dmart")
}

// AddRestoreAllFlags menambahkan flags khusus untuk restore all databases
// Flags: --force, --continue-on-error, --exclude-db, --exclude-db-file, --skip-system-dbs
func AddRestoreAllFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("force", false, "Force restore tanpa konfirmasi interaktif")
	cmd.Flags().Bool("continue-on-error", false, "Lanjutkan restore meski ada error (default: stop on error)")
	cmd.Flags().Bool("skip-system-dbs", true, "Skip system databases (mysql, sys, information_schema, performance_schema)")

	// Filter flags
	cmd.Flags().StringSlice("exclude-db", []string{}, "Daftar database yang akan di-exclude dari restore")
	cmd.Flags().String("exclude-db-file", "", "File berisi daftar database yang akan di-exclude (satu per baris)")
}

// AddRestoreSelectionFlags menambahkan flags khusus untuk restore selection (CSV-based)
// Flags: --csv, --force, --continue-on-error
func AddRestoreSelectionFlags(cmd *cobra.Command) {
	cmd.Flags().String("csv", "", "Path CSV: filename,db_name,enc_key,grants_file (wajib)")
	cmd.MarkFlagRequired("csv")

	cmd.Flags().Bool("force", false, "Bypass konfirmasi")
	cmd.Flags().Bool("continue-on-error", false, "Lanjutkan meskipun terjadi error pada salah satu file")
}

// ============================================================================
// Composite Flag Functions - Menggabungkan beberapa flag groups
// ============================================================================

// AddRestoreSingleFlags menambahkan semua flags untuk restore single command
func AddRestoreSingleFlags(cmd *cobra.Command) {
	AddRestoreCommonFlags(cmd)
	AddRestoreFileFlags(cmd)
	AddRestoreTargetFlags(cmd, false) // target-db optional (bisa auto-detect)
	cmd.Flags().Bool("force", false, "Force restore tanpa konfirmasi interaktif")
	cmd.Flags().Bool("continue-on-error", false, "Lanjutkan restore meski ada error (default: stop on error)")
	AddRestoreGrantsFlags(cmd)
	AddRestoreDryRunFlag(cmd)
}

// AddRestorePrimaryAllFlags menambahkan semua flags untuk restore primary command
func AddRestorePrimaryAllFlags(cmd *cobra.Command) {
	AddRestoreCommonFlags(cmd)
	AddRestoreFileFlags(cmd)
	AddRestoreTargetFlags(cmd, false) // target-db optional (bisa auto-detect)
	cmd.Flags().Bool("force", false, "Force restore tanpa konfirmasi interaktif")
	cmd.Flags().Bool("continue-on-error", false, "Lanjutkan restore meski ada error (default: stop on error)")
	AddRestorePrimaryFlags(cmd)
	AddRestoreGrantsFlags(cmd)
	AddRestoreDryRunFlag(cmd)
}

// AddRestoreAllAllFlags menambahkan semua flags untuk restore all command
func AddRestoreAllAllFlags(cmd *cobra.Command) {
	AddRestoreCommonFlags(cmd)
	AddRestoreFileFlags(cmd)

	// Restore all tidak pakai target-db, tapi butuh backup-dir
	cmd.Flags().Bool("skip-backup", false, "Skip backup sebelum restore")
	cmd.Flags().StringP("backup-dir", "b", "", "Direktori output untuk backup pre-restore (default: dari config)")
	cmd.Flags().Bool("drop-target", false, "Drop semua database non-sistem sebelum restore")

	AddRestoreAllFlags(cmd)
	AddRestoreDryRunFlag(cmd)
}

// AddRestoreSelectionAllFlags menambahkan semua flags untuk restore selection command
func AddRestoreSelectionAllFlags(cmd *cobra.Command) {
	AddRestoreCommonFlags(cmd)
	AddRestoreSelectionFlags(cmd)

	// Selection juga butuh target flags tapi tanpa --target-db (dibaca dari CSV)
	cmd.Flags().Bool("drop-target", true, "Drop target database sebelum restore")
	cmd.Flags().Bool("skip-backup", false, "Skip backup database target sebelum restore")
	cmd.Flags().String("backup-dir", "", "Direktori output untuk backup pre-restore (default: dari config)")

	AddRestoreDryRunFlag(cmd)
}
