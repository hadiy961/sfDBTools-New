package writer

import (
	"fmt"
	"io"
	"strings"
	"time"

	"sfDBTools/pkg/ui"
)

// databaseMonitorWriter wraps an io.Writer to track mysqldump progress per database.
type databaseMonitorWriter struct {
	target    io.Writer
	spinner   *ui.SpinnerWithElapsed
	currentDB string
	startTime time.Time
}

func newDatabaseMonitorWriter(target io.Writer, spinner *ui.SpinnerWithElapsed) *databaseMonitorWriter {
	return &databaseMonitorWriter{target: target, spinner: spinner}
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
		w.spinner.SuspendAndRun(func() {
			fmt.Println(msg)
		})
	}

	w.currentDB = newDB
	w.startTime = now
	w.spinner.UpdateMessage(fmt.Sprintf("Processing %s", newDB))
}

func (w *databaseMonitorWriter) Finish(success bool) {
	if w.currentDB != "" && success {
		duration := time.Since(w.startTime)
		msg := fmt.Sprintf("✓ Database %s (%s)", w.currentDB, duration.Round(time.Millisecond))
		w.spinner.SuspendAndRun(func() {
			fmt.Println(msg)
		})
	}
}
