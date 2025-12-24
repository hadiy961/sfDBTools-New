// File : internal/restore/modes/all.go
// Deskripsi : Executor untuk restore all databases dengan streaming filtering
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-18
// Last Modified : 2025-12-18

package modes

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"sfDBTools/internal/restore/helpers"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/ui"
	"sort"
	"strings"
	"time"
)

// AllExecutor implements restore for all databases dengan streaming filtering
type AllExecutor struct {
	service RestoreService
}

// NewAllExecutor creates a new AllExecutor
func NewAllExecutor(svc RestoreService) *AllExecutor {
	return &AllExecutor{service: svc}
}

// Execute executes all databases restore dengan streaming processing
func (e *AllExecutor) Execute(ctx context.Context) (*types.RestoreResult, error) {
	startTime := time.Now()
	opts := e.service.GetAllOptions()

	result := &types.RestoreResult{
		TargetDB:   "ALL_DATABASES",
		SourceFile: opts.File,
		Success:    false,
	}

	logger := e.service.GetLogger()
	logger.Info("Memulai proses restore all databases")
	e.service.SetRestoreInProgress("ALL_DATABASES")
	defer e.service.ClearRestoreInProgress()

	// Konfirmasi sudah dilakukan di setup (DisplayConfirmation).

	// Note: do not close the target DB client here. Some helper operations
	// (e.g., DropAllDatabases) may still need an active client. Closing the
	// client prematurely causes "sql: database is closed" errors.

	// Tidak melakukan bulk-drop di sini. Drop akan dilakukan selektif
	// hanya untuk database yang memang akan di-restore.

	// Dry run mode - hanya analisis file tanpa restore
	if opts.DryRun {
		logger.Info("Mode DRY-RUN: Analisis file dump tanpa restore...")
		return e.executeDryRun(ctx, opts, result)
	}

	// Execute actual restore dengan streaming
	if err := e.executeStreamingRestore(ctx, opts); err != nil {
		result.Error = err
		return result, err
	}

	// Restore user grants if available (optional)
	if !opts.SkipGrants {
		result.GrantsFile = opts.GrantsFile
		grantsRestored, err := e.service.RestoreUserGrantsIfAvailable(ctx, opts.GrantsFile)
		if err != nil {
			logger.Errorf("Gagal restore user grants: %v", err)
			result.GrantsRestored = false
		} else {
			result.GrantsRestored = grantsRestored
		}
	}

	result.Success = true
	result.Duration = time.Since(startTime).Round(time.Second).String()
	logger.Info("Restore all databases berhasil")

	return result, nil
}

// executeDryRun melakukan analisis file dump tanpa restore
func (e *AllExecutor) executeDryRun(ctx context.Context, opts *types.RestoreAllOptions, result *types.RestoreResult) (*types.RestoreResult, error) {
	logger := e.service.GetLogger()
	start := time.Now()
	logger.Info("Membuka dan menganalisis file dump...")

	reader, closers, err := helpers.OpenAndPrepareReader(opts.File, opts.EncryptionKey)
	if err != nil {
		result.Error = err
		return result, err
	}
	defer helpers.CloseReaders(closers)

	scanner := bufio.NewScanner(reader)
	configureScanner(scanner)

	dbStats := make(map[string]int)
	// tidak lagi menampilkan daftar database yang di-skip; filtering tetap berlaku secara internal
	var currentDB string

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "USE `") {
			currentDB = extractDBName(line)
			if currentDB != "" {
				skipReason := shouldSkipDatabase(currentDB, opts)
				if skipReason == "" {
					dbStats[currentDB] = 0
				}
			}
		}

		if currentDB != "" && dbStats[currentDB] >= 0 {
			dbStats[currentDB]++
		}
	}

	if err := scanner.Err(); err != nil {
		result.Error = fmt.Errorf("error membaca file dump: %w", err)
		return result, result.Error
	}

	// Print hasil analisis
	ui.PrintSuccess("\n📊 Hasil Analisis File Dump:")
	ui.PrintInfo(fmt.Sprintf("Total database yang akan di-restore: %d", len(dbStats)))

	if len(dbStats) > 0 {
		ui.PrintInfo("\nDatabase yang akan di-restore:")
		for db, lines := range dbStats {
			ui.PrintInfo(fmt.Sprintf("  • %s (%d baris)", db, lines))
		}
	}

	result.Success = true
	result.Duration = time.Since(start).Round(time.Second).String()
	return result, nil
}

