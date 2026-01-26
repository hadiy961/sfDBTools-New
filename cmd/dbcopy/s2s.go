// File : cmd/dbcopy/s2s.go
// Deskripsi : Subcommand db-copy s2s (secondary -> secondary)
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

// CmdCopyS2S: secondary -> secondary
var CmdCopyS2S = &cobra.Command{
	Use:   "s2s",
	Short: "Copy secondary -> secondary",
	Long: `Copy database secondary ke secondary.

Rule-based (automation):
  --client-code, --source-instance, --target-instance

Override eksplisit:
  --source-db / --target-db`,
	Run: func(cmd *cobra.Command, args []string) {
		runner.Run(cmd, func() error {
			return dbcopy.ExecuteCopyS2S(cmd, appdeps.Deps)
		})
	},
}

func init() {
	CmdCopyS2S.Flags().StringP("client-code", "C", "", "Client code")
	CmdCopyS2S.Flags().String("source-instance", "", "Instance secondary sumber")
	CmdCopyS2S.Flags().String("target-instance", "", "Instance secondary target")

	CmdCopyS2S.Flags().String("source-db", "", "Override nama database source (secondary)")
	CmdCopyS2S.Flags().String("target-db", "", "Override nama database target (secondary)")
}
