// File : internal/restore/setup_single.go
// Deskripsi : Setup untuk restore single database mode
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-30
// Last Modified :  2026-01-05
package restore

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"sfDBTools/internal/app/restore/helpers"
	restoremodel "sfDBTools/internal/app/restore/model"
	"sfDBTools/pkg/helper"
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/ui"
)

// SetupRestoreSession melakukan setup untuk restore single session
func (s *Service) SetupRestoreSession(ctx context.Context) error {
	ui.Headers("Restore Single Database")
	if s.RestoreOpts == nil {
		return fmt.Errorf("opsi single tidak tersedia")
	}

	allowInteractive := !s.RestoreOpts.Force

	if err := s.prepareRestoreSinglePrereqs(ctx, allowInteractive); err != nil {
		return err
	}

	s.warnRestoreSingle()

	if s.RestoreOpts.Force {
		return nil
	}

	return s.confirmRestoreSingleLoop()
}

func (s *Service) prepareRestoreSinglePrereqs(ctx context.Context, allowInteractive bool) error {
	if err := s.resolveTargetProfile(&s.RestoreOpts.Profile, allowInteractive); err != nil {
		return fmt.Errorf("gagal resolve target profile: %w", err)
	}

	if err := s.connectToTargetDatabase(ctx); err != nil {
		return fmt.Errorf("gagal koneksi ke database target: %w", err)
	}

	if err := s.resolveBackupFile(&s.RestoreOpts.File, allowInteractive); err != nil {
		return fmt.Errorf("gagal resolve file backup: %w", err)
	}

	if err := s.resolveEncryptionKey(s.RestoreOpts.File, &s.RestoreOpts.EncryptionKey, allowInteractive); err != nil {
		return fmt.Errorf("gagal resolve encryption key: %w", err)
	}

	if err := s.resolveTargetDatabaseSingle(ctx); err != nil {
		return fmt.Errorf("gagal resolve target database: %w", err)
	}

	if err := s.promptSkipGrantsSingle(allowInteractive); err != nil {
		return err
	}

	if !s.RestoreOpts.SkipGrants {
		if err := s.resolveGrantsFile(&s.RestoreOpts.SkipGrants, &s.RestoreOpts.GrantsFile, s.RestoreOpts.File, allowInteractive, s.RestoreOpts.StopOnError); err != nil {
			return fmt.Errorf("gagal resolve grants file: %w", err)
		}
	} else {
		s.Log.Info("Skip restore user grants (skip-grants=true)")
	}

	if err := s.resolveInteractiveSafetyOptions(&s.RestoreOpts.DropTarget, &s.RestoreOpts.SkipBackup, allowInteractive); err != nil {
		return err
	}

	if !s.RestoreOpts.SkipBackup {
		if s.RestoreOpts.BackupOptions == nil {
			s.RestoreOpts.BackupOptions = &restoremodel.RestoreBackupOptions{}
		}
		s.setupBackupOptions(s.RestoreOpts.BackupOptions, s.RestoreOpts.EncryptionKey, allowInteractive)
	}

	if err := s.resolveTicketNumber(&s.RestoreOpts.Ticket, allowInteractive); err != nil {
		return fmt.Errorf("gagal resolve ticket number: %w", err)
	}

	return nil
}

func (s *Service) promptSkipGrantsSingle(allowInteractive bool) error {
	if !allowInteractive {
		return nil
	}

	skip, err := input.AskYesNo("Skip restore user grants?", s.RestoreOpts.SkipGrants)
	if err != nil {
		return fmt.Errorf("gagal resolve opsi skip-grants: %w", err)
	}
	s.RestoreOpts.SkipGrants = skip
	return nil
}

func (s *Service) warnRestoreSingle() {
	if s.RestoreOpts.Force || s.RestoreOpts.DryRun {
		return
	}

	ui.PrintWarning("⚠️  Restore single database akan menimpa target jika drop-target aktif")
}

func (s *Service) confirmRestoreSingleLoop() error {
	for {
		ui.PrintSubHeader("Konfirmasi Restore")
		ui.FormatTable([]string{"Parameter", "Value"}, s.restoreSingleSummaryRows())

		action, err := input.SelectSingleFromList(
			[]string{"Lanjutkan", "Ubah opsi", "Batalkan"},
			"Pilih aksi",
		)
		if err != nil {
			return fmt.Errorf("gagal memilih aksi konfirmasi: %w", err)
		}

		switch action {
		case "Lanjutkan":
			if strings.TrimSpace(s.RestoreOpts.Ticket) == "" {
				ui.PrintError("Ticket number wajib diisi sebelum melanjutkan.")
				continue
			}
			return nil
		case "Batalkan":
			return fmt.Errorf("restore dibatalkan oleh user")
		case "Ubah opsi":
			if err := s.editRestoreSingleOptionsInteractive(); err != nil {
				return err
			}
		}
	}
}

func (s *Service) restoreSingleSummaryRows() [][]string {
	rows := [][]string{
		{"Source File", filepath.Base(s.RestoreOpts.File)},
		{"Target Host", fmt.Sprintf("%s:%d", s.Profile.DBInfo.Host, s.Profile.DBInfo.Port)},
		{"Target Database", s.RestoreOpts.TargetDB},
		{"Drop Target", fmt.Sprintf("%v", s.RestoreOpts.DropTarget)},
		{"Skip Backup", fmt.Sprintf("%v", s.RestoreOpts.SkipBackup)},
		{"Skip Grants", fmt.Sprintf("%v", s.RestoreOpts.SkipGrants)},
		{"Ticket Number", s.RestoreOpts.Ticket},
	}

	grantsVal := "Tidak ada"
	if s.RestoreOpts.SkipGrants {
		grantsVal = "Skipped"
	} else if s.RestoreOpts.GrantsFile != "" {
		grantsVal = filepath.Base(s.RestoreOpts.GrantsFile)
	}
	rows = append(rows, []string{"Grants File", grantsVal})

	if !s.RestoreOpts.SkipBackup && s.RestoreOpts.BackupOptions != nil {
		rows = append(rows, []string{"Backup Directory", s.RestoreOpts.BackupOptions.OutputDir})
	}

	return rows
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
