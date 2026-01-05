package flags

import "github.com/spf13/cobra"

func AddScriptEncryptFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("file", "f", "", "Path file entrypoint (.sh) (wajib)")
	cmd.Flags().StringP("key", "k", "", "Encryption key (opsional, jika kosong pakai env SFDB_SCRIPT_KEY atau prompt)")
	cmd.Flags().String("encryption-key", "", "(deprecated) Gunakan --key atau -k")
	_ = cmd.Flags().MarkHidden("encryption-key")
	cmd.Flags().StringP("mode", "m", "bundle", "Mode encrypt: bundle|single")
	cmd.Flags().String("output", "", "Path output file .sftools (opsional, jika kosong pakai config YAML/otomatis)")
	cmd.Flags().Bool("delete-source", false, "Hapus sumber setelah encrypt berhasil (single=file, bundle=folder) (opsional)")
}

func AddScriptRunFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("file", "f", "", "Path file .sftools (wajib)")
	cmd.Flags().StringP("key", "k", "", "Encryption key (opsional, jika kosong pakai env SFDB_SCRIPT_KEY atau prompt)")
	cmd.Flags().String("encryption-key", "", "(deprecated) Gunakan --key atau -k")
	_ = cmd.Flags().MarkHidden("encryption-key")
}

func AddScriptExtractFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("file", "f", "", "Path file .sftools (wajib)")
	cmd.Flags().StringP("key", "k", "", "Encryption key (opsional, jika kosong pakai env SFDB_SCRIPT_KEY atau prompt)")
	cmd.Flags().String("encryption-key", "", "(deprecated) Gunakan --key atau -k")
	_ = cmd.Flags().MarkHidden("encryption-key")
	cmd.Flags().StringP("out-dir", "o", "", "Directory output untuk hasil extract (wajib)")
}

func AddScriptInfoFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("file", "f", "", "Path file .sftools (wajib)")
	cmd.Flags().StringP("key", "k", "", "Encryption key (opsional, jika kosong pakai env SFDB_SCRIPT_KEY atau prompt)")
	cmd.Flags().String("encryption-key", "", "(deprecated) Gunakan --key atau -k")
	_ = cmd.Flags().MarkHidden("encryption-key")
}
