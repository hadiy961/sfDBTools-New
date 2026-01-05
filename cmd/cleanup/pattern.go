package cleanupcmd

import (
	"sfDBTools/internal/app/cleanup"
	defaultVal "sfDBTools/internal/cli/defaults"
	appdeps "sfDBTools/internal/cli/deps"
	"sfDBTools/internal/cli/flags"

	"github.com/spf13/cobra"
)

// CmdCleanupPattern menjalankan pembersihan berdasarkan pola glob tertentu
var CmdCleanupPattern = &cobra.Command{
	Use:   "pattern",
	Short: "Hapus file backup yang cocok dengan pola tertentu",
	Long: `Membersihkan file backup lama yang cocok dengan pola glob tertentu (contoh: "**/*.sql.gz").
Hanya file yang lebih tua dari kebijakan retensi yang akan dipertimbangkan.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := cleanup.ExecuteCleanup(cmd, appdeps.Deps, "pattern"); err != nil {
			appdeps.Deps.Logger.Error("cleanup by pattern gagal: " + err.Error())
		}
	},
}

func init() {
	defaultOpts := defaultVal.DefaultCleanupOptions()

	// Add pattern-specific flag
	CmdCleanupPattern.Flags().StringVarP(&defaultOpts.Pattern, "pattern", "p", "", "Pola glob untuk memilih file (contoh: '**/*.sql.gz')")
	CmdCleanupPattern.MarkFlagRequired("pattern")

	// Add common cleanup flags
	flags.AddCleanupFlags(CmdCleanupPattern, &defaultOpts)

	CmdCleanupMain.AddCommand(CmdCleanupPattern)
}
