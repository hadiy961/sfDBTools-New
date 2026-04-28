package copy

import (
	"context"
	"fmt"
	"sync"
	"time"

	profileconn "sfdbtools/internal/app/profile/connection"
	"sfdbtools/internal/shared/database"
	"sfdbtools/internal/ui/progress"
)

// CopyTableResult menyimpan status eksekusi kloning per tabel.
type CopyTableResult struct {
	SourceDB     string
	SourceTable  string
	TargetDB     string
	TargetTable  string
	Status       string
	VerifyStatus string
	Duration     time.Duration
	Error        error
}

// CopyTablesConcurrent melakukan penyalinan banyak tabel secara paralel menggunakan worker pool.
func (s *Service) CopyTablesConcurrent(ctx context.Context, opts CopyTablesConcurrentOptions) ([]CopyTableResult, error) {
	var results []CopyTableResult
	var mu sync.Mutex

	type tableTask struct {
		index int
		name  string
	}
	taskChan := make(chan tableTask, len(opts.SourceTables))
	wg := sync.WaitGroup{}

	numWorkers := opts.Workers
	if numWorkers > len(opts.SourceTables) {
		numWorkers = len(opts.SourceTables)
	}
	if numWorkers < 1 {
		numWorkers = 1
	}

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range taskChan {
				select {
				case <-ctx.Done():
					return
				default:
				}

				// Prepare specific options for this table
				tableOpts := opts.CopyTableOptions
				tableOpts.SourceTable = task.name
				if tableOpts.TargetTable == "" {
					tableOpts.TargetTable = task.name // default to same name
				}

				start := time.Now()
				s.log.Infof("[%d/%d] Kloning %s.%s -> %s.%s ...", task.index+1, len(opts.SourceTables), tableOpts.SourceDB, tableOpts.SourceTable, tableOpts.TargetDB, tableOpts.TargetTable)

				_, _, verifyStatus, err := s.CopyTable(ctx, tableOpts)
				duration := time.Since(start).Round(time.Millisecond)

				status := "Sukses"
				if err != nil {
					status = "Gagal"
					s.log.Errorf("  ❌ Error %s: %v", task.name, err)
				} else {
					s.log.Infof("  ✅ %s Berhasil (%s)", task.name, duration)
				}

				mu.Lock()
				results = append(results, CopyTableResult{
					SourceDB:     tableOpts.SourceDB,
					SourceTable:  tableOpts.SourceTable,
					TargetDB:     tableOpts.TargetDB,
					TargetTable:  tableOpts.TargetTable,
					Status:       status,
					VerifyStatus: verifyStatus,
					Duration:     duration,
					Error:        err,
				})
				mu.Unlock()
			}
		}()
	}

	for i, tbl := range opts.SourceTables {
		taskChan <- tableTask{index: i, name: tbl}
	}
	close(taskChan)
	wg.Wait()

	return results, nil
}

// CopyTable melakukan penyalinan tabel spesifik.
func (s *Service) CopyTable(ctx context.Context, opts CopyTableOptions) (string, string, string, error) {
	client, err := profileconn.ConnectWithProfile(s.cfg, opts.Profile, "")
	if err != nil {
		return "", "", "", fmt.Errorf("gagal koneksi ke database: %w", err)
	}
	defer client.Close()

	// Validation
	exists, err := s.validateTableExists(ctx, client, opts.SourceDB, opts.SourceTable)
	if err != nil || !exists {
		return "", "", "", fmt.Errorf("tabel sumber %s.%s tidak ditemukan", opts.SourceDB, opts.SourceTable)
	}

	if opts.TargetDB == "" {
		opts.TargetDB = opts.SourceDB
	}

	// Create DB if not exists
	if err := client.CreateDatabaseIfNotExists(ctx, opts.TargetDB); err != nil {
		return "", "", "", err
	}

	targetExists, err := s.validateTableExists(ctx, client, opts.TargetDB, opts.TargetTable)
	if err != nil {
		return "", "", "", err
	}

	if targetExists {
		// Gunakan helper konsolidasi untuk konfirmasi overwrite
		var err error
		opts.Force, opts.BackupFirst, err = s.ConfirmOverwriteInteractive(
			fmt.Sprintf("%s.%s", opts.TargetDB, opts.TargetTable),
			opts.NonInteractive,
			opts.Force,
			opts.BackupFirst,
		)
		if err != nil {
			return "", "", "", err
		}

		if opts.BackupFirst {
			s.log.Infof("Menjalankan backup pengamanan untuk tabel: %s.%s", opts.TargetDB, opts.TargetTable)
			if err := s.runSafetyTableBackup(ctx, opts.Profile, client, opts.TargetDB, opts.TargetTable); err != nil {
				return "", "", "", fmt.Errorf("gagal melakukan backup pengamanan tabel: %w", err)
			}
		}
	}

	s.log.Debugf("Memulai copy tabel: %s.%s -> %s.%s", opts.SourceDB, opts.SourceTable, opts.TargetDB, opts.TargetTable)

	spin := progress.NewSpinnerWithElapsed(fmt.Sprintf("Copying table %s.%s", opts.SourceDB, opts.SourceTable))
	spin.Start()
	defer spin.Stop()

	if opts.Force || opts.BackupFirst {
		dropSQL := fmt.Sprintf("DROP TABLE IF EXISTS `%s`.`%s` ", opts.TargetDB, opts.TargetTable)
		_, _ = client.ExecContextWithRetry(ctx, dropSQL)
	}

	createSQL := fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s`.`%s` LIKE `%s`.`%s` ", opts.TargetDB, opts.TargetTable, opts.SourceDB, opts.SourceTable)
	if _, err := client.ExecContextWithRetry(ctx, createSQL); err != nil {
		return "", "", "", fmt.Errorf("gagal membuat struktur tabel target: %w", err)
	}

	if !opts.SchemaOnly {
		insertSQL := fmt.Sprintf("INSERT INTO `%s`.`%s` SELECT * FROM `%s`.`%s` ", opts.TargetDB, opts.TargetTable, opts.SourceDB, opts.SourceTable)
		if _, err := client.ExecContextWithRetry(ctx, insertSQL); err != nil {
			return "", "", "", fmt.Errorf("gagal menyalin data tabel: %w", err)
		}
	}

	// 6. Copy Grants (if enabled)
	if opts.IncludeGrants {
		// Untuk tabel tunggal, kita tetap menyalin grants database-level sebagai pendekatan aman
		if err := s.CopyGrants(ctx, opts.Profile, opts.SourceDB, opts.TargetDB); err != nil {
			s.log.Warnf("Gagal menyalin hak akses user: %v", err)
		}
	}

	// 7. Verify Checksum
	verifyStatus := "-"
	if opts.Verify && !opts.SchemaOnly {
		ok, err := s.VerifyChecksum(ctx, client, opts.SourceDB, opts.SourceTable, opts.TargetDB, opts.TargetTable)
		if err != nil {
			verifyStatus = "Error Verifikasi"
			s.log.Warnf("Gagal verifikasi checksum %s: %v", opts.SourceTable, err)
		} else if ok {
			verifyStatus = "Cocok"
		} else {
			verifyStatus = "Gagal (Mismatch)"
		}
	}

	return opts.TargetDB, opts.TargetTable, verifyStatus, nil
}

func (s *Service) validateTableExists(ctx context.Context, client *database.Client, dbName, tableName string) (bool, error) {
	query := "SELECT TABLE_NAME FROM information_schema.TABLES WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ? LIMIT 1"
	rows, err := client.QueryContextWithRetry(ctx, query, dbName, tableName)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	return rows.Next(), nil
}
