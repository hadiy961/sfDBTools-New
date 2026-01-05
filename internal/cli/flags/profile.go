package flags

import (
	defaultVal "sfDBTools/internal/cli/defaults"

	"github.com/spf13/cobra"
)

// ProfileCreate - Flag untuk membuat profil baru
func ProfileCreate(cmd *cobra.Command) {
	// Ambil default dari defaultVal
	AddDBInfoFlags(cmd)
	defaultOptions := defaultVal.DefaultProfileCreateOptions()

	// Tambahkan flag spesifik untuk pembuatan profil
	cmd.Flags().StringP("profile", "n", defaultOptions.ProfileInfo.Name, "Nama profil yang akan dibuat")
	cmd.Flags().StringP("output-dir", "o", defaultOptions.OutputDir, "Direktori output untuk menyimpan profil yang dibuat (opsional)")
	cmd.Flags().StringP("profile-key", "k", "", "Kunci enkripsi untuk mengenkripsi file profil (ENV: SFDB_TARGET_PROFILE_KEY atau SFDB_SOURCE_PROFILE_KEY)")

	// SSH tunnel (opsional)
	cmd.Flags().Bool("ssh", false, "Aktifkan koneksi database melalui SSH tunnel")
	cmd.Flags().String("ssh-host", "", "SSH bastion host")
	cmd.Flags().Int("ssh-port", 22, "SSH port")
	cmd.Flags().String("ssh-user", "", "SSH username")
	cmd.Flags().String("ssh-password", "", "SSH password (opsional)")
	cmd.Flags().String("ssh-identity-file", "", "Path ke SSH private key (opsional)")
	cmd.Flags().Int("ssh-local-port", 0, "Local port untuk SSH tunnel (0 = otomatis)")
}

// ProfileEdit - Flag untuk mengedit profil yang ada
func ProfileEdit(cmd *cobra.Command) {
	// Ambil default dari defaultVal
	AddDBInfoFlags(cmd)
	// defaultOptions tidak lagi dipakai (flag --interactive dihapus)

	// Tambahkan flag spesifik untuk mengedit profil
	cmd.Flags().StringP("profile", "f", "", "Nama file profil yang akan diedit")
	// new-name: apabila diberikan, lakukan rename saat menyimpan
	cmd.Flags().StringP("new-name", "N", "", "Nama baru untuk file profil (akan merename file saat menyimpan)")
	cmd.Flags().StringP("profile-key", "k", "", "Kunci enkripsi untuk mendekripsi/enkripsi file profil (ENV: SFDB_TARGET_PROFILE_KEY atau SFDB_SOURCE_PROFILE_KEY)")

	// SSH tunnel (opsional)
	cmd.Flags().Bool("ssh", false, "Aktifkan koneksi database melalui SSH tunnel")
	cmd.Flags().String("ssh-host", "", "SSH bastion host")
	cmd.Flags().Int("ssh-port", 22, "SSH port")
	cmd.Flags().String("ssh-user", "", "SSH username")
	cmd.Flags().String("ssh-password", "", "SSH password (opsional)")
	cmd.Flags().String("ssh-identity-file", "", "Path ke SSH private key (opsional)")
	cmd.Flags().Int("ssh-local-port", 0, "Local port untuk SSH tunnel (0 = otomatis)")
}

// ProfileShow - Flag untuk menampilkan profil yang ada
func ProfileShow(cmd *cobra.Command) {
	// Ambil default dari defaultVal
	defaultOptions := defaultVal.DefaultProfileShowOptions()

	// Tambahkan flag spesifik untuk melihat profil
	cmd.Flags().StringP("profile", "f", "", "Nama file profil yang akan ditampilkan")
	cmd.Flags().StringP("profile-key", "k", "", "Kunci enkripsi untuk mendekripsi file profil (ENV: SFDB_TARGET_PROFILE_KEY atau SFDB_SOURCE_PROFILE_KEY)")
	cmd.Flags().BoolP("reveal-password", "r", defaultOptions.RevealPassword, "Tampilkan password secara jelas saat menampilkan profil")
}

// ProfileDelete - Flag untuk menghapus profil yang ada
func ProfileDelete(cmd *cobra.Command) {
	// Tambahkan flag spesifik untuk menghapus profil
	cmd.Flags().StringSliceP("profile", "f", []string{}, "Nama file profil yang akan dihapus (bisa multiple)")
	cmd.Flags().BoolP("force", "F", false, "Hapus profil tanpa konfirmasi")
}
