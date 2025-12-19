// File : internal/restore/setup.go
// Deskripsi : Setup dan validasi untuk restore operations
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-16
// Last Modified : 2025-12-17

package restore

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sfDBTools/internal/restore/display"
	"sfDBTools/internal/restore/helpers"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/helper"
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/ui"
	"strings"
)

// SetupRestoreSession melakukan setup untuk restore single session
func (s *Service) SetupRestoreSession(ctx context.Context) error {
	ui.PrintHeader("Restore Single Database")

	// 1. Resolve backup file
	if err := s.resolveBackupFile(&s.RestoreOpts.File); err != nil {
		return fmt.Errorf("gagal resolve file backup: %w", err)
	}

	// 2. Resolve encryption key if needed
	if err := s.resolveEncryptionKey(s.RestoreOpts.File, &s.RestoreOpts.EncryptionKey); err != nil {
		return fmt.Errorf("gagal resolve encryption key: %w", err)
	}

	// 3. Resolve target profile
	if err := s.resolveTargetProfile(&s.RestoreOpts.Profile); err != nil {
		return fmt.Errorf("gagal resolve target profile: %w", err)
	}

	// 4. Connect to target database
	if err := s.connectToTargetDatabase(ctx); err != nil {
		return fmt.Errorf("gagal koneksi ke database target: %w", err)
	}

	// 5. Resolve target database name
	if err := s.resolveTargetDatabaseSingle(ctx); err != nil {
		return fmt.Errorf("gagal resolve target database: %w", err)
	}

	// 6. Resolve ticket number
	if err := s.resolveTicketNumber(&s.RestoreOpts.Ticket); err != nil {
		return fmt.Errorf("gagal resolve ticket number: %w", err)
	}

	// 7. Resolve grants file
	if err := s.resolveGrantsFile(s.RestoreOpts.SkipGrants, &s.RestoreOpts.GrantsFile, s.RestoreOpts.File); err != nil {
		return fmt.Errorf("gagal resolve grants file: %w", err)
	}

	// 8. Setup backup options if not skipped
	if !s.RestoreOpts.SkipBackup {
		if s.RestoreOpts.BackupOptions == nil {
			s.RestoreOpts.BackupOptions = &types.RestoreBackupOptions{}
		}
		s.setupBackupOptions(s.RestoreOpts.BackupOptions, s.RestoreOpts.EncryptionKey)
	}

	// 9. Display confirmation
	confirmOpts := map[string]string{
		"Source File":     s.RestoreOpts.File,
		"Target Database": s.RestoreOpts.TargetDB,
		"Target Host":     fmt.Sprintf("%s:%d", s.Profile.DBInfo.Host, s.Profile.DBInfo.Port),
		"Drop Target":     fmt.Sprintf("%v", s.RestoreOpts.DropTarget),
		"Skip Backup":     fmt.Sprintf("%v", s.RestoreOpts.SkipBackup),
		"Ticket Number":   s.RestoreOpts.Ticket,
	}
	if !s.RestoreOpts.SkipBackup && s.RestoreOpts.BackupOptions != nil {
		confirmOpts["Backup Directory"] = s.RestoreOpts.BackupOptions.OutputDir
	}
	if s.RestoreOpts.GrantsFile != "" {
		confirmOpts["Grants File"] = filepath.Base(s.RestoreOpts.GrantsFile)
	} else {
		confirmOpts["Grants File"] = "Tidak ada"
	}

	if !s.RestoreOpts.Force {
		if err := display.DisplayConfirmation(confirmOpts); err != nil {
			return err
		}
	}

	return nil
}

