// File : internal/crypto/audit/logger.go
// Deskripsi : Audit logging untuk crypto operations (security trail)
// Author : Hadiyatna Muflihun
// Tanggal : 21 Januari 2026
// Last Modified : 21 Januari 2026
package audit

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"time"

	applog "sfdbtools/internal/services/log"
)

// Operation represents crypto operation type
type Operation string

const (
	OpEncryptFile Operation = "encrypt_file"
	OpDecryptFile Operation = "decrypt_file"
	OpEncryptText Operation = "encrypt_text"
	OpDecryptText Operation = "decrypt_text"
	OpBase64Enc   Operation = "base64_encode"
	OpBase64Dec   Operation = "base64_decode"
	OpEnvEncode   Operation = "env_encode"
)

// Event represents crypto operation audit event
type Event struct {
	Timestamp time.Time
	Operation Operation
	User      string
	Hostname  string
	InputPath string
	Success   bool
	Error     string
}

// LogOperation logs crypto operation untuk audit trail.
// Non-fatal: gagal logging tidak menggagalkan operasi crypto.
func LogOperation(logger applog.Logger, op Operation, inputPath string, success bool, err error) {
	if logger == nil {
		return
	}

	event := Event{
		Timestamp: time.Now(),
		Operation: op,
		InputPath: inputPath,
		Success:   success,
	}

	// Get username (best effort)
	if u, e := user.Current(); e == nil {
		event.User = u.Username
	} else {
		event.User = "unknown"
	}

	// Get hostname (best effort)
	if h, e := os.Hostname(); e == nil {
		event.Hostname = h
	} else {
		event.Hostname = "unknown"
	}

	// Format error message
	if err != nil {
		event.Error = err.Error()
	}

	// Log dengan context
	logMsg := fmt.Sprintf("AUDIT: operation=%s user=%s@%s input=%s success=%t",
		event.Operation,
		event.User,
		event.Hostname,
		filepath.Base(event.InputPath), // Only basename untuk privacy
		event.Success,
	)

	if !event.Success && event.Error != "" {
		logMsg += fmt.Sprintf(" error=%q", event.Error)
	}

	if event.Success {
		logger.Info(logMsg)
	} else {
		logger.Warn(logMsg)
	}
}
