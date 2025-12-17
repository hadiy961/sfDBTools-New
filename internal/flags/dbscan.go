package flags

import (

	"github.com/spf13/cobra"
)

// AddDbScanFilterFlags mendaftarkan flags untuk perintah `dbscan filter`.
// Parsing dilakukan terpisah di pkg/parsing.
func AddDbScanFilterFlags(cmd *cobra.Command) {
	// Profile (Manual string flags karena parsing dilakukan manual nanti)
	cmd.Flags().String("profile", "", "Path ke file profil database terenkripsi")
	cmd.Flags().String("profile-key", "", "Encryption key untuk decrypt file profil database")

	// Filters (Simple version without struct binding)
	AddFilterFlagsSimple(cmd)
}

// AddDbScanAllFlags mendaftarkan flags minimal untuk command `dbscan all`.
// Hanya flag yang diminta: --background, --exclude-system, --profile, --profile-key
func AddDbScanAllFlags(cmd *cobra.Command) {
	// Profile (Manual definition agar tidak butuh struct binding)
	cmd.Flags().String("profile", "", "Path ke file konfigurasi database (encrypted)")
	cmd.Flags().String("profile-key", "", "Encryption key untuk decrypt file profil database")

	// Filter option
	cmd.Flags().Bool("exclude-system", true, "Kecualikan system databases")

	cmd.Flags().Bool("background", false, "Jalankan scanning di background (async mode)")
	cmd.Flags().Bool("show-options", true, "Tampilkan opsi scanning yang digunakan sebelum eksekusi")
}
