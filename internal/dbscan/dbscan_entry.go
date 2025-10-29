package dbscan

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/process"
	"sfDBTools/pkg/ui"
	"time"
)

// ExecuteScanCommand adalah entry point untuk database scan
func (s *Service) ExecuteScanCommand(config types.ScanEntryConfig) error {
	ctx := context.Background()
	s.ScanOptions.Mode = config.Mode
	// Jika background mode, spawn sebagai daemon process
	if s.ScanOptions.Background {
		if s.ScanOptions.ProfileInfo.Path == "" {
			return fmt.Errorf("background mode memerlukan file konfigurasi database")
		}

		// Check jika sudah running dalam daemon mode
		if os.Getenv(consts.ENV_DAEMON_MODE) == "1" {
			// Sudah dalam daemon mode, jalankan actual work
			return s.ExecuteScanInBackground(ctx, config)
		}

		// Spawn new process sebagai daemon
		return s.spawnDaemonProcess(config)
	}

	// Setup connections
	sourceClient, targetClient, dbFiltered, cleanup, err := s.setupScanConnections(ctx, config.HeaderTitle, config.ShowOptions)
	if err != nil {
		return err
	}
	defer cleanup()

	// Lakukan scanning (foreground mode dengan UI output)
	result, err := s.ExecuteScan(ctx, sourceClient, targetClient, dbFiltered, false)
	if err != nil {
		s.Logger.Error(config.LogPrefix + " gagal: " + err.Error())
		return err
	}

	// Tampilkan hasil
	s.DisplayScanResult(result)

	// Print success message jika ada
	if config.SuccessMsg != "" {
		ui.PrintSuccess(config.SuccessMsg)
	}

	return nil
}

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

	pid, logFile, err := process.SpawnDaemon(executable, args, env, logDir, pidFile)
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

// setupScanConnections melakukan setup koneksi source dan target database
// Returns: sourceClient, targetClient, dbFiltered, cleanupFunc, error
func (s *Service) setupScanConnections(ctx context.Context, headerTitle string, showOptions bool) (*database.Client, *database.Client, []string, func(), error) {
	// Jika mode rescan, gunakan PrepareRescanSession
	if s.ScanOptions.Mode == "rescan" {
		sourceClient, targetClient, dbFiltered, err := s.PrepareRescanSession(ctx, headerTitle, showOptions)
		if err != nil {
			return nil, nil, nil, nil, err
		}

		// Force enable SaveToDB untuk rescan karena kita perlu update error_message
		s.ScanOptions.SaveToDB = true

		// Cleanup function untuk close semua connections
		cleanup := func() {
			if sourceClient != nil {
				sourceClient.Close()
			}
			if targetClient != nil {
				targetClient.Close()
			}
		}

		return sourceClient, targetClient, dbFiltered, cleanup, nil
	}

	// Setup session (koneksi database source) untuk mode normal
	sourceClient, dbFiltered, err := s.PrepareScanSession(ctx, headerTitle, showOptions)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	// Koneksi ke target database untuk menyimpan hasil scan
	var targetClient *database.Client
	if s.ScanOptions.SaveToDB {
		targetClient, err = s.ConnectToTargetDB(ctx)
		if err != nil {
			s.Logger.Warn("Gagal koneksi ke target database, hasil scan tidak akan disimpan: " + err.Error())
			s.ScanOptions.SaveToDB = false
		}
	}

	// Cleanup function untuk close semua connections
	cleanup := func() {
		if sourceClient != nil {
			sourceClient.Close()
		}
		if targetClient != nil {
			targetClient.Close()
		}
	}

	return sourceClient, targetClient, dbFiltered, cleanup, nil
}
