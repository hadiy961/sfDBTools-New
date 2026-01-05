// File : internal/app/dbscan/executor.go
// Deskripsi : Background executor dan local sizing logic
// Author : Hadiyatna Muflihun
// Tanggal : 15 Oktober 2025
// Last Modified : 5 Januari 2026

package dbscan

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"sfDBTools/internal/app/dbscan/helpers"
	dbscanmodel "sfDBTools/internal/app/dbscan/model"
)

// executeScanInBackground menjalankan scanning dalam mode background (tanpa UI).
// Menggunakan file locking untuk mencegah multiple instance berjalan bersamaan.
func (s *Service) executeScanInBackground(ctx context.Context, config dbscanmodel.ScanEntryConfig) error {
	_ = config
	scanID := fmt.Sprintf("scan_%s", time.Now().Format("20060102_150405"))

	s.Log.Infof("[%s] === START BACKGROUND SCAN ===", scanID)

	// Setup connections (silent mode)
	sourceClient, dbFiltered, cleanup, err := s.setupScanConnections(ctx, "", false)
	if err != nil {
		s.Log.Errorf("[%s] Setup failed: %v", scanID, err)
		return err
	}
	defer cleanup()

	// Acquire lock to ensure single instance
	unlock, err := s.acquireBackgroundLock(scanID)
	if err != nil {
		return err
	}
	defer unlock()

	// Setup graceful shutdown context
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	s.setupSignalHandler(scanID, cancel)

	// Execute Scan
	s.Log.Infof("[%s] Scanning %d database...", scanID, len(dbFiltered))
	result, detailsMap, err := s.executeScanWithClients(runCtx, sourceClient, dbFiltered, true)
	if err != nil {
		s.Log.Errorf("[%s] Scan failed: %v", scanID, err)
		s.logErrorToFile(scanID, err)
		return err
	}

	// Log results
	helpers.LogScanResult(result, s.Log, scanID)
	if len(detailsMap) > 0 {
		helpers.LogDetailResults(detailsMap, s.Log)
	}

	s.Log.Infof("[%s] === FINISHED BACKGROUND SCAN ===", scanID)
	return nil
}

// acquireBackgroundLock mengelola file locking dan PID file
func (s *Service) acquireBackgroundLock(scanID string) (func(), error) {
	logDir := filepath.Join("logs", "dbscan")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		s.Log.Warnf("[%s] Failed to create log dir: %v", scanID, err)
	}

	lockFile := filepath.Join(logDir, "dbscan_background.lock")
	pidFile := filepath.Join(logDir, "dbscan_background.pid")

	lf, err := os.OpenFile(lockFile, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		s.Log.Warnf("[%s] Failed to open lock file: %v", scanID, err)
		return nil, nil // Proceed carefully? Or fail? Logic suggests we should probably just warn if we can't lock, but let's try to lock if opened.
	}

	// Try exclusive non-blocking lock
	if err := syscall.Flock(int(lf.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		lf.Close()
		return nil, fmt.Errorf("background process already running (locked: %s)", lockFile)
	}

	// Write PID
	_ = os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0644)

	// Return unlock function
	return func() {
		_ = os.Remove(pidFile)
		_ = syscall.Flock(int(lf.Fd()), syscall.LOCK_UN)
		lf.Close()
	}, nil
}

// setupSignalHandler menangani SIGINT/SIGTERM untuk graceful shutdown
func (s *Service) setupSignalHandler(scanID string, cancel context.CancelFunc) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		s.Log.Warnf("[%s] Received signal %s - shutting down...", scanID, sig)
		cancel()
	}()
}

// logErrorToFile menulis error ke file log terpisah via ErrorLog
func (s *Service) logErrorToFile(scanID string, err error) {
	if s.ErrorLog == nil {
		return
	}
	logFile := s.ErrorLog.Log(map[string]interface{}{
		"scanid": scanID,
		"type":   "background_scan",
	}, err)
	if logFile != "" {
		s.Log.Infof("[%s] Error details saved to: %s", scanID, logFile)
	}
}

// LocalScanSizes menghitung ukuran folder database secara langsung
func (s *Service) LocalScanSizes(ctx context.Context, datadir string, dbNames []string) (map[string]int64, error) {
	s.Log.Infof("Starting local size scan in: %s", datadir)
	sizes := make(map[string]int64, len(dbNames))

	for _, dbName := range dbNames {
		if ctx.Err() != nil {
			return sizes, ctx.Err()
		}

		dbPath := filepath.Join(datadir, dbName)
		size, err := dirSize(ctx, dbPath)
		if err != nil {
			s.Log.Warnf("Failed to calculate size for %s: %v", dbName, err)
			sizes[dbName] = 0
			continue
		}
		sizes[dbName] = size
	}
	return sizes, nil
}

// dirSize helper untuk menghitung ukuran direktori rekursif
func dirSize(ctx context.Context, root string) (int64, error) {
	var total int64
	info, err := os.Stat(root)
	if err != nil {
		return 0, err
	}
	if !info.IsDir() {
		return info.Size(), nil
	}

	err = filepath.WalkDir(root, func(_ string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if d.Type().IsRegular() {
			if fi, err := d.Info(); err == nil {
				total += fi.Size()
			}
		}
		return nil
	})
	return total, err
}
