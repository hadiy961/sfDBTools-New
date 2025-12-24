package flags

import (
	"sfDBTools/internal/types/types_backup"

	"github.com/spf13/cobra"
)

func addBackupCommonFlags(cmd *cobra.Command, opts *types_backup.BackupDBOptions) {
	// Profile flags (shared)
	AddProfileFlags(cmd, &opts.Profile)

	// Encryption (shared)
	AddEncryptionFlags(cmd, &opts.Encryption)

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
	cmd.Flags().StringArray("exclude-db", opts.Filter.ExcludeDatabases, "Daftar database yang akan dikecualikan. Dapat dikombinasi dengan --exclude-db-file.")
	cmd.Flags().String("exclude-db-file", opts.Filter.ExcludeDBFile, "File berisi daftar database yang akan dikecualikan (satu per baris). Dapat dikombinasi dengan --exclude-db.")
}

func addBackupAllModeExcludeFlags(cmd *cobra.Command, opts *types_backup.BackupDBOptions) {
}

func addBackupCompanionFlags(cmd *cobra.Command, opts *types_backup.BackupDBOptions) {
	cmd.Flags().BoolVar(&opts.IncludeDmart, "include-dmart", opts.IncludeDmart, "Backup juga database <database>_dmart jika tersedia")
}

func addBackupCommonModeFlags(cmd *cobra.Command, defaultOpts *types_backup.BackupDBOptions) {
	cmd.Flags().StringVarP(&defaultOpts.OutputDir, "output-dir", "o", defaultOpts.OutputDir, "Direktori output untuk menyimpan file backup")
	cmd.Flags().StringVarP(&defaultOpts.File.Filename, "filename", "f", "", "Nama file backup")
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

	// Custom filename untuk output file (mode all menghasilkan satu file)
	cmd.Flags().StringVar(&opts.File.Filename, "filename", opts.File.Filename, "Nama file backup (custom, tanpa ekstensi)")

	// Exclude filters (tidak ada include)
	addBackupAllModeExcludeFlags(cmd, opts)
}

func AddBackupFlgs(cmd *cobra.Command, opts *types_backup.BackupDBOptions, mode string) {
	if mode == "single" {
		AddProfileFlags(cmd, &opts.Profile)
		AddEncryptionFlags(cmd, &opts.Encryption)
		SingleBackupFlags(cmd, opts)
	} else if mode == "primary" {
		AddProfileFlags(cmd, &opts.Profile)
		AddEncryptionFlags(cmd, &opts.Encryption)
		PrimaryBackupFlags(cmd, opts)
	} else if mode == "secondary" {
		AddProfileFlags(cmd, &opts.Profile)
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
