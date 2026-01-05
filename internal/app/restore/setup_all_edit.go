// File : internal/restore/setup_all_edit.go
// Deskripsi : Handler interaktif untuk mengubah opsi restore-all
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-30
// Last Modified :  2026-01-05
package restore

import (
	"context"
	"fmt"
	restoremodel "sfDBTools/internal/app/restore/model"
	"sfDBTools/pkg/input"
	"strings"
)

func (s *Service) editRestoreAllOptionsInteractive() error {
	options := []string{
		"Profile",
		"File Backup",
		"File Grants",
		"Nomor Ticket",
		"Drop Target",
		"Continue on Error",
		"Skip Grants",
		"Skip Backup",
		"Pre-backup Directory",
		"Kembali",
	}

	choice, err := input.SelectSingleFromList(options, "Pilih opsi yang ingin diubah")
	if err != nil {
		return fmt.Errorf("gagal memilih opsi untuk diubah: %w", err)
	}

	switch choice {
	case "Kembali":
		return nil
	case "Profile":
		return s.changeAllProfile()
	case "File Backup":
		return s.changeAllBackupFile()
	case "File Grants":
		return s.changeAllGrantsFile()
	case "Nomor Ticket":
		return s.changeAllTicket()
	case "Drop Target":
		return s.changeAllDropTarget()
	case "Continue on Error":
		return s.changeAllContinueOnError()
	case "Skip Grants":
		return s.changeAllSkipGrants()
	case "Skip Backup":
		return s.changeAllSkipBackup()
	case "Pre-backup Directory":
		return s.changeAllBackupDirectory()
	}

	return nil
}

func (s *Service) changeAllProfile() error {
	s.RestoreAllOpts.Profile.Path = ""
	s.RestoreAllOpts.Profile.EncryptionKey = ""

	if err := s.resolveTargetProfile(&s.RestoreAllOpts.Profile, true); err != nil {
		return fmt.Errorf("gagal mengubah profile target: %w", err)
	}

	_ = s.Close()
	if err := s.connectToTargetDatabase(context.Background()); err != nil {
		return fmt.Errorf("gagal koneksi ke database target baru: %w", err)
	}
	return nil
}

func (s *Service) changeAllBackupFile() error {
	s.RestoreAllOpts.File = ""
	if err := s.resolveBackupFile(&s.RestoreAllOpts.File, true); err != nil {
		return fmt.Errorf("gagal mengubah file backup: %w", err)
	}
	if err := s.resolveEncryptionKey(s.RestoreAllOpts.File, &s.RestoreAllOpts.EncryptionKey, true); err != nil {
		return fmt.Errorf("gagal resolve encryption key: %w", err)
	}
	if s.RestoreAllOpts.SkipGrants {
		return nil
	}
	if err := s.resolveGrantsFile(&s.RestoreAllOpts.SkipGrants, &s.RestoreAllOpts.GrantsFile, s.RestoreAllOpts.File, true, s.RestoreAllOpts.StopOnError); err != nil {
		return fmt.Errorf("gagal resolve grants file: %w", err)
	}
	return nil
}

func (s *Service) changeAllGrantsFile() error {
	s.RestoreAllOpts.SkipGrants = false
	if err := s.resolveGrantsFile(&s.RestoreAllOpts.SkipGrants, &s.RestoreAllOpts.GrantsFile, s.RestoreAllOpts.File, true, s.RestoreAllOpts.StopOnError); err != nil {
		return fmt.Errorf("gagal mengubah file grants: %w", err)
	}
	return nil
}

func (s *Service) changeAllTicket() error {
	if err := s.resolveTicketNumber(&s.RestoreAllOpts.Ticket, true); err != nil {
		return fmt.Errorf("gagal mengubah ticket number: %w", err)
	}
	return nil
}

func (s *Service) changeAllDropTarget() error {
	val, err := input.AskYesNo("Drop semua database non-sistem sebelum restore?", s.RestoreAllOpts.DropTarget)
	if err != nil {
		return fmt.Errorf("gagal mengubah opsi drop-target: %w", err)
	}
	s.RestoreAllOpts.DropTarget = val
	return nil
}

func (s *Service) changeAllContinueOnError() error {
	defaultContinue := !s.RestoreAllOpts.StopOnError
	cont, err := input.AskYesNo("Lanjutkan meski ada error? (continue-on-error)", defaultContinue)
	if err != nil {
		return fmt.Errorf("gagal mengubah opsi continue-on-error: %w", err)
	}
	s.RestoreAllOpts.StopOnError = !cont
	return nil
}

func (s *Service) changeAllSkipGrants() error {
	skip, err := input.AskYesNo("Skip restore user grants?", s.RestoreAllOpts.SkipGrants)
	if err != nil {
		return fmt.Errorf("gagal mengubah opsi skip-grants: %w", err)
	}
	s.RestoreAllOpts.SkipGrants = skip
	if skip {
		s.RestoreAllOpts.GrantsFile = ""
		return nil
	}
	if err := s.resolveGrantsFile(&s.RestoreAllOpts.SkipGrants, &s.RestoreAllOpts.GrantsFile, s.RestoreAllOpts.File, true, s.RestoreAllOpts.StopOnError); err != nil {
		return fmt.Errorf("gagal resolve grants file: %w", err)
	}
	return nil
}

func (s *Service) changeAllSkipBackup() error {
	skip, err := input.AskYesNo("Skip backup sebelum restore?", s.RestoreAllOpts.SkipBackup)
	if err != nil {
		return fmt.Errorf("gagal mengubah opsi skip-backup: %w", err)
	}
	s.RestoreAllOpts.SkipBackup = skip
	if skip {
		return nil
	}

	if s.RestoreAllOpts.BackupOptions == nil {
		s.RestoreAllOpts.BackupOptions = &restoremodel.RestoreBackupOptions{}
	}
	s.setupBackupOptions(s.RestoreAllOpts.BackupOptions, s.RestoreAllOpts.EncryptionKey, true)
	return nil
}

func (s *Service) changeAllBackupDirectory() error {
	if s.RestoreAllOpts.BackupOptions == nil {
		s.RestoreAllOpts.BackupOptions = &restoremodel.RestoreBackupOptions{}
	}

	current := s.RestoreAllOpts.BackupOptions.OutputDir
	if strings.TrimSpace(current) == "" {
		current = s.Config.Backup.Output.BaseDirectory
		if current == "" {
			current = "./backups"
		}
	}

	dir, err := input.AskString("Direktori backup pre-restore", current, nil)
	if err != nil {
		return fmt.Errorf("gagal mengubah direktori backup: %w", err)
	}

	dir = strings.TrimSpace(dir)
	if dir == "" {
		dir = current
	}
	s.RestoreAllOpts.BackupOptions.OutputDir = dir
	return nil
}
