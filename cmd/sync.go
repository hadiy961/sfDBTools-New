package cmd

import (
	"context"
	"fmt"
	"sfdbtools/internal/app/sync"
	appdeps "sfdbtools/internal/cli/deps"
	"sfdbtools/internal/shared/database"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var syncDryRun bool

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sinkronisasi data dengan remote hub (Postgres/MySQL)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !database.GetSettingBool("sync_enabled") {
			return fmt.Errorf("sinkronisasi dinonaktifkan")
		}

		if syncDryRun {
			fmt.Println(color.YellowString("[DRY-RUN] Simulating synchronization..."))
			time.Sleep(1 * time.Second)
			fmt.Println(color.GreenString("[DRY-RUN] No changes applied."))
			return nil
		}

		remote, err := database.ConnectToRemoteHub()
		if err != nil {
			return err
		}
		defer remote.Close()

		local, _ := database.GetSQLite()
		clientCode := appdeps.Deps.Config.General.ClientCode
		
		svc := sync.NewService(local, remote, clientCode)
		ctx, cancel := context.WithTimeout(cmd.Context(), 120*time.Second)
		defer cancel()

		fmt.Println(color.CyanString("Memulai sinkronisasi otomatis..."))
		return svc.RunSync(ctx, database.GetSetting("sync_mode"))
	},
}

func init() {
	syncCmd.Flags().BoolVarP(&syncDryRun, "dry-run", "d", false, "Preview changes without applying them")
	rootCmd.AddCommand(syncCmd)
}
