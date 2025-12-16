package flags

import (
	"sfDBTools/internal/types"
	"sfDBTools/internal/types/types_backup"
	flagsbackup "sfDBTools/pkg/flags/flags_backup"

	"github.com/spf13/cobra"
)

// AddBackupFlags
func AddBackupFlags(cmd *cobra.Command, opts *types_backup.BackupDBOptions) {
	AddProfileFlags(cmd, &opts.Profile)

	// Filters
	cmd.Flags().Bool("exclude-system", opts.Filter.ExcludeSystem, "Kecualikan system databases (information_schema, mysql, dll)")
	cmd.Flags().StringArray("exclude-db", opts.Filter.ExcludeDatabases, "Daftar database yang akan dikecualikan. Dapat dikombinasi dengan --exclude-db-file.")
	cmd.Flags().String("exclude-db-file", opts.Filter.ExcludeDBFile, "File berisi daftar database yang akan dikecualikan (satu per baris). Dapat dikombinasi dengan --exclude-db.")
	cmd.Flags().StringArray("db", opts.Filter.IncludeDatabases, "Daftar database yang akan di-backup. Dapat dikombinasi dengan --db-file.")
	cmd.Flags().String("db-file", opts.Filter.IncludeFile, "File berisi daftar database yang akan di-backup (satu per baris). Dapat dikombinasi dengan --db.")

	// Compression
	AddCompressionFlags(cmd, &opts.Compression)

	// Encryption
	AddEncryptionFlags(cmd, &opts.Encryption)

	// Capture GTID
	cmd.Flags().Bool("capture-gtid", opts.CaptureGTID, "Tangkap informasi GTID saat melakukan backup")

	// Exclude User
	cmd.Flags().Bool("exclude-user", opts.ExcludeUser, "Exclude user grants dari export (default: false = export user)")

	// Dry Run
	cmd.Flags().Bool("dry-run", opts.DryRun, "Jalankan backup dalam mode dry-run (tidak benar-benar membuat file backup)")

	// Output Directory
	cmd.Flags().String("output-dir", opts.OutputDir, "Direktori output untuk menyimpan file backup")
	cmd.Flags().Bool("force", opts.Force, "Tampilkan opsi backup sebelum eksekusi")

	// Ticket (wajib)
	cmd.Flags().String("ticket", opts.Ticket, "Ticket number untuk request backup (wajib)")
	cmd.MarkFlagRequired("ticket")
}

// AddEncryptionFlags menambahkan flags untuk konfigurasi enkripsi.
func AddEncryptionFlags(cmd *cobra.Command, opts *types.EncryptionOptions) {
	cmd.Flags().StringP("encryption-key", "K", opts.Key, "Kunci enkripsi yang digunakan untuk mengenkripsi file backup")
}

// AddCompressionFlags menambahkan flags untuk konfigurasi kompresi.
func AddCompressionFlags(cmd *cobra.Command, opts *types.CompressionOptions) {
	cmd.Flags().StringP("compress-type", "C", opts.Type, "Tipe kompresi yang digunakan (gzip, zstd, xz, pgzip, zlib, none)")
	cmd.Flags().Int("compress-level", opts.Level, "Tingkat kompresi yang digunakan (1-9)")
}

