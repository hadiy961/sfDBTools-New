// File : internal/dbscan/dbscan_executor.go
// Deskripsi : Eksekutor utama untuk database scanning dan menyimpan hasil ke database
// Author : Hadiyatna Muflihun
// Tanggal : 15 Oktober 2025
// Last Modified : 16 Oktober 2025

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
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/ui"

	"github.com/dustin/go-humanize"
)

// ExecuteScanInBackground menjalankan scanning tanpa UI output (pure logging)
// Ini adalah "background" mode dalam artian tidak ada interaksi UI, bukan goroutine
// Process tetap berjalan sampai selesai, cocok untuk cron job atau automation
func (s *Service) ExecuteScanInBackground(ctx context.Context, config types.ScanEntryConfig) error {
	// Generate unique scan ID
	scanID := fmt.Sprintf("scan_%s", time.Now().Format("20060102_150405"))

	s.Logger.Infof("[%s] ========================================", scanID)
	s.Logger.Infof("[%s] DATABASE SCANNING - BACKGROUND MODE", scanID)
	s.Logger.Infof("[%s] ========================================", scanID)
	s.Logger.Infof("[%s] Memulai background scanning...", scanID)

	// Setup connections
	sourceClient, targetClient, dbFiltered, cleanup, err := s.setupScanConnections(ctx, "", false)
	if err != nil {
		s.Logger.Errorf("[%s] Gagal setup session: %v", scanID, err)
		return err
	}
	defer cleanup()

	// Ensure logs dir, create lockfile and pid file, and acquire exclusive lock (flock)
	logDir := filepath.Join("logs", "dbscan")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		s.Logger.Warnf("[%s] Gagal membuat log dir: %v", scanID, err)
	}

	lockFile := filepath.Join(logDir, "dbscan_background.lock")
	pidFile := filepath.Join(logDir, "dbscan_background.pid")

	// Open (or create) lock file
	lf, err := os.OpenFile(lockFile, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		s.Logger.Warnf("[%s] Gagal membuka lock file %s: %v", scanID, lockFile, err)
	} else {
		// Try to acquire exclusive non-blocking lock
		if err := syscall.Flock(int(lf.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
			// Couldn't acquire lock -> another process likely running
			lf.Close()
			s.Logger.Warnf("[%s] Tidak dapat memperoleh lock, proses background lain mungkin sedang berjalan (lockfile=%s)", scanID, lockFile)
			return fmt.Errorf("background process sudah berjalan (lockfile=%s)", lockFile)
		}
		s.Logger.Infof("[%s] Berhasil memperoleh lock: %s", scanID, lockFile)
	}

	// Write own PID to pid file
	if err := os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0644); err != nil {
		s.Logger.Warnf("[%s] Gagal menulis pid file: %v", scanID, err)
	} else {
		s.Logger.Infof("[%s] Menulis PID file: %s", scanID, pidFile)
	}

	// Create cancellable context to support graceful shutdown
	runCtx, cancel := context.WithCancel(ctx)

	// Setup signal handler to cancel context (graceful shutdown)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		s.Logger.Warnf("[%s] Menerima sinyal %s - memulai graceful shutdown", scanID, sig.String())
		cancel()
	}()

	// Cleanup: remove pidfile, release flock, close lock file, stop signal notifications
	defer func() {
		// Remove pid file
		if err := os.Remove(pidFile); err != nil {
			s.Logger.Warnf("[%s] Gagal menghapus pid file %s: %v", scanID, pidFile, err)
		} else {
			s.Logger.Infof("[%s] PID file %s dihapus", scanID, pidFile)
		}

		// Release flock and close file
		if lf != nil {
			if err := syscall.Flock(int(lf.Fd()), syscall.LOCK_UN); err != nil {
				s.Logger.Warnf("[%s] Gagal melepaskan lock %s: %v", scanID, lockFile, err)
			} else {
				s.Logger.Infof("[%s] Lock dilepas: %s", scanID, lockFile)
			}
			lf.Close()
		}

		signal.Stop(sigs)
		cancel()
	}()

	// Lakukan scanning dengan background mode (pure logging)
	s.Logger.Infof("[%s] Scanning %d database...", scanID, len(dbFiltered))
	result, err := s.ExecuteScan(runCtx, sourceClient, targetClient, dbFiltered, true)
	if err != nil {
		s.Logger.Errorf("[%s] Scanning gagal: %v", scanID, err)
		return err
	}

	// Log hasil
	s.Logger.Infof("[%s] ========================================", scanID)
	s.Logger.Infof("[%s] HASIL SCANNING", scanID)
	s.Logger.Infof("[%s] ========================================", scanID)
	s.Logger.Infof("[%s] Total Database  : %d", scanID, result.TotalDatabases)
	s.Logger.Infof("[%s] Berhasil        : %d", scanID, result.SuccessCount)
	s.Logger.Infof("[%s] Gagal           : %d", scanID, result.FailedCount)
	s.Logger.Infof("[%s] Durasi          : %s", scanID, result.Duration)

	if len(result.Errors) > 0 {
		s.Logger.Warnf("[%s] Terdapat %d error saat scanning:", scanID, len(result.Errors))
		for i, errMsg := range result.Errors {
			s.Logger.Warnf("[%s]   %d. %s", scanID, i+1, errMsg)
		}
	}

	s.Logger.Infof("[%s] Background scanning selesai dengan sukses.", scanID)
	s.Logger.Infof("[%s] ========================================", scanID)

	return nil
}

