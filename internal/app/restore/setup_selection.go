// File : internal/restore/setup_selection.go
// Deskripsi : Setup untuk restore selection (CSV) mode
// Author : Hadiyatna Muflihun
// Tanggal : 30 Desember 2025
// Last Modified : 26 Januari 2026
package restore

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	backupfile "sfdbtools/internal/app/backup/helpers/file"
	"sfdbtools/internal/app/restore/display"
	restoremodel "sfdbtools/internal/app/restore/model"
	"sfdbtools/internal/shared/runtimecfg"
	"sfdbtools/internal/ui/print"
	"strings"
)

func (s *Service) collectSelectionTargetDBs(csvPath string) ([]string, error) {
	csvPath = strings.TrimSpace(csvPath)
	if csvPath == "" {
		return nil, fmt.Errorf("path CSV wajib diisi (--csv)")
	}

	csvDir := filepath.Dir(csvPath)
	f, err := os.Open(csvPath)
	if err != nil {
		return nil, fmt.Errorf("gagal membuka CSV: %w", err)
	}
	defer f.Close()

	r := csv.NewReader(bufio.NewReader(f))
	r.TrimLeadingSpace = true
	r.FieldsPerRecord = -1

	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("gagal membaca CSV: %w", err)
	}
	if len(records) == 0 {
		return []string{}, nil
	}

	startIdx := 0
	if len(records[0]) >= 1 && strings.EqualFold(strings.Trim(strings.TrimSpace(records[0][0]), " '"), "filename") {
		startIdx = 1
	}

	unique := map[string]struct{}{}

	get := func(rec []string, idx int) string {
		if idx < len(rec) {
			return strings.Trim(strings.TrimSpace(rec[idx]), " '")
		}
		return ""
	}

	for i := startIdx; i < len(records); i++ {
		rec := records[i]
		if len(rec) == 0 {
			continue
		}

		file := get(rec, 0)
		if file == "" {
			continue
		}
		if !filepath.IsAbs(file) {
			file = filepath.Join(csvDir, file)
		}

		dbName := strings.TrimSpace(get(rec, 1))
		if dbName == "" {
			dbName = backupfile.ExtractDatabaseNameFromFile(file)
		}
		dbName = strings.TrimSpace(dbName)
		if dbName == "" {
			continue
		}
		unique[dbName] = struct{}{}
	}

	out := make([]string, 0, len(unique))
	for db := range unique {
		out = append(out, db)
	}
	return out, nil
}

// SetupRestoreSelectionSession melakukan setup untuk restore selection (CSV)
func (s *Service) SetupRestoreSelectionSession(ctx context.Context) error {
	print.PrintAppHeader("Restore Selection (CSV)")
	if s.RestoreSelOpts == nil {
		return fmt.Errorf("opsi selection tidak tersedia")
	}
	nonInteractive := s.RestoreSelOpts.Force || runtimecfg.IsQuiet()
	allowInteractive := !nonInteractive

	// 1. Resolve CSV path (interaktif jika kosong, kecuali mode non-interaktif --skip-confirm/--quiet)
	if err := s.resolveSelectionCSV(&s.RestoreSelOpts.CSV, allowInteractive); err != nil {
		return err
	}

	// 2. Resolve target profile
	if err := s.resolveTargetProfile(&s.RestoreSelOpts.Profile, allowInteractive); err != nil {
		return fmt.Errorf("gagal resolve target profile: %w", err)
	}

	// 3. Connect to target database
	if err := s.connectToTargetDatabase(ctx); err != nil {
		return fmt.Errorf("gagal koneksi ke database target: %w", err)
	}

	// 4. Resolve ticket number
	if err := s.resolveTicketNumber(&s.RestoreSelOpts.Ticket, allowInteractive); err != nil {
		return fmt.Errorf("gagal resolve ticket number: %w", err)
	}

	// 5. Interaktif: pilih backup pre-restore & drop target
	targets, err := s.collectSelectionTargetDBs(s.RestoreSelOpts.CSV)
	if err != nil {
		return err
	}
	if err := s.resolveInteractiveSafetyOptionsForTargets(ctx, targets, &s.RestoreSelOpts.DropTarget, &s.RestoreSelOpts.SkipBackup, allowInteractive); err != nil {
		return err
	}

	// 6. Setup backup options if not skipped
	if !s.RestoreSelOpts.SkipBackup {
		if s.RestoreSelOpts.BackupOptions == nil {
			s.RestoreSelOpts.BackupOptions = &restoremodel.RestoreBackupOptions{}
		}
		// In selection mode, encryption for backup uses profile's encryption by default (if any)
		s.setupBackupOptions(s.RestoreSelOpts.BackupOptions, s.Profile.EncryptionKey, allowInteractive)
	}

	// 7. Confirmation (concise)
	confirmOpts := map[string]string{
		"CSV File":          filepath.Base(s.RestoreSelOpts.CSV),
		"Target Host":       fmt.Sprintf("%s:%d", s.Profile.DBInfo.Host, s.Profile.DBInfo.Port),
		"Drop Target":       fmt.Sprintf("%v", s.RestoreSelOpts.DropTarget),
		"Skip Backup":       fmt.Sprintf("%v", s.RestoreSelOpts.SkipBackup),
		"Dry Run":           fmt.Sprintf("%v", s.RestoreSelOpts.DryRun),
		"Continue on Error": fmt.Sprintf("%v", !s.RestoreSelOpts.StopOnError),
		"Ticket Number":     s.RestoreSelOpts.Ticket,
	}
	if !s.RestoreSelOpts.SkipBackup && s.RestoreSelOpts.BackupOptions != nil {
		confirmOpts["Backup Directory"] = s.RestoreSelOpts.BackupOptions.OutputDir
	}

	if !s.RestoreSelOpts.Force && !runtimecfg.IsQuiet() {
		if err := display.DisplayConfirmation(confirmOpts); err != nil {
			return err
		}
	}

	return nil
}