// resolveTargetDatabaseSingle resolve nama database target untuk single mode
func (s *Service) resolveTargetDatabaseSingle(ctx context.Context) error {
	if s.RestoreOpts.TargetDB != "" {
		// Validate not primary database
		if err := helpers.ValidateNotPrimaryDatabase(ctx, s.TargetClient, s.RestoreOpts.TargetDB); err != nil {
			return err
		}
		s.Log.Infof("Target database: %s", s.RestoreOpts.TargetDB)
		return nil
	}

	// Get list of databases
	databases, err := s.TargetClient.GetNonSystemDatabases(ctx)
	if err != nil {
		return fmt.Errorf("gagal mendapatkan list database: %w", err)
	}

	// Extract suggested database name from file
	suggestedDBName := helper.ExtractDatabaseNameFromFile(s.RestoreOpts.File)

	// Build options
	options := []string{}

	if suggestedDBName != "" {
		options = append(options, fmt.Sprintf("📝 [ Buat database baru: %s ] (dari nama file)", suggestedDBName))
	}

	options = append(options, "⌨️  [ Input nama database baru secara manual ]")

	if len(databases) > 0 {
		options = append(options, "────────────────────────")
		options = append(options, databases...)
	}

	// Show selection
	selectedDB, err := input.SelectSingleFromList(options, "Pilih target database untuk restore")
	if err != nil {
		return fmt.Errorf("gagal memilih database: %w", err)
	}

	// Handle selection
	if strings.HasPrefix(selectedDB, "📝 [ Buat database baru:") {
		s.RestoreOpts.TargetDB = suggestedDBName
		s.Log.Infof("Akan membuat database baru: %s", s.RestoreOpts.TargetDB)
	} else if selectedDB == "⌨️  [ Input nama database baru secara manual ]" {
		dbName, err := input.AskString("Masukkan nama database baru: ", suggestedDBName, func(ans interface{}) error {
			str, ok := ans.(string)
			if !ok {
				return fmt.Errorf("input tidak valid")
			}
			if strings.TrimSpace(str) == "" {
				return fmt.Errorf("nama database tidak boleh kosong")
			}
			if !helper.IsValidDatabaseName(str) {
				return fmt.Errorf("nama database hanya boleh berisi huruf, angka, underscore, dan dash")
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("gagal mendapatkan nama database: %w", err)
		}
		s.RestoreOpts.TargetDB = strings.TrimSpace(dbName)
		s.Log.Infof("Akan membuat database baru: %s", s.RestoreOpts.TargetDB)
	} else if selectedDB == "────────────────────────" {
		return errors.New("pilihan tidak valid")
	} else {
		s.RestoreOpts.TargetDB = selectedDB
	}

	// Validate not primary database
	if err := helpers.ValidateNotPrimaryDatabase(ctx, s.TargetClient, s.RestoreOpts.TargetDB); err != nil {
		return err
	}

	s.Log.Infof("Target database: %s", s.RestoreOpts.TargetDB)
	return nil
}

// SetupRestorePrimarySession melakukan setup untuk restore primary session
func (s *Service) SetupRestorePrimarySession(ctx context.Context) error {
	ui.PrintHeader("Restore Primary Database")

	// 1. Resolve backup file
	if err := s.resolveBackupFile(&s.RestorePrimaryOpts.File); err != nil {
		return fmt.Errorf("gagal resolve file backup: %w", err)
	}

	// 2. Resolve encryption key if needed
	if err := s.resolveEncryptionKey(s.RestorePrimaryOpts.File, &s.RestorePrimaryOpts.EncryptionKey); err != nil {
		return fmt.Errorf("gagal resolve encryption key: %w", err)
	}

	// 3. Resolve target profile
	if err := s.resolveTargetProfile(&s.RestorePrimaryOpts.Profile); err != nil {
		return fmt.Errorf("gagal resolve target profile: %w", err)
	}

	// 4. Connect to target database
	if err := s.connectToTargetDatabase(ctx); err != nil {
		return fmt.Errorf("gagal koneksi ke database target: %w", err)
	}

	// 5. Resolve target database name
	if err := s.resolveTargetDatabasePrimary(ctx); err != nil {
		return fmt.Errorf("gagal resolve target database: %w", err)
	}

	// 6. Resolve ticket number
	if err := s.resolveTicketNumber(&s.RestorePrimaryOpts.Ticket); err != nil {
		return fmt.Errorf("gagal resolve ticket number: %w", err)
	}

	// 7. Resolve grants file
	if err := s.resolveGrantsFile(s.RestorePrimaryOpts.SkipGrants, &s.RestorePrimaryOpts.GrantsFile, s.RestorePrimaryOpts.File); err != nil {
		return fmt.Errorf("gagal resolve grants file: %w", err)
	}

	// 8. Validasi password aplikasi
	if err := s.validateApplicationPassword(); err != nil {
		return fmt.Errorf("validasi password aplikasi gagal: %w", err)
	}

	// 9. Setup backup options if not skipped
	if !s.RestorePrimaryOpts.SkipBackup {
		if s.RestorePrimaryOpts.BackupOptions == nil {
			s.RestorePrimaryOpts.BackupOptions = &types.RestoreBackupOptions{}
		}
		s.setupBackupOptions(s.RestorePrimaryOpts.BackupOptions, s.RestorePrimaryOpts.EncryptionKey)
	}

	// 10. Display confirmation
	confirmOpts := map[string]string{
		"Target Profile":  filepath.Base(s.RestorePrimaryOpts.Profile.Path),
		"Database Server": fmt.Sprintf("%s:%d", s.Profile.DBInfo.Host, s.Profile.DBInfo.Port),
		"Target Database": s.RestorePrimaryOpts.TargetDB,
		"Backup File":     filepath.Base(s.RestorePrimaryOpts.File),
		"Ticket Number":   s.RestorePrimaryOpts.Ticket,
		"Drop Target":     fmt.Sprintf("%v", s.RestorePrimaryOpts.DropTarget),
		"Skip Backup":     fmt.Sprintf("%v", s.RestorePrimaryOpts.SkipBackup),
	}

	if s.RestorePrimaryOpts.IncludeDmart {
		companionStatus := "Auto-detect"
		if s.RestorePrimaryOpts.CompanionFile != "" {
			companionStatus = filepath.Base(s.RestorePrimaryOpts.CompanionFile)
		}
		confirmOpts["Companion (dmart)"] = companionStatus
	}

	if s.RestorePrimaryOpts.GrantsFile != "" {
		confirmOpts["Grants File"] = filepath.Base(s.RestorePrimaryOpts.GrantsFile)
	}

	if !s.RestorePrimaryOpts.Force {
		if err := display.DisplayConfirmation(confirmOpts); err != nil {
			return err
		}
	}

	return nil
}

// resolveTargetDatabasePrimary resolve nama database target untuk primary mode
func (s *Service) resolveTargetDatabasePrimary(ctx context.Context) error {
	if s.RestorePrimaryOpts.TargetDB != "" {
		s.Log.Infof("Target database: %s", s.RestorePrimaryOpts.TargetDB)
		return nil
	}

	// Extract from filename
	basename := filepath.Base(s.RestorePrimaryOpts.File)
	dbName := helper.ExtractDatabaseNameFromFile(basename)

	if dbName != "" {
		s.RestorePrimaryOpts.TargetDB = dbName
		s.Log.Infof("Target database (dari filename): %s", dbName)

		// Confirm with user unless force is enabled
		if !s.RestorePrimaryOpts.Force {
			confirm, err := input.PromptConfirm(fmt.Sprintf("Target database: %s. Lanjutkan?", dbName))
			if err != nil {
				return fmt.Errorf("gagal mendapatkan konfirmasi: %w", err)
			}
			if !confirm {
				// Ask for manual input
				dbName, err = input.PromptString("Masukkan nama target database")
				if err != nil {
					return fmt.Errorf("gagal mendapatkan nama database: %w", err)
				}
				s.RestorePrimaryOpts.TargetDB = dbName
			}
		}
	} else {
		// Manual input
		dbName, err := input.PromptString("Masukkan nama target database")
		if err != nil {
			return fmt.Errorf("gagal mendapatkan nama database: %w", err)
		}
		s.RestorePrimaryOpts.TargetDB = dbName
	}

	s.Log.Infof("Target database: %s", s.RestorePrimaryOpts.TargetDB)
	return nil
}

// SetupRestoreAllSession melakukan setup untuk restore all databases session
func (s *Service) SetupRestoreAllSession(ctx context.Context) error {
	ui.Headers("Restore All Databases")

	// 1. Resolve backup file
	if err := s.resolveBackupFile(&s.RestoreAllOpts.File); err != nil {
		return fmt.Errorf("gagal resolve file backup: %w", err)
	}

	// 2. Resolve encryption key if needed
	if err := s.resolveEncryptionKey(s.RestoreAllOpts.File, &s.RestoreAllOpts.EncryptionKey); err != nil {
		return fmt.Errorf("gagal resolve encryption key: %w", err)
	}

	// 3. Resolve target profile
	if err := s.resolveTargetProfile(&s.RestoreAllOpts.Profile); err != nil {
		return fmt.Errorf("gagal resolve target profile: %w", err)
	}

	// 4. Connect to target database
	if err := s.connectToTargetDatabase(ctx); err != nil {
		return fmt.Errorf("gagal koneksi ke database target: %w", err)
	}

	// 5. Resolve ticket number
	if err := s.resolveTicketNumber(&s.RestoreAllOpts.Ticket); err != nil {
		return fmt.Errorf("gagal resolve ticket number: %w", err)
	}

	// 6. Setup backup options if not skipped
	if !s.RestoreAllOpts.SkipBackup {
		if s.RestoreAllOpts.BackupOptions == nil {
			s.RestoreAllOpts.BackupOptions = &types.RestoreBackupOptions{}
		}
		s.setupBackupOptions(s.RestoreAllOpts.BackupOptions, s.RestoreAllOpts.EncryptionKey)
	}

	// 7. Display confirmation
	confirmOpts := map[string]string{
		"Source File":       filepath.Base(s.RestoreAllOpts.File),
		"Target Host":       fmt.Sprintf("%s:%d", s.Profile.DBInfo.Host, s.Profile.DBInfo.Port),
		"Skip System DBs":   fmt.Sprintf("%v", s.RestoreAllOpts.SkipSystemDBs),
		"Skip Backup":       fmt.Sprintf("%v", s.RestoreAllOpts.SkipBackup),
		"Dry Run":           fmt.Sprintf("%v", s.RestoreAllOpts.DryRun),
		"Continue on Error": fmt.Sprintf("%v", !s.RestoreAllOpts.StopOnError),
		"Ticket Number":     s.RestoreAllOpts.Ticket,
	}

	if len(s.RestoreAllOpts.ExcludeDBs) > 0 {
		confirmOpts["Excluded DBs"] = strings.Join(s.RestoreAllOpts.ExcludeDBs, ", ")
	}

	if !s.RestoreAllOpts.SkipBackup && s.RestoreAllOpts.BackupOptions != nil {
		confirmOpts["Backup Directory"] = s.RestoreAllOpts.BackupOptions.OutputDir
	}

	if !s.RestoreAllOpts.Force {
		if err := display.DisplayConfirmation(confirmOpts); err != nil {
			return err
		}
	}

	return nil
}

// SetupRestoreSelectionSession melakukan setup untuk restore selection (CSV)
func (s *Service) SetupRestoreSelectionSession(ctx context.Context) error {
	ui.Headers("Restore Selection (CSV)")

	// 1. Pastikan CSV path terisi
	if s.RestoreSelOpts == nil || strings.TrimSpace(s.RestoreSelOpts.CSV) == "" {
		return fmt.Errorf("path CSV wajib diisi (--csv)")
	}

	// 2. Resolve target profile
	if err := s.resolveTargetProfile(&s.RestoreSelOpts.Profile); err != nil {
		return fmt.Errorf("gagal resolve target profile: %w", err)
	}

	// 3. Connect to target database
	if err := s.connectToTargetDatabase(ctx); err != nil {
		return fmt.Errorf("gagal koneksi ke database target: %w", err)
	}

	// 4. Resolve ticket number
	if err := s.resolveTicketNumber(&s.RestoreSelOpts.Ticket); err != nil {
		return fmt.Errorf("gagal resolve ticket number: %w", err)
	}

	// 5. Setup backup options if not skipped
	if !s.RestoreSelOpts.SkipBackup {
		if s.RestoreSelOpts.BackupOptions == nil {
			s.RestoreSelOpts.BackupOptions = &types.RestoreBackupOptions{}
		}
		// In selection mode, encryption for backup uses profile's encryption by default (if any)
		s.setupBackupOptions(s.RestoreSelOpts.BackupOptions, s.Profile.EncryptionKey)
	}

	// 6. Confirmation (concise)
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