// ExecuteScan melakukan scanning database dan menyimpan hasilnya
// Parameter isBackground menentukan apakah output menggunakan logger (true) atau UI (false)
func (s *Service) ExecuteScan(ctx context.Context, sourceClient *database.Client, targetClient *database.Client, dbNames []string, isBackground bool) (*types.ScanResult, error) {
	startTime := time.Now()

	if isBackground {
		s.Logger.Info("Memulai proses scanning database...")
		s.Logger.Infof("Total database yang akan di-scan: %d", len(dbNames))
	}
	ui.PrintSubHeader("Memulai Proses Scanning Database")

	if len(dbNames) == 0 {
		return nil, fmt.Errorf("tidak ada database untuk di-scan")
	}

	// Get Original max_statement_time
	// Set and restore session max_statement_time via helper
	s.Logger.Info("Mengatur max_statement_time (session) untuk mencegah query lama...")
	restore, originalMaxStatementTime, err := database.WithSessionMaxStatementTime(ctx, sourceClient, 0)
	if err != nil {
		s.Logger.Warn("Setup max_statement_time (session) gagal: " + err.Error())
	} else {
		s.Logger.Infof("Original (session) max_statement_time: %f detik", originalMaxStatementTime)
		// Pastikan nilai dikembalikan pada akhir fungsi
		defer func(orig float64) {
			bctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if rerr := restore(bctx); rerr != nil {
				s.Logger.Warn("Gagal mengembalikan max_statement_time (session): " + rerr.Error())
			} else {
				s.Logger.Info("max_statement_time (session) berhasil dikembalikan.")
			}
		}(originalMaxStatementTime)
	}

	// Mulai scanning dan simpan segera; hentikan pada kegagalan simpan
	s.Logger.Info("Memulai pengumpulan detail database...")

	// Opsi LocalScan: hanya metrik SIZE yang diambil dari filesystem (datadir),
	// metrik lain tetap melalui query seperti biasa.
	var localSizes map[string]int64
	if s.ScanOptions.LocalScan {
		s.Logger.Info("Mode Local Scan diaktifkan: ukuran database diambil dari datadir, metrik lain via query.")

		// Coba ambil datadir dari source terlebih dahulu; fallback ke target bila perlu
		var datadir string
		if sourceClient != nil {
			if err := sourceClient.DB().QueryRowContext(ctx, "SELECT @@datadir").Scan(&datadir); err != nil {
				s.Logger.Warnf("Gagal mendapatkan datadir dari source: %v", err)
			}
		}
		if datadir == "" && targetClient != nil {
			if err := targetClient.DB().QueryRowContext(ctx, "SELECT @@datadir").Scan(&datadir); err != nil {
				s.Logger.Warnf("Gagal mendapatkan datadir dari target: %v", err)
			}
		}
		if datadir == "" {
			return nil, fmt.Errorf("tidak dapat menentukan datadir dari source maupun target")
		}

		// Hitung ukuran setiap database dari datadir
		sizes, err := s.LocalScanSizes(ctx, datadir, dbNames)
		if err != nil {
			return nil, fmt.Errorf("gagal melakukan local size scan: %v", err)
		}
		localSizes = sizes
		s.Logger.Infof("Local size scan selesai untuk %d database.", len(localSizes))
	}

	// Kumpulan detail jika nanti ingin ditampilkan
	detailsMap := make(map[string]types.DatabaseDetailInfo)

	// Counter
	successCount := 0
	failedCount := 0
	var errors []string

	// Ambil server info untuk penyimpanan
	if err := sourceClient.DB().QueryRowContext(ctx, "SELECT @@hostname, @@port").Scan(&s.ScanOptions.ProfileInfo.DBInfo.Host, &s.ScanOptions.ProfileInfo.DBInfo.Port); err != nil {
		s.Logger.Warnf("Gagal mendapatkan datadir dari target: %v", err)
	}

	serverHost := s.ScanOptions.ProfileInfo.DBInfo.Host
	serverPort := s.ScanOptions.ProfileInfo.DBInfo.Port

	saveEnabled := s.ScanOptions.SaveToDB && targetClient != nil

	// Kumpulkan detail sambil langsung menyimpan setiap hasil
	// Siapkan opsi untuk override size menggunakan hasil local scan (jika ada)
	var collectOpts *database.DetailCollectOptions
	if s.ScanOptions.LocalScan {
		sizeProvider := func(ctx context.Context, dbName string) (int64, error) {
			if localSizes == nil {
				return 0, nil
			}
			if sz, ok := localSizes[dbName]; ok {
				return sz, nil
			}
			return 0, nil
		}
		collectOpts = &database.DetailCollectOptions{SizeProvider: sizeProvider}
	}

	detailsMap, collectErr := sourceClient.CollectDatabaseDetailsWithOptions(ctx, dbNames, s.Logger, collectOpts, func(detail types.DatabaseDetailInfo) error {
		// Simpan ke map untuk pelaporan/penampilan opsional
		detailsMap[detail.DatabaseName] = detail

		// Jika LocalScan aktif dan ada ukuran lokal, pastikan nilai size sudah sesuai (idempotent)
		if s.ScanOptions.LocalScan && localSizes != nil {
			if sz, ok := localSizes[detail.DatabaseName]; ok {
				detail.SizeBytes = sz
				detail.SizeHuman = humanize.Bytes(uint64(sz))
			}
		}

		// Jika tidak perlu simpan ke DB, anggap sukses koleksi
		if !saveEnabled {
			successCount++
			return nil
		}

		// Lakukan penyimpanan segera
		if err := targetClient.SaveDatabaseDetail(ctx, detail, serverHost, serverPort); err != nil {
			failedCount++
			errors = append(errors, fmt.Sprintf("%s: %v", detail.DatabaseName, err))
			// Log dan hentikan proses dengan mengembalikan error
			s.Logger.Errorf("Gagal menyimpan database %s: %v", detail.DatabaseName, err)
			return err
		}
		successCount++
		return nil
	})

	if collectErr != nil {
		s.Logger.Errorf("Proses scanning dihentikan: %v", collectErr)

		duration := time.Since(startTime)
		return &types.ScanResult{
			TotalDatabases: len(dbNames),
			SuccessCount:   successCount,
			FailedCount:    failedCount,
			Duration:       duration.String(),
			Errors:         errors,
		}, collectErr
	}

	// Tampilkan hasil jika diminta (hanya untuk foreground)
	if s.ScanOptions.DisplayResults && !isBackground {
		s.DisplayDetailResults(detailsMap)
	}

	// Log detail untuk background mode
	if isBackground && s.ScanOptions.DisplayResults {
		s.LogDetailResults(detailsMap)
	}

	duration := time.Since(startTime)

	return &types.ScanResult{
		TotalDatabases: len(dbNames),
		SuccessCount:   successCount,
		FailedCount:    failedCount,
		Duration:       duration.String(),
		Errors:         errors,
	}, nil
}

func (s *Service) LocalScan(ctx context.Context, sourceClient *database.Client, datadir string) error {
	// Deprecated signature kept temporarily; redirect to the new implementation requiring dbNames.
	return fmt.Errorf("LocalScan(ctx, *Client, datadir) deprecated - gunakan LocalScan(ctx, datadir, dbNames)")
}

// LocalScan menghitung ukuran setiap database (folder) langsung dari filesystem datadir.
// Hanya ukuran (size) yang dihitung; metrik lain tetap dikumpulkan melalui query.
func (s *Service) LocalScanSizes(ctx context.Context, datadir string, dbNames []string) (map[string]int64, error) {
	s.Logger.Infof("Memulai local size scan di datadir: %s", datadir)

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
			s.Logger.Warnf("Gagal menghitung size untuk %s: %v", dbName, err)
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
