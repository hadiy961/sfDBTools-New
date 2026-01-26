// File : cmd/dbcopy/p2p.go
// Deskripsi : Subcommand db-copy p2p (primary -> primary)
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

// CmdCopyP2P: primary -> primary
var CmdCopyP2P = &cobra.Command{
	Use:   "p2p",
	Short: "Copy primary -> primary",
	Long: `Copy database primary ke database primary (umumnya beda server/profile).

Rule-based:
  --client-code dan (opsional) --target-client-code

Override eksplisit:
  --source-db / --target-db`,
	Run: func(cmd *cobra.Command, args []string) {
		runner.Run(cmd, func() error {
			return dbcopy.ExecuteCopyP2P(cmd, appdeps.Deps)
		})
	},
}

func init() {
	CmdCopyP2P.Flags().StringP("client-code", "C", "", "Client code source (untuk membentuk nama primary source)")
	CmdCopyP2P.Flags().String("target-client-code", "", "Client code target (default: sama dengan --client-code)")

	CmdCopyP2P.Flags().String("source-db", "", "Override nama database source (primary)")
	CmdCopyP2P.Flags().String("target-db", "", "Override nama database target (primary)")
}
