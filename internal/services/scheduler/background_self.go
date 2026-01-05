// File : internal/services/scheduler/background_self.go
// Deskripsi : Helper untuk menjalankan ulang proses saat ini via systemd-run (background)
// Author : Hadiyatna Muflihun
// Tanggal : 3 Januari 2026
// Last Modified : 5 Januari 2026
package scheduler

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"
)

type SpawnSelfOptions struct {
	UnitPrefix    string
	Mode          RunMode
	EnvFile       string
	WorkDir       string
	Collect       bool
	NoAskPass     bool
	WrapWithFlock bool
	FlockPath     string
}

func ensureDaemonArg(args []string) []string {
	for _, a := range args {
		if a == "--daemon" || strings.HasPrefix(a, "--daemon=") {
			return args
		}
	}
	return append(args, "--daemon")
}

// SpawnSelfInBackground menjalankan ulang proses saat ini sebagai transient systemd unit.
// Secara default akan menambahkan flag --daemon bila belum ada.
// Jika WrapWithFlock aktif, command akan dibungkus dengan `flock -x <lock> -- <executable> <args...>`.
func SpawnSelfInBackground(ctx context.Context, opts SpawnSelfOptions) (*SystemdRunResult, error) {
	if strings.TrimSpace(opts.UnitPrefix) == "" {
		return nil, fmt.Errorf("unit prefix tidak boleh kosong")
	}

	executable, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("gagal mendapatkan executable path: %w", err)
	}

	args := ensureDaemonArg(os.Args[1:])

	unit := fmt.Sprintf("%s-%s", opts.UnitPrefix, time.Now().Format("20060102-150405"))

	cmd := executable
	cmdArgs := args
	if opts.WrapWithFlock {
		if strings.TrimSpace(opts.FlockPath) == "" {
			return nil, fmt.Errorf("flock path tidak boleh kosong")
		}
		cmd = "flock"
		cmdArgs = append([]string{"-x", opts.FlockPath, "--", executable}, args...)
	}

	res, err := SystemdRun(ctx, SystemdRunOptions{
		UnitName:  unit,
		Mode:      opts.Mode,
		EnvFile:   opts.EnvFile,
		WorkDir:   opts.WorkDir,
		Collect:   opts.Collect,
		NoAskPass: opts.NoAskPass,
	}, cmd, cmdArgs...)
	if err != nil {
		return nil, err
	}
	return res, nil
}
