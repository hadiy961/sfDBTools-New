package verify

import (
	"sfdbtools/internal/app/backup/model/types_backup"
	"sfdbtools/internal/services/config"
	applog "sfdbtools/internal/services/log"
)

// PostBackupCheck menjalankan verifikasi ringan setelah backup selesai
// Default: checksum + size check saja (tanpa header/footer)
func PostBackupCheck(filePath string, cfg appconfig.VerificationConfig, logger applog.Logger) *types_backup.VerificationResult {
	minSize, err := ParseMinFileSize(cfg.MinFileSize)
	if err != nil {
		logger.Warnf("Invalid min_file_size config '%s', defaulting to 0: %v", cfg.MinFileSize, err)
		minSize = 0
	}

	opts := CheckOptions{
		Checksum:     true,
		ChecksumAlgo: cfg.ChecksumAlgorithm,
		HeaderFooter: cfg.HeaderFooterCheck,
		SizeCheck:    true,
		MinFileSize:  minSize,
	}

	result, err := Check(filePath, opts, logger)
	if err != nil {
		logger.Errorf("Post-backup verification error: %v", err)
		// Return partial or failed result if check failed entirely (e.g., file unreadable)
		if result == nil {
			result = &types_backup.VerificationResult{
				VerifyStatus:  "failed",
				FailureReason: err.Error(),
			}
		}
	}

	return result
}
