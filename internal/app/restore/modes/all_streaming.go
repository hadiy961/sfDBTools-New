// File : internal/restore/modes/all_streaming.go
// Deskripsi : Streaming restore execution untuk AllExecutor
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-30
// Last Modified :  2026-01-05
package modes

import (
	"context"
	"fmt"
	"io"
	"sfDBTools/internal/app/restore/helpers"
	restoremodel "sfDBTools/internal/app/restore/model"
	"sfDBTools/pkg/ui"
	"sort"
	"strings"
	"time"
)

// executeStreamingRestore melakukan restore dengan streaming processing
func (e *AllExecutor) executeStreamingRestore(ctx context.Context, opts *restoremodel.RestoreAllOptions) error {
	logger := e.service.GetLogger()

	// 1) Kumpulkan daftar DB target dari dump (pass-1)
	logger.Info("Menganalisis file dump untuk mengumpulkan daftar database target...")
	targetDBs, err := collectDatabasesToRestore(ctx, opts)
	if err != nil {
		return err
	}
	if len(targetDBs) == 0 {
		return fmt.Errorf("tidak ada database yang akan di-restore (semua tersaring)")
	}

	// Siapkan daftar terurut untuk pemrosesan deterministik
	names := make([]string, 0, len(targetDBs))
	for name := range targetDBs {
		names = append(names, name)
	}
	sort.Strings(names)
	logger.Infof("Hasil analisis: %d database akan diproses untuk restore", len(names))
	if len(names) > 0 {
		logger.Infof("Database ditemukan dari file dump (%d): %s", len(names), strings.Join(names, ", "))
	}

	// 2) Pre-restore backup dan/atau drop target
	if err := e.handlePreRestoreOperations(ctx, opts, names); err != nil {
		return err
	}

	// 3) Mulai proses restore (pass-2)
	return e.performStreamingRestore(ctx, opts)
}

// handlePreRestoreOperations handles backup and drop operations before restore
func (e *AllExecutor) handlePreRestoreOperations(ctx context.Context, opts *restoremodel.RestoreAllOptions, names []string) error {
	logger := e.service.GetLogger()

	// Prepare backup and drop lists
	toBackup, toDrop, existsCount, err := prepareTargetDatabases(ctx, e.service, names, opts.SkipBackup, opts.DropTarget, opts.StopOnError)
	if err != nil {
		return err
	}

	// Execute bulk backup
	if !opts.SkipBackup {
		backupHelper := &bulkBackupHelper{
			service:     e.service,
			ctx:         ctx,
			databases:   toBackup,
			backupOpts:  opts.BackupOptions,
			stopOnError: opts.StopOnError,
		}
		if _, err := backupHelper.executeBackup(); err != nil {
			return err
		}
	}

	// Execute bulk drop
	if opts.DropTarget {
		notExists := len(names) - existsCount
		logger.Infof("Ringkasan drop: %d dari %d database akan di-drop; %d belum ada", len(toDrop), len(names), notExists)

		dropHelper := &bulkDropHelper{
			service:     e.service,
			ctx:         ctx,
			databases:   toDrop,
			stopOnError: opts.StopOnError,
		}
		if _, err := dropHelper.executeDrop(); err != nil {
			return err
		}
	}

	return nil
}

// performStreamingRestore executes the actual streaming restore
func (e *AllExecutor) performStreamingRestore(ctx context.Context, opts *restoremodel.RestoreAllOptions) error {
	logger := e.service.GetLogger()
	restoreStart := time.Now()
	spin := ui.NewSpinnerWithElapsed("Memulai proses restore...")
	spin.Start()

	// Buat pipe reader untuk streaming processing
	pipeReader, pipeWriter := io.Pipe()

	// Channel untuk menerima stats dan progress updates
	statsCh := make(chan *restoreStats, 1)
	progressCh := make(chan string, 100)
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
	go e.handleProgressUpdates(spin, progressCh, progressDone)

	// Build mysql args
	profile := e.service.GetProfile()
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

	// Tunggu progress updater selesai
	<-progressDone

	// Stop spinner dan print summary
	spin.Stop()
	restoreDuration := time.Since(restoreStart).Round(time.Millisecond)
	logger.Infof("Durasi restore: %s", restoreDuration)

	if stats := <-statsCh; stats != nil {
		ui.PrintSuccess(fmt.Sprintf("✓ Total database restored: %d", stats.RestoredCount))
	}

	return nil
}

// handleProgressUpdates handles spinner updates from progress channel
func (e *AllExecutor) handleProgressUpdates(spin *ui.SpinnerWithElapsed, progressCh <-chan string, done chan<- bool) {
	var lastDB string
	var startTime time.Time

	for currentDB := range progressCh {
		if currentDB != lastDB {
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
		spin.UpdateMessage(fmt.Sprintf("Restoring database: %s", currentDB))
	}

	if lastDB != "" {
		duration := time.Since(startTime)
		msg := fmt.Sprintf("✓ Database %s restored (%s)", lastDB, duration.Round(time.Millisecond))
		spin.SuspendAndRun(func() {
			fmt.Println(msg)
		})
	}
	done <- true
}
