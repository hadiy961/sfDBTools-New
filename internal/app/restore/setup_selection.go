// File : internal/restore/setup_selection.go
// Deskripsi : Setup untuk restore selection (CSV) mode
// Author : Hadiyatna Muflihun
// Tanggal : 30 Desember 2025
// Last Modified : 5 Januari 2026
package restore

import (
	"context"
	"fmt"
	"path/filepath"
	"sfDBTools/internal/app/restore/display"
	restoremodel "sfDBTools/internal/app/restore/model"
	"sfDBTools/pkg/ui"
)

// SetupRestoreSelectionSession melakukan setup untuk restore selection (CSV)
func (s *Service) SetupRestoreSelectionSession(ctx context.Context) error {
	ui.Headers("Restore Selection (CSV)")
	if s.RestoreSelOpts == nil {
		return fmt.Errorf("opsi selection tidak tersedia")
	}
	allowInteractive := !s.RestoreSelOpts.Force

	// 1. Resolve CSV path (interaktif jika kosong, kecuali --force)
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
	if err := s.resolveInteractiveSafetyOptions(&s.RestoreSelOpts.DropTarget, &s.RestoreSelOpts.SkipBackup, allowInteractive); err != nil {
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

	if !s.RestoreSelOpts.Force {
		if err := display.DisplayConfirmation(confirmOpts); err != nil {
			return err
		}
	}

	return nil
}
