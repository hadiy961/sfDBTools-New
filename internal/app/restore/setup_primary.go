// File : internal/restore/setup_primary.go
// Deskripsi : Setup untuk restore primary database mode
// Author : Hadiyatna Muflihun
// Tanggal : 30 Desember 2025
// Last Modified : 6 Januari 2026
package restore

import (
	"context"
	"fmt"
	"sfdbtools/internal/app/restore/helpers"
	restoremodel "sfdbtools/internal/app/restore/model"
	"sfdbtools/internal/ui/print"
	"sfdbtools/internal/ui/prompt"
)

// SetupRestorePrimarySession melakukan setup untuk restore primary session
func (s *Service) SetupRestorePrimarySession(ctx context.Context) error {
	print.PrintAppHeader("Restore Primary Database")
	if s.RestorePrimaryOpts == nil {
		return fmt.Errorf("opsi primary tidak tersedia")
	}
	allowInteractive := !s.RestorePrimaryOpts.Force

	// Steps 1-4: File, encryption, profile, dan koneksi
	if err := s.setupBasicRequirements(ctx, &basicSetupOptions{
		file:          &s.RestorePrimaryOpts.File,
		encryptionKey: &s.RestorePrimaryOpts.EncryptionKey,
		profile:       &s.RestorePrimaryOpts.Profile,
		interactive:   allowInteractive,
	}); err != nil {
		return err
	}

	// Step 5: Resolve dan validasi target database primary
	if err := s.resolveAndValidatePrimaryDB(ctx); err != nil {
		return err
	}

	// Steps 6-9: Ticket, safety options, grants, dan companion
	if err := s.setupPostDatabaseOptions(ctx, &postDatabaseSetupOptions{
		ticket:       &s.RestorePrimaryOpts.Ticket,
		dropTarget:   &s.RestorePrimaryOpts.DropTarget,
		skipBackup:   &s.RestorePrimaryOpts.SkipBackup,
		skipGrants:   &s.RestorePrimaryOpts.SkipGrants,
		grantsFile:   &s.RestorePrimaryOpts.GrantsFile,
		backupFile:   s.RestorePrimaryOpts.File,
		stopOnError:  s.RestorePrimaryOpts.StopOnError,
		includeDmart: s.RestorePrimaryOpts.IncludeDmart,
		interactive:  allowInteractive,
	}); err != nil {
		return err
	}

	// Step 10-12: Backup options, password, confirmation
	return s.finalizePrimarySetup(allowInteractive)
}

// finalizePrimarySetup menyelesaikan setup dengan backup options, password, dan confirmation
func (s *Service) finalizePrimarySetup(allowInteractive bool) error {
	if !s.RestorePrimaryOpts.SkipBackup {
		if s.RestorePrimaryOpts.BackupOptions == nil {
			s.RestorePrimaryOpts.BackupOptions = &restoremodel.RestoreBackupOptions{}
		}
		s.setupBackupOptions(s.RestorePrimaryOpts.BackupOptions, s.RestorePrimaryOpts.EncryptionKey, allowInteractive)
	}

	if err := s.validateApplicationPassword(); err != nil {
		return fmt.Errorf("validasi password aplikasi gagal: %w", err)
	}

	return s.displayPrimaryConfirmation()
}

// setupBasicRequirements melakukan setup dasar: file, encryption, profile, dan koneksi
func (s *Service) setupBasicRequirements(ctx context.Context, opts *basicSetupOptions) error {
	if err := s.resolveBackupFile(opts.file, opts.interactive); err != nil {
		return fmt.Errorf("gagal resolve file backup: %w", err)
	}

	if err := s.resolveEncryptionKey(*opts.file, opts.encryptionKey, opts.interactive); err != nil {
		return fmt.Errorf("gagal resolve encryption key: %w", err)
	}

	if err := s.resolveTargetProfile(opts.profile, opts.interactive); err != nil {
		return fmt.Errorf("gagal resolve target profile: %w", err)
	}

	if err := s.connectToTargetDatabase(ctx); err != nil {
		return fmt.Errorf("gagal koneksi ke database target: %w", err)
	}

	return nil
}

// setupPostDatabaseOptions melakukan setup setelah database terdeteksi
func (s *Service) setupPostDatabaseOptions(_ context.Context, opts *postDatabaseSetupOptions) error {
	if err := s.resolveTicketNumber(opts.ticket, opts.interactive); err != nil {
		return fmt.Errorf("gagal resolve ticket number: %w", err)
	}

	if err := s.resolveInteractiveSafetyOptions(opts.dropTarget, opts.skipBackup, opts.interactive); err != nil {
		return err
	}

	if err := s.resolveGrantsFile(opts.skipGrants, opts.grantsFile, opts.backupFile, opts.interactive, opts.stopOnError); err != nil {
		return fmt.Errorf("gagal resolve grants file: %w", err)
	}

	if opts.includeDmart {
		if err := s.DetectOrSelectCompanionFile(); err != nil {
			return fmt.Errorf("gagal deteksi companion database: %w", err)
		}
		if err := s.validateCompanionFile(s.RestorePrimaryOpts, opts.interactive); err != nil {
			return err
		}
	}

	return nil
}

