package flags

import (
	"sfDBTools/internal/types"

	"github.com/spf13/cobra"
)

// AddBackupFlags
func AddBackupFlags(cmd *cobra.Command, opts *types.BackupDBOptions) {
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

	// Dry Run
	cmd.Flags().Bool("dry-run", opts.DryRun, "Jalankan backup dalam mode dry-run (tidak benar-benar membuat file backup)")

	// Output Directory
	cmd.Flags().String("output-dir", opts.OutputDir, "Direktori output untuk menyimpan file backup")
	cmd.Flags().Bool("force", opts.Force, "Tampilkan opsi backup sebelum eksekusi")
}

// AddEncryptionFlags menambahkan flags untuk konfigurasi enkripsi.
func AddEncryptionFlags(cmd *cobra.Command, opts *types.EncryptionOptions) {
	cmd.Flags().Bool("encrypt", opts.Enabled, "Aktifkan enkripsi untuk file backup")
	cmd.Flags().String("encryption-key", opts.Key, "Kunci enkripsi yang digunakan untuk mengenkripsi file backup")
}

// AddCompressionFlags menambahkan flags untuk konfigurasi kompresi.
func AddCompressionFlags(cmd *cobra.Command, opts *types.CompressionOptions) {
	cmd.Flags().Bool("compress", opts.Enabled, "Aktifkan kompresi untuk file backup")
	cmd.Flags().String("compression-type", opts.Type, "Tipe kompresi yang digunakan (gzip, zstd, xz, pgzip, zlib)")
	cmd.Flags().Int("compression-level", opts.Level, "Tingkat kompresi yang digunakan (1-9)")
}
