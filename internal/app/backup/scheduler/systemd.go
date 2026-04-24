package scheduler

import (
	"context"
	"os/exec"
)

// SystemdManager handles systemd daemon reloads and service toggles.
type SystemdManager struct{}

func NewSystemdManager() *SystemdManager {
	return &SystemdManager{}
}

func (s *SystemdManager) DaemonReload(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "systemctl", "daemon-reload")
	return cmd.Run()
}

func (s *SystemdManager) EnableAndStartTimer(ctx context.Context, timerName string) error {
	cmd := exec.CommandContext(ctx, "systemctl", "enable", "--now", timerName)
	return cmd.Run()
}

func (s *SystemdManager) DisableAndStopTimer(ctx context.Context, timerName string) error {
	cmd := exec.CommandContext(ctx, "systemctl", "disable", "--now", timerName)
	return cmd.Run()
}
