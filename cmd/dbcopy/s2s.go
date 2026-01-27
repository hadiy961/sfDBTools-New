// File : cmd/dbcopy/s2s.go
// Deskripsi : Subcommand db-copy s2s (secondary -> secondary)
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
			if appdeps.Deps == nil || appdeps.Deps.Logger == nil || appdeps.Deps.Config == nil {
				return fmt.Errorf("dependencies tidak tersedia")
			}

			// Parse flags
			opts, err := helpers.ParseS2SFlags(cmd)
			if err != nil {
				return err
			}

			// Validate options
			if err := helpers.ValidateS2SOptions(opts); err != nil {
				return err
			}

			// Execute via mode executor
			svc := dbcopy.NewService(appdeps.Deps.Logger, appdeps.Deps.Config)
			exec := modes.NewS2SExecutor(appdeps.Deps.Logger, svc, opts)

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
	CmdCopyS2S.Flags().StringP("client-code", "C", "", "Client code")
	CmdCopyS2S.Flags().String("source-instance", "", "Instance secondary sumber")
	CmdCopyS2S.Flags().String("target-instance", "", "Instance secondary target")

	CmdCopyS2S.Flags().Bool("prebackup-target", true, "Backup target sebelum overwrite (safety)")

	CmdCopyS2S.Flags().String("source-db", "", "Override nama database source (secondary)")
	CmdCopyS2S.Flags().String("target-db", "", "Override nama database target (secondary)")
}
