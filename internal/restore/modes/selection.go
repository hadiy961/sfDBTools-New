// File : internal/restore/modes/selection.go
// Deskripsi : Executor untuk restore selection berbasis CSV
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-19

package modes

import (
	"bufio"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"sfDBTools/internal/types"
	"sfDBTools/pkg/helper"
	"sfDBTools/pkg/ui"
)

type selectionExecutor struct {
	svc RestoreService
}

func NewSelectionExecutor(svc RestoreService) RestoreExecutor { return &selectionExecutor{svc: svc} }

func (e *selectionExecutor) Execute(ctx context.Context) (*types.RestoreResult, error) {
	opts := e.svc.GetSelectionOptions()
	if opts == nil {
		return nil, fmt.Errorf("opsi selection tidak tersedia")
	}

	start := time.Now()
	logger := e.svc.GetLogger()

	// Parse CSV entries
	entries, err := e.readCSV(opts.CSV)
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("CSV tidak memiliki data restore")
	}

	client := e.svc.GetTargetClient()
	tracker := newBatchRestoreTracker(len(entries))

	for idx, ent := range entries {
		select {
		case <-ctx.Done():
			return &types.RestoreResult{Success: false, SourceFile: filepath.Base(opts.CSV), Duration: time.Since(start).String()}, ctx.Err()
		default:
		}

		// Resolve db name from file if empty
		dbName := strings.TrimSpace(ent.DBName)
		if dbName == "" {
			dbName = helper.ExtractDatabaseNameFromFile(ent.File)
			if dbName == "" {
				tracker.recordFailure()
				msg := fmt.Sprintf("[%d/%d] %s: gagal infer nama database dari filename", idx+1, tracker.total, filepath.Base(ent.File))
				logger.Warn(msg)
				if opts.StopOnError {
					return nil, errors.New(msg)
				}
				continue
			}
		}

		// Encrypted file requires enc key
		if e.isEncryptedFile(ent.File) && strings.TrimSpace(ent.EncKey) == "" {
			tracker.recordFailure()
			msg := fmt.Sprintf("[%d/%d] %s: file terenkripsi, enc_key wajib diisi", idx+1, tracker.total, filepath.Base(ent.File))
			logger.Warn(msg)
			if opts.StopOnError {
				return nil, errors.New(msg)
			}
			continue
		}

		// Log summary line for this entry
		logger.Infof("[%d/%d] Restore %s → %s", idx+1, tracker.total, filepath.Base(ent.File), dbName)

		if opts.DryRun {
			// For dry-run, only check file exists
			if _, err := os.Stat(ent.File); err != nil {
				tracker.recordFailure()
				msg := fmt.Sprintf("file tidak ditemukan: %s", ent.File)
				logger.Warn(msg)
				if opts.StopOnError {
					return nil, errors.New(msg)
				}
				continue
			}
			tracker.recordSuccess()
			continue
		}

		// Ensure connection is healthy
		if pingErr := client.Ping(ctx); pingErr != nil {
			logger.Warnf("koneksi database tidak sehat: %v", pingErr)
		}

		// Use batch restore helper
		restorer := &singleDatabaseRestore{
			service:       e.svc,
			ctx:           ctx,
			dbName:        dbName,
			sourceFile:    ent.File,
			encryptionKey: ent.EncKey,
			grantsFile:    ent.GrantsFile,
			skipBackup:    opts.SkipBackup,
			dropTarget:    opts.DropTarget,
			stopOnError:   opts.StopOnError,
			backupOpts:    opts.BackupOptions,
		}

		if err := restorer.execute(); err != nil {
			tracker.recordFailure()
			logger.Warnf("restore gagal (%s): %v", dbName, err)
			if opts.StopOnError {
				return nil, err
			}
			continue
		}

		tracker.recordSuccess()
	}

	// Print summary
	summary := tracker.getSummary()
	if tracker.isAllSuccess() {
		ui.PrintSuccess("Hasil: " + summary)
	} else {
		ui.PrintWarning("Hasil: " + summary)
	}

	return &types.RestoreResult{
		Success:    tracker.isAllSuccess(),
		SourceFile: opts.CSV,
		Duration:   time.Since(start).String(),
	}, nil
}

func (e *selectionExecutor) readCSV(path string) ([]types.RestoreSelectionEntry, error) {
	if strings.TrimSpace(path) == "" {
		return nil, errors.New("path CSV wajib diisi (--csv)")
	}
	csvDir := filepath.Dir(path)
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("gagal membuka CSV: %w", err)
	}
	defer f.Close()

	r := csv.NewReader(bufio.NewReader(f))
	r.TrimLeadingSpace = true
	// Allow variable number of fields per record (some rows may include extra blanks)
	r.FieldsPerRecord = -1

	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("gagal membaca CSV: %w", err)
	}
	if len(records) == 0 {
		return []types.RestoreSelectionEntry{}, nil
	}

	entries := make([]types.RestoreSelectionEntry, 0, len(records))

	startIdx := 0
	// Header-aware: if first row starts with filename, assume header
	if len(records[0]) >= 1 && strings.EqualFold(strings.Trim(strings.TrimSpace(records[0][0]), " '"), "filename") {
		startIdx = 1
	}

	for i := startIdx; i < len(records); i++ {
		rec := records[i]
		if len(rec) == 0 {
			// Skip empty lines
			continue
		}

		// Map columns robustly:
		// 0: filename (required)
		// 1: db_name (optional)
		// 2: enc_key (optional)
		// last: grants_file (optional; handle extra blank columns)

		get := func(idx int) string {
			if idx < len(rec) {
				return strings.Trim(strings.TrimSpace(rec[idx]), " '")
			}
			return ""
		}

		file := get(0)
		if file == "" {
			// Skip if file path missing
			continue
		}
		if !filepath.IsAbs(file) {
			file = filepath.Join(csvDir, file)
		}

		dbName := get(1)
		encKey := get(2)
		grants := ""
		if len(rec) >= 4 {
			// Take the last field as grants_file to tolerate extra empty columns
			grants = strings.Trim(strings.TrimSpace(rec[len(rec)-1]), " '")
		}
		if grants != "" && !filepath.IsAbs(grants) {
			grants = filepath.Join(csvDir, grants)
		}

		entry := types.RestoreSelectionEntry{
			File:       file,
			DBName:     dbName,
			EncKey:     encKey,
			GrantsFile: grants,
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

func (e *selectionExecutor) isEncryptedFile(path string) bool {
	return helper.IsEncryptedFile(path)
}
