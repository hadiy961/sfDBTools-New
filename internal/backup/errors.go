// File : internal/backup/errors.go
// Deskripsi : Error handling untuk backup operations
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-05
// Last Modified : 2025-12-05

package backup

import (
	"fmt"
	"sfDBTools/internal/applog"
	"sfDBTools/internal/cleanup"
	"sfDBTools/pkg/errorlog"
	"sfDBTools/pkg/ui"
)

// BackupErrorHandler menyimpan dependencies untuk error handling
type BackupErrorHandler struct {
	Logger   applog.Logger
	ErrorLog *errorlog.ErrorLogger
	ShowUI   bool
}

// NewBackupErrorHandler membuat error handler baru
func NewBackupErrorHandler(logger applog.Logger, errorLog *errorlog.ErrorLogger, showUI bool) *BackupErrorHandler {
	return &BackupErrorHandler{
		Logger:   logger,
		ErrorLog: errorLog,
		ShowUI:   showUI,
	}
}

// logError mencatat error ke log file
func (h *BackupErrorHandler) logError(logMetadata map[string]interface{}, stderrOutput string, err error) {
	if h.ErrorLog != nil {
		h.ErrorLog.LogWithOutput(logMetadata, stderrOutput, err)
	}
}

// HandleDatabaseBackupError menangani error untuk single database backup
func (h *BackupErrorHandler) HandleDatabaseBackupError(
	filePath string,
	dbName string,
	err error,
	stderrOutput string,
	logMetadata map[string]interface{},
) string {
	// Cleanup dan log
	cleanup.CleanupFailedBackup(filePath, h.Logger)
	h.logError(logMetadata, stderrOutput, err)

	// Build error message
	errorMsg := fmt.Sprintf("gagal backup database %s: %v", dbName, err)
	if stderrOutput != "" {
		errorMsg = fmt.Sprintf("%s\nDetail: %s", errorMsg, stderrOutput)
	}

	// UI feedback
	if h.ShowUI {
		ui.PrintError(fmt.Sprintf("✗ Database %s gagal di-backup", dbName))
	}

	h.Logger.Error(errorMsg)
	return errorMsg
}

// HandleCombinedBackupError menangani error untuk combined backup
func (h *BackupErrorHandler) HandleCombinedBackupError(
	filePath string,
	err error,
	stderrOutput string,
	logMetadata map[string]interface{},
) error {
	// Cleanup dan log
	cleanup.CleanupFailedBackup(filePath, h.Logger)
	h.logError(logMetadata, stderrOutput, err)

	// Show UI error details
	if h.ShowUI {
		ui.PrintHeader("ERROR : Mysqldump Gagal dijalankan")
		if stderrOutput != "" {
			ui.PrintSubHeader("Detail Error dari mysqldump:")
			ui.PrintError(stderrOutput)
		}
	}

	h.Logger.Errorf("Gagal menjalankan mysqldump: %v", err)
	return fmt.Errorf("gagal menjalankan mysqldump: %w", err)
}
