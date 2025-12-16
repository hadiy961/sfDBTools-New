// File : internal/dbscan/dbscan_main.go
// Deskripsi : Service utama untuk database scanning dengan BaseService pattern
// Author : Hadiyatna Muflihun
// Tanggal : 15 Oktober 2025
// Last Modified : 16 Desember 2025

package dbscan

import (
	"errors"

	"sfDBTools/internal/appconfig"
	"sfDBTools/internal/applog"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/errorlog"
	"sfDBTools/pkg/servicehelper"
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
