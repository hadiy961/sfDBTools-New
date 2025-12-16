// File : internal/dbscan/dbscan_executor.go
// Deskripsi : Background executor untuk database scanning
// Author : Hadiyatna Muflihun
// Tanggal : 15 Oktober 2025
// Last Modified : 16 Desember 2025

package dbscan

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"sfDBTools/internal/types"
	"sfDBTools/pkg/dbscanhelper"
)

// executeScanInBackground menjalankan scanning tanpa UI output (pure logging)
// Ini adalah "background" mode dalam artian tidak ada interaksi UI, bukan goroutine
// Process tetap berjalan sampai selesai, cocok untuk cron job atau automation
func (s *Service) executeScanInBackground(ctx context.Context, config types.ScanEntryConfig) error {
	// Generate unique scan ID
	scanID := fmt.Sprintf("scan_%s", time.Now().Format("20060102_150405"))

	s.Log.Infof("[%s] ========================================", scanID)
	s.Log.Infof("[%s] DATABASE SCANNING - BACKGROUND MODE", scanID)
	s.Log.Infof("[%s] ========================================", scanID)
	s.Log.Infof("[%s] Memulai background scanning...", scanID)

	// Setup connections
	sourceClient, targetClient, dbFiltered, cleanup, err := s.setupScanConnections(ctx, "", false)
	if err != nil {
		s.Log.Errorf("[%s] Gagal setup session: %v", scanID, err)
		return err
	}
	defer cleanup()

	// Ensure logs dir, create lockfile and pid file, and acquire exclusive lock (flock)
	logDir := filepath.Join("logs", "dbscan")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		s.Log.Warnf("[%s] Gagal membuat log dir: %v", scanID, err)
	}

	lockFile := filepath.Join(logDir, "dbscan_background.lock")
	pidFile := filepath.Join(logDir, "dbscan_background.pid")

	// Open (or create) lock file
	lf, err := os.OpenFile(lockFile, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		s.Log.Warnf("[%s] Gagal membuka lock file %s: %v", scanID, lockFile, err)
	} else {
		// Try to acquire exclusive non-blocking lock
		if err := syscall.Flock(int(lf.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
			// Couldn't acquire lock -> another process likely running
			lf.Close()
			s.Log.Warnf("[%s] Tidak dapat memperoleh lock, proses background lain mungkin sedang berjalan (lockfile=%s)", scanID, lockFile)
			return fmt.Errorf("background process sudah berjalan (lockfile=%s)", lockFile)
		}
		s.Log.Infof("[%s] Berhasil memperoleh lock: %s", scanID, lockFile)
	}

	// Write own PID to pid file
	if err := os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0644); err != nil {
		s.Log.Warnf("[%s] Gagal menulis pid file: %v", scanID, err)
	} else {
		s.Log.Infof("[%s] Menulis PID file: %s", scanID, pidFile)
	}

	// Create cancellable context to support graceful shutdown
	runCtx, cancel := context.WithCancel(ctx)

	// Setup signal handler to cancel context (graceful shutdown)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		s.Log.Warnf("[%s] Menerima sinyal %s - memulai graceful shutdown", scanID, sig.String())
		cancel()
	}()

	// Cleanup: remove pidfile, release flock, close lock file, stop signal notifications
	defer func() {
		// Remove pid file
		if err := os.Remove(pidFile); err != nil {
			s.Log.Warnf("[%s] Gagal menghapus pid file %s: %v", scanID, pidFile, err)
		} else {
			s.Log.Infof("[%s] PID file %s dihapus", scanID, pidFile)
		}

		// Release flock and close file
		if lf != nil {
			if err := syscall.Flock(int(lf.Fd()), syscall.LOCK_UN); err != nil {
				s.Log.Warnf("[%s] Gagal melepaskan lock %s: %v", scanID, lockFile, err)
			} else {
				s.Log.Infof("[%s] Lock dilepas: %s", scanID, lockFile)
			}
			lf.Close()
		}

		signal.Stop(sigs)
		cancel()
	}()

	// Lakukan scanning dengan background mode (pure logging)
	s.Log.Infof("[%s] Scanning %d database...", scanID, len(dbFiltered))
	result, detailsMap, err := s.executeScanWithClients(runCtx, sourceClient, targetClient, dbFiltered, true)
	if err != nil {
		s.Log.Errorf("[%s] Scanning gagal: %v", scanID, err)

		// Log ke error log file terpisah
		logFile := s.ErrorLog.Log(map[string]interface{}{
			"scanid": scanID,
			"type":   "background_scan",
		}, err)
		if logFile != "" {
			s.Log.Infof("[%s] ℹ Error details tersimpan di: %s", scanID, logFile)
		}

		return err
	}

	// Log hasil menggunakan helper
	dbscanhelper.LogScanResult(result, s.Log, scanID)
	if len(detailsMap) > 0 {
		dbscanhelper.LogDetailResults(detailsMap, s.Log)
	}

	s.Log.Infof("[%s] Background scanning selesai dengan sukses.", scanID)
	s.Log.Infof("[%s] ========================================", scanID)

	return nil
}

// LocalScanSizes menghitung ukuran setiap database (folder) langsung dari filesystem datadir.
// Hanya ukuran (size) yang dihitung; metrik lain tetap dikumpulkan melalui query.
func (s *Service) LocalScanSizes(ctx context.Context, datadir string, dbNames []string) (map[string]int64, error) {
	s.Log.Infof("Memulai local size scan di datadir: %s", datadir)

	sizes := make(map[string]int64, len(dbNames))
	for _, dbName := range dbNames {
		select {
		case <-ctx.Done():
			return sizes, ctx.Err()
		default:
		}

		dbPath := filepath.Join(datadir, dbName)
		size, err := dirSize(ctx, dbPath)
		if err != nil {
			// Catat per-db error, tapi lanjutkan database lain
			s.Log.Warnf("Gagal menghitung size untuk %s: %v", dbName, err)
			sizes[dbName] = 0
			continue
		}
		sizes[dbName] = size
	}

	// Catatan: Perhitungan ini tidak mencakup shared tablespace (mis. ibdata1),
	// sehingga total bisa lebih kecil dibandingkan information_schema.
	// Ini sesuai permintaan: hanya menghitung dari direktori per-database.
	return sizes, nil
}

// dirSize menghitung total ukuran semua file regular di dalam sebuah direktori secara rekursif.
func dirSize(ctx context.Context, root string) (int64, error) {
	var total int64

	info, err := os.Stat(root)
	if err != nil {
		return 0, err
	}
	if !info.IsDir() {
		// Jika bukan direktori, kembalikan ukuran file (kasus langka)
		return info.Size(), nil
	}

	walkErr := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// cek pembatalan secara berkala
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if d.Type().IsRegular() {
			fi, ferr := d.Info()
			if ferr != nil {
				return ferr
			}
			total += fi.Size()
		}
		return nil
	})
	if walkErr != nil {
		return 0, walkErr
	}
	return total, nil
}
