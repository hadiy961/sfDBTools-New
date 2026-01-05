// File : internal/services/scheduler/systemd_run.go
// Deskripsi : Helper systemd-run untuk menjalankan perintah sebagai transient unit
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
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

type RunMode int

const (
	RunModeAuto RunMode = iota
	RunModeSystem
	RunModeUser
)

type SystemdRunOptions struct {
	UnitName  string
	Mode      RunMode
	EnvFile   string
	WorkDir   string
	Collect   bool
	NoAskPass bool
}

type SystemdRunResult struct {
	UnitName string
	Mode     RunMode
	Output   string
}

func SystemdRun(ctx context.Context, opts SystemdRunOptions, command string, args ...string) (*SystemdRunResult, error) {
	if runtime.GOOS != "linux" {
		return nil, fmt.Errorf("systemd hanya didukung di Linux")
	}
	if strings.TrimSpace(opts.UnitName) == "" {
		return nil, fmt.Errorf("unit name tidak boleh kosong")
	}
	if strings.TrimSpace(command) == "" {
		return nil, fmt.Errorf("command tidak boleh kosong")
	}

	mode := opts.Mode
	if mode == RunModeAuto {
		if os.Geteuid() == 0 {
			mode = RunModeSystem
		} else {
			mode = RunModeUser
		}
	}

	runArgs := []string{"--unit", opts.UnitName}
	if mode == RunModeUser {
		runArgs = append(runArgs, "--user")
	}
	if opts.Collect {
		runArgs = append(runArgs, "--collect")
	}
	if opts.NoAskPass {
		runArgs = append(runArgs, "--no-ask-password")
	}
	if strings.TrimSpace(opts.WorkDir) != "" {
		runArgs = append(runArgs, fmt.Sprintf("--property=WorkingDirectory=%s", opts.WorkDir))
	}
	if strings.TrimSpace(opts.EnvFile) != "" {
		runArgs = append(runArgs, fmt.Sprintf("--property=EnvironmentFile=-%s", opts.EnvFile))
	}

	runArgs = append(runArgs, "--", command)
	runArgs = append(runArgs, args...)

	cmd := exec.CommandContext(ctx, "systemd-run", runArgs...)
	out, err := cmd.CombinedOutput()
	res := &SystemdRunResult{UnitName: opts.UnitName, Mode: mode, Output: string(out)}
	if err != nil {
		return res, fmt.Errorf("systemd-run gagal: %w (output: %s)", err, strings.TrimSpace(string(out)))
	}
	return res, nil
}
