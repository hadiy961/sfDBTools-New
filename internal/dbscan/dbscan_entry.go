package dbscan

import (
	"context"
	"fmt"
	"os"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/ui"
)

// ExecuteScanCommand adalah entry point untuk database scan
func (s *Service) ExecuteScanCommand(config types.ScanEntryConfig) error {
	ctx := context.Background()
	s.ScanOptions.Mode = config.Mode
	s.ScanOptions.LocalScan = (config.Mode == "all-local")
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
