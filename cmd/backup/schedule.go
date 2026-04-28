package backupcmd

import (
	"fmt"
	"sfdbtools/internal/app/backup/scheduler"
	appdeps "sfdbtools/internal/cli/deps"

	"github.com/spf13/cobra"
)

var CmdBackupSchedule = &cobra.Command{
	Use:   "schedule",
	Short: "Kelola penjadwalan backup (systemd timers)",
	Long:  "Mengelola penjadwalan backup secara otomatis menggunakan systemd timers, dikonfigurasi melalui config.yaml",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var CmdBackupScheduleSync = &cobra.Command{
	Use:   "sync",
	Short: "Sinkronisasi config.yaml menjadi systemd timers",
	RunE: func(cmd *cobra.Command, args []string) error {
		deps := appdeps.Deps
		if deps == nil || deps.Config == nil {
			return fmt.Errorf("konfigurasi aplikasi belum dimuat")
		}

		mgr := scheduler.NewManager(deps.Config, deps.Logger)
		if err := mgr.Sync(cmd.Context()); err != nil {
			return fmt.Errorf("gagal sync scheduler: %w", err)
		}

		deps.Logger.Info("Sinkronisasi jadwal backup berhasil.")
		return nil
	},
}

var CmdBackupScheduleStatus = &cobra.Command{
	Use:   "status",
	Short: "Lihat status seluruh job scheduler backup",
	RunE: func(cmd *cobra.Command, args []string) error {
		deps := appdeps.Deps
		if deps == nil || deps.Config == nil {
			return fmt.Errorf("konfigurasi aplikasi belum dimuat")
		}

		mgr := scheduler.NewManager(deps.Config, deps.Logger)
		if err := mgr.ShowStatus(cmd.Context()); err != nil {
			return fmt.Errorf("gagal menampilkan status scheduler: %w", err)
		}

		return nil
	},
}

func init() {
	CmdBackupSchedule.AddCommand(CmdBackupScheduleSync)
	CmdBackupSchedule.AddCommand(CmdBackupScheduleStatus)
}
