// File : internal/restore/setup_all.go
// Deskripsi : Setup untuk restore all databases mode
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-30
// Last Modified :  2026-01-05
package restore

import (
	"context"
	"fmt"
	"path/filepath"
	restoremodel "sfDBTools/internal/app/restore/model"
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/ui"
	"strings"
)

// SetupRestoreAllSession melakukan setup untuk restore all databases session
func (s *Service) SetupRestoreAllSession(ctx context.Context) error {
	ui.Headers("Restore All Databases")
	allowInteractive := !s.RestoreAllOpts.Force

	if err := s.prepareRestoreAllPrereqs(ctx, allowInteractive); err != nil {
		return err
	}

	s.warnRestoreAll()

	if s.RestoreAllOpts.Force {
		return nil
	}

	return s.confirmRestoreAllLoop()
}

func (s *Service) prepareRestoreAllPrereqs(ctx context.Context, allowInteractive bool) error {
	if err := s.resolveTargetProfile(&s.RestoreAllOpts.Profile, allowInteractive); err != nil {
		return fmt.Errorf("gagal resolve target profile: %w", err)
	}

	if err := s.connectToTargetDatabase(ctx); err != nil {
		return fmt.Errorf("gagal koneksi ke database target: %w", err)
	}

	if err := s.resolveBackupFile(&s.RestoreAllOpts.File, allowInteractive); err != nil {
		return fmt.Errorf("gagal resolve file backup: %w", err)
	}

	if err := s.resolveEncryptionKey(s.RestoreAllOpts.File, &s.RestoreAllOpts.EncryptionKey, allowInteractive); err != nil {
		return fmt.Errorf("gagal resolve encryption key: %w", err)
	}

	if err := s.promptSkipGrantsIfInteractive(allowInteractive); err != nil {
		return err
	}

	if !s.RestoreAllOpts.SkipGrants {
		if err := s.resolveGrantsFile(&s.RestoreAllOpts.SkipGrants, &s.RestoreAllOpts.GrantsFile, s.RestoreAllOpts.File, allowInteractive, s.RestoreAllOpts.StopOnError); err != nil {
			return fmt.Errorf("gagal resolve grants file: %w", err)
		}
	} else {
		s.Log.Info("Skip restore user grants (skip-grants=true)")
	}

	if err := s.resolveInteractiveSafetyOptions(&s.RestoreAllOpts.DropTarget, &s.RestoreAllOpts.SkipBackup, allowInteractive); err != nil {
		return err
	}

	if !s.RestoreAllOpts.SkipBackup {
		if s.RestoreAllOpts.BackupOptions == nil {
			s.RestoreAllOpts.BackupOptions = &restoremodel.RestoreBackupOptions{}
		}
		s.setupBackupOptions(s.RestoreAllOpts.BackupOptions, s.RestoreAllOpts.EncryptionKey, allowInteractive)
	}

	if allowInteractive {
		defaultContinue := !s.RestoreAllOpts.StopOnError
		cont, err := input.AskYesNo("Lanjutkan meski ada error? (continue-on-error)", defaultContinue)
		if err != nil {
			return fmt.Errorf("gagal resolve opsi continue-on-error: %w", err)
		}
		s.RestoreAllOpts.StopOnError = !cont
	}

	if err := s.resolveTicketNumber(&s.RestoreAllOpts.Ticket, allowInteractive); err != nil {
		return fmt.Errorf("gagal resolve ticket number: %w", err)
	}

	return nil
}

func (s *Service) promptSkipGrantsIfInteractive(allowInteractive bool) error {
	if !allowInteractive {
		return nil
	}

	defaultSkip := s.RestoreAllOpts.SkipGrants
	skip, err := input.AskYesNo("Skip restore user grants?", defaultSkip)
	if err != nil {
		return fmt.Errorf("gagal resolve opsi skip-grants: %w", err)
	}
	s.RestoreAllOpts.SkipGrants = skip
	return nil
}

func (s *Service) warnRestoreAll() {
	if s.RestoreAllOpts.Force || s.RestoreAllOpts.DryRun {
		return
	}

	ui.PrintWarning("⚠️  PERINGATAN: Operasi ini akan restore SEMUA database dari file dump!")
	ui.PrintWarning("    Database yang sudah ada akan ditimpa (jika drop-target aktif)")
	if len(s.RestoreAllOpts.ExcludeDBs) > 0 {
		s.Log.Infof("Database yang akan di-exclude: %v", s.RestoreAllOpts.ExcludeDBs)
	}
	if s.RestoreAllOpts.SkipSystemDBs {
		s.Log.Info("System databases (mysql, sys, information_schema, performance_schema) akan di-skip")
	}
}

func (s *Service) confirmRestoreAllLoop() error {
	for {
		ui.PrintSubHeader("Konfirmasi Restore")
		ui.FormatTable([]string{"Parameter", "Value"}, s.restoreAllSummaryRows())

		action, err := input.SelectSingleFromList(
			[]string{"Lanjutkan", "Ubah opsi", "Batalkan"},
			"Pilih aksi",
		)
		if err != nil {
			return fmt.Errorf("gagal memilih aksi konfirmasi: %w", err)
		}

		switch action {
		case "Lanjutkan":
			if strings.TrimSpace(s.RestoreAllOpts.Ticket) == "" {
				ui.PrintError("Ticket number wajib diisi sebelum melanjutkan.")
				continue
			}
			return nil
		case "Batalkan":
			return fmt.Errorf("restore dibatalkan oleh user")
		case "Ubah opsi":
			if err := s.editRestoreAllOptionsInteractive(); err != nil {
				return err
			}
		}
	}
}

func (s *Service) restoreAllSummaryRows() [][]string {
	rows := [][]string{
		{"Source File", filepath.Base(s.RestoreAllOpts.File)},
		{"Target Host", fmt.Sprintf("%s:%d", s.Profile.DBInfo.Host, s.Profile.DBInfo.Port)},
		{"Skip System DBs", fmt.Sprintf("%v", s.RestoreAllOpts.SkipSystemDBs)},
		{"Drop Target", fmt.Sprintf("%v", s.RestoreAllOpts.DropTarget)},
		{"Skip Backup", fmt.Sprintf("%v", s.RestoreAllOpts.SkipBackup)},
		{"Skip Grants", fmt.Sprintf("%v", s.RestoreAllOpts.SkipGrants)},
		{"Dry Run", fmt.Sprintf("%v", s.RestoreAllOpts.DryRun)},
		{"Continue on Error", fmt.Sprintf("%v", !s.RestoreAllOpts.StopOnError)},
		{"Ticket Number", s.RestoreAllOpts.Ticket},
	}

	grantsVal := "Tidak ada"
	if s.RestoreAllOpts.SkipGrants {
		grantsVal = "Skipped"
	} else if s.RestoreAllOpts.GrantsFile != "" {
		grantsVal = filepath.Base(s.RestoreAllOpts.GrantsFile)
	}
	rows = append(rows, []string{"Grants File", grantsVal})

	if !s.RestoreAllOpts.SkipBackup && s.RestoreAllOpts.BackupOptions != nil {
		rows = append(rows, []string{"Backup Directory", s.RestoreAllOpts.BackupOptions.OutputDir})
	}
	if len(s.RestoreAllOpts.ExcludeDBs) > 0 {
		rows = append(rows, []string{"Excluded DBs", strings.Join(s.RestoreAllOpts.ExcludeDBs, ", ")})
	}

	return rows
}
