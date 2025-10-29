// File : internal/dbscan/dbscan_main.go
// Deskripsi : Service utama untuk database scanning
// Author : Hadiyatna Muflihun
// Tanggal : 15 Oktober 2025
// Last Modified : 15 Oktober 2025

package dbscan

import (
	"sfDBTools/internal/appconfig"
	"sfDBTools/internal/applog"
	"sfDBTools/internal/types"
)

// Service adalah service untuk database scanning
type Service struct {
	Logger      applog.Logger
	Config      *appconfig.Config
	ScanOptions types.ScanOptions
}

// NewService membuat instance baru dari Service
func NewDBScanService(logger applog.Logger, config *appconfig.Config) *Service {
	return &Service{
		Logger: logger,
		Config: config,
	}
}

// SetScanOptions mengatur opsi scan
func (s *Service) SetScanOptions(opts types.ScanOptions) {
	s.ScanOptions = opts
}
