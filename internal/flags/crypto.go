package flags

import (
	"github.com/spf13/cobra"
)

// AddBase64EncodeFlags mendaftarkan flags untuk base64 encode
func AddBase64EncodeFlags(cmd *cobra.Command) {
	cmd.Flags().String("text", "", "Teks input (opsional, default baca stdin jika ada)")
	cmd.Flags().StringP("out", "o", "", "File output (opsional)")
}

// AddBase64DecodeFlags mendaftarkan flags untuk base64 decode
func AddBase64DecodeFlags(cmd *cobra.Command) {
	cmd.Flags().String("data", "", "Input base64 (opsional, default baca stdin jika ada)")
	cmd.Flags().StringP("out", "o", "", "File output (opsional)")
}

// AddEncryptFileFlags mendaftarkan flags untuk encrypt file
func AddEncryptFileFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("in", "i", "", "Path file input (wajib)")
	cmd.Flags().StringP("out", "o", "", "Path file output (wajib)")
	cmd.Flags().StringP("key", "k", "", "Encryption key (opsional, jika kosong pakai env atau prompt)")
}

// AddDecryptFileFlags mendaftarkan flags untuk decrypt file
func AddDecryptFileFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("in", "i", "", "Path file input terenkripsi (wajib)")
	cmd.Flags().StringP("out", "o", "", "Path file output hasil dekripsi (wajib)")
	cmd.Flags().StringP("key", "k", "", "Encryption key (opsional, jika kosong pakai env atau prompt)")
}

// AddEncryptTextFlags mendaftarkan flags untuk encrypt text
func AddEncryptTextFlags(cmd *cobra.Command) {
	cmd.Flags().String("text", "", "Teks input (kosongkan untuk baca dari stdin)")
	cmd.Flags().StringP("out", "o", "", "File output biner (opsional)")
	cmd.Flags().StringP("key", "k", "", "Encryption key (opsional, jika kosong pakai env atau prompt)")
}

// AddDecryptTextFlags mendaftarkan flags untuk decrypt text
func AddDecryptTextFlags(cmd *cobra.Command) {
	cmd.Flags().String("data", "", "Input terenkripsi base64 (kosongkan untuk baca dari stdin)")
	cmd.Flags().StringP("out", "o", "", "File output plaintext (opsional)")
	cmd.Flags().StringP("key", "k", "", "Encryption key (opsional, jika kosong pakai env atau prompt)")
}
