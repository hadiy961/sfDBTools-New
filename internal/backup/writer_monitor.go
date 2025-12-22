package backup

import (
	"fmt"
	"io"
	"sfDBTools/pkg/ui"
	"strings"
	"time"
)

// =============================================================================
// Database Monitor Writer
// =============================================================================

// databaseMonitorWriter membungkus io.Writer untuk memantau progress per database
type databaseMonitorWriter struct {
	target    io.Writer
	spinner   *ui.SpinnerWithElapsed
	currentDB string
	startTime time.Time
}

func newDatabaseMonitorWriter(target io.Writer, spinner *ui.SpinnerWithElapsed) *databaseMonitorWriter {
	return &databaseMonitorWriter{
		target:  target,
		spinner: spinner,
	}
}

var dbMarker = []byte("-- Current Database: ")

func (w *databaseMonitorWriter) Write(p []byte) (n int, err error) {
	// Scan marker
	if idx := strings.Index(string(p), string(dbMarker)); idx != -1 {
		// Marker found, try to extract DB name
		// Format: -- Current Database: `dbname`
		remainder := p[idx+len(dbMarker):]

		// Cari backtick pembuka
		if startTick := strings.IndexByte(string(remainder), '`'); startTick != -1 {
			remainder = remainder[startTick+1:]
			// Cari backtick penutup
			if endTick := strings.IndexByte(string(remainder), '`'); endTick != -1 {
				dbName := string(remainder[:endTick])

				if dbName != w.currentDB {
					w.onDatabaseSwitch(dbName)
				}
			}
		}
	}

	return w.target.Write(p)
}

func (w *databaseMonitorWriter) onDatabaseSwitch(newDB string) {
	now := time.Now()

	// Report previous DB success
	if w.currentDB != "" {
		duration := now.Sub(w.startTime)
		msg := fmt.Sprintf("✓ Database %s (%s)", w.currentDB, duration.Round(time.Millisecond))

		// Gunakan SuspendAndRun untuk print ke stdout tanpa ganggu spinner
		w.spinner.SuspendAndRun(func() {
			fmt.Println(msg)
		})
	}

	// Update state
	w.currentDB = newDB
	w.startTime = now

	// Update spinner
	w.spinner.UpdateMessage(fmt.Sprintf("Processing %s", newDB))
}

func (w *databaseMonitorWriter) Finish(success bool) {
	// Report last DB
	if w.currentDB != "" && success {
		duration := time.Since(w.startTime)
		msg := fmt.Sprintf("✓ Database %s (%s)", w.currentDB, duration.Round(time.Millisecond))

		w.spinner.SuspendAndRun(func() {
			fmt.Println(msg)
		})
	}
}
