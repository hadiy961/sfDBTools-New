// File : internal/restore/setup_single.go
// Deskripsi : Setup untuk restore single database mode
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-30
// Last Modified : 2025-12-30

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
	ui.Headers("Restore Single Database")
	allowInteractive := !s.RestoreOpts.Force

	// 1. Resolve backup file
	if err := s.resolveBackupFile(&s.RestoreOpts.File, allowInteractive); err != nil {
		return fmt.Errorf("gagal resolve file backup: %w", err)
	}

	// 2. Resolve encryption key if needed
	if err := s.resolveEncryptionKey(s.RestoreOpts.File, &s.RestoreOpts.EncryptionKey, allowInteractive); err != nil {
		return fmt.Errorf("gagal resolve encryption key: %w", err)
	}

	// 3. Resolve target profile
	if err := s.resolveTargetProfile(&s.RestoreOpts.Profile, allowInteractive); err != nil {
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
	if err := s.resolveTicketNumber(&s.RestoreOpts.Ticket, allowInteractive); err != nil {
		return fmt.Errorf("gagal resolve ticket number: %w", err)
	}

	// 7. Interaktif: pilih backup pre-restore & drop target
	if err := s.resolveInteractiveSafetyOptions(&s.RestoreOpts.DropTarget, &s.RestoreOpts.SkipBackup, allowInteractive); err != nil {
		return err
	}

	// 8. Resolve grants file
	if err := s.resolveGrantsFile(&s.RestoreOpts.SkipGrants, &s.RestoreOpts.GrantsFile, s.RestoreOpts.File, allowInteractive, s.RestoreOpts.StopOnError); err != nil {
		return fmt.Errorf("gagal resolve grants file: %w", err)
	}

	// 9. Setup backup options if not skipped
	if !s.RestoreOpts.SkipBackup {
		if s.RestoreOpts.BackupOptions == nil {
			s.RestoreOpts.BackupOptions = &types.RestoreBackupOptions{}
		}
		s.setupBackupOptions(s.RestoreOpts.BackupOptions, s.RestoreOpts.EncryptionKey, allowInteractive)
	}

	// 10. Display confirmation
	confirmOpts := map[string]string{
		"Source File":     s.RestoreOpts.File,
		"Target Database": s.RestoreOpts.TargetDB,
		"Target Host":     fmt.Sprintf("%s:%d", s.Profile.DBInfo.Host, s.Profile.DBInfo.Port),
		"Drop Target":     fmt.Sprintf("%v", s.RestoreOpts.DropTarget),
		"Skip Backup":     fmt.Sprintf("%v", s.RestoreOpts.SkipBackup),
		"Skip Grants":     fmt.Sprintf("%v", s.RestoreOpts.SkipGrants),
		"Ticket Number":   s.RestoreOpts.Ticket,
	}
	if !s.RestoreOpts.SkipBackup && s.RestoreOpts.BackupOptions != nil {
		confirmOpts["Backup Directory"] = s.RestoreOpts.BackupOptions.OutputDir
	}
	if s.RestoreOpts.SkipGrants {
		confirmOpts["Grants File"] = "Skipped"
	} else if s.RestoreOpts.GrantsFile != "" {
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

	if s.RestoreOpts.Force {
		return fmt.Errorf("target database wajib diisi (--target-db) pada mode non-interaktif (--force)")
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
