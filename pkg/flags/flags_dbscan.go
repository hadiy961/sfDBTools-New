package flags

import (
	"sfDBTools/internal/types"

	"github.com/spf13/cobra"
)

// AddDbScanFilterFlags mendaftarkan flags untuk perintah `dbscan filter`.
// Parsing dilakukan terpisah di pkg/parsing.
func AddDbScanFilterFlags(cmd *cobra.Command) {
	// Profile
	cmd.Flags().String("profile", "", "Path ke file profil database terenkripsi")
	cmd.Flags().String("profile-key", "", "Encryption key untuk decrypt file profil database")

	// Filters
	cmd.Flags().Bool("exclude-system", true, "Kecualikan system databases (information_schema, mysql, dll)")
	cmd.Flags().String("exclude-db", "", "Daftar database yang akan dikecualikan (comma-separated)")
	cmd.Flags().String("exclude-db-file", "", "File berisi daftar database yang akan dikecualikan (satu per baris)")
	cmd.Flags().String("db", "", "Daftar database yang akan di-scan (comma-separated)")
	cmd.Flags().String("db-file", "", "File berisi daftar database yang akan di-scan (satu per baris)")
}

// AddDbScanAllFlags mendaftarkan flags minimal untuk command `dbscan all`.
// Hanya flag yang diminta: --background, --exclude-system, --profile, --profile-key
func AddDbScanAllFlags(cmd *cobra.Command, opts *types.ScanOptions) {
	// Profile
	cmd.Flags().StringVar(&opts.ProfileInfo.Path, "profile", opts.ProfileInfo.Path,
		"Path ke file konfigurasi database (encrypted)")
	cmd.Flags().StringVar(&opts.Encryption.Key, "profile-key", opts.Encryption.Key,
		"Encryption key untuk decrypt file profil database")

	// Filter option
	cmd.Flags().BoolVar(&opts.ExcludeSystem, "exclude-system", opts.ExcludeSystem,
		"Kecualikan system databases")

	cmd.Flags().BoolVar(&opts.Background, "background", opts.Background,
		"Jalankan scanning di background (async mode)")
	cmd.Flags().BoolVar(&opts.ShowOptions, "show-options", opts.ShowOptions, "Tampilkan opsi scanning yang digunakan sebelum eksekusi")
}
