// File : cmd/dbcopy/p2p.go
// Deskripsi : Subcommand db-copy p2p (primary -> primary)
// Author : Hadiyatna Muflihun
// Tanggal : 26 Januari 2026
// Last Modified : 26 Januari 2026
package dbcopycmd

import (
	"fmt"
	"sfdbtools/internal/app/dbcopy"
	"sfdbtools/internal/app/dbcopy/helpers"
	"sfdbtools/internal/app/dbcopy/modes"
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
			if appdeps.Deps == nil || appdeps.Deps.Logger == nil || appdeps.Deps.Config == nil {
				return fmt.Errorf("dependencies tidak tersedia")
			}

			// Parse flags
			opts, err := helpers.ParseP2PFlags(cmd)
			if err != nil {
				return err
			}

			// Validate options
			if err := helpers.ValidateP2POptions(opts); err != nil {
				return err
			}

			// Execute via mode executor
			svc := dbcopy.NewService(appdeps.Deps.Logger, appdeps.Deps.Config)
			exec := modes.NewP2PExecutor(appdeps.Deps.Logger, svc, opts)

			ctx, cancel := svc.SetupContext()
			defer cancel()

			result, err := exec.Execute(ctx)
			if err != nil {
				return err
			}

			if result.Success {
				appdeps.Deps.Logger.Infof("✓ %s", result.Message)
			}
			return nil
		})
	},
}

func init() {
	CmdCopyP2P.Flags().StringP("client-code", "C", "", "Client code source (untuk membentuk nama primary source)")
	CmdCopyP2P.Flags().String("target-client-code", "", "Client code target (default: sama dengan --client-code)")

	CmdCopyP2P.Flags().String("source-db", "", "Override nama database source (primary)")
	CmdCopyP2P.Flags().String("target-db", "", "Override nama database target (primary)")
}
