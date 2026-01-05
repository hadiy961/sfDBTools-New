// File : internal/restore/setup_single_edit.go
// Deskripsi : Handler interaktif untuk mengubah opsi restore-single
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

func (s *Service) editRestoreSingleOptionsInteractive() error {
	options := []string{
		"Profile",
		"File Backup",
		"Target Database",
		"File Grants",
		"Nomor Ticket",
		"Drop Target",
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
		return s.changeSingleProfile()
	case "File Backup":
		return s.changeSingleBackupFile()
	case "Target Database":
		return s.changeSingleTargetDatabase()
	case "File Grants":
		return s.changeSingleGrantsFile()
	case "Nomor Ticket":
		return s.changeSingleTicket()
	case "Drop Target":
		return s.changeSingleDropTarget()
	case "Skip Grants":
		return s.changeSingleSkipGrants()
	case "Skip Backup":
		return s.changeSingleSkipBackup()
	case "Pre-backup Directory":
		return s.changeSingleBackupDirectory()
	}

	return nil
}

func (s *Service) changeSingleProfile() error {
	s.RestoreOpts.Profile.Path = ""
	s.RestoreOpts.Profile.EncryptionKey = ""

	if err := s.resolveTargetProfile(&s.RestoreOpts.Profile, true); err != nil {
		return fmt.Errorf("gagal mengubah profile target: %w", err)
	}

	_ = s.Close()
	if err := s.connectToTargetDatabase(context.Background()); err != nil {
		return fmt.Errorf("gagal koneksi ke database target baru: %w", err)
	}
	return nil
}

func (s *Service) changeSingleBackupFile() error {
	s.RestoreOpts.File = ""
	if err := s.resolveBackupFile(&s.RestoreOpts.File, true); err != nil {
		return fmt.Errorf("gagal mengubah file backup: %w", err)
	}
	if err := s.resolveEncryptionKey(s.RestoreOpts.File, &s.RestoreOpts.EncryptionKey, true); err != nil {
		return fmt.Errorf("gagal resolve encryption key: %w", err)
	}
	if s.RestoreOpts.SkipGrants {
		return nil
	}
	if err := s.resolveGrantsFile(&s.RestoreOpts.SkipGrants, &s.RestoreOpts.GrantsFile, s.RestoreOpts.File, true, s.RestoreOpts.StopOnError); err != nil {
		return fmt.Errorf("gagal resolve grants file: %w", err)
	}
	return nil
}

func (s *Service) changeSingleTargetDatabase() error {
	s.RestoreOpts.TargetDB = ""
	if err := s.resolveTargetDatabaseSingle(context.Background()); err != nil {
		return fmt.Errorf("gagal mengubah target database: %w", err)
	}
	return nil
}

func (s *Service) changeSingleGrantsFile() error {
	s.RestoreOpts.SkipGrants = false
	if err := s.resolveGrantsFile(&s.RestoreOpts.SkipGrants, &s.RestoreOpts.GrantsFile, s.RestoreOpts.File, true, s.RestoreOpts.StopOnError); err != nil {
		return fmt.Errorf("gagal mengubah file grants: %w", err)
	}
	return nil
}

func (s *Service) changeSingleTicket() error {
	if err := s.resolveTicketNumber(&s.RestoreOpts.Ticket, true); err != nil {
		return fmt.Errorf("gagal mengubah ticket number: %w", err)
	}
	return nil
}

func (s *Service) changeSingleDropTarget() error {
	val, err := input.AskYesNo("Drop target database sebelum restore?", s.RestoreOpts.DropTarget)
	if err != nil {
		return fmt.Errorf("gagal mengubah opsi drop-target: %w", err)
	}
	s.RestoreOpts.DropTarget = val
	return nil
}

func (s *Service) changeSingleSkipGrants() error {
	skip, err := input.AskYesNo("Skip restore user grants?", s.RestoreOpts.SkipGrants)
	if err != nil {
		return fmt.Errorf("gagal mengubah opsi skip-grants: %w", err)
	}
	s.RestoreOpts.SkipGrants = skip
	if skip {
		s.RestoreOpts.GrantsFile = ""
		return nil
	}
	if err := s.resolveGrantsFile(&s.RestoreOpts.SkipGrants, &s.RestoreOpts.GrantsFile, s.RestoreOpts.File, true, s.RestoreOpts.StopOnError); err != nil {
		return fmt.Errorf("gagal resolve grants file: %w", err)
	}
	return nil
}

func (s *Service) changeSingleSkipBackup() error {
	skip, err := input.AskYesNo("Skip backup sebelum restore?", s.RestoreOpts.SkipBackup)
	if err != nil {
		return fmt.Errorf("gagal mengubah opsi skip-backup: %w", err)
	}
	s.RestoreOpts.SkipBackup = skip
	if skip {
		return nil
	}

	if s.RestoreOpts.BackupOptions == nil {
		s.RestoreOpts.BackupOptions = &restoremodel.RestoreBackupOptions{}
	}
	s.setupBackupOptions(s.RestoreOpts.BackupOptions, s.RestoreOpts.EncryptionKey, true)
	return nil
}

func (s *Service) changeSingleBackupDirectory() error {
	if s.RestoreOpts.BackupOptions == nil {
		s.RestoreOpts.BackupOptions = &restoremodel.RestoreBackupOptions{}
	}

	current := s.RestoreOpts.BackupOptions.OutputDir
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
	s.RestoreOpts.BackupOptions.OutputDir = dir
	return nil
}
