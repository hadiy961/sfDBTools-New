// File : cmd/backup/schedule.go
// Deskripsi : Command untuk mengelola scheduler backup (systemd timer)
// Author : Hadiyatna Muflihun
// Tanggal : 2026-01-02
// Last Modified : 2026-01-05
package backupcmd

import (
	"fmt"
	"os"
	"sfdbtools/internal/app/backup/scheduler"
	appdeps "sfdbtools/internal/cli/deps"
	"sfdbtools/internal/shared/consts"

	"github.com/spf13/cobra"
)

// CmdBackupSchedule adalah root command untuk scheduler backup berbasis systemd.
var CmdBackupSchedule = &cobra.Command{
	Use:   "schedule",
	Short: "Kelola scheduler backup (systemd timer)",
	Long: `Kelola scheduler backup berbasis systemd.

Perintah ini TIDAK menjalankan daemon internal. sfdbtools akan:
- Membuat/meng-update unit systemd (service + timer)
- Enable/disable timer berdasarkan config.yaml

Eksekusi job dipaksa serial (antri) via global lock (flock).`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	CmdBackupSchedule.AddCommand(cmdBackupScheduleStart)
	CmdBackupSchedule.AddCommand(cmdBackupScheduleStop)
	CmdBackupSchedule.AddCommand(cmdBackupScheduleStatus)
	CmdBackupSchedule.AddCommand(cmdBackupScheduleHistory)
	CmdBackupSchedule.AddCommand(cmdBackupScheduleRun)
}

var cmdBackupScheduleStart = &cobra.Command{
	Use:   "start",
	Short: "Aktifkan scheduler backup (enable timer)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if appdeps.Deps == nil {
			cmd.PrintErrln("✗ Dependencies tidak tersedia. Pastikan aplikasi diinisialisasi dengan benar.")
			return fmt.Errorf("dependencies tidak tersedia")
		}
		jobName, _ := cmd.Flags().GetString("job")
		return scheduler.Start(cmd.Context(), appdeps.Deps, jobName)
	},
}

var cmdBackupScheduleStop = &cobra.Command{
	Use:   "stop",
	Short: "Nonaktifkan scheduler backup (disable timer)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if appdeps.Deps == nil {
			cmd.PrintErrln("✗ Dependencies tidak tersedia. Pastikan aplikasi diinisialisasi dengan benar.")
			return fmt.Errorf("dependencies tidak tersedia")
		}
		jobName, _ := cmd.Flags().GetString("job")
		killRunning, _ := cmd.Flags().GetBool("kill-running")
		return scheduler.Stop(cmd.Context(), appdeps.Deps, jobName, killRunning)
	},
}

var cmdBackupScheduleStatus = &cobra.Command{
	Use:   "status",
	Short: "Tampilkan status scheduler backup",
	RunE: func(cmd *cobra.Command, args []string) error {
		if appdeps.Deps == nil {
			cmd.PrintErrln("✗ Dependencies tidak tersedia. Pastikan aplikasi diinisialisasi dengan benar.")
			return fmt.Errorf("dependencies tidak tersedia")
		}
		jobName, _ := cmd.Flags().GetString("job")
		return scheduler.Status(cmd.Context(), appdeps.Deps, jobName)
	},
}

var cmdBackupScheduleHistory = &cobra.Command{
	Use:   "history",
	Short: "Tampilkan history eksekusi job scheduler (journalctl)",
	Long: `Tampilkan history eksekusi job scheduler melalui systemd journal.

Secara default akan menampilkan log dari unit service (eksekusi backup).
Gunakan --timer untuk melihat event timer (trigger scheduler).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if appdeps.Deps == nil {
			cmd.PrintErrln("✗ Dependencies tidak tersedia. Pastikan aplikasi diinisialisasi dengan benar.")
			return fmt.Errorf("dependencies tidak tersedia")
		}
		jobName, _ := cmd.Flags().GetString("job")
		since, _ := cmd.Flags().GetString("since")
		until, _ := cmd.Flags().GetString("until")
		lines, _ := cmd.Flags().GetInt("lines")
		follow, _ := cmd.Flags().GetBool("follow")
		useTimer, _ := cmd.Flags().GetBool("timer")
		return scheduler.History(cmd.Context(), appdeps.Deps, jobName, scheduler.HistoryOptions{
			Since:  since,
			Until:  until,
			Lines:  lines,
			Follow: follow,
			Timer:  useTimer,
		})
	},
}

// cmdBackupScheduleRun diekseksekusi oleh unit systemd (service) untuk menjalankan satu job.
var cmdBackupScheduleRun = &cobra.Command{
	Use:   "run",
	Short: "Jalankan satu job scheduler (internal/systemd)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if appdeps.Deps == nil {
			cmd.PrintErrln("✗ Dependencies tidak tersedia. Pastikan aplikasi diinisialisasi dengan benar.")
			os.Exit(consts.ExitCodePermanentError)
		}
		jobName, _ := cmd.Flags().GetString("job")
		err := scheduler.RunJob(cmd.Context(), appdeps.Deps, jobName)
		if err != nil {
			// Semantic exit codes untuk systemd restart policy
			if scheduler.IsTransientError(err) {
				appdeps.Deps.Logger.Warnf("Transient error detected, exit dengan code %d untuk systemd retry", consts.ExitCodeTransientError)
				os.Exit(consts.ExitCodeTransientError)
			}
			// Permanent error - systemd should not retry
			appdeps.Deps.Logger.Errorf("Permanent error detected, exit dengan code %d", consts.ExitCodePermanentError)
			os.Exit(consts.ExitCodePermanentError)
		}
		return nil
	},
	Hidden: true,
}

func init() {
	cmdBackupScheduleStart.Flags().String("job", "", "Nama job scheduler (kosong = semua job enabled)")

	cmdBackupScheduleStop.Flags().String("job", "", "Nama job scheduler (kosong = semua job enabled)")
	cmdBackupScheduleStop.Flags().Bool("kill-running", false, "Hentikan proses backup yang sedang berjalan untuk job ini (butuh --job)")

	cmdBackupScheduleStatus.Flags().String("job", "", "Nama job scheduler (kosong = semua job)")

	cmdBackupScheduleHistory.Flags().String("job", "", "Nama job scheduler (wajib)")
	_ = cmdBackupScheduleHistory.MarkFlagRequired("job")
	cmdBackupScheduleHistory.Flags().String("since", "", "Filter log mulai waktu tertentu (format journalctl, contoh: 'today', '2026-01-02 08:00:00')")
	cmdBackupScheduleHistory.Flags().String("until", "", "Filter log sampai waktu tertentu")
	cmdBackupScheduleHistory.Flags().Int("lines", 200, "Jumlah baris terakhir yang ditampilkan")
	cmdBackupScheduleHistory.Flags().Bool("follow", false, "Ikuti log secara realtime (journalctl -f)")
	cmdBackupScheduleHistory.Flags().Bool("timer", false, "Tampilkan log unit timer (bukan service)")

	cmdBackupScheduleRun.Flags().String("job", "", "Nama job scheduler (wajib)")
	_ = cmdBackupScheduleRun.MarkFlagRequired("job")
}
