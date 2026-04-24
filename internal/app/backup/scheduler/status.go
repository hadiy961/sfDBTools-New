package scheduler

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"sfdbtools/internal/ui/print"
	"sfdbtools/internal/ui/table"
)

// ShowStatus displays the status of all configured backup jobs
func (m *Manager) ShowStatus(ctx context.Context) error {
	if len(m.Config.Backup.Scheduler.Jobs) == 0 {
		m.Log.Info("Tidak ada konfigurasi job scheduler di config.yaml.")
		return nil
	}

	headers := []string{"Job Name", "Enabled (YAML)", "Schedule (Cron)", "Systemd Timer Status"}
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
		
		// Check systemd status
		statusCmd := exec.CommandContext(ctx, "systemctl", "is-active", timerName)
		output, _ := statusCmd.CombinedOutput()
		sysStatus := strings.TrimSpace(string(output))
		
		if sysStatus == "active" {
			sysStatus = "Active"
		} else if sysStatus == "inactive" {
			sysStatus = "Inactive"
		}

		rows = append(rows, []string{
			job.Name,
			enabled,
			job.Schedule,
			sysStatus,
		})
	}

	print.PrintSubHeader("Status Scheduler Backup (Systemd Timers)")
	table.Render(headers, rows)
	
	m.Log.Info("Untuk melihat waktu eksekusi (Next Run), jalankan: systemctl list-timers | grep sfdbtools")
	
	return nil
}
