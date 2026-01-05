package flags

import (
	"sfDBTools/internal/types"

	"github.com/spf13/cobra"
)

// CleanupFlags mendefinisikan flag untuk perintah cleanup
func AddCleanupFlags(cmd *cobra.Command, opts *types.CleanupOptions) {
	cmd.Flags().IntVar(&opts.Days, "days", opts.Days,
		"Jumlah hari untuk menyimpan backup. Backup yang lebih tua dari ini akan dihapus.")
	cmd.Flags().BoolVar(&opts.Background, "background", opts.Background,
		"Jalankan pembersihan di background (async mode)")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", opts.DryRun,
		"Tampilkan pratinjau tanpa menghapus file")
}
