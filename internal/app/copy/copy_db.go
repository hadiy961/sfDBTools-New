package copy

import (
	"context"
	"fmt"
	"strings"
	"time"

	copyexec "sfdbtools/internal/app/copy/execution"
	profileconn "sfdbtools/internal/app/profile/connection"
	"sfdbtools/internal/ui/prompt"
)

// CopyDatabaseResult menyimpan status eksekusi kloning per database.
type CopyDatabaseResult struct {
	SourceDB string
	TargetDB string
	Method   string
	Duration time.Duration
	Status   string
	Error    error
}

// CopyDatabases melakukan penyalinan banyak database secara berurutan.
func (s *Service) CopyDatabases(ctx context.Context, opts CopyDatabasesOptions) ([]CopyDatabaseResult, error) {
	var results []CopyDatabaseResult
	methodLabel := "Piping"
	if opts.UseConcurrent {
		methodLabel = "Concurrent"
	}

	for i, db := range opts.SourceDBs {
		// Check for graceful shutdown
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		default:
		}

		currTarget := opts.TargetDBIfSingle
		if len(opts.SourceDBs) > 1 {
			currTarget = db + opts.Suffix
		}

		start := time.Now()
		s.log.Infof("[%d/%d] Kloning %s -> %s [%s]...", i+1, len(opts.SourceDBs), db, currTarget, methodLabel)

		dbOpts := opts.CopyDatabaseOptions
		dbOpts.SourceDB = db
		dbOpts.TargetDB = currTarget

		finalTarget, err := s.CopyDatabase(ctx, dbOpts)
		duration := time.Since(start).Round(time.Second)

		status := "Sukses"
		if err != nil {
			status = "Gagal"
			s.log.Errorf("  ❌ Error: %v", err)
		} else {
			s.log.Infof("  ✅ Berhasil: %s (%s)", finalTarget, duration)
		}

		results = append(results, CopyDatabaseResult{
			SourceDB: db,
			TargetDB: currTarget,
			Method:   methodLabel,
			Duration: duration,
			Status:   status,
			Error:    err,
		})
	}

	return results, nil
}

// CopyDatabase melakukan penyalinan satu database utuh.
func (s *Service) CopyDatabase(ctx context.Context, opts CopyDatabaseOptions) (string, error) {
	client, err := profileconn.ConnectWithProfile(s.cfg, opts.Profile, "")
	if err != nil {
		return "", fmt.Errorf("gagal koneksi ke database: %w", err)
	}
	defer client.Close()

	// Pre-flight checks
	exists, err := client.CheckDatabaseExists(ctx, opts.SourceDB)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", fmt.Errorf("database sumber '%s' tidak ditemukan", opts.SourceDB)
	}

	if opts.TargetDB == "" {
		if opts.NonInteractive {
			opts.TargetDB = fmt.Sprintf("%s_copy_%s", opts.SourceDB, time.Now().Format("20060102"))
			s.log.Infof("Nama database target otomatis: %s", opts.TargetDB)
		} else {
			opts.TargetDB, err = prompt.AskText("Masukkan nama database target:", prompt.WithDefault(fmt.Sprintf("%s_copy_%s", opts.SourceDB, time.Now().Format("20060102"))))
			if err != nil {
				return "", err
			}
		}
	}

	if strings.EqualFold(opts.SourceDB, opts.TargetDB) {
		return "", fmt.Errorf("database target tidak boleh sama dengan database sumber")
	}

	targetExists, err := client.CheckDatabaseExists(ctx, opts.TargetDB)
	if err != nil {
		return "", err
	}

	if targetExists {
		var err error
		opts.Force, opts.BackupFirst, err = s.ConfirmOverwriteInteractive(opts.TargetDB, opts.NonInteractive, opts.Force, opts.BackupFirst)
		if err != nil {
			return "", err
		}

		// Jalankan backup target jika diminta
		if opts.BackupFirst {
			s.log.Infof("Menjalankan backup pengamanan untuk target: %s", opts.TargetDB)
			if err := s.runSafetyBackup(ctx, opts.Profile, client, opts.TargetDB); err != nil {
				return "", fmt.Errorf("gagal melakukan backup pengamanan: %w", err)
			}
		}
	}

	// Route to Concurrent Engine if requested
	if opts.UseConcurrent && !opts.SchemaOnly {
		return s.CopyDatabaseConcurrent(ctx, opts)
	}

	if err := client.CreateDatabaseIfNotExists(ctx, opts.TargetDB); err != nil {
		return "", err
	}

	// Smart Overwrite (Clean up objects if target already exists)
	if targetExists && (opts.Force || opts.BackupFirst) {
		if err := s.SmartDropDatabaseObjects(ctx, client, opts.TargetDB); err != nil {
			s.log.Warnf("Gagal membersihkan database target secara bersih: %v", err)
		}
	}

	methodName := "Piping"
	s.log.Debugf("Memulai copy database: %s -> %s [Metode: %s]", opts.SourceDB, opts.TargetDB, methodName)

	// Execution
	// Discover total tables for progress tracking in Piping mode
	totalTables := 0
	if objects, err := s.DiscoverTablesAndViews(ctx, client, opts.SourceDB); err == nil {
		for _, obj := range objects {
			if obj.Type == TableTypeBaseTable {
				totalTables++
			}
		}
	}

	extraDumpArgs := ""
	if !opts.SkipRoutines {
		extraDumpArgs += " --routines"
	}
	if !opts.SkipEvents {
		extraDumpArgs += " --events"
	}
	if !opts.SkipTriggers {
		extraDumpArgs += " --triggers"
	}

	if err := copyexec.ExecutePiping(ctx, s.log, copyexec.PipingOptions{
		Profile:      opts.Profile,
		SourceDB:     opts.SourceDB,
		TargetDB:     opts.TargetDB,
		SchemaOnly:   opts.SchemaOnly,
		BaseDumpArgs: s.cfg.Backup.MysqlDumpArgs + extraDumpArgs,
		LimitSpeed:   opts.LimitSpeed,
		Force:        opts.Force,
		TotalTables:  totalTables,
	}); err != nil {
		return "", err
	}

	// 6. Copy Grants (if enabled)
	if opts.IncludeGrants {
		if err := s.CopyGrants(ctx, opts.Profile, opts.SourceDB, opts.TargetDB); err != nil {
			s.log.Warnf("Gagal menyalin hak akses user: %v", err)
		}
	}

	// 7. Verify Checksum (for non-concurrent)
	if opts.Verify && !opts.SchemaOnly {
		s.log.Info("Memulai verifikasi database...")
		objects, _ := s.DiscoverTablesAndViews(ctx, client, opts.SourceDB)
		for _, obj := range objects {
			if obj.Type == TableTypeBaseTable {
				ok, _ := s.VerifyChecksum(ctx, client, opts.SourceDB, obj.Name, opts.TargetDB, obj.Name)
				if !ok {
					s.log.Warnf("Checksum mismatch pada tabel %s", obj.Name)
				}
			}
		}
	}

	return opts.TargetDB, nil
}

