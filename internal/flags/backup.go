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

	// Non Interactive
	cmd.Flags().BoolVarP(&opts.NonInteractive, "non-interactive", "n", opts.NonInteractive, "Tidak melakukan interaksi (mode skrip)")

	// Dry Run
	cmd.Flags().BoolVarP(&opts.DryRun, "dry-run", "d", opts.DryRun, "Jalankan backup dalam mode dry-run (tidak benar-benar membuat file backup)")

	// Backup Directory (output)
	cmd.Flags().StringVarP(&opts.OutputDir, "backup-dir", "o", opts.OutputDir, "Direktori output untuk menyimpan file backup (default: dari config)")

	// Filename (tanpa ekstensi, optional - default auto dari config/pattern)
	cmd.Flags().StringVarP(&opts.File.Filename, "filename", "f", opts.File.Filename, "Menentukan nama file untuk backup (Tanpa ekstensi)")

	// Ticket
	cmd.Flags().StringVarP(&opts.Ticket, "ticket", "t", opts.Ticket, "Ticket number untuk request backup")

	// Exclude flags (default mengikuti config)
	cmd.Flags().BoolVarP(&opts.Filter.ExcludeData, "exclude-data", "x", opts.Filter.ExcludeData, "Mengecualikan data dari pencadangan (schema only)")
	cmd.Flags().BoolVarP(&opts.Filter.ExcludeSystem, "exclude-system", "S", opts.Filter.ExcludeSystem, "Mengecualikan sistem database")
	cmd.Flags().BoolVarP(&opts.Filter.ExcludeEmpty, "exclude-empty", "E", opts.Filter.ExcludeEmpty, "Mengecualikan database kosong")

	// Compression flags
	cmd.Flags().BoolP("skip-compress", "C", !opts.Compression.Enabled, "Melewati proses kompresi pada file backup (default: dari config)")
	cmd.Flags().StringVarP(&opts.Compression.Type, "compress", "c", opts.Compression.Type, "Menentukan jenis kompresi (gzip, zstd, xz, zlib, pgzip, none)")
	cmd.Flags().IntVarP(&opts.Compression.Level, "compress-level", "l", opts.Compression.Level, "Menentukan level kompresi (1-9) (default: dari config)")

	// Encryption skip flag
	cmd.Flags().Bool("skip-encrypt", !opts.Encryption.Enabled, "Melewati proses enkripsi pada file backup (default: dari config)")
}

func addBackupIncludeFilterFlags(cmd *cobra.Command, opts *types_backup.BackupDBOptions) {
	cmd.Flags().StringArray("db", opts.Filter.IncludeDatabases, "Daftar database yang akan di-backup. Dapat dikombinasi dengan --db-file.")
	cmd.Flags().String("db-file", opts.Filter.IncludeFile, "File berisi daftar database yang akan di-backup (satu per baris). Dapat dikombinasi dengan --db.")
}

func addBackupAllModeExcludeFlags(cmd *cobra.Command, opts *types_backup.BackupDBOptions) {
}

func addBackupCompanionFlags(cmd *cobra.Command, opts *types_backup.BackupDBOptions) {
	cmd.Flags().BoolVar(&opts.IncludeDmart, "include-dmart", opts.IncludeDmart, "Backup juga database <database>_dmart jika tersedia")
}

// AddBackupFlags
func AddBackupFlags(cmd *cobra.Command, opts *types_backup.BackupDBOptions) {
	addBackupCommonFlags(cmd, opts)

	// Filters: hanya include (tidak expose exclude-db/exclude-db-file)
	addBackupIncludeFilterFlags(cmd, opts)

	// Ticket wajib divalidasi saat runtime (khususnya untuk --non-interactive)
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
	addBackupAllModeExcludeFlags(cmd, opts)
}

func AddBackupFlgs(cmd *cobra.Command, opts *types_backup.BackupDBOptions, mode string) {
	// Semua mode backup memakai flags common yang sama.
	addBackupCommonFlags(cmd, opts)

	switch mode {
	case "single":
		SingleBackupFlags(cmd, opts)
	case "primary":
		PrimaryBackupFlags(cmd, opts)
	case "secondary":
		SecondaryBackupFlags(cmd, opts)
	default:
		// combined/separated/filter/all punya fungsi flag masing-masing.
	}
}

func SingleBackupFlags(cmd *cobra.Command, defaultOpts *types_backup.BackupDBOptions) {
	// Menambahkan flag spesifik untuk backup single database
	cmd.Flags().StringVar(&defaultOpts.DBName, "database", defaultOpts.DBName, "Nama database yang akan di-backup")
}

func SecondaryBackupFlags(cmd *cobra.Command, defaultOpts *types_backup.BackupDBOptions) {
	// Menambahkan flag spesifik untuk backup secondary database (tanpa --database flag)
	addBackupCompanionFlags(cmd, defaultOpts)
	cmd.Flags().StringVar(&defaultOpts.ClientCode, "client-code", defaultOpts.ClientCode, "Filter database secondary berdasarkan client code (contoh: adaro)")
	cmd.Flags().StringVar(&defaultOpts.Instance, "instance", defaultOpts.Instance, "Filter database secondary berdasarkan instance (contoh: 1, 2, 3)")
}

func PrimaryBackupFlags(cmd *cobra.Command, defaultOpts *types_backup.BackupDBOptions) {
	// Menambahkan flag spesifik untuk backup primary database (tanpa --database flag)
	addBackupCompanionFlags(cmd, defaultOpts)
	cmd.Flags().StringVar(&defaultOpts.ClientCode, "client-code", defaultOpts.ClientCode, "Filter database primary berdasarkan client code (contoh: adaro)")
}