// resolveAndValidatePrimaryDB resolve dan validasi target database untuk primary mode
func (s *Service) resolveAndValidatePrimaryDB(ctx context.Context) error {
	if err := s.resolveTargetDatabasePrimary(ctx); err != nil {
		return fmt.Errorf("gagal resolve target database: %w", err)
	}

	if err := helpers.ValidatePrimaryDatabaseName(s.RestorePrimaryOpts.TargetDB); err != nil {
		if s.RestorePrimaryOpts.Force {
			return err
		}
		return s.retryPrimaryDatabaseInput(err)
	}

	return nil
}

// retryPrimaryDatabaseInput memberikan kesempatan user untuk input ulang nama database primary
func (s *Service) retryPrimaryDatabaseInput(initialErr error) error {
	for {
		retry, askErr := prompt.Confirm("Nama target database tidak valid. Input ulang?", true)
		if askErr != nil || !retry {
			return initialErr
		}

		prefix := inferPrimaryPrefixFromTargetOrFile(s.RestorePrimaryOpts.TargetDB, s.RestorePrimaryOpts.File)
		defaultClientCode := extractClientCodeFromDB(s.RestorePrimaryOpts.TargetDB, prefix)

		newClientCode, inErr := prompt.AskText(
			"Masukkan client-code target (contoh: tes123_tes)",
			prompt.WithDefault(defaultClientCode),
			prompt.WithValidator(validatePrimaryClientCodeInput(prefix)),
		)
		if inErr != nil {
			return fmt.Errorf("gagal mendapatkan nama database: %w", inErr)
		}

		s.RestorePrimaryOpts.TargetDB = buildPrimaryTargetDBFromClientCode(prefix, newClientCode)
		s.Log.Infof("Target database: %s", s.RestorePrimaryOpts.TargetDB)

		if vErr := helpers.ValidatePrimaryDatabaseName(s.RestorePrimaryOpts.TargetDB); vErr == nil {
			break
		} else {
			initialErr = vErr
		}
	}
	return nil
}

// displayPrimaryConfirmation menampilkan konfirmasi untuk restore primary
func (s *Service) displayPrimaryConfirmation() error {
	if s.RestorePrimaryOpts.Force {
		return nil
	}

	builder := NewConfirmationBuilder().
		AddFile("Target Profile", s.RestorePrimaryOpts.Profile.Path).
		AddHostPort("Database Server", s.Profile.DBInfo.Host, s.Profile.DBInfo.Port).
		Add("Target Database", s.RestorePrimaryOpts.TargetDB).
		AddFile("Backup File", s.RestorePrimaryOpts.File).
		Add("Ticket Number", s.RestorePrimaryOpts.Ticket).
		AddBool("Drop Target", s.RestorePrimaryOpts.DropTarget).
		AddBool("Skip Backup", s.RestorePrimaryOpts.SkipBackup).
		AddGrants(s.RestorePrimaryOpts.SkipGrants, s.RestorePrimaryOpts.GrantsFile).
		AddCompanion(s.RestorePrimaryOpts.IncludeDmart, s.RestorePrimaryOpts.CompanionFile)

	return builder.Display()
}

// resolveTargetDatabasePrimary resolve nama database target untuk primary mode
func (s *Service) resolveTargetDatabasePrimary(_ context.Context) error {
	if s.RestorePrimaryOpts.TargetDB != "" {
		s.Log.Infof("Target database: %s", s.RestorePrimaryOpts.TargetDB)
		return nil
	}

	if s.RestorePrimaryOpts.Force {
		return fmt.Errorf("client-code wajib diisi (--client-code) pada mode non-interaktif (--force)")
	}

	defaultClientCode := extractDefaultClientCodeFromFile(s.RestorePrimaryOpts.File)
	prefix := inferPrimaryPrefixFromTargetOrFile("", s.RestorePrimaryOpts.File)

	clientCode, err := prompt.AskText(
		"Masukkan client-code target (contoh: tes123_tes)",
		prompt.WithDefault(defaultClientCode),
		prompt.WithValidator(validatePrimaryClientCodeInput(prefix)),
	)
	if err != nil {
		return fmt.Errorf("gagal mendapatkan client-code: %w", err)
	}

	s.RestorePrimaryOpts.TargetDB = buildPrimaryTargetDBFromClientCode(prefix, clientCode)
	s.Log.Infof("Target database: %s", s.RestorePrimaryOpts.TargetDB)
	return nil
}
