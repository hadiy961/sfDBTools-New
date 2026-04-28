package scheduler

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	appconfig "sfdbtools/internal/services/config"
	applog "sfdbtools/internal/services/log"
)

type Manager struct {
	Config  *appconfig.Config
	Log     applog.Logger
	Systemd *SystemdManager
}

func NewManager(cfg *appconfig.Config, log applog.Logger) *Manager {
	return &Manager{
		Config:  cfg,
		Log:     log,
		Systemd: NewSystemdManager(),
	}
}

type TemplateData struct {
	JobName     string
	ExecPath    string
	Mode        string
	OutputMode  string
	Ticket      string
	Profile     string
	IncludeFile string
	OutputDir   string
	OnCalendar  string
}

func (m *Manager) Sync(ctx context.Context) error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	systemdDir := "/etc/systemd/system"
	if _, err := os.Stat(systemdDir); os.IsNotExist(err) {
		return fmt.Errorf("systemd directory %s does not exist, scheduling might not be supported on this system", systemdDir)
	}

	serviceTmpl, err := template.New("service").Parse(serviceTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse service template: %w", err)
	}

	timerTmpl, err := template.New("timer").Parse(timerTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse timer template: %w", err)
	}

	m.Log.Info("Syncing scheduler jobs to systemd...")

	// To keep track of generated timers for cleanup if a job is removed (future enhancement)
	// For now, we'll just generate/enable what's in the config.

	for _, job := range m.Config.Backup.Scheduler.Jobs {
		if job.Name == "" {
			m.Log.Warn("Found job with empty name, skipping...")
			continue
		}

		serviceName := fmt.Sprintf("sfdbtools-backup-%s.service", job.Name)
		timerName := fmt.Sprintf("sfdbtools-backup-%s.timer", job.Name)

		servicePath := filepath.Join(systemdDir, serviceName)
		timerPath := filepath.Join(systemdDir, timerName)

		if !job.Enabled {
			// Disable timer if it exists
			m.Log.Infof("Job '%s' is disabled, ensuring timer is stopped...", job.Name)
			m.Systemd.DisableAndStopTimer(ctx, timerName)
			continue
		}

		onCalendar, err := ConvertCronToSystemd(job.Schedule)
		if err != nil {
			m.Log.Warnf("Failed to parse schedule for job '%s': %v", job.Name, err)
			continue
		}

		data := TemplateData{
			JobName:     job.Name,
			ExecPath:    execPath,
			Mode:        job.Mode,
			OutputMode:  job.OutputMode,
			Ticket:      job.Ticket,
			Profile:     job.Profile,
			IncludeFile: job.IncludeFile,
			OutputDir:   job.Output.BaseDirectory,
			OnCalendar:  onCalendar,
		}

		// Generate Service File
		var svcBuf bytes.Buffer
		if err := serviceTmpl.Execute(&svcBuf, data); err != nil {
			m.Log.Errorf("Failed to execute service template for job '%s': %v", job.Name, err)
			continue
		}
		if err := os.WriteFile(servicePath, svcBuf.Bytes(), 0644); err != nil {
			m.Log.Errorf("Failed to write service file for job '%s': %v", job.Name, err)
			continue
		}

		// Generate Timer File
		var tmrBuf bytes.Buffer
		if err := timerTmpl.Execute(&tmrBuf, data); err != nil {
			m.Log.Errorf("Failed to execute timer template for job '%s': %v", job.Name, err)
			continue
		}
		if err := os.WriteFile(timerPath, tmrBuf.Bytes(), 0644); err != nil {
			m.Log.Errorf("Failed to write timer file for job '%s': %v", job.Name, err)
			continue
		}

		m.Log.Infof("Generated systemd units for job '%s'", job.Name)
	}

	m.Log.Info("Reloading systemd daemon...")
	if err := m.Systemd.DaemonReload(ctx); err != nil {
		m.Log.Warnf("Failed to reload systemd daemon: %v", err)
	}

	// Enable and start timers
	for _, job := range m.Config.Backup.Scheduler.Jobs {
		if !job.Enabled {
			continue
		}
		timerName := fmt.Sprintf("sfdbtools-backup-%s.timer", job.Name)
		m.Log.Infof("Enabling and starting timer for job '%s'...", job.Name)
		if err := m.Systemd.EnableAndStartTimer(ctx, timerName); err != nil {
			m.Log.Errorf("Failed to enable/start timer for job '%s': %v", job.Name, err)
		}
	}

	m.Log.Info("Scheduler sync complete.")
	return nil
}
