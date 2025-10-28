package flags

import "github.com/spf13/cobra"

func AddDBScanAllFlags(cmd *cobra.Command) {
	ProfileSelect(cmd)
	cmd.Flags().BoolP("exclude-system", "e", true, "Mengecualikan database sistem dari hasil scan")
	cmd.Flags().BoolP("save-to-db", "s", true, "Menyimpan hasil scan ke database")
	cmd.Flags().BoolP("background", "b", false, "Menjalankan Scanning di mode background")
}
