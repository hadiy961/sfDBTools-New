package copy

import (
	"context"
	"fmt"
	"sync"

	copyexec "sfdbtools/internal/app/copy/execution"
	profileconn "sfdbtools/internal/app/profile/connection"
	"sfdbtools/internal/domain"
)

// CopyDatabaseConcurrent menyalin seluruh isi database menggunakan multi-threading worker pool.
// Metode ini jauh lebih cepat untuk database dengan banyak tabel.
func (s *Service) CopyDatabaseConcurrent(ctx context.Context, profile *domain.ProfileInfo, sourceDB, targetDB string, workers int, limitSpeed int64, force, backupFirst, includeGrants, verify, skipRoutines, skipEvents, skipTriggers, nonInteractive bool) (string, error) {
	client, err := profileconn.ConnectWithProfile(s.cfg, profile, "")
	if err != nil {
		return "", fmt.Errorf("gagal koneksi ke database: %w", err)
	}
	defer client.Close()

	// 1. Discovery Objek (Gunakan Optimized SHOW commands)
	s.log.Infof("Memulai discovery objek database '%s'...", sourceDB)
	allObjects, err := s.DiscoverTablesAndViews(ctx, client, sourceDB)
	if err != nil {
		return "", err
	}

	var baseTables []string
	for _, obj := range allObjects {
		if obj.Type == TableTypeBaseTable {
			baseTables = append(baseTables, obj.Name)
		}
	}

	s.log.Infof("Ditemukan %d tabel untuk disalin.", len(baseTables))

	// 2. Setup Target (Create & Smart Clean)
	if err := client.CreateDatabaseIfNotExists(ctx, targetDB); err != nil {
		return "", err
	}
	if force || backupFirst {
		if err := s.SmartDropDatabaseObjects(ctx, client, targetDB); err != nil {
			s.log.Warnf("Gagal membersihkan database target: %v", err)
		}
	}

	// 3. Step A: Salin Struktur (Schema Only)
	s.log.Info("Menyalin struktur database (schema-only)...")
	extraDumpArgs := ""
	if !skipRoutines {
		extraDumpArgs += " --routines"
	}
	if !skipEvents {
		extraDumpArgs += " --events"
	}
	if !skipTriggers {
		extraDumpArgs += " --triggers"
	}

	err = copyexec.ExecutePiping(ctx, s.log, copyexec.PipingOptions{
		Profile:      profile,
		SourceDB:     sourceDB,
		TargetDB:     targetDB,
		SchemaOnly:   true,
		BaseDumpArgs: s.cfg.Backup.MysqlDumpArgs + extraDumpArgs,
		LimitSpeed:   0,
	})
	if err != nil {
		return "", fmt.Errorf("gagal menyalin struktur schema: %w", err)
	}

	// 4. Step B: Salin Data (Concurrent)
	if len(baseTables) > 0 {
		s.log.Infof("Menyalin data tabel menggunakan %d workers...", workers)

		// Matikan checks di target agar load data cepat dan tidak error FK
		if _, err := client.ExecContextWithRetry(ctx, "SET FOREIGN_KEY_CHECKS=0"); err != nil {
			return "", err
		}
		if _, err := client.ExecContextWithRetry(ctx, "SET UNIQUE_CHECKS=0"); err != nil {
			return "", err
		}
		defer func() {
			_, _ = client.ExecContextWithRetry(ctx, "SET FOREIGN_KEY_CHECKS=1")
			_, _ = client.ExecContextWithRetry(ctx, "SET UNIQUE_CHECKS=1")
		}()

		tableChan := make(chan string, len(baseTables))
		errChan := make(chan error, len(baseTables))
		var wg sync.WaitGroup

		limitPerWorker := int64(0)
		if limitSpeed > 0 {
			limitPerWorker = limitSpeed / int64(workers)
			if limitPerWorker < 1024*1024 {
				limitPerWorker = 1024 * 1024
			}
		}

		for w := 1; w <= workers; w++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				for tbl := range tableChan {
					select {
					case <-ctx.Done():
						return
					default:
					}

					s.log.Debugf("[Worker %d] Menyalin data tabel: %s", workerID, tbl)
					err := copyexec.ExecutePiping(ctx, s.log, copyexec.PipingOptions{
						Profile:    profile,
						SourceDB:   sourceDB,
						TargetDB:   targetDB,
						TableName:  tbl,
						SchemaOnly: false,
						// Gunakan flag --no-create-info karena tabel sudah ada dari step schema
						BaseDumpArgs: s.cfg.Backup.MysqlDumpArgs + " --no-create-info",
						LimitSpeed:   limitPerWorker,
					})
					if err != nil {
						errChan <- fmt.Errorf("tabel %s gagal: %w", tbl, err)
					}
				}
			}(w)
		}

		for _, tbl := range baseTables {
			tableChan <- tbl
		}
		close(tableChan)

		wg.Wait()
		close(errChan)

		if len(errChan) > 0 {
			firstErr := <-errChan
			return "", fmt.Errorf("terjadi error saat concurrent data copy: %w (dan %d error lainnya)", firstErr, len(errChan))
		}

		// 4.5 Data Integrity Check (if enabled)
		if verify {
			s.log.Info("Memulai verifikasi checksum seluruh tabel...")
			for _, tbl := range baseTables {
				ok, err := s.VerifyChecksum(ctx, client, sourceDB, tbl, targetDB, tbl)
				if err != nil {
					s.log.Warnf("Gagal verifikasi checksum %s: %v", tbl, err)
				} else if !ok {
					s.log.Errorf("Mismatch checksum pada tabel: %s", tbl)
				}
			}
		}
	}

	// 5. Step C: Copy Grants (if enabled)
	if includeGrants {
		if err := s.CopyGrants(ctx, profile, sourceDB, targetDB); err != nil {
			s.log.Warnf("Gagal menyalin hak akses user: %v", err)
		}
	}

	return targetDB, nil
}
