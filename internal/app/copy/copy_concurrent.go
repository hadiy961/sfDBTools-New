package copy

import (
	"context"
	"fmt"
	"strings"
	"sync"

	copyexec "sfdbtools/internal/app/copy/execution"
	profileconn "sfdbtools/internal/app/profile/connection"
)

// CopyDatabaseConcurrent menyalin seluruh isi database menggunakan multi-threading worker pool.
func (s *Service) CopyDatabaseConcurrent(ctx context.Context, opts CopyDatabaseOptions) (string, error) {
	client, err := profileconn.ConnectWithProfile(s.cfg, opts.Profile, "")
	if err != nil {
		return "", fmt.Errorf("gagal koneksi ke database: %w", err)
	}
	defer client.Close()

	// 1. Discovery Objek
	s.log.Infof("Memulai discovery objek database '%s'…", opts.SourceDB)
	allObjects, err := s.DiscoverTablesAndViews(ctx, client, opts.SourceDB)
	if err != nil {
		return "", err
	}

	var baseTables []string
	var viewNames []string
	for _, obj := range allObjects {
		if obj.Type == TableTypeBaseTable {
			baseTables = append(baseTables, obj.Name)
		} else {
			viewNames = append(viewNames, obj.Name)
		}
	}

	totalTables := len(baseTables)
	s.log.Infof("Ditemukan %d tabel dan %d views untuk disalin.", totalTables, len(viewNames))

	var procNames, funcNames, eventNames, triggerNames []string
	if !opts.SkipRoutines {
		procNames, funcNames, _ = s.DiscoverRoutines(ctx, client, opts.SourceDB)
	}
	if !opts.SkipEvents {
		eventNames, _ = s.DiscoverEvents(ctx, client, opts.SourceDB)
	}
	if !opts.SkipTriggers {
		triggerNames, _ = s.DiscoverTriggers(ctx, client, opts.SourceDB)
	}

	// 2. Setup Target
	if err := client.CreateDatabaseIfNotExists(ctx, opts.TargetDB); err != nil {
		return "", err
	}
	if opts.Force || opts.BackupFirst {
		if err := s.SmartDropDatabaseObjects(ctx, client, opts.TargetDB); err != nil {
			s.log.Warnf("Gagal membersihkan database target: %v", err)
		}
	}

	// 3. Step A: Salin Struktur Tabel SAJA
	if totalTables > 0 {
		s.log.Infof("Menyalin struktur %d tabel…", totalTables)

		var completedSchema int
		var mu sync.Mutex
		tableChan := make(chan string, totalTables)
		errChan := make(chan error, totalTables)
		var wg sync.WaitGroup

		for w := 1; w <= opts.Workers; w++ {
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
						Profile:      opts.Profile,
						SourceDB:     opts.SourceDB,
						TargetDB:     opts.TargetDB,
						TableName:    tbl,
						SchemaOnly:   true,
						BaseDumpArgs: s.cfg.Backup.MysqlDumpArgs + " --no-data --skip-triggers --skip-routines --skip-events",
						HideProgress: true,
						Force:        opts.Force,
					})

					mu.Lock()
					completedSchema++
					percent := float64(completedSchema) * 100 / float64(totalTables)
					s.log.Infof("[%d/%d] Struktur: %s (%.1f%%)", completedSchema, totalTables, tbl, percent)
					if err != nil {
						errChan <- fmt.Errorf("struktur tabel %s gagal: %w", tbl, err)
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

		if len(errChan) > 0 {
			return "", <-errChan
		}
	}

	// 4. Step B: Salin Data (Concurrent)
	if totalTables > 0 {
		s.log.Infof("Menyalin data %d tabel menggunakan %d workers…", totalTables, opts.Workers)

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
		if opts.LimitSpeed > 0 {
			limitPerWorker = opts.LimitSpeed / int64(opts.Workers)
			if limitPerWorker < 1024*1024 {
				limitPerWorker = 1024 * 1024
			}
		}

		var completedData int
		var mu sync.Mutex

		for w := 1; w <= opts.Workers; w++ {
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
						Profile:      opts.Profile,
						SourceDB:     opts.SourceDB,
						TargetDB:     opts.TargetDB,
						TableName:    tbl,
						SchemaOnly:   false,
						BaseDumpArgs: sanitizedBaseArgs + " --no-create-info --skip-triggers --skip-routines --skip-events",
						LimitSpeed:   limitPerWorker,
						HideProgress: true,
						Force:        opts.Force,
					})

					mu.Lock()
					completedData++
					percent := float64(completedData) * 100 / float64(totalTables)
					s.log.Infof("[%d/%d] Data: %s (%.1f%%)", completedData, totalTables, tbl, percent)
					if err != nil {
						errChan <- fmt.Errorf("data tabel %s gagal: %w", tbl, err)
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
		close(errChan)

		if len(errChan) > 0 {
			return "", <-errChan
		}
	}

	// 5. Step C: Salin Objek Pelengkap
	s.log.Info("Memasang objek tambahan (Views, Triggers, Routines, Events)…")

	// 5.1 Salin Triggers (Bulk)
	if !opts.SkipTriggers && len(triggerNames) > 0 {
		err = copyexec.ExecutePiping(ctx, s.log, copyexec.PipingOptions{
			Profile:      opts.Profile,
			SourceDB:     opts.SourceDB,
			TargetDB:     opts.TargetDB,
			SchemaOnly:   true,
			BaseDumpArgs: s.cfg.Backup.MysqlDumpArgs + " --no-create-info --triggers --skip-routines --skip-events",
			Label:        fmt.Sprintf("Memasang %d Triggers", len(triggerNames)),
			Force:        opts.Force,
		})
		if err != nil {
			s.log.Warnf("Gagal menyalin Triggers: %v", err)
		}
	}

	// 5.2 Salin Procedures & Functions
	if !opts.SkipRoutines {
		if len(procNames) > 0 {
			s.log.Infof("Memasang %d Procedures…", len(procNames))
			for i, name := range procNames {
				percent := float64(i+1) * 100 / float64(len(procNames))
				s.log.Infof("[%d/%d] Procedure: %s (%.1f%%)", i+1, len(procNames), name, percent)
				if err := s.copyIndividualObject(ctx, client, opts.Profile, opts.SourceDB, opts.TargetDB, "PROCEDURE", name); err != nil {
					s.log.Warnf("Gagal menyalin procedure %s: %v", name, err)
				}
			}
		}
		if len(funcNames) > 0 {
			s.log.Infof("Memasang %d Functions…", len(funcNames))
			for i, name := range funcNames {
				percent := float64(i+1) * 100 / float64(len(funcNames))
				s.log.Infof("[%d/%d] Function: %s (%.1f%%)", i+1, len(funcNames), name, percent)
				if err := s.copyIndividualObject(ctx, client, opts.Profile, opts.SourceDB, opts.TargetDB, "FUNCTION", name); err != nil {
					s.log.Warnf("Gagal menyalin function %s: %v", name, err)
				}
			}
		}
	}

	// 5.3 Salin Events
	if !opts.SkipEvents && len(eventNames) > 0 {
		s.log.Infof("Memasang %d Events…", len(eventNames))
		for i, name := range eventNames {
			percent := float64(i+1) * 100 / float64(len(eventNames))
			s.log.Infof("[%d/%d] Event: %s (%.1f%%)", i+1, len(eventNames), name, percent)
			if err := s.copyIndividualObject(ctx, client, opts.Profile, opts.SourceDB, opts.TargetDB, "EVENT", name); err != nil {
				s.log.Warnf("Gagal menyalin event %s: %v", name, err)
			}
		}
	}

	// 5.4 Salin Views (Bulk)
	if len(viewNames) > 0 {
		err = copyexec.ExecutePiping(ctx, s.log, copyexec.PipingOptions{
			Profile:      opts.Profile,
			SourceDB:     opts.SourceDB,
			TargetDB:     opts.TargetDB,
			SchemaOnly:   true,
			BaseDumpArgs: s.cfg.Backup.MysqlDumpArgs + " --no-create-info --skip-triggers --skip-routines --skip-events",
			Label:        fmt.Sprintf("Memasang %d Views", len(viewNames)),
			Force:        opts.Force,
		})
		if err != nil {
			s.log.Warnf("Gagal menyalin Views: %v", err)
		}
	}

	// 6. Data Integrity Check
	if opts.Verify && totalTables > 0 {
		s.log.Info("Memulai verifikasi checksum seluruh tabel…")
		for _, tbl := range baseTables {
			ok, err := s.VerifyChecksum(ctx, client, opts.SourceDB, tbl, opts.TargetDB, tbl)
			if err != nil {
				s.log.Warnf("Gagal verifikasi checksum %s: %v", tbl, err)
			} else if !ok {
				s.log.Errorf("Mismatch checksum pada tabel: %s", tbl)
			}
		}
	}

	// 7. Step D: Copy Grants
	if opts.IncludeGrants {
		if err := s.CopyGrants(ctx, opts.Profile, opts.SourceDB, opts.TargetDB); err != nil {
			s.log.Warnf("Gagal menyalin hak akses user: %v", err)
		}
	}

	return opts.TargetDB, nil
}


// sanitizeArgsForData membuang flag yang tidak diinginkan untuk fase penyalinan data murni.
func (s *Service) sanitizeArgsForData(args string) string {
	fields := strings.Fields(args)
	var filtered []string
	skipNext := false
	for i, f := range fields {
		if skipNext {
			skipNext = false
			continue
		}
		l := strings.ToLower(f)

		// List flags yang harus dibuang bersama nilainya
		if l == "--databases" || l == "-b" {
			skipNext = true
			continue
		}

		// List flags yang harus dibuang (standalone atau prefix)
		if l == "--routines" || l == "-r" ||
			l == "--triggers" || l == "-t" ||
			l == "--events" || l == "-e" ||
			l == "--all-databases" || l == "-a" ||
			strings.HasPrefix(l, "--set-gtid-purged") {
			continue
		}

		// Jika dalam format --databases=db1 (dengan =)
		if strings.HasPrefix(l, "--databases=") {
			continue
		}

		filtered = append(filtered, fields[i])
	}
	return strings.Join(filtered, " ")
}
