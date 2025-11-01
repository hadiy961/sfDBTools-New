package process

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sfDBTools/pkg/fsops"
	"syscall"
	"time"
)

// CheckAndReadPID untuk memeriksa apakah pidfile ada dan proses masih hidup; mengembalikan pid dan true jika berjalan.
func CheckAndReadPID(pidFile string) (int, bool) {
	if pidFile == "" {
		return 0, false
	}
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return 0, false
	}
	var pid int
	if _, err := fmt.Sscanf(string(data), "%d", &pid); err != nil {
		_ = os.Remove(pidFile)
		return 0, false
	}
	if err := syscall.Kill(pid, 0); err == nil {
		return pid, true
	}
	// jika tidak berjalan, hapus pidfile yang usang
	_ = os.Remove(pidFile)
	return 0, false
}

// SpawnDaemon starts a background process with stdout/stderr redirected to a log file under logDir.
// If pidFile is provided, it will be written after start; if an active pid is found, returns error.
func SpawnDaemon(executable string, args []string, env []string, logDir string, pidFile string, mode string) (pid int, logFile string, err error) {
	if _, running := CheckAndReadPID(pidFile); running {
		return 0, "", fmt.Errorf("background process sudah berjalan (pidfile=%s)", pidFile)
	}

	if _, err := fsops.EnsureDir(logDir); err != nil {
		// fallback: no log dir
		logDir = ""
	}

	var out *os.File
	if logDir != "" {
		logFile = filepath.Join(logDir, fmt.Sprintf("%s_%s.log", mode, timestamp()))
		f, ferr := os.Create(logFile)
		if ferr == nil {
			out = f
			defer func() {
				if out != nil {
					out.Close()
				}
			}()
		} else {
			logFile = ""
		}
	}

	cmd := exec.Command(executable, args...)
	if len(env) > 0 {
		cmd.Env = append(os.Environ(), env...)
	}
	if out != nil {
		cmd.Stdout = out
		cmd.Stderr = out
	}
	if err := cmd.Start(); err != nil {
		return 0, logFile, fmt.Errorf("gagal memulai background process: %w", err)
	}
	if pidFile != "" {
		_ = os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", cmd.Process.Pid)), 0644)
	}
	return cmd.Process.Pid, logFile, nil
}

func timestamp() string { return time.Now().Format("20060102_150405") }
