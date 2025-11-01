package dbscan

import (
	"fmt"
	"os"
	"path/filepath"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/process"
	"sfDBTools/pkg/ui"
	"time"
)

// spawnDaemonProcess spawns new process sebagai background daemon
func (s *Service) spawnDaemonProcess(config types.ScanEntryConfig) error {
	// Get executable path
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("gagal mendapatkan executable path: %w", err)
	}

	// Scan ID hanya untuk tampilan
	scanID := fmt.Sprintf("scan_%s", time.Now().Format("20060102_150405"))
	logDir := filepath.Join("logs", "dbscan")
	// PID file path (fixed)
	pidFile := filepath.Join(logDir, "dbscan_background.pid")

	args := os.Args[1:] // pass through args
	env := []string{"SFDB_DAEMON_MODE=1"}

	pid, logFile, err := process.SpawnDaemon(executable, args, env, logDir, pidFile, "dbscan")
	if err != nil {
		return err
	}

	ui.PrintHeader("DATABASE SCANNING - BACKGROUND MODE")
	ui.PrintSuccess(fmt.Sprintf("Background process dimulai dengan PID: %d", pid))
	ui.PrintInfo(fmt.Sprintf("Scan ID: %s", ui.ColorText(scanID, ui.ColorCyan)))
	if logFile != "" {
		ui.PrintInfo(fmt.Sprintf("Log file: %s", ui.ColorText(logFile, ui.ColorCyan)))
		ui.PrintInfo(fmt.Sprintf("PID file: %s", ui.ColorText(pidFile, ui.ColorCyan)))
		ui.PrintInfo(fmt.Sprintf("Monitor dengan: tail -f %s", logFile))
	} else {
		ui.PrintInfo("Logs akan ditulis ke system logger")
	}
	ui.PrintInfo("Process berjalan di background. Gunakan 'ps aux | grep sfdbtools' untuk check status.")
	return nil
}
