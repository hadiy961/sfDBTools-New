// File : cmd/cleanup/schedule.go
// Deskripsi : Command untuk mengelola scheduler cleanup (systemd timer)
// Author : Hadiyatna Muflihun
// Tanggal : 2 Januari 2026
// Last Modified : 5 Januari 2026
package cleanupcmd

import (
	"sfdbtools/internal/app/cleanup/scheduler"
	appdeps "sfdbtools/internal/cli/deps"

	"github.com/spf13/cobra"
)

// CmdCleanupSchedule adalah root command untuk scheduler cleanup berbasis systemd.
var CmdCleanupSchedule = &cobra.Command{
	Use:   "schedule",
	Short: "Kelola scheduler cleanup (systemd timer)",
	Long: `Kelola scheduler cleanup berbasis systemd.

Perintah ini TIDAK menjalankan daemon internal. sfdbtools akan:
- Membuat/meng-update unit systemd (service + timer)
- Enable/disable timer berdasarkan config.yaml

Eksekusi dipaksa serial (antri) via global lock (flock), sehingga aman terhadap scheduled backup.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	CmdCleanupSchedule.AddCommand(cmdCleanupScheduleStart)
	CmdCleanupSchedule.AddCommand(cmdCleanupScheduleStop)
	CmdCleanupSchedule.AddCommand(cmdCleanupScheduleStatus)
	CmdCleanupSchedule.PersistentFlags().Bool("dry-run", false, "Jalankan scheduler cleanup dalam mode dry-run (tidak menghapus file)")
}

var cmdCleanupScheduleStart = &cobra.Command{
	Use:   "start",
	Short: "Aktifkan scheduler cleanup (enable timer)",
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		return scheduler.Start(cmd.Context(), appdeps.Deps, dryRun)
	},
}

var cmdCleanupScheduleStop = &cobra.Command{
	Use:   "stop",
	Short: "Nonaktifkan scheduler cleanup (disable timer)",
	RunE: func(cmd *cobra.Command, args []string) error {
		killRunning, _ := cmd.Flags().GetBool("kill-running")
		return scheduler.Stop(cmd.Context(), appdeps.Deps, killRunning)
	},
}

var cmdCleanupScheduleStatus = &cobra.Command{
	Use:   "status",
	Short: "Tampilkan status scheduler cleanup",
	RunE: func(cmd *cobra.Command, args []string) error {
		return scheduler.Status(cmd.Context(), appdeps.Deps)
	},
}

func init() {
	cmdCleanupScheduleStop.Flags().Bool("kill-running", false, "Hentikan proses cleanup yang sedang berjalan")
}
