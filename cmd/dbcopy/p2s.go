// File : cmd/dbcopy/p2s.go
// Deskripsi : Subcommand db-copy p2s (primary -> secondary)
// Author : Hadiyatna Muflihun
// Tanggal : 26 Januari 2026
// Last Modified : 26 Januari 2026
package dbcopycmd

import (
	"sfdbtools/internal/app/dbcopy"
	appdeps "sfdbtools/internal/cli/deps"
	"sfdbtools/internal/cli/runner"

	"github.com/spf13/cobra"
)

// CmdCopyP2S: primary -> secondary
var CmdCopyP2S = &cobra.Command{
	Use:   "p2s",
	Short: "Copy primary -> secondary",
	Long: `Copy database primary ke database secondary.

Mode rule-based (automation):
  --client-code dan --instance

Mode eksplisit:
  --source-db dan --target-db`,
	Run: func(cmd *cobra.Command, args []string) {
		runner.Run(cmd, func() error {
			return dbcopy.ExecuteCopyP2S(cmd, appdeps.Deps)
		})
	},
}

func init() {
	CmdCopyP2S.Flags().StringP("client-code", "C", "", "Client code (untuk membentuk nama primary/secondary)")
	CmdCopyP2S.Flags().StringP("instance", "I", "", "Instance secondary target (suffix setelah _secondary_)")

	CmdCopyP2S.Flags().String("source-db", "", "Override nama database source (primary)")
	CmdCopyP2S.Flags().String("target-db", "", "Override nama database target (secondary)")
}
