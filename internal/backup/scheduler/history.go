// File : internal/backup/scheduler/history.go
// Deskripsi : Helper untuk menampilkan history job scheduler (journalctl)
// Author : Hadiyatna Muflihun
// Tanggal : 2026-01-02
// Last Modified : 2026-01-05
package scheduler
import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	appdeps "sfDBTools/internal/cli/deps"
	"sfDBTools/internal/services/scheduler"
)

type HistoryOptions struct {
	Since  string
	Until  string
	Lines  int
	Follow bool
	Timer  bool
}

func History(ctx context.Context, deps *appdeps.Dependencies, jobName string, opt HistoryOptions) error {
	if err := schedulerutil.EnsureLinux(); err != nil {
		return err
	}
	jobName = strings.TrimSpace(jobName)
	if jobName == "" {
		return fmt.Errorf("job wajib diisi")
	}
	if opt.Lines <= 0 {
		opt.Lines = 200
	}

	unit := fmt.Sprintf("sfdbtools-backup@%s.service", jobName)
	if opt.Timer {
		unit = fmt.Sprintf("sfdbtools-backup@%s.timer", jobName)
	}

	args := []string{"-u", unit, "--no-pager"}
	if strings.TrimSpace(opt.Since) != "" {
		args = append(args, "--since", strings.TrimSpace(opt.Since))
	}
	if strings.TrimSpace(opt.Until) != "" {
		args = append(args, "--until", strings.TrimSpace(opt.Until))
	}
	if opt.Follow {
		args = append(args, "-f")
	} else {
		args = append(args, "-n", fmt.Sprintf("%d", opt.Lines))
	}

	cmd := exec.CommandContext(ctx, "journalctl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		// Jika tidak punya permission baca journal, berikan hint.
		if deps != nil && deps.Logger != nil {
			deps.Logger.Warn("Jika muncul error permission, coba jalankan dengan sudo atau pastikan user masuk group 'systemd-journal'.")
		}
		return fmt.Errorf("journalctl gagal: %w", err)
	}
	return nil
}