// executeStreamingRestore melakukan restore dengan streaming processing
func (e *AllExecutor) executeStreamingRestore(ctx context.Context, opts *types.RestoreAllOptions) error {
	logger := e.service.GetLogger()
	// 1) Kumpulkan daftar DB target dari dump (pass-1)
	logger.Info("Menganalisis file dump untuk mengumpulkan daftar database target...")
	targetDBs, err := e.collectDatabasesToRestore(ctx, opts)
	if err != nil {
		return err
	}
	if len(targetDBs) == 0 {
		return fmt.Errorf("tidak ada database yang akan di-restore (semua tersaring)")
	}

	// Siapkan daftar terurut untuk pemrosesan deterministik (tanpa menampilkan nama)
	names := make([]string, 0, len(targetDBs))
	for name := range targetDBs {
		names = append(names, name)
	}
	sort.Strings(names)
	logger.Infof("Hasil analisis: %d database akan diproses untuk restore", len(names))
	// Tampilkan daftar database hasil analisis dalam satu baris (comma-separated)
	if len(names) > 0 {
		logger.Infof("Database ditemukan dari file dump (%d): %s", len(names), strings.Join(names, ", "))
	}

	// 2) Pre-restore backup (opsional) dan/atau drop target.
	// Untuk safety, backup harus dilakukan SEBELUM drop.
	client := e.service.GetTargetClient()
	if client == nil {
		return fmt.Errorf("client database tidak tersedia")
	}

	// Kumpulkan info existence sekali (dipakai untuk backup dan drop)
	toBackup := make([]string, 0, len(names))
	toDrop := make([]string, 0, len(names))
	existsCount := 0
	logger.Info("Memeriksa keberadaan database pada server target...")
	for _, dbName := range names {
		exists, chkErr := client.CheckDatabaseExists(ctx, dbName)
		if chkErr != nil {
			if opts.StopOnError {
				return fmt.Errorf("gagal mengecek database %s: %w", dbName, chkErr)
			}
			logger.Warnf("Gagal cek database %s: %v (lanjut)", dbName, chkErr)
			continue
		}
		if exists {
			existsCount++
			if !opts.SkipBackup {
				toBackup = append(toBackup, dbName)
			}
			if opts.DropTarget {
				toDrop = append(toDrop, dbName)
			}
		}
	}
	logger.Infof("Ringkasan pengecekan: %d dari %d database sudah ada", existsCount, len(names))

	if !opts.SkipBackup {
		if len(toBackup) == 0 {
			logger.Info("Pre-restore backup: tidak ada database existing untuk dibackup")
		} else {
			outDir := ""
			if opts.BackupOptions != nil {
				outDir = opts.BackupOptions.OutputDir
			}
			if outDir == "" {
				outDir = "(default dari config)"
			}
			logger.Infof("Pre-restore backup (single-file): membackup %d database ke %s", len(toBackup), outDir)
			backupStart := time.Now()
			backupFile, bErr := e.service.BackupDatabasesSingleFileIfNeeded(ctx, toBackup, false, opts.BackupOptions)
			backupDuration := time.Since(backupStart).Round(time.Millisecond)
			if bErr != nil {
				if opts.StopOnError {
					return fmt.Errorf("backup pre-restore gagal: %w", bErr)
				}
				logger.Warnf("backup pre-restore gagal (durasi: %s): %v (lanjut)", backupDuration, bErr)
			} else if backupFile != "" {
				logger.Infof("Pre-restore backup selesai (durasi: %s): %s", backupDuration, backupFile)
			} else {
				logger.Infof("Pre-restore backup selesai (durasi: %s)", backupDuration)
			}
		}
	}

	if opts.DropTarget {
		notExists := len(names) - existsCount
		logger.Infof("Ringkasan drop: %d dari %d database akan di-drop; %d belum ada", len(toDrop), len(names), notExists)

		// Tampilkan daftar database yang akan di-drop pada satu baris (comma-separated)
		if len(toDrop) > 0 {
			logger.Infof("Database yang akan di-drop (%d): %s", len(toDrop), strings.Join(toDrop, ", "))
		} else {
			logger.Info("Tidak ada database yang perlu di-drop")
		}

		// Lanjutkan drop setelah ringkasan ditampilkan (tanpa spinner tambahan, tetap ukur durasi)
		dropStart := time.Now()
		droppedCount := 0
		for _, dbName := range toDrop {
			if dropErr := client.DropDatabase(ctx, dbName); dropErr != nil {
				if opts.StopOnError {
					return fmt.Errorf("gagal drop database %s: %w", dbName, dropErr)
				}
				logger.Warnf("Gagal drop database %s: %v (lanjut)", dbName, dropErr)
				continue
			}
			droppedCount++
		}
		dropDuration := time.Since(dropStart).Round(time.Millisecond)
		logger.Infof("Berhasil drop %d/%d database (durasi: %s)", droppedCount, len(toDrop), dropDuration)
	}

	// 3) Mulai proses restore (pass-2)
	restoreStart := time.Now()
	spin := ui.NewSpinnerWithElapsed("Memulai proses restore...")
	spin.Start()

	// Buat pipe reader untuk streaming processing
	pipeReader, pipeWriter := io.Pipe()

	// Channel untuk menerima stats dan progress updates
	statsCh := make(chan *restoreStats, 1)
	progressCh := make(chan string, 100) // Buffered to prevent blocking
	errCh := make(chan error, 1)
	progressDone := make(chan bool)

	// Process file dalam goroutine terpisah
	go func() {
		defer pipeWriter.Close()
		defer close(progressCh)
		stats, err := e.processStreamWithFiltering(ctx, opts, pipeWriter, progressCh)
		if err != nil {
			pipeWriter.CloseWithError(err)
			errCh <- err
			return
		}
		statsCh <- stats
		errCh <- nil
	}()

	// Goroutine untuk update spinner dari progress updates
	go func() {
		var lastDB string
		var startTime time.Time

		for currentDB := range progressCh {
			// Handle DB Switch
			if currentDB != lastDB {
				// Jika ada database sebelumnya yang sedang diproses, berarti sudah selesai
				if lastDB != "" {
					duration := time.Since(startTime)
					successMsg := fmt.Sprintf("✓ Database %s restored (%s)", lastDB, duration.Round(time.Millisecond))
					spin.SuspendAndRun(func() {
						fmt.Println(successMsg)
					})
				}
				lastDB = currentDB
				startTime = time.Now()
			}

			// Update Spinner Message
			spin.UpdateMessage(fmt.Sprintf("Restoring database: %s", currentDB))
		}

		// Handle database terakhir setelah loop selesai
		if lastDB != "" {
			duration := time.Since(startTime)
			msg := fmt.Sprintf("✓ Database %s restored (%s)", lastDB, duration.Round(time.Millisecond))
			spin.SuspendAndRun(func() {
				fmt.Println(msg)
			})
		}
		progressDone <- true
	}()

	// Build mysql args
	profile := e.service.GetProfile()

	// Gunakan opsi yang meningkatkan ketahanan streaming terhadap error/koneksi.
	// - --force: jangan keluar saat terjadi error per-statement (hindari broken pipe)
	// - --reconnect: coba reconnect jika koneksi TCP terputus saat proses panjang
	// - --max_allowed_packet: cegah error paket besar dari dump
	extraArgs := []string{"--force", "--reconnect", "--max_allowed_packet=1073741824"}
	args := helpers.BuildMySQLArgs(profile, "", extraArgs...)

	// Execute mysql dengan streaming input
	if err := helpers.ExecuteMySQLCommand(ctx, args, pipeReader); err != nil {
		spin.Stop()
		return fmt.Errorf("gagal restore: %w", err)
	}

	// Check processing error
	if err := <-errCh; err != nil {
		spin.Stop()
		return fmt.Errorf("error processing file: %w", err)
	}

	// Tunggu progress updater selesai (print last DB)
	<-progressDone

	// Stop spinner sebelum print summary
	spin.Stop()
	restoreDuration := time.Since(restoreStart).Round(time.Millisecond)
	logger.Infof("Durasi restore: %s", restoreDuration)

	// Print summary stats
	if stats := <-statsCh; stats != nil {
		ui.PrintSuccess(fmt.Sprintf("✓ Total database restored: %d", stats.RestoredCount))
	}

	return nil
}

