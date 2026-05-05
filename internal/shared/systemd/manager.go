package systemd

import (
	"os"
	"os/exec"
	"text/template"
)

const serviceTemplate = `[Unit]
Description=SFDBTools Auto Sync Service
After=network.target

[Service]
Type=oneshot
ExecStart={{.ExecPath}} sync --auto
User={{.User}}
Group={{.Group}}

[Install]
WantedBy=multi-user.target
`

const timerTemplate = `[Unit]
Description=SFDBTools Auto Sync Timer

[Timer]
OnBootSec=5min
OnUnitActiveSec={{.Interval}}min
Unit=sfdbtools-sync.service

[Install]
WantedBy=timers.target
`

type Manager struct {
	execPath string
	user     string
	group    string
}

func NewManager() (*Manager, error) {
	execPath, err := os.Executable()
	if err != nil {
		return nil, err
	}
	return &Manager{
		execPath: execPath,
		user:     "root", // Default to root as per requirements (/etc/sfDBTools)
		group:    "root",
	}, nil
}

func (m *Manager) InstallTimer(intervalMin int) error {
	if intervalMin <= 0 {
		intervalMin = 15
	}

	// 1. Create Service File
	servicePath := "/etc/systemd/system/sfdbtools-sync.service"
	if err := m.writeTemplate(servicePath, serviceTemplate, map[string]any{
		"ExecPath": m.execPath,
		"User":     m.user,
		"Group":    m.group,
	}); err != nil {
		return err
	}

	// 2. Create Timer File
	timerPath := "/etc/systemd/system/sfdbtools-sync.timer"
	if err := m.writeTemplate(timerPath, timerTemplate, map[string]any{
		"Interval": intervalMin,
	}); err != nil {
		return err
	}

	// 3. Reload, Enable, and Start
	exec.Command("systemctl", "daemon-reload").Run()
	exec.Command("systemctl", "enable", "sfdbtools-sync.timer").Run()
	return exec.Command("systemctl", "restart", "sfdbtools-sync.timer").Run()
}

func (m *Manager) RemoveTimer() error {
	exec.Command("systemctl", "stop", "sfdbtools-sync.timer").Run()
	exec.Command("systemctl", "disable", "sfdbtools-sync.timer").Run()
	
	os.Remove("/etc/systemd/system/sfdbtools-sync.timer")
	os.Remove("/etc/systemd/system/sfdbtools-sync.service")
	
	return exec.Command("systemctl", "daemon-reload").Run()
}

func (m *Manager) writeTemplate(path, tmplStr string, data any) error {
	tmpl, err := template.New("systemd").Parse(tmplStr)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.Execute(f, data)
}
