package flags

import (
	"sfDBTools/internal/types/types_backup"

	"github.com/spf13/cobra"
)

func addBackupCommonFlags(cmd *cobra.Command, opts *types_backup.BackupDBOptions) {
	// Profile flags (shared)
	AddProfileFlags(cmd, &opts.Profile)

	// Compression (shared)
	AddCompressionFlags(cmd, &opts.Compression)

	// Encryption (shared)
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

func addBackupIncludeFilterFlags(cmd *cobra.Command, opts *types_backup.BackupDBOptions) {
	cmd.Flags().StringArray("db", opts.Filter.IncludeDatabases, "Daftar database yang akan di-backup. Dapat dikombinasi dengan --db-file.")
	cmd.Flags().String("db-file", opts.Filter.IncludeFile, "File berisi daftar database yang akan di-backup (satu per baris). Dapat dikombinasi dengan --db.")
}

func addBackupExcludeFilterFlags(cmd *cobra.Command, opts *types_backup.BackupDBOptions) {
	cmd.Flags().Bool("exclude-system", opts.Filter.ExcludeSystem, "Kecualikan system databases (information_schema, mysql, dll)")
	cmd.Flags().StringArray("exclude-db", opts.Filter.ExcludeDatabases, "Daftar database yang akan dikecualikan. Dapat dikombinasi dengan --exclude-db-file.")
	cmd.Flags().String("exclude-db-file", opts.Filter.ExcludeDBFile, "File berisi daftar database yang akan dikecualikan (satu per baris). Dapat dikombinasi dengan --exclude-db.")
}

func addBackupAllModeExcludeFlags(cmd *cobra.Command, opts *types_backup.BackupDBOptions) {
	cmd.Flags().Bool("exclude-data", opts.Filter.ExcludeData, "Exclude data, hanya backup struktur database")
	cmd.Flags().Bool("exclude-empty", opts.Filter.ExcludeEmpty, "Exclude database yang kosong (tidak ada tabel)")
}

func addBackupCompanionFlags(cmd *cobra.Command, opts *types_backup.BackupDBOptions) {
	cmd.Flags().BoolVar(&opts.IncludeDmart, "include-dmart", opts.IncludeDmart, "Backup juga database <database>_dmart jika tersedia")
	cmd.Flags().BoolVar(&opts.IncludeTemp, "include-temp", opts.IncludeTemp, "Backup juga database <database>_temp jika tersedia")
	cmd.Flags().BoolVar(&opts.IncludeArchive, "include-archive", opts.IncludeArchive, "Backup juga database <database>_archive jika tersedia")
}

func addBackupCommonModeFlags(cmd *cobra.Command, defaultOpts *types_backup.BackupDBOptions) {
	cmd.Flags().StringVarP(&defaultOpts.OutputDir, "output-dir", "o", defaultOpts.OutputDir, "Direktori output untuk menyimpan file backup")
	cmd.Flags().StringVarP(&defaultOpts.File.Filename, "filename", "f", "", "Nama file backup")
	cmd.Flags().BoolVarP(&defaultOpts.ExcludeUser, "exclude-user", "e", defaultOpts.ExcludeUser, "Exclude user grants dari export")
	cmd.Flags().BoolVar(&defaultOpts.Filter.ExcludeData, "exclude-data", defaultOpts.Filter.ExcludeData, "Backup hanya struktur database tanpa data")
	cmd.Flags().StringVar(&defaultOpts.Ticket, "ticket", defaultOpts.Ticket, "Ticket number untuk request backup")
	cmd.Flags().BoolVar(&defaultOpts.Force, "force", defaultOpts.Force, "Tampilkan opsi backup sebelum eksekusi")
}

// AddBackupFlags
func AddBackupFlags(cmd *cobra.Command, opts *types_backup.BackupDBOptions) {
	addBackupCommonFlags(cmd, opts)

	// Filters (Custom implementation for backup struct compatibility)
	addBackupExcludeFilterFlags(cmd, opts)
	addBackupIncludeFilterFlags(cmd, opts)

	// Ticket (wajib)
	cmd.Flags().SetAnnotation("ticket", cobra.BashCompOneRequiredFlag, []string{"true"})
	cmd.MarkFlagRequired("ticket")
}

// AddBackupFilterFlags menambahkan flags untuk backup filter (tanpa exclude flags)
func AddBackupFilterFlags(cmd *cobra.Command, opts *types_backup.BackupDBOptions) {
	addBackupCommonFlags(cmd, opts)

	// Hanya include filters (tidak ada exclude)
	addBackupIncludeFilterFlags(cmd, opts)
}

// AddBackupAllFlags menambahkan flags untuk backup all (dengan exclude flags, tanpa include)
func AddBackupAllFlags(cmd *cobra.Command, opts *types_backup.BackupDBOptions) {
	addBackupCommonFlags(cmd, opts)

	// Exclude filters (tidak ada include)
	cmd.Flags().Bool("exclude-system", opts.Filter.ExcludeSystem, "Kecualikan system databases (information_schema, mysql, dll)")
	addBackupAllModeExcludeFlags(cmd, opts)
}

func AddBackupFlgs(cmd *cobra.Command, opts *types_backup.BackupDBOptions, mode string) {
	if mode == "single" {
		AddProfileFlags(cmd, &opts.Profile)
		AddCompressionFlags(cmd, &opts.Compression)
		AddEncryptionFlags(cmd, &opts.Encryption)
		SingleBackupFlags(cmd, opts)
	} else if mode == "primary" {
		AddProfileFlags(cmd, &opts.Profile)
		AddCompressionFlags(cmd, &opts.Compression)
		AddEncryptionFlags(cmd, &opts.Encryption)
		PrimaryBackupFlags(cmd, opts)
	} else if mode == "secondary" {
		AddProfileFlags(cmd, &opts.Profile)
		AddCompressionFlags(cmd, &opts.Compression)
		AddEncryptionFlags(cmd, &opts.Encryption)
		SecondaryBackupFlags(cmd, opts)
	} else {
		AddBackupFlags(cmd, opts)
	}
}

func SingleBackupFlags(cmd *cobra.Command, defaultOpts *types_backup.BackupDBOptions) {
	// Menambahkan flag spesifik untuk backup single database
	cmd.Flags().StringVarP(&defaultOpts.DBName, "database", "d", defaultOpts.DBName, "Nama database yang akan di-backup")
	addBackupCommonModeFlags(cmd, defaultOpts)
}

func SecondaryBackupFlags(cmd *cobra.Command, defaultOpts *types_backup.BackupDBOptions) {
	// Menambahkan flag spesifik untuk backup secondary database (tanpa --database flag)
	addBackupCommonModeFlags(cmd, defaultOpts)
	addBackupCompanionFlags(cmd, defaultOpts)
	cmd.Flags().StringVar(&defaultOpts.ClientCode, "client-code", defaultOpts.ClientCode, "Filter database secondary berdasarkan client code (contoh: adaro)")
	cmd.Flags().StringVar(&defaultOpts.Instance, "instance", defaultOpts.Instance, "Filter database secondary berdasarkan instance (contoh: 1, 2, 3)")
}

func PrimaryBackupFlags(cmd *cobra.Command, defaultOpts *types_backup.BackupDBOptions) {
	// Menambahkan flag spesifik untuk backup primary database (tanpa --database flag)
	addBackupCommonModeFlags(cmd, defaultOpts)
	addBackupCompanionFlags(cmd, defaultOpts)
	cmd.Flags().StringVar(&defaultOpts.ClientCode, "client-code", defaultOpts.ClientCode, "Filter database primary berdasarkan client code (contoh: adaro)")
}
