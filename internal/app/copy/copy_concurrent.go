package copy

import (
	"context"
	"fmt"
	"strings"
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
	s.log.Infof("Memulai discovery objek database '%s'விற்கு...", sourceDB)
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

	totalTables := len(baseTables)
	s.log.Infof("Ditemukan %d tabel untuk disalin.", totalTables)

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
	if totalTables > 0 {
		s.log.Infof("Menyalin data tabel menggunakan %d workers...", workers)

		// Sanitize base args: Buang flags yang memicu pembuatan objek karena sudah dibuat di Step A
		sanitizedBaseArgs := s.sanitizeArgsForData(s.cfg.Backup.MysqlDumpArgs)

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

		tableChan := make(chan string, totalTables)
		errChan := make(chan error, totalTables)
		var wg sync.WaitGroup

		limitPerWorker := int64(0)
		if limitSpeed > 0 {
			limitPerWorker = limitSpeed / int64(workers)
			if limitPerWorker < 1024*1024 {
				limitPerWorker = 1024*1024
			}
		}

		// Progress tracking
		var completedCount int
		var mu sync.Mutex

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

					err := copyexec.ExecutePiping(ctx, s.log, copyexec.PipingOptions{
						Profile:      profile,
						SourceDB:     sourceDB,
						TargetDB:     targetDB,
						TableName:    tbl,
						SchemaOnly:   false,
						BaseDumpArgs: sanitizedBaseArgs + " --no-create-info --skip-triggers --skip-routines --skip-events",
						LimitSpeed:   limitPerWorker,
						HideProgress: true, // Sembunyikan per-table spinner agar tidak berantakan
					})

					mu.Lock()
					completedCount++
					percentage := float64(completedCount) * 100 / float64(totalTables)
					// Tampilkan progress global yang lebih informatif
					fmt.Printf("\r  ⏳ Progres: %d/%d tabel selesai (%.1f%%) [Worker %d: %s]   ", completedCount, totalTables, percentage, workerID, tbl)
					if err != nil {
						fmt.Printf("\n  ❌ Error pada tabel %s: %v\n", tbl, err)
						errChan <- fmt.Errorf("tabel %s gagal: %w", tbl, err)
					}
					mu.Unlock()
				}
			}(w)
		}

		for _, tbl := range baseTables {
			tableChan <- tbl
		}
		close(tableChan)

		wg.Wait()
		fmt.Println() // Newline after progress bar
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

// sanitizeArgsForData membuang flag yang tidak diinginkan untuk fase penyalinan data murni.
func (s *Service) sanitizeArgsForData(args string) string {
	fields := strings.Fields(args)
	var filtered []string
	for _, f := range fields {
		l := strings.ToLower(f)
		if l == "--routines" || l == "-r" ||
			l == "--triggers" || l == "-t" ||
			l == "--events" || l == "-e" ||
			l == "--databases" || l == "-b" ||
			l == "--all-databases" || l == "-a" ||
			strings.HasPrefix(l, "--set-gtid-purged") {
			continue
		}
		filtered = append(filtered, f)
	}
	return strings.Join(filtered, " ")
}
