package flags

import (
	defaultVal "sfDBTools/internal/defaultval"

	"github.com/spf13/cobra"
)

// ProfileCreate - Flag untuk membuat profil baru
func ProfileCreate(cmd *cobra.Command) {
	// Ambil default dari defaultVal
	AddDBInfoFlags(cmd)
	defaultOptions := defaultVal.DefaultProfileCreateOptions()

	// Tambahkan flag spesifik untuk pembuatan profil
	cmd.Flags().StringP("profil", "n", defaultOptions.ProfileInfo.Name, "Nama profil yang akan dibuat")
	cmd.Flags().StringP("output-dir", "o", defaultOptions.OutputDir, "Direktori output untuk menyimpan profil yang dibuat (opsional)")
	cmd.Flags().StringP("key", "k", "", "Kunci enkripsi untuk mengenkripsi file profil")
	cmd.Flags().BoolP("interactive", "i", defaultOptions.Interactive, "Mode interaktif untuk memasukkan informasi profil")
}

// ProfileEdit - Flag untuk mengedit profil yang ada
func ProfileEdit(cmd *cobra.Command) {
	// Ambil default dari defaultVal
	AddDBInfoFlags(cmd)
	defaultOptions := defaultVal.DefaultProfileCreateOptions()

	// Tambahkan flag spesifik untuk mengedit profil
	cmd.Flags().StringP("file", "f", "", "Nama file profil yang akan diedit")
	// new-name: apabila diberikan, lakukan rename saat menyimpan
	cmd.Flags().StringP("new-name", "N", "", "Nama baru untuk file profil (akan merename file saat menyimpan)")
	cmd.Flags().StringP("key", "k", "", "Kunci enkripsi untuk mendekripsi/enkripsi file profil")
	cmd.Flags().BoolP("interactive", "i", defaultOptions.Interactive, "Mode interaktif untuk mengedit informasi profil")
}

// ProfileShow - Flag untuk menampilkan profil yang ada
func ProfileShow(cmd *cobra.Command) {
	// Ambil default dari defaultVal
	defaultOptions := defaultVal.DefaultProfileShowOptions()

	// Tambahkan flag spesifik untuk melihat profil
	cmd.Flags().StringP("file", "f", "", "Nama file profil yang akan ditampilkan")
	cmd.Flags().StringP("key", "k", "", "kunci enkripsi untuk mendekripsi file profil")
	cmd.Flags().BoolP("reveal-password", "r", defaultOptions.RevealPassword, "Tampilkan password secara jelas saat menampilkan profil")
}

// ProfileDelete - Flag untuk menghapus profil yang ada
func ProfileDelete(cmd *cobra.Command) {
	// Tambahkan flag spesifik untuk menghapus profil
	cmd.Flags().StringP("file", "f", "", "Nama file profil yang akan dihapus")
	cmd.Flags().BoolP("force", "F", false, "Hapus profil tanpa konfirmasi")
}
