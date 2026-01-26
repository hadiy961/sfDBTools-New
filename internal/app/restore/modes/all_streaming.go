// File : internal/restore/modes/all_streaming.go
// Deskripsi : Streaming restore execution untuk AllExecutor
// Author : Hadiyatna Muflihun
// Tanggal : 30 Desember 2025
// Last Modified : 26 Januari 2026
package modes

import (
	"context"
	"fmt"
	"io"
	"sfdbtools/internal/app/restore/helpers"
	restoremodel "sfdbtools/internal/app/restore/model"
	"sfdbtools/internal/ui/print"
	"sfdbtools/internal/ui/progress"
	"sort"
	"strings"
	"time"
)

func isSSLMismatchServerNotSupport(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "tls/ssl error") && strings.Contains(msg, "server does not support")
}

func hasSkipSSLArg(args []string) bool {
	for _, a := range args {
		if strings.TrimSpace(strings.ToLower(a)) == "--skip-ssl" {
			return true
		}
	}
	return false
}

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

	profile := e.service.GetProfile()
	baseExtraArgs := []string{"--force", "--reconnect", "--max_allowed_packet=1073741824"}

	runOnce := func(withSkipSSL bool) (*restoreStats, time.Duration, error) {
		restoreStart := time.Now()
		spin := progress.NewSpinnerWithElapsed("Memulai proses restore...")
		spin.Start()
		defer spin.Stop()

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

		extraArgs := append([]string{}, baseExtraArgs...)
		if withSkipSSL {
			extraArgs = append([]string{"--skip-ssl"}, extraArgs...)
		}
		args := helpers.BuildMySQLArgs(profile, "", extraArgs...)

		err := helpers.ExecuteMySQLCommand(ctx, args, pipeReader)
		if err != nil {
			// Hentikan processing goroutine secepat mungkin.
			_ = pipeReader.CloseWithError(err)
			_ = <-errCh
			<-progressDone
			return nil, time.Since(restoreStart).Round(time.Millisecond), fmt.Errorf("gagal restore: %w", err)
		}

		// Check processing error
		if perr := <-errCh; perr != nil {
			<-progressDone
			return nil, time.Since(restoreStart).Round(time.Millisecond), fmt.Errorf("error processing file: %w", perr)
		}

		// Tunggu progress updater selesai
		<-progressDone

		duration := time.Since(restoreStart).Round(time.Millisecond)
		var stats *restoreStats
		select {
		case stats = <-statsCh:
		default:
			stats = nil
		}
		return stats, duration, nil
	}

	stats, dur, err := runOnce(false)
	if err != nil {
		if isSSLMismatchServerNotSupport(err) {
			logger.Warn("Restore gagal karena SSL mismatch; mencoba ulang dengan --skip-ssl")
			stats2, dur2, err2 := runOnce(true)
			if err2 == nil {
				logger.Infof("Durasi restore: %s", dur2)
				if stats2 != nil {
					print.PrintSuccess(fmt.Sprintf("✓ Total database restored: %d", stats2.RestoredCount))
				}
				return nil
			}
			return err2
		}
		return err
	}

	logger.Infof("Durasi restore: %s", dur)
	if stats != nil {
		print.PrintSuccess(fmt.Sprintf("✓ Total database restored: %d", stats.RestoredCount))
	}
	return nil
}

// handleProgressUpdates handles spinner updates from progress channel
func (e *AllExecutor) handleProgressUpdates(spin *progress.Spinner, progressCh <-chan string, done chan<- bool) {
	var lastDB string
	var startTime time.Time

	for currentDB := range progressCh {
		if currentDB != lastDB {
			if lastDB != "" {
				duration := time.Since(startTime)
				successMsg := fmt.Sprintf("✓ Database %s restored (%s)", lastDB, duration.Round(time.Millisecond))
				progress.RunWithSpinnerSuspended(func() { fmt.Println(successMsg) })
			}
			lastDB = currentDB
			startTime = time.Now()
		}
		spin.Update(fmt.Sprintf("Restoring database: %s", currentDB))
	}

	if lastDB != "" {
		duration := time.Since(startTime)
		msg := fmt.Sprintf("✓ Database %s restored (%s)", lastDB, duration.Round(time.Millisecond))
		progress.RunWithSpinnerSuspended(func() { fmt.Println(msg) })
	}
	done <- true
}
