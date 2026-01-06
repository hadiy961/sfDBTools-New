package writer

import (
	"fmt"
	"io"
	"strings"
	"time"

	applog "sfdbtools/internal/services/log"
	"sfdbtools/internal/ui/progress"
)

// databaseMonitorWriter wraps an io.Writer to track mysqldump progress per database.
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
	if idx := strings.Index(string(p), string(dbMarker)); idx != -1 {
		remainder := p[idx+len(dbMarker):]

		if startTick := strings.IndexByte(string(remainder), '`'); startTick != -1 {
			remainder = remainder[startTick+1:]
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
