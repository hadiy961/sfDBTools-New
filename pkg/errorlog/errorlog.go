// File : pkg/errorlog/errorlog.go
// Deskripsi : Error logging utility untuk semua fitur (restore, backup, dbscan)
// Author : Hadiyatna Muflihun
// Tanggal : 2025-11-10
// Last Modified : 2025-11-10

package errorlog

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"sfDBTools/internal/applog"
)

// ErrorLogger mencatat error ke file log terpisah untuk setiap fitur
type ErrorLogger struct {
	Logger  applog.Logger
	LogDir  string
	Feature string // "restore", "backup", "dbscan"
}

// NewErrorLogger membuat instance ErrorLogger baru
func NewErrorLogger(logger applog.Logger, logDir, feature string) *ErrorLogger {
	if logDir == "" {
		logDir = "/var/log/sfDBTools"
	}
	return &ErrorLogger{
		Logger:  logger,
		LogDir:  logDir,
		Feature: feature,
	}
}

// Log mencatat error sederhana ke file log
func (el *ErrorLogger) Log(details map[string]interface{}, err error) string {
	if err == nil {
		return ""
	}

	logFile := el.getLogFilePath()
	if err := os.MkdirAll(el.LogDir, 0755); err != nil {
		el.Logger.Warnf("Gagal membuat log directory: %v", err)
		return ""
	}

	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		el.Logger.Warnf("Gagal membuka error log file: %v", err)
		return ""
	}
	defer f.Close()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logEntry := el.formatLogEntry(timestamp, details, fmt.Sprintf("%v", err), "")

	if _, err := f.WriteString(logEntry); err != nil {
		el.Logger.Warnf("Gagal menulis ke error log file: %v", err)
		return ""
	}

	return logFile
}

// LogWithOutput mencatat error dengan output detail ke file log
func (el *ErrorLogger) LogWithOutput(details map[string]interface{}, output string, err error) string {
	if err == nil {
		return ""
	}

	logFile := el.getLogFilePath()
	if err := os.MkdirAll(el.LogDir, 0755); err != nil {
		el.Logger.Warnf("Gagal membuat log directory: %v", err)
		return ""
	}

	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		el.Logger.Warnf("Gagal membuka error log file: %v", err)
		return ""
	}
	defer f.Close()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logEntry := el.formatLogEntry(timestamp, details, fmt.Sprintf("%v", err), output)

	if _, err := f.WriteString(logEntry); err != nil {
		el.Logger.Warnf("Gagal menulis ke error log file: %v", err)
		return ""
	}

	return logFile
}

// getLogFilePath mengembalikan path file log berdasarkan feature
func (el *ErrorLogger) getLogFilePath() string {
	dateStr := time.Now().Format("2006-01-02")
	return filepath.Join(el.LogDir, fmt.Sprintf("sfDBTools_%s_error_%s.log", dateStr, el.Feature))
}

// formatLogEntry membuat format log entry
func (el *ErrorLogger) formatLogEntry(timestamp string, details map[string]interface{}, errMsg, output string) string {
	entry := fmt.Sprintf("[%s]", timestamp)

	// Tambahkan details
	for key, val := range details {
		entry += fmt.Sprintf(" %s: %v |", key, val)
	}

	entry = entry[:len(entry)-1] + "\n"
	entry += fmt.Sprintf("Error: %s\n", errMsg)

	if output != "" {
		entry += fmt.Sprintf("Output:\n%s\n", output)
	}

	entry += "---\n"
	return entry
}
