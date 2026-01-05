// File : internal/services/scheduler/systemd.go
// Deskripsi : Helper systemd (systemctl) untuk scheduler
// Author : Hadiyatna Muflihun
// Tanggal : 2 Januari 2026
// Last Modified : 5 Januari 2026
package schedulerutil

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func EnsureLinux() error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("scheduler systemd hanya didukung di Linux")
	}
	return nil
}

func Systemctl(ctx context.Context, args ...string) error {
	cmd := exec.CommandContext(ctx, "systemctl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("systemctl %s gagal: %w", strings.Join(args, " "), err)
	}
	return nil
}

func SystemctlOutput(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "systemctl", args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}
