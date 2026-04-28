package verify

import (
	"sfdbtools/internal/app/backup/model/types_backup"
	applog "sfdbtools/internal/services/log"
)

// Wrapper functions for backwards compatibility and ease of use

// Check menjalankan verifikasi pada satu file menggunakan pipeline engine default
func Check(filePath string, opts CheckOptions, logger applog.Logger) (*types_backup.VerificationResult, error) {
	engine := NewEngine()
	return engine.Check(filePath, opts, logger)
}

// CheckBatch menjalankan verifikasi pada batch files menggunakan pipeline engine default
func CheckBatch(dirPath string, opts CheckOptions, logger applog.Logger) (map[string]*types_backup.VerificationResult, error) {
	engine := NewEngine()
	return engine.CheckBatch(dirPath, opts, logger)
}

// CheckFromMetadata menjalankan verifikasi dari info metadata
func CheckFromMetadata(metaPath string, opts CheckOptions, logger applog.Logger) (*types_backup.VerificationResult, error) {
	engine := NewEngine()
	return engine.CheckFromMetadata(metaPath, opts, logger)
}
