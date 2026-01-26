// File : cmd/dbcopy/p2s.go
// Deskripsi : Subcommand db-copy p2s (primary -> secondary)
// Author : Hadiyatna Muflihun
// Tanggal : 26 Januari 2026
// Last Modified : 26 Januari 2026
package dbcopycmd

import (
	"sfdbtools/internal/app/dbcopy"
	"sfdbtools/internal/app/dbcopy/helpers"
	"sfdbtools/internal/app/dbcopy/modes"
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
			if appdeps.Deps == nil || appdeps.Deps.Logger == nil || appdeps.Deps.Config == nil {
				return runner.ErrDependencyNotAvailable
			}

			// Parse flags
			opts, err := helpers.ParseP2SFlags(cmd)
			if err != nil {
				return err
			}

			// Validate options
			if err := helpers.ValidateP2SOptions(opts); err != nil {
				return err
			}

			// Execute via mode executor
			svc := dbcopy.NewService(appdeps.Deps.Logger, appdeps.Deps.Config)
			exec := modes.NewP2SExecutor(appdeps.Deps.Logger, svc, opts)

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
	CmdCopyP2S.Flags().StringP("client-code", "C", "", "Client code (untuk membentuk nama primary/secondary)")
	CmdCopyP2S.Flags().StringP("instance", "I", "", "Instance secondary target (suffix setelah _secondary_)")

	CmdCopyP2S.Flags().String("source-db", "", "Override nama database source (primary)")
	CmdCopyP2S.Flags().String("target-db", "", "Override nama database target (secondary)")
}
