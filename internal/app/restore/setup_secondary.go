// File : internal/restore/setup_secondary.go
// Deskripsi : Setup untuk restore secondary database mode (main setup flow only)
// Author : Hadiyatna Muflihun
// Tanggal : 30 Desember 2025
// Last Modified : 5 Januari 2026
package restore

import (
	"context"
	"fmt"
	"path/filepath"
	restoremodel "sfdbtools/internal/app/restore/model"
	"sfdbtools/internal/ui/print"
	"strings"
)

// SetupRestoreSecondarySession melakukan setup untuk restore secondary session
func (s *Service) SetupRestoreSecondarySession(ctx context.Context) error {
	print.PrintAppHeader("Restore Secondary Database")
	if s.RestoreSecondaryOpts == nil {
		return fmt.Errorf("opsi secondary tidak tersedia")
	}
	allowInteractive := !s.RestoreSecondaryOpts.Force

	// Step 1-5: Resolve dari source, file (jika dari file), profile, connection, client code, encryption
	if err := s.setupSecondaryBasics(ctx, allowInteractive); err != nil {
		return err
	}

	// Step 6-8: Resolve primary DB atau prefix, instance, build target DB
	if err := s.setupSecondaryDatabase(ctx, allowInteractive); err != nil {
		return err
	}

	// Step 9-11: Ticket, safety options, backup options
	return s.finalizeSecondarySetup(ctx, allowInteractive)
}

// setupSecondaryBasics melakukan setup dasar untuk secondary restore
func (s *Service) setupSecondaryBasics(ctx context.Context, allowInteractive bool) error {
	opts := s.RestoreSecondaryOpts

	if err := s.resolveSecondaryFrom(opts, allowInteractive); err != nil {
		return err
	}

	if opts.From == "file" {
		if err := s.resolveBackupFile(&opts.File, allowInteractive); err != nil {
			return fmt.Errorf("gagal resolve file backup: %w", err)
		}
	}

	if err := s.resolveTargetProfile(&opts.Profile, allowInteractive); err != nil {
		return fmt.Errorf("gagal resolve target profile: %w", err)
	}

	if err := s.connectToTargetDatabase(ctx); err != nil {
		return fmt.Errorf("gagal koneksi ke database target: %w", err)
	}

	if err := s.resolveSecondaryClientCode(opts, allowInteractive); err != nil {
		return err
	}

	if err := s.resolveSecondaryEncryptionKey(opts, allowInteractive); err != nil {
		return err
	}

	if err := s.resolveSecondaryCompanionFile(allowInteractive); err != nil {
		return fmt.Errorf("gagal resolve companion (dmart) file: %w", err)
	}

	// Validate companion file encryption key if applicable
	if opts.IncludeDmart && strings.TrimSpace(opts.CompanionFile) != "" {
		if err := s.resolveEncryptionKey(opts.CompanionFile, &opts.EncryptionKey, allowInteractive); err != nil {
			return fmt.Errorf("gagal resolve encryption key untuk dmart file: %w", err)
		}
	}

	return nil
}

// setupSecondaryDatabase setup nama database secondary (primary DB + instance)
func (s *Service) setupSecondaryDatabase(ctx context.Context, allowInteractive bool) error {
	opts := s.RestoreSecondaryOpts
	primaryDB := ""

	if opts.From == "primary" {
		if err := s.resolveSecondaryPrimaryDB(ctx, opts); err != nil {
			return err
		}
		primaryDB = opts.PrimaryDB
	} else {
		prefix, err := s.resolveSecondaryPrefixForFileMode(ctx, opts)
		if err != nil {
			return err
		}
		primaryDB = buildPrimaryTargetDBFromClientCode(prefix, opts.ClientCode)
		opts.PrimaryDB = primaryDB
	}

	if err := s.resolveSecondaryInstance(ctx, opts, primaryDB, allowInteractive); err != nil {
		return err
	}

	opts.TargetDB = secondaryDBName(primaryDB, opts.Instance)
	return nil
}

// finalizeSecondarySetup menyelesaikan setup secondary dengan ticket, safety options, dan confirmation
func (s *Service) finalizeSecondarySetup(ctx context.Context, allowInteractive bool) error {
	opts := s.RestoreSecondaryOpts

	if err := s.resolveTicketNumber(&opts.Ticket, allowInteractive); err != nil {
		return fmt.Errorf("gagal resolve ticket number: %w", err)
	}

	if err := s.resolveInteractiveSafetyOptions(&opts.DropTarget, &opts.SkipBackup, allowInteractive); err != nil {
		return err
	}

	if opts.BackupOptions == nil {
		opts.BackupOptions = &restoremodel.RestoreBackupOptions{}
	}

	if !opts.SkipBackup || opts.From == "primary" {
		s.setupBackupOptions(opts.BackupOptions, opts.EncryptionKey, allowInteractive)
	}

	if opts.File != "" {
		opts.File = filepath.Clean(opts.File)
	}

	return s.displaySecondaryConfirmation()
}

// displaySecondaryConfirmation menampilkan konfirmasi untuk restore secondary
func (s *Service) displaySecondaryConfirmation() error {
	if s.RestoreSecondaryOpts.Force {
		return nil
	}

	opts := s.RestoreSecondaryOpts
	builder := NewConfirmationBuilder().
		Add("From", opts.From).
		Add("Client Code", opts.ClientCode).
		Add("Instance", opts.Instance).
		Add("Primary Database", opts.PrimaryDB).
		Add("Target Database", opts.TargetDB).
		AddHostPort("Target Host", s.Profile.DBInfo.Host, s.Profile.DBInfo.Port).
		AddBool("Drop Target", opts.DropTarget).
		AddBool("Skip Backup", opts.SkipBackup).
		Add("Ticket Number", opts.Ticket)

	if opts.From == "file" {
		builder.AddFile("Source File", opts.File)
	}

	builder.AddCompanion(opts.IncludeDmart, opts.CompanionFile)

	if opts.BackupOptions != nil && (opts.From == "primary" || !opts.SkipBackup) {
		builder.AddBackupDir(opts.SkipBackup, opts.BackupOptions.OutputDir)
	}

	return builder.Display()
}
