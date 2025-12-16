// File : internal/dbscan/service.go
// Deskripsi : Service implementation untuk database scanning operations
// Author : Hadiyatna Muflihun
// Tanggal : 16 Desember 2025
// Last Modified : 16 Desember 2025

package dbscan

import (
	"context"
	"fmt"
	"os"

	"sfDBTools/internal/types"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/dbscanhelper"
	"sfDBTools/pkg/ui"
)

// ExecuteScan adalah entry point untuk database scan
func (s *Service) ExecuteScan(config types.ScanEntryConfig) error {
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
			return s.executeScanInBackground(ctx, config)
		}

		// Spawn new process sebagai daemon
		return dbscanhelper.SpawnScanDaemon(config)
	}

	// Setup connections
	sourceClient, targetClient, dbFiltered, cleanup, err := s.setupScanConnections(ctx, config.HeaderTitle, config.ShowOptions)
	if err != nil {
		return err
	}
	defer cleanup()

	// Lakukan scanning (foreground mode dengan UI output)
	result, detailsMap, err := s.executeScanWithClients(ctx, sourceClient, targetClient, dbFiltered, false)
	if err != nil {
		s.Log.Error(config.LogPrefix + " gagal: " + err.Error())
		return err
	}

	// Tampilkan hasil
	if s.ScanOptions.DisplayResults {
		dbscanhelper.DisplayScanResult(result)
		if len(detailsMap) > 0 {
			dbscanhelper.DisplayDetailResults(detailsMap)
		}
	}

	// Print success message jika ada
	if config.SuccessMsg != "" {
		ui.PrintSuccess(config.SuccessMsg)
	}

	return nil
}

// executeScanWithClients melakukan scanning dengan koneksi yang sudah tersedia
func (s *Service) executeScanWithClients(
	ctx context.Context,
	sourceClient *database.Client,
	targetClient *database.Client,
	dbNames []string,
	isBackground bool,
) (*types.ScanResult, map[string]types.DatabaseDetailInfo, error) {
	if !isBackground {
		ui.PrintSubHeader("Memulai Proses Scanning Database")
	}

	// Ambil server info
	serverHost := s.ScanOptions.ProfileInfo.DBInfo.Host
	serverPort := s.ScanOptions.ProfileInfo.DBInfo.Port

	if err := sourceClient.DB().QueryRowContext(ctx, "SELECT @@hostname, @@port").
		Scan(&serverHost, &serverPort); err != nil {
		s.Log.Warnf("Gagal mendapatkan server info: %v", err)
	}

	// Siapkan local sizes jika local scan mode
	var localSizes map[string]int64
	if s.ScanOptions.LocalScan {
		s.Log.Info("Mode Local Scan diaktifkan: ukuran database diambil dari datadir")
		datadir, err := s.getDataDir(ctx, sourceClient, targetClient)
		if err != nil {
			return nil, nil, err
		}

		sizes, err := s.LocalScanSizes(ctx, datadir, dbNames)
		if err != nil {
			return nil, nil, fmt.Errorf("gagal melakukan local size scan: %w", err)
		}
		localSizes = sizes
		s.Log.Infof("Local size scan selesai untuk %d database.", len(localSizes))
	}

	// Execute scan menggunakan helper
	opts := dbscanhelper.ScanExecutorOptions{
		SaveToDB:      s.ScanOptions.SaveToDB,
		LocalScan:     s.ScanOptions.LocalScan,
		DisplayResult: s.ScanOptions.DisplayResults,
		IsBackground:  isBackground,
		Logger:        s.Log,
		LocalSizes:    localSizes,
	}

	result, detailsMap, err := dbscanhelper.ExecuteScanWithSave(
		ctx, sourceClient, targetClient, dbNames,
		serverHost, serverPort, opts,
	)

	if err != nil && s.ErrorLog != nil {
		logFile := s.ErrorLog.Log(map[string]interface{}{
			"type": "scan_execution",
			"mode": s.ScanOptions.Mode,
		}, err)
		if logFile != "" {
			s.Log.Infof("ℹ Error details tersimpan di: %s", logFile)
		}
	}

	return result, detailsMap, err
}

// getDataDir mendapatkan datadir dari source atau target client
func (s *Service) getDataDir(ctx context.Context, sourceClient, targetClient *database.Client) (string, error) {
	var datadir string
	if sourceClient != nil {
		if err := sourceClient.DB().QueryRowContext(ctx, "SELECT @@datadir").Scan(&datadir); err != nil {
			s.Log.Warnf("Gagal mendapatkan datadir dari source: %v", err)
		}
	}
	if datadir == "" && targetClient != nil {
		if err := targetClient.DB().QueryRowContext(ctx, "SELECT @@datadir").Scan(&datadir); err != nil {
			s.Log.Warnf("Gagal mendapatkan datadir dari target: %v", err)
		}
	}
	if datadir == "" {
		return "", fmt.Errorf("tidak dapat menentukan datadir dari source maupun target")
	}
	return datadir, nil
}