// restoreStats menyimpan statistik restore
type restoreStats struct {
	RestoredCount int
}

// processStreamWithFiltering membaca file, filter, dan tulis ke MySQL stdin
func (e *AllExecutor) processStreamWithFiltering(ctx context.Context, opts *types.RestoreAllOptions, output io.Writer, progressCh chan<- string) (*restoreStats, error) {
	reader, closers, err := helpers.OpenAndPrepareReader(opts.File, opts.EncryptionKey)
	if err != nil {
		return nil, err
	}
	defer helpers.CloseReaders(closers)

	scanner := bufio.NewScanner(reader)
	configureScanner(scanner)

	var currentDB string
	skipCurrentDB := false
	restoredDBs := make(map[string]bool)
	// tidak lagi melaporkan 'skipped' ke user; tetap gunakan flag internal untuk melewati system DB

	// Buffer writes to reduce syscall overhead and speed up piping to mysql
	bufWriter := bufio.NewWriterSize(output, 4*1024*1024) // 4MB buffer
	defer bufWriter.Flush()

	for scanner.Scan() {
		line := scanner.Text()

		// Deteksi statement USE `db_name`;
		if strings.HasPrefix(line, "USE `") {
			newDB := extractDBName(line)
			if newDB != "" {
				currentDB = newDB

				// Check apakah DB harus di-skip
				skipReason := shouldSkipDatabase(currentDB, opts)
				skipCurrentDB = (skipReason != "")

				if !skipCurrentDB {
					if !restoredDBs[currentDB] {
						restoredDBs[currentDB] = true
						// Kirim update progress DB ke spinner
						progressCh <- currentDB
					}
				}
			}
		}

		// Skip baris jika database sedang di-skip
		if skipCurrentDB {
			continue
		}

		// Tulis baris ke MySQL stdin (buffered) dengan newline
		if _, err := bufWriter.WriteString(line + "\n"); err != nil {
			return nil, fmt.Errorf("gagal menulis ke mysql stdin: %w", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error membaca file dump: %w", err)
	}

	return &restoreStats{
		RestoredCount: len(restoredDBs),
	}, nil
}

// configureScanner mengkonfigurasi scanner dengan buffer besar untuk handle INSERT panjang
func configureScanner(scanner *bufio.Scanner) {
	const maxCapacity = 100 * 1024 * 1024 // 100MB max per line
	buf := make([]byte, 0, 4*1024*1024)   // Initial 4MB to reduce resizes
	scanner.Buffer(buf, maxCapacity)
}

// extractDBName mengekstrak nama database dari statement USE `db_name`;
func extractDBName(line string) string {
	start := strings.Index(line, "`")
	if start == -1 {
		return ""
	}
	start++ // skip first backtick

	end := strings.Index(line[start:], "`")
	if end != -1 {
		return line[start : start+end]
	}
	return ""
}

// shouldSkipDatabase checks apakah database harus di-skip (returns reason string or empty)
func shouldSkipDatabase(dbName string, opts *types.RestoreAllOptions) string {
	for _, excluded := range opts.ExcludeDBs {
		if dbName == excluded {
			return "excluded by user"
		}
	}
	if opts.SkipSystemDBs && database.IsSystemDatabase(dbName) {
		return "system database"
	}
	return ""
}

// collectDatabasesToRestore melakukan pass awal untuk mengumpulkan daftar DB target dari dump
func (e *AllExecutor) collectDatabasesToRestore(ctx context.Context, opts *types.RestoreAllOptions) (map[string]struct{}, error) {
	reader, closers, err := helpers.OpenAndPrepareReader(opts.File, opts.EncryptionKey)
	if err != nil {
		return nil, err
	}
	defer helpers.CloseReaders(closers)

	scanner := bufio.NewScanner(reader)
	configureScanner(scanner)

	targets := make(map[string]struct{})
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "USE `") {
			db := extractDBName(line)
			if db == "" {
				continue
			}
			if reason := shouldSkipDatabase(db, opts); reason == "" {
				targets[db] = struct{}{}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error membaca file dump saat mengumpulkan DB: %w", err)
	}
	return targets, nil
}
