// File : pkg/flags/flags_restore.go
// Deskripsi : Flags definitions untuk restore commands
// Author : Hadiyatna Muflihun
// Tanggal : 2025-11-05
// Last Modified : 2025-11-05

package flags

import (
	"sfDBTools/internal/types"

	"github.com/spf13/cobra"
)

// AddRestoreSingleFlags menambahkan flags untuk restore single command
func AddRestoreSingleFlags(cmd *cobra.Command, opts *types.RestoreOptions) {
	// Source backup file
	cmd.Flags().StringVarP(&opts.SourceFile, "source", "s", "", "Lokasi file backup source (required)")
	cmd.MarkFlagRequired("source")

	// Target profile dan authentication
	cmd.Flags().StringVarP(&opts.TargetProfile, "profile", "p", "", "Profile database target untuk restore (ENV: SFDB_TARGET_PROFILE)")
	cmd.Flags().String("profile-key", "", "Kunci enkripsi profile database target (ENV: SFDB_TARGET_PROFILE_KEY)")

	// Encryption key untuk decrypt backup
	cmd.Flags().StringVarP(&opts.EncryptionKey, "encryption-key", "k", "", "Kunci enkripsi untuk decrypt file backup (ENV: SFDB_BACKUP_ENCRYPTION_KEY)")

	// Target database name
	cmd.Flags().StringVarP(&opts.TargetDB, "target-db", "d", "", "Nama database target untuk restore (jika kosong, gunakan nama dari backup file)")

	// Verification
	cmd.Flags().BoolVar(&opts.VerifyChecksum, "verify-checksum", true, "Verifikasi checksum sebelum restore")

	// Force dan dry-run
	cmd.Flags().BoolVar(&opts.Force, "force", false, "Skip confirmation prompt")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "Simulasi restore tanpa eksekusi (untuk testing)")

	// Show options
	cmd.Flags().BoolVar(&opts.ShowOptions, "show-options", false, "Tampilkan opsi restore sebelum eksekusi")

	// Safety backup
	cmd.Flags().BoolVar(&opts.BackupBeforeRestore, "backup-before-restore", false, "Backup database terlebih dahulu sebelum restore (safety backup)")
}

// AddRestoreAllFlags menambahkan flags untuk restore all command
func AddRestoreAllFlags(cmd *cobra.Command, opts *types.RestoreOptions) {
	// Source backup file
	cmd.Flags().StringVarP(&opts.SourceFile, "source", "s", "", "Lokasi file combined backup source (required)")
	cmd.MarkFlagRequired("source")

	// Target profile dan authentication
	cmd.Flags().StringVarP(&opts.TargetProfile, "profile", "p", "", "Profile database target untuk restore (ENV: SFDB_TARGET_PROFILE)")
	cmd.Flags().String("profile-key", "", "Kunci enkripsi profile database target (ENV: SFDB_TARGET_PROFILE_KEY)")

	// Encryption key untuk decrypt backup
	cmd.Flags().StringVarP(&opts.EncryptionKey, "encryption-key", "k", "", "Kunci enkripsi untuk decrypt file backup (ENV: SFDB_BACKUP_ENCRYPTION_KEY)")

	// Verification
	cmd.Flags().BoolVar(&opts.VerifyChecksum, "verify-checksum", true, "Verifikasi checksum sebelum restore")

	// Force dan dry-run
	cmd.Flags().BoolVar(&opts.Force, "force", false, "Skip confirmation prompt")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "Simulasi restore tanpa eksekusi (untuk testing)")

	// Show options
	cmd.Flags().BoolVar(&opts.ShowOptions, "show-options", false, "Tampilkan opsi restore sebelum eksekusi")

	// Safety backup
	cmd.Flags().BoolVar(&opts.BackupBeforeRestore, "backup-before-restore", false, "Backup database terlebih dahulu sebelum restore (safety backup)")
}

// AddRestoreMultiFlags menambahkan flags untuk restore multi command (placeholder)
func AddRestoreMultiFlags(cmd *cobra.Command, opts *types.RestoreOptions) {
	// Placeholder untuk fitur restore multi
	cmd.Flags().StringVarP(&opts.SourceFile, "source", "s", "", "Pattern atau list file backup sources (future feature)")
	cmd.Flags().StringVarP(&opts.TargetProfile, "profile", "p", "", "Profile database target untuk restore")
	cmd.Flags().String("profile-key", "", "Kunci enkripsi profile database target")
	cmd.Flags().StringVarP(&opts.EncryptionKey, "encryption-key", "k", "", "Kunci enkripsi untuk decrypt file backup")
	cmd.Flags().BoolVar(&opts.VerifyChecksum, "verify-checksum", true, "Verifikasi checksum sebelum restore")
	cmd.Flags().BoolVar(&opts.Force, "force", false, "Skip confirmation prompt")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "Simulasi restore tanpa eksekusi")
	cmd.Flags().BoolVar(&opts.ShowOptions, "show-options", false, "Tampilkan opsi restore sebelum eksekusi")
}
