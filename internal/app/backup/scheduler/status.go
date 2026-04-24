package scheduler

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"sfdbtools/internal/ui/print"
	"sfdbtools/internal/ui/table"
)

// getTimerProperty queries a specific property from a systemd timer using systemctl show.
func getTimerProperty(ctx context.Context, timerName, property string) string {
	cmd := exec.CommandContext(ctx, "systemctl", "show", timerName, "--property="+property, "--value")
	out, err := cmd.Output()
	if err != nil {
		return "-"
	}
	val := strings.TrimSpace(string(out))
	if val == "" || val == "0" || val == "n/a" {
		return "-"
	}
	return val
}

// getServiceProperty queries a specific property from a systemd service using systemctl show.
func getServiceProperty(ctx context.Context, serviceName, property string) string {
	cmd := exec.CommandContext(ctx, "systemctl", "show", serviceName, "--property="+property, "--value")
	out, err := cmd.Output()
	if err != nil {
		return "-"
	}
	val := strings.TrimSpace(string(out))
	if val == "" || val == "0" || val == "n/a" {
		return "-"
	}
	return val
}

// ShowStatus displays the status of all configured backup jobs
func (m *Manager) ShowStatus(ctx context.Context) error {
	if len(m.Config.Backup.Scheduler.Jobs) == 0 {
		m.Log.Info("Tidak ada konfigurasi job scheduler di config.yaml.")
		return nil
	}

	headers := []string{"Job Name", "Enabled", "Schedule (Cron)", "Timer Status", "Next Run", "Last Trigger", "Last Result"}
	var rows [][]string

	for _, job := range m.Config.Backup.Scheduler.Jobs {
		if job.Name == "" {
			continue
		}

		enabled := "No"
		if job.Enabled {
			enabled = "Yes"
		}

		timerName := fmt.Sprintf("sfdbtools-backup-%s.timer", job.Name)
		serviceName := fmt.Sprintf("sfdbtools-backup-%s.service", job.Name)

		// Timer active status
		statusCmd := exec.CommandContext(ctx, "systemctl", "is-active", timerName)
		statusOut, _ := statusCmd.CombinedOutput()
		sysStatus := strings.TrimSpace(string(statusOut))
		switch sysStatus {
		case "active":
			sysStatus = "Active ✓"
		case "inactive":
			sysStatus = "Inactive"
		case "failed":
			sysStatus = "Failed ✗"
		default:
			sysStatus = "Not Found"
		}

		// Next scheduled run (from timer)
		nextRun := getTimerProperty(ctx, timerName, "NextElapseUSecRealtime")
		if nextRun == "-" {
			nextRun = getTimerProperty(ctx, timerName, "NextElapseUSec")
		}

		// Last time timer triggered
		lastTrigger := getTimerProperty(ctx, timerName, "LastTriggerUSec")

		// Last service execution result
		lastResult := getServiceProperty(ctx, serviceName, "Result")
		switch lastResult {
		case "success":
			lastResult = "success ✓"
		case "exit-code":
			lastResult = "failed ✗"
		case "-":
			lastResult = "never run"
		}

		rows = append(rows, []string{
			job.Name,
			enabled,
			job.Schedule,
			sysStatus,
			nextRun,
			lastTrigger,
			lastResult,
		})
	}

	print.PrintSubHeader("Status Scheduler Backup (Systemd Timers)")
	table.Render(headers, rows)

	// Print actionable log commands per job
	fmt.Println()
	fmt.Println("  📖 Cara memantau & memverifikasi job:")
	fmt.Println()
	for _, job := range m.Config.Backup.Scheduler.Jobs {
		if job.Name == "" {
			continue
		}
		serviceName := fmt.Sprintf("sfdbtools-backup-%s.service", job.Name)
		fmt.Printf("  [%s]\n", job.Name)
		fmt.Printf("    Lihat log terakhir  : journalctl -u %s -n 50 --no-pager\n", serviceName)
		fmt.Printf("    Pantau real-time    : journalctl -f -u %s\n", serviceName)
		fmt.Printf("    Cek hasil eksekusi  : systemctl status %s\n", serviceName)
		fmt.Printf("    Jalankan manual     : systemctl start %s\n", serviceName)
		fmt.Println()
	}

	fmt.Println("  ⏱  Semua timer aktif:")
	fmt.Println("    systemctl list-timers | grep sfdbtools")
	fmt.Println()

	return nil
}