// AddBackupFilterFlags menambahkan flags untuk backup filter (tanpa exclude flags)
func AddBackupFilterFlags(cmd *cobra.Command, opts *types_backup.BackupDBOptions) {
	AddProfileFlags(cmd, &opts.Profile)

	// Hanya include filters (tidak ada exclude)
	cmd.Flags().StringArray("db", opts.Filter.IncludeDatabases, "Daftar database yang akan di-backup. Dapat dikombinasi dengan --db-file.")
	cmd.Flags().String("db-file", opts.Filter.IncludeFile, "File berisi daftar database yang akan di-backup (satu per baris). Dapat dikombinasi dengan --db.")

	// Compression
	AddCompressionFlags(cmd, &opts.Compression)

	// Encryption
	AddEncryptionFlags(cmd, &opts.Encryption)

	// Capture GTID
	cmd.Flags().Bool("capture-gtid", opts.CaptureGTID, "Tangkap informasi GTID saat melakukan backup")

	// Exclude User
	cmd.Flags().Bool("exclude-user", opts.ExcludeUser, "Exclude user grants dari export (default: false = export user)")

	// Dry Run
	cmd.Flags().Bool("dry-run", opts.DryRun, "Jalankan backup dalam mode dry-run (tidak benar-benar membuat file backup)")

	// Output Directory
	cmd.Flags().String("output-dir", opts.OutputDir, "Direktori output untuk menyimpan file backup")
	cmd.Flags().Bool("force", opts.Force, "Tampilkan opsi backup sebelum eksekusi")

	// Ticket
	cmd.Flags().String("ticket", opts.Ticket, "Ticket number untuk request backup")
}

// AddBackupAllFlags menambahkan flags untuk backup all (dengan exclude flags, tanpa include)
func AddBackupAllFlags(cmd *cobra.Command, opts *types_backup.BackupDBOptions) {
	AddProfileFlags(cmd, &opts.Profile)

	// Exclude filters (tidak ada include)
	cmd.Flags().Bool("exclude-system", opts.Filter.ExcludeSystem, "Kecualikan system databases (information_schema, mysql, dll)")
	cmd.Flags().StringArray("exclude-db", opts.Filter.ExcludeDatabases, "Daftar database yang akan dikecualikan. Dapat dikombinasi dengan --exclude-db-file.")
	cmd.Flags().String("exclude-db-file", opts.Filter.ExcludeDBFile, "File berisi daftar database yang akan dikecualikan (satu per baris). Dapat dikombinasi dengan --exclude-db.")

	// Exclude options
	cmd.Flags().Bool("exclude-data", opts.Filter.ExcludeData, "Exclude data, hanya backup struktur database")
	cmd.Flags().Bool("exclude-empty", opts.Filter.ExcludeEmpty, "Exclude database yang kosong (tidak ada tabel)")

	// Compression
	AddCompressionFlags(cmd, &opts.Compression)

	// Encryption
	AddEncryptionFlags(cmd, &opts.Encryption)

	// Capture GTID
	cmd.Flags().Bool("capture-gtid", opts.CaptureGTID, "Tangkap informasi GTID saat melakukan backup")

	// Exclude User
	cmd.Flags().Bool("exclude-user", opts.ExcludeUser, "Exclude user grants dari export (default: false = export user)")

	// Dry Run
	cmd.Flags().Bool("dry-run", opts.DryRun, "Jalankan backup dalam mode dry-run (tidak benar-benar membuat file backup)")

	// Output Directory
	cmd.Flags().String("output-dir", opts.OutputDir, "Direktori output untuk menyimpan file backup")
	cmd.Flags().Bool("force", opts.Force, "Tampilkan opsi backup sebelum eksekusi")

	// Ticket
	cmd.Flags().String("ticket", opts.Ticket, "Ticket number untuk request backup")
}

func AddBackupFlgs(cmd *cobra.Command, opts *types_backup.BackupDBOptions, mode string) {
	if mode == "single" {
		AddProfileFlags(cmd, &opts.Profile)
		AddCompressionFlags(cmd, &opts.Compression)
		AddEncryptionFlags(cmd, &opts.Encryption)
		flagsbackup.SingleBackupFlags(cmd, opts)
	} else if mode == "primary" {
		AddProfileFlags(cmd, &opts.Profile)
		AddCompressionFlags(cmd, &opts.Compression)
		AddEncryptionFlags(cmd, &opts.Encryption)
		flagsbackup.PrimaryBackupFlags(cmd, opts)
	} else if mode == "secondary" {
		AddProfileFlags(cmd, &opts.Profile)
		AddCompressionFlags(cmd, &opts.Compression)
		AddEncryptionFlags(cmd, &opts.Encryption)
		flagsbackup.SecondaryBackupFlags(cmd, opts)
	} else {
		AddBackupFlags(cmd, opts)
	}
}
