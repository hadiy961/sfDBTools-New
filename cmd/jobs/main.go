// File : cmd/jobs/main.go
// Deskripsi : Command terpusat untuk monitoring job background & scheduler systemd
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 5 Januari 2026
package jobs

import (
	"errors"
	"fmt"
	"os"

	internaljobs "sfDBTools/internal/jobs"
	"sfDBTools/internal/services/scheduler"
	"sfDBTools/internal/ui/print"
	"sfDBTools/pkg/validation"

	"github.com/spf13/cobra"
)

var CmdJobs = &cobra.Command{
	Use:   "jobs",
	Short: "Monitoring terpusat untuk background jobs & scheduler",
	Long: `Monitoring terpusat untuk:
- Background job via systemd-run (backup/cleanup/dbscan)
- Scheduled job via systemd timer (backup schedule + cleanup schedule)

Default akan menampilkan daftar service dan timer yang relevan tanpa user perlu mengetik systemctl/journalctl manual.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		print.PrintAppHeader("SYSTEMD JOBS")
		if internaljobs.IsInteractiveAllowed() {
			scope, err := getScope(cmd)
			if err != nil {
				return err
			}
			return internaljobs.RunInteractive(cmd.Context(), scope, os.Geteuid() == 0)
		}
		scope, err := getScope(cmd)
		if err != nil {
			return err
		}
		return internaljobs.PrintListBody(cmd.Context(), scope, os.Geteuid() == 0)
	},
}

var cmdJobsList = &cobra.Command{
	Use:   "list",
	Short: "Tampilkan daftar unit service & timer",
	RunE: func(cmd *cobra.Command, args []string) error {
		print.PrintAppHeader("SYSTEMD JOBS")
		print.PrintSubHeader("List")
		scope, err := getScope(cmd)
		if err != nil {
			return err
		}
		// Jika user tidak set --scope dan mode interaktif allowed, tawarkan pilih scope (user/system/both/auto)
		if internaljobs.IsInteractiveAllowed() && !flagChanged(cmd, "scope") {
			return internaljobs.RunInteractiveList(cmd.Context(), scope, os.Geteuid() == 0)
		}
		return internaljobs.PrintListBody(cmd.Context(), scope, os.Geteuid() == 0)
	},
}

var cmdJobsStatus = &cobra.Command{
	Use:   "status [unit]",
	Short: "Tampilkan status detail unit",
	Args:  cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		print.PrintAppHeader("SYSTEMD JOBS")
		print.PrintSubHeader("Status")
		scope, err := getScope(cmd)
		if err != nil {
			return err
		}
		scopeSet := flagChanged(cmd, "scope")
		if len(args) == 0 {
			if internaljobs.IsInteractiveAllowed() {
				return internaljobs.RunInteractiveStatus(cmd.Context(), scope, scopeSet)
			}
			return fmt.Errorf("mode non-interaktif (--quiet): unit wajib diisi; contoh: sfdbtools jobs status <unit>: %w", validation.ErrNonInteractive)
		}
		out, used, err := internaljobs.Status(cmd.Context(), scope, args[0])
		if err != nil {
			return err
		}
		print.PrintInfo(fmt.Sprintf("Scope: %s", used))
		fmt.Print(out)
		return nil
	},
}

var cmdJobsLogs = &cobra.Command{
	Use:   "logs [unit]",
	Short: "Tampilkan log unit (journalctl)",
	Args:  cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		print.PrintAppHeader("SYSTEMD JOBS")
		print.PrintSubHeader("Logs")
		scope, err := getScope(cmd)
		if err != nil {
			return err
		}
		scopeSet := flagChanged(cmd, "scope")
		lines, _ := cmd.Flags().GetInt("lines")
		follow, _ := cmd.Flags().GetBool("follow")
		if len(args) == 0 {
			if internaljobs.IsInteractiveAllowed() {
				return internaljobs.RunInteractiveLogs(cmd.Context(), scope, scopeSet, lines, follow, cmd.Flags().Changed("lines"), cmd.Flags().Changed("follow"))
			}
			return fmt.Errorf("mode non-interaktif (--quiet): unit wajib diisi; contoh: sfdbtools jobs logs <unit>: %w", validation.ErrNonInteractive)
		}
		out, used, err := internaljobs.Logs(cmd.Context(), scope, args[0], lines, follow)
		if err != nil {
			return err
		}
		print.PrintInfo(fmt.Sprintf("Scope: %s", used))
		fmt.Print(out)
		return nil
	},
}

var cmdJobsStop = &cobra.Command{
	Use:   "stop [unit]",
	Short: "Stop unit",
	Args:  cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		print.PrintAppHeader("SYSTEMD JOBS")
		print.PrintSubHeader("Stop")
		scope, err := getScope(cmd)
		if err != nil {
			return err
		}
		scopeSet := flagChanged(cmd, "scope")
		if len(args) == 0 {
			if internaljobs.IsInteractiveAllowed() {
				return internaljobs.RunInteractiveStop(cmd.Context(), scope, scopeSet)
			}
			return fmt.Errorf("mode non-interaktif (--quiet): unit wajib diisi; contoh: sfdbtools jobs stop <unit>: %w", validation.ErrNonInteractive)
		}
		out, used, err := internaljobs.Stop(cmd.Context(), scope, args[0])
		if err != nil {
			return err
		}
		print.PrintInfo(fmt.Sprintf("Scope: %s", used))
		fmt.Print(out)
		print.PrintSuccess("Stop command dikirim")
		return nil
	},
}

var cmdJobsRemove = &cobra.Command{
	Use:   "remove [unit]",
	Short: "Hapus job (stop/disable + reset-failed + optional purge unit file)",
	Args:  cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		print.PrintAppHeader("SYSTEMD JOBS")
		print.PrintSubHeader("Remove")
		scope, err := getScope(cmd)
		if err != nil {
			return err
		}
		scopeSet := flagChanged(cmd, "scope")
		purge, _ := cmd.Flags().GetBool("purge")
		if len(args) == 0 {
			if internaljobs.IsInteractiveAllowed() {
				out, used, err := internaljobs.RunInteractiveRemove(cmd.Context(), scope, scopeSet, os.Geteuid() == 0, purge, cmd.Flags().Changed("purge"))
				if err != nil {
					if errors.Is(err, validation.ErrUserCancelled) {
						return nil
					}
					return err
				}
				_ = used
				if out != "" {
					fmt.Print(out)
				}
				print.PrintSuccess("Remove selesai")
				return nil
			}
			return fmt.Errorf("mode non-interaktif (--quiet): unit wajib diisi; contoh: sfdbtools jobs remove <unit>: %w", validation.ErrNonInteractive)
		}
		out, used, err := internaljobs.Remove(cmd.Context(), args[0], internaljobs.RemoveOptions{Scope: scope, Purge: purge})
		if err != nil {
			return err
		}
		print.PrintInfo(fmt.Sprintf("Scope: %s", used))
		if out != "" {
			fmt.Print(out)
		}
		print.PrintSuccess("Remove selesai")
		return nil
	},
}

func init() {
	CmdJobs.PersistentFlags().String("scope", "auto", "Scope unit: auto|user|system|both")

	cmdJobsLogs.Flags().Int("lines", 200, "Jumlah baris log")
	cmdJobsLogs.Flags().Bool("follow", false, "Ikuti log realtime")
	cmdJobsRemove.Flags().Bool("purge", false, "Hapus unit file dari /etc/systemd/system (hanya scope=system, butuh sudo)")

	CmdJobs.AddCommand(cmdJobsList)
	CmdJobs.AddCommand(cmdJobsStatus)
	CmdJobs.AddCommand(cmdJobsLogs)
	CmdJobs.AddCommand(cmdJobsStop)
	CmdJobs.AddCommand(cmdJobsRemove)
}

func getScope(cmd *cobra.Command) (scheduler.Scope, error) {
	val, _ := cmd.Flags().GetString("scope")
	s, err := scheduler.NormalizeScope(val)
	if err != nil {
		return "", err
	}
	return s, nil
}

func flagChanged(cmd *cobra.Command, name string) bool {
	if f := cmd.Flag(name); f != nil {
		return f.Changed
	}
	if f := cmd.Flags().Lookup(name); f != nil {
		return f.Changed
	}
	if f := cmd.InheritedFlags().Lookup(name); f != nil {
		return f.Changed
	}
	return false
}
