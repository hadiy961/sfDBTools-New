// File : internal/app/dbscan/service.go
// Deskripsi : Service utama implementation untuk database scanning operations
// Author : Hadiyatna Muflihun
// Tanggal : 15 Oktober 2025
// Last Modified : 5 Januari 2026
package dbscan

import (
	"context"
	"errors"
	"fmt"

	"sfdbtools/internal/app/dbscan/helpers"
	dbscanmodel "sfdbtools/internal/app/dbscan/model"
	appconfig "sfdbtools/internal/services/config"
	applog "sfdbtools/internal/services/log"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/database"
	"sfdbtools/internal/shared/errorlog"
	"sfdbtools/internal/shared/servicehelper"
	"sfdbtools/internal/ui/print"
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
	ScanOptions dbscanmodel.ScanOptions
}

// NewDBScanService membuat instance baru dari Service dengan proper dependency injection
func NewDBScanService(config *appconfig.Config, logger applog.Logger, opts dbscanmodel.ScanOptions) *Service {
	logDir := config.Log.Output.File.Dir
	if logDir == "" {
		logDir = consts.DefaultLogDir
	}

	return &Service{
		Config:      config,
		Log:         logger,
		ErrorLog:    errorlog.NewErrorLogger(logger, logDir, "dbscan"),
		ScanOptions: opts,
	}
}

// ExecuteScan adalah entry point utama untuk database scan
func (s *Service) ExecuteScan(config dbscanmodel.ScanEntryConfig) error {
	ctx := context.Background()
	s.ScanOptions.Mode = config.Mode
	s.ScanOptions.LocalScan = (config.Mode == "all-local")

	// Setup connections
	sourceClient, dbFiltered, cleanup, err := s.setupScanConnections(ctx, config.HeaderTitle, config.ShowOptions)
	if err != nil {
		return err
	}
	defer cleanup()

	// Lakukan scanning dengan UI output
	result, detailsMap, err := s.executeScanWithClients(ctx, sourceClient, dbFiltered)
	if err != nil {
		s.Log.Error(config.LogPrefix + " gagal: " + err.Error())
		return err
	}

	// Tampilkan hasil
	if s.ScanOptions.DisplayResults {
		helpers.DisplayScanResult(result)
		if len(detailsMap) > 0 {
			helpers.DisplayDetailResults(detailsMap)
		}
	}

	// Print success message jika ada
	if config.SuccessMsg != "" {
		print.PrintSuccess(config.SuccessMsg)
	}

	return nil
}

// executeScanWithClients melakukan scanning dengan koneksi yang sudah tersedia
func (s *Service) executeScanWithClients(
	ctx context.Context,
	sourceClient *database.Client,
	dbNames []string,
) (*dbscanmodel.ScanResult, map[string]dbscanmodel.DatabaseDetailInfo, error) {
	print.PrintSubHeader("Memulai Proses Scanning Database")

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
		datadir, err := s.getDataDir(ctx, sourceClient)
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
	opts := helpers.ScanExecutorOptions{
		LocalScan:     s.ScanOptions.LocalScan,
		DisplayResult: s.ScanOptions.DisplayResults,
		Logger:        s.Log,
		LocalSizes:    localSizes,
	}

	result, detailsMap, err := helpers.ExecuteScanWithSave(
		ctx, sourceClient, dbNames,
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
func (s *Service) getDataDir(ctx context.Context, sourceClient *database.Client) (string, error) {
	var datadir string
	if sourceClient != nil {
		_ = sourceClient.DB().QueryRowContext(ctx, "SELECT @@datadir").Scan(&datadir)
	}

	if datadir == "" {
		return "", fmt.Errorf("tidak dapat menentukan datadir dari source maupun target")
	}
	return datadir, nil
}
