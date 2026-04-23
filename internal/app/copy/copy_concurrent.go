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
func (s *Service) CopyDatabaseConcurrent(ctx context.Context, profile *domain.ProfileInfo, sourceDB, targetDB string, workers int, limitSpeed int64, force, backupFirst, includeGrants, verify, skipRoutines, skipEvents, skipTriggers, nonInteractive bool) (string, error) {
	client, err := profileconn.ConnectWithProfile(s.cfg, profile, "")
	if err != nil {
		return "", fmt.Errorf("gagal koneksi ke database: %w", err)
	}
	defer client.Close()

	// 1. Discovery Objek
	s.log.Infof("Memulai discovery objek database '%s'…", sourceDB)
	allObjects, err := s.DiscoverTablesAndViews(ctx, client, sourceDB)
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
	if !skipRoutines {
		procNames, funcNames, _ = s.DiscoverRoutines(ctx, client, sourceDB)
	}
	if !skipEvents {
		eventNames, _ = s.DiscoverEvents(ctx, client, sourceDB)
	}
	if !skipTriggers {
		triggerNames, _ = s.DiscoverTriggers(ctx, client, sourceDB)
	}

	// 2. Setup Target
	if err := client.CreateDatabaseIfNotExists(ctx, targetDB); err != nil {
		return "", err
	}
	if force || backupFirst {
		if err := s.SmartDropDatabaseObjects(ctx, client, targetDB); err != nil {
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
						SchemaOnly:   true,
						BaseDumpArgs: s.cfg.Backup.MysqlDumpArgs + " --no-data --skip-triggers --skip-routines --skip-events",
						HideProgress: true,
						Force:        force,
					})

					mu.Lock()
					completedSchema++
					percent := float64(completedSchema) * 100 / float64(totalTables)
					fmt.Printf("\r  ⏳ Progres Struktur: %d/%d tabel (%.1f%%) [Worker %d: %s]   ", completedSchema, totalTables, percent, workerID, tbl)
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
		fmt.Println()

		if len(errChan) > 0 {
			return "", <-errChan
		}
	}

	// 4. Step B: Salin Data (Concurrent)
	if totalTables > 0 {
		s.log.Infof("Menyalin data %d tabel menggunakan %d workers…", totalTables, workers)

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
				limitPerWorker = 1024 * 1024
			}
		}

		var completedData int
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
						HideProgress: true,
						Force:        force,
					})

					mu.Lock()
					completedData++
					percent := float64(completedData) * 100 / float64(totalTables)
					fmt.Printf("\r  ⏳ Progres Data: %d/%d tabel selesai (%.1f%%) [Worker %d: %s]   ", completedData, totalTables, percent, workerID, tbl)
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
		fmt.Println()
		close(errChan)

		if len(errChan) > 0 {
			return "", <-errChan
		}
	}

	// 5. Step C: Salin Objek Pelengkap
	s.log.Info("Memasang objek tambahan (Views, Triggers, Routines, Events)…")

	// 5.1 Salin Triggers (Bulk)
	if !skipTriggers && len(triggerNames) > 0 {
		err = copyexec.ExecutePiping(ctx, s.log, copyexec.PipingOptions{
			Profile:      profile,
			SourceDB:     sourceDB,
			TargetDB:     targetDB,
			SchemaOnly:   true,
			BaseDumpArgs: s.cfg.Backup.MysqlDumpArgs + " --no-create-info --triggers --skip-routines --skip-events",
			Label:        fmt.Sprintf("Memasang %d Triggers", len(triggerNames)),
			Force:        force,
		})
		if err != nil { s.log.Warnf("Gagal menyalin Triggers: %v", err) }
	}

	// 5.2 Salin Procedures & Functions
	if !skipRoutines {
		if len(procNames) > 0 {
			s.log.Infof("Memasang %d Procedures…", len(procNames))
			for i, name := range procNames {
				percent := float64(i+1) * 100 / float64(len(procNames))
				fmt.Printf("\r  ⏳ Progres Procedures: %d/%d (%.1f%%) [%s]           ", i+1, len(procNames), percent, name)
				if err := s.copyIndividualObject(ctx, client, profile, sourceDB, targetDB, "PROCEDURE", name); err != nil {
					s.log.Warnf("\nGagal menyalin procedure %s: %v", name, err)
				}
			}
			fmt.Println()
		}
		if len(funcNames) > 0 {
			s.log.Infof("Memasang %d Functions…", len(funcNames))
			for i, name := range funcNames {
				percent := float64(i+1) * 100 / float64(len(funcNames))
				fmt.Printf("\r  ⏳ Progres Functions: %d/%d (%.1f%%) [%s]            ", i+1, len(funcNames), percent, name)
				if err := s.copyIndividualObject(ctx, client, profile, sourceDB, targetDB, "FUNCTION", name); err != nil {
					s.log.Warnf("\nGagal menyalin function %s: %v", name, err)
				}
			}
			fmt.Println()
		}
	}

	// 5.3 Salin Events
	if !skipEvents && len(eventNames) > 0 {
		s.log.Infof("Memasang %d Events…", len(eventNames))
		for i, name := range eventNames {
			percent := float64(i+1) * 100 / float64(len(eventNames))
			fmt.Printf("\r  ⏳ Progres Events: %d/%d (%.1f%%) [%s]               ", i+1, len(eventNames), percent, name)
			if err := s.copyIndividualObject(ctx, client, profile, sourceDB, targetDB, "EVENT", name); err != nil {
				s.log.Warnf("\nGagal menyalin event %s: %v", name, err)
			}
		}
		fmt.Println()
	}

	// 5.4 Salin Views (Bulk)
	if len(viewNames) > 0 {
		err = copyexec.ExecutePiping(ctx, s.log, copyexec.PipingOptions{
			Profile:      profile,
			SourceDB:     sourceDB,
			TargetDB:     targetDB,
			SchemaOnly:   true,
			BaseDumpArgs: s.cfg.Backup.MysqlDumpArgs + " --no-create-info --skip-triggers --skip-routines --skip-events",
			Label:        fmt.Sprintf("Memasang %d Views", len(viewNames)),
			Force:        force,
		})
		if err != nil { s.log.Warnf("Gagal menyalin Views: %v", err) }
	}

	// 6. Data Integrity Check
	if verify && totalTables > 0 {
		s.log.Info("Memulai verifikasi checksum seluruh tabel…")
		for _, tbl := range baseTables {
			ok, err := s.VerifyChecksum(ctx, client, sourceDB, tbl, targetDB, tbl)
			if err != nil {
				s.log.Warnf("Gagal verifikasi checksum %s: %v", tbl, err)
			} else if !ok {
				s.log.Errorf("Mismatch checksum pada tabel: %s", tbl)
			}
		}
	}

	// 7. Step D: Copy Grants
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
