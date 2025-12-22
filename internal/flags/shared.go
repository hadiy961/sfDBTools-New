package flags

import (
	"sfDBTools/internal/types"

	"github.com/spf13/cobra"
)

// AddProfileFlags mendaftarkan flag standar untuk memilih profile database source.
// Flag: --profile, --profile-key
func AddProfileFlags(cmd *cobra.Command, opts *types.ProfileInfo) {
	cmd.Flags().StringVarP(&opts.Path, "profile", "p", opts.Path, "Path ke file profil database terenkripsi")
	cmd.Flags().StringVarP(&opts.EncryptionKey, "profile-key", "P", opts.EncryptionKey, "Encryption key untuk decrypt file profil database")
}

// AddEncryptionFlags mendaftarkan flag untuk opsi enkripsi output.
// Flag: --encryption-key
func AddEncryptionFlags(cmd *cobra.Command, opts *types.EncryptionOptions) {
	cmd.Flags().StringVarP(&opts.Key, "encryption-key", "K", opts.Key, "Kunci enkripsi yang digunakan")
}

// AddCompressionFlags mendaftarkan flag untuk opsi kompresi.
// Flag: --compress-type, --compress-level
func AddCompressionFlags(cmd *cobra.Command, opts *types.CompressionOptions) {
	cmd.Flags().StringVarP(&opts.Type, "compress-type", "C", opts.Type, "Tipe kompresi (gzip, zstd, xz, pgzip, zlib, none)")
	cmd.Flags().IntVar(&opts.Level, "compress-level", opts.Level, "Tingkat kompresi (1-9)")
}

// AddFilterFlags mendaftarkan flag untuk filtering database (Include/Exclude).
// Flag: --db, --db-file, --exclude-db, --exclude-db-file, --exclude-system
func AddFilterFlags(cmd *cobra.Command, opts *types.FilterOptions) {
	cmd.Flags().StringArrayVar(&opts.IncludeDatabases, "db", opts.IncludeDatabases, "Daftar database yang akan di-include (comma-separated). Dapat dikombinasi dengan --db-file.")
	cmd.Flags().StringVar(&opts.IncludeFile, "db-file", opts.IncludeFile, "File berisi daftar database yang akan di-include (satu per baris).")

	cmd.Flags().StringArrayVar(&opts.ExcludeDatabases, "exclude-db", opts.ExcludeDatabases, "Daftar database yang akan dikecualikan (comma-separated).")
	cmd.Flags().StringVar(&opts.ExcludeDBFile, "excludeDBFile", opts.ExcludeDBFile, "File berisi daftar database yang akan dikecualikan.")

	// Alias flag agar kompatibel dengan berbagai penamaan di command yang berbeda jika perlu
	// Namun standarisasi lebih baik: gunakan exclude-file di parsing jika ingin konsisten,
	// atau kita tambahkan flag alias disini.
	cmd.Flags().StringVar(&opts.ExcludeDBFile, "exclude-file", opts.ExcludeDBFile, "Alias untuk --exclude-db-file")

	cmd.Flags().BoolVar(&opts.ExcludeSystem, "exclude-system", opts.ExcludeSystem, "Kecualikan system databases (information_schema, mysql, dll)")
}

// AddFilterFlagsSimple mendaftarkan flag filter versi string pointer (untuk struct yang belum menggunakan FilterOptions)
// Ini berguna untuk transisi `dbscan` yang struct-nya masih flat.
func AddFilterFlagsSimple(cmd *cobra.Command) {
	cmd.Flags().String("db", "", "Daftar database yang akan di-scan/backup (comma-separated).")
	cmd.Flags().String("db-file", "", "File berisi daftar database yang akan di-scan/backup.")
	cmd.Flags().String("exclude-db", "", "Daftar database yang akan dikecualikan.")
	cmd.Flags().String("exclude-file", "", "File berisi daftar database yang akan dikecualikan.")
	cmd.Flags().Bool("exclude-system", true, "Kecualikan system databases")
}

// AddDBInfoFlags mendaftarkan flag koneksi database standar.
// Flag: --host, --port, --user, --password
func AddDBInfoFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("host", "H", "", "Database host")
	cmd.Flags().IntP("port", "P", 0, "Database port")
	cmd.Flags().StringP("user", "U", "", "Database username")
	cmd.Flags().StringP("password", "p", "", "Database password")
}
