package writer

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"time"

	applog "sfdbtools/internal/services/log"
	"sfdbtools/internal/ui/progress"
)

// databaseMonitorWriter wraps an io.Writer untuk tracking progress dump per database.
type databaseMonitorWriter struct {
	target    io.Writer
	spinner   *progress.Spinner
	log       applog.Logger
	currentDB string
	startTime time.Time
}

func newDatabaseMonitorWriter(target io.Writer, spinner *progress.Spinner, log applog.Logger) *databaseMonitorWriter {
	return &databaseMonitorWriter{target: target, spinner: spinner, log: log}
}

var dbMarker = []byte("-- Current Database: ")

func (w *databaseMonitorWriter) Write(p []byte) (n int, err error) {
	// Parse database marker dengan robust error handling
	dbName, parseErr := w.parseDBMarker(p)
	if parseErr != nil {
		// Log parse error hanya jika bukan "marker not found" (non-fatal)
		if w.log != nil && parseErr.Error() != "marker not found" {
			w.log.Debugf("Monitor: parse error (continuing): %v", parseErr)
		}
	} else if dbName != w.currentDB {
		// Database switch detected
		w.onDatabaseSwitch(dbName)
	}

	return w.target.Write(p)
}

// parseDBMarker parses database name dari mysqldump output dengan robust error handling.
// Support MySQL escaped backticks (e.g., `my“db` = my`db).
func (w *databaseMonitorWriter) parseDBMarker(data []byte) (string, error) {
	idx := bytes.Index(data, dbMarker)
	if idx == -1 {
		return "", fmt.Errorf("marker not found")
	}

	remainder := data[idx+len(dbMarker):]

	// Extract quoted identifier (handle escaped backticks)
	dbName, err := extractQuotedIdentifier(remainder)
	if err != nil {
		return "", fmt.Errorf("failed to extract db name: %w", err)
	}

	// Validate db name
	if strings.TrimSpace(dbName) == "" {
		return "", fmt.Errorf("empty database name")
	}

	return dbName, nil
}

// extractQuotedIdentifier extracts MySQL quoted identifier dengan support untuk escaped backticks.
// MySQL escaping: `my“db` = my`db (double backtick = single backtick literal)
func extractQuotedIdentifier(data []byte) (string, error) {
	// Find opening backtick
	start := bytes.IndexByte(data, '`')
	if start == -1 {
		return "", fmt.Errorf("no opening backtick")
	}

	var name strings.Builder
	i := start + 1
	for i < len(data) {
		if data[i] == '`' {
			// Check for escaped backtick (double ``)
			if i+1 < len(data) && data[i+1] == '`' {
				name.WriteByte('`')
				i += 2
				continue
			}
			// Closing backtick found
			return name.String(), nil
		}
		name.WriteByte(data[i])
		i++
	}

	return "", fmt.Errorf("no closing backtick")
}

func (w *databaseMonitorWriter) onDatabaseSwitch(newDB string) {
	now := time.Now()

	if w.currentDB != "" {
		duration := now.Sub(w.startTime)
		msg := fmt.Sprintf("✓ Database %s (%s)", w.currentDB, duration.Round(time.Millisecond))
		if w.log != nil {
			w.log.Infof("%s", msg)
		}
	}

	w.currentDB = newDB
	w.startTime = now
	if w.log != nil {
		w.log.Infof("Processing database: %s", newDB)
	}
	if w.spinner != nil {
		w.spinner.Update(fmt.Sprintf("Processing %s", newDB))
	}
}

func (w *databaseMonitorWriter) Finish(success bool) {
	if w.currentDB != "" && success {
		duration := time.Since(w.startTime)
		msg := fmt.Sprintf("✓ Database %s (%s)", w.currentDB, duration.Round(time.Millisecond))
		if w.log != nil {
			w.log.Infof("%s", msg)
		}
	}
}
