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
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/ui"
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

	e.service.LogInfo("Memulai proses restore all databases")
	e.service.SetRestoreInProgress("ALL_DATABASES")
	defer e.service.ClearRestoreInProgress()

	// Validasi: konfirmasi user jika bukan force mode atau dry-run
	if !opts.Force && !opts.DryRun {
		ui.PrintWarning("⚠️  PERINGATAN: Operasi ini akan restore SEMUA database dari file dump!")
		ui.PrintWarning("    Database yang sudah ada akan ditimpa (jika drop-target aktif)")

		if len(opts.ExcludeDBs) > 0 {
			e.service.LogInfof("Database yang akan di-exclude: %v", opts.ExcludeDBs)
		}
		if opts.SkipSystemDBs {
			e.service.LogInfo("System databases (mysql, sys, information_schema, performance_schema) akan di-skip")
		}

		confirm, err := input.AskYesNo("\nLanjutkan proses restore?", false)
		if err != nil || !confirm {
			e.service.LogInfo("Restore dibatalkan oleh user")
			result.Error = fmt.Errorf("restore dibatalkan oleh user")
			return result, result.Error
		}
	}

	// Dry run mode - hanya analisis file tanpa restore
	if opts.DryRun {
		e.service.LogInfo("Mode DRY-RUN: Analisis file dump tanpa restore...")
		return e.executeDryRun(ctx, opts, result)
	}

	// Execute actual restore dengan streaming
	if err := e.executeStreamingRestore(ctx, opts); err != nil {
		result.Error = err
		return result, err
	}

	result.Success = true
	result.Duration = time.Since(startTime).Round(time.Second).String()
	e.service.LogInfo("Restore all databases berhasil")

	return result, nil
}

// executeDryRun melakukan analisis file dump tanpa restore
func (e *AllExecutor) executeDryRun(ctx context.Context, opts *types.RestoreAllOptions, result *types.RestoreResult) (*types.RestoreResult, error) {
	e.service.LogInfo("Membuka dan menganalisis file dump...")

	reader, closers, err := helpers.OpenAndPrepareReader(opts.File, opts.EncryptionKey)
	if err != nil {
		result.Error = err
		return result, err
	}
	defer helpers.CloseReaders(closers)

	scanner := bufio.NewScanner(reader)
	configureScanner(scanner)

	dbStats := make(map[string]int)
	skippedDBs := make(map[string]string)
	var currentDB string

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "USE `") {
			currentDB = extractDBName(line)
			if currentDB != "" {
				skipReason := shouldSkipDatabase(currentDB, opts)
				if skipReason != "" {
					skippedDBs[currentDB] = skipReason
				} else {
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

	if len(skippedDBs) > 0 {
		ui.PrintWarning(fmt.Sprintf("\nTotal database yang di-skip: %d", len(skippedDBs)))
		for db, reason := range skippedDBs {
			ui.PrintWarning(fmt.Sprintf("  • %s (%s)", db, reason))
		}
	}

	result.Success = true
	result.Duration = time.Since(time.Now()).Round(time.Second).String()
	return result, nil
}

// executeStreamingRestore melakukan restore dengan streaming processing
func (e *AllExecutor) executeStreamingRestore(ctx context.Context, opts *types.RestoreAllOptions) error {
	spin := ui.NewSpinnerWithElapsed("Memulai proses restore...")
	spin.Start()

	// Buat pipe reader untuk streaming processing
	pipeReader, pipeWriter := io.Pipe()

	// Channel untuk menerima stats dan progress updates
	statsCh := make(chan *restoreStats, 1)
	progressCh := make(chan string, 10)
	errCh := make(chan error, 1)

	// Process file dalam goroutine terpisah
	go func() {
		defer pipeWriter.Close()
		defer close(progressCh)
		stats, err := e.processStreamWithFiltering(opts, pipeWriter, progressCh)
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
		for dbName := range progressCh {
			spin.UpdateMessage(fmt.Sprintf("Restoring database: %s", dbName))
		}
	}()

	// Build mysql args
	profile := e.service.GetProfile()
	extraArgs := []string{}
	if !opts.StopOnError {
		extraArgs = append(extraArgs, "--force")
	}
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

	// Stop spinner sebelum print summary
	spin.Stop()

	// Print summary stats
	if stats := <-statsCh; stats != nil {
		ui.PrintSuccess(fmt.Sprintf("✓ Total database restored: %d", stats.RestoredCount))
		if stats.SkippedCount > 0 {
			ui.PrintInfo(fmt.Sprintf("⏭️  Total database skipped: %d", stats.SkippedCount))
		}
	}

	return nil
}

// restoreStats menyimpan statistik restore
type restoreStats struct {
	RestoredCount int
	SkippedCount  int
}

// processStreamWithFiltering membaca file, filter, dan tulis ke MySQL stdin
func (e *AllExecutor) processStreamWithFiltering(opts *types.RestoreAllOptions, output io.Writer, progressCh chan<- string) (*restoreStats, error) {
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
	skippedDBs := make(map[string]bool)

	for scanner.Scan() {
		line := scanner.Text()

		// Deteksi statement USE `db_name`;
		if strings.HasPrefix(line, "USE `") {
			currentDB = extractDBName(line)
			if currentDB == "" {
				continue
			}

			// Check apakah DB harus di-skip
			skipReason := shouldSkipDatabase(currentDB, opts)
			skipCurrentDB = (skipReason != "")

			if skipCurrentDB {
				if !skippedDBs[currentDB] {
					skippedDBs[currentDB] = true
				}
			} else {
				if !restoredDBs[currentDB] {
					restoredDBs[currentDB] = true
					// Kirim update progress ke spinner
					select {
					case progressCh <- currentDB:
					default:
						// Skip jika channel penuh
					}
				}
			}
		}

		// Skip baris jika database sedang di-skip
		if skipCurrentDB {
			continue
		}

		// Tulis baris ke MySQL stdin (dengan newline)
		if _, err := io.WriteString(output, line+"\n"); err != nil {
			return nil, fmt.Errorf("gagal menulis ke mysql stdin: %w", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error membaca file dump: %w", err)
	}

	return &restoreStats{
		RestoredCount: len(restoredDBs),
		SkippedCount:  len(skippedDBs),
	}, nil
}

// configureScanner mengkonfigurasi scanner dengan buffer besar untuk handle INSERT panjang
func configureScanner(scanner *bufio.Scanner) {
	const maxCapacity = 100 * 1024 * 1024 // 100MB max per line
	buf := make([]byte, 0, 1024*1024)     // Initial 1MB
	scanner.Buffer(buf, maxCapacity)
}

// extractDBName mengekstrak nama database dari statement USE `db_name`;
func extractDBName(line string) string {
	parts := strings.Split(line, "`")
	if len(parts) >= 2 {
		return parts[1]
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
