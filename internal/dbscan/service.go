// File : internal/dbscan/service.go
// Deskripsi : Service utama implementation untuk database scanning operations
// Author : Hadiyatna Muflihun
// Tanggal : 15 Oktober 2025
// Last Modified : 17 Desember 2025

package dbscan

import (
	"context"
	"errors"
	"fmt"
	"os"

	"sfDBTools/internal/appconfig"
	"sfDBTools/internal/applog"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/dbscanhelper"
	"sfDBTools/pkg/errorlog"
	"sfDBTools/pkg/servicehelper"
	"sfDBTools/pkg/ui"
)

// Error definitions
var (
	ErrInvalidScanMode = errors.New("mode scan tidak valid")
)

// Service adalah service untuk database scanning
type Service struct {
	servicehelper.BaseService

	Config      *appconfig.Config
	Log         applog.Logger
	ErrorLog    *errorlog.ErrorLogger
	ScanOptions types.ScanOptions
}

// NewDBScanService membuat instance baru dari Service dengan proper dependency injection
func NewDBScanService(config *appconfig.Config, logger applog.Logger, opts types.ScanOptions) *Service {
	logDir := config.Log.Output.File.Dir
	if logDir == "" {
		logDir = "/var/log/sfDBTools"
	}

	return &Service{
		Config:      config,
		Log:         logger,
		ErrorLog:    errorlog.NewErrorLogger(logger, logDir, "dbscan"),
		ScanOptions: opts,
	}
}

// ExecuteScan adalah entry point utama untuk database scan
func (s *Service) ExecuteScan(config types.ScanEntryConfig) error {
	ctx := context.Background()
	s.ScanOptions.Mode = config.Mode
	s.ScanOptions.LocalScan = (config.Mode == "all-local")

	// Jika background mode, spawn sebagai daemon process atau jalankan background task
	if s.ScanOptions.Background {
		return s.handleBackgroundExecution(ctx, config)
	}

	// Setup connections (Foreground Mode)
	sourceClient, targetClient, dbFiltered, cleanup, err := s.setupScanConnections(ctx, config.HeaderTitle, config.ShowOptions)
	if err != nil {
		return err
	}
	defer cleanup()

	// Lakukan scanning dengan UI output
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

// handleBackgroundExecution menangani logika eksekusi background/daemon
func (s *Service) handleBackgroundExecution(ctx context.Context, config types.ScanEntryConfig) error {
	if s.ScanOptions.ProfileInfo.Path == "" {
		return fmt.Errorf("background mode memerlukan file konfigurasi database")
	}

	// Check jika sudah running dalam daemon mode (env flag set)
	if os.Getenv(consts.ENV_DAEMON_MODE) == "1" {
		return s.executeScanInBackground(ctx, config)
	}

	// Spawn new process sebagai daemon
	return dbscanhelper.SpawnScanDaemon(config)
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
	// Coba ambil dari source
	if sourceClient != nil {
		_ = sourceClient.DB().QueryRowContext(ctx, "SELECT @@datadir").Scan(&datadir)
	}
	// Fallback ke target jika source gagal/tidak ada
	if datadir == "" && targetClient != nil {
		_ = targetClient.DB().QueryRowContext(ctx, "SELECT @@datadir").Scan(&datadir)
	}
	
	if datadir == "" {
		return "", fmt.Errorf("tidak dapat menentukan datadir dari source maupun target")
	}
	return datadir, nil
}