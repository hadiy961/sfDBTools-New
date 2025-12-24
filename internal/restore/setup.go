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
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/helper"
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/ui"
	"strings"
)

func inferPrimaryPrefixFromTargetOrFile(targetDB string, filePath string) string {
	targetLower := strings.ToLower(strings.TrimSpace(targetDB))
	if strings.HasPrefix(targetLower, consts.PrimaryPrefixBiznet) {
		return consts.PrimaryPrefixBiznet
	}
	if strings.HasPrefix(targetLower, consts.PrimaryPrefixNBC) {
		return consts.PrimaryPrefixNBC
	}

	inferred := helper.ExtractDatabaseNameFromFile(filepath.Base(filePath))
	inferredLower := strings.ToLower(strings.TrimSpace(inferred))
	if strings.HasPrefix(inferredLower, consts.PrimaryPrefixBiznet) {
		return consts.PrimaryPrefixBiznet
	}
	if strings.HasPrefix(inferredLower, consts.PrimaryPrefixNBC) {
		return consts.PrimaryPrefixNBC
	}

	return consts.PrimaryPrefixNBC
}

func buildPrimaryTargetDBFromClientCode(prefix string, clientCode string) string {
	cc := strings.TrimSpace(clientCode)
	ccLower := strings.ToLower(cc)
	if strings.HasPrefix(ccLower, consts.PrimaryPrefixNBC) || strings.HasPrefix(ccLower, consts.PrimaryPrefixBiznet) {
		return cc
	}
	return prefix + cc
}

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

// SetupRestorePrimarySession melakukan setup untuk restore primary session
func (s *Service) SetupRestorePrimarySession(ctx context.Context) error {
	ui.Headers("Restore Primary Database")
	allowInteractive := !s.RestorePrimaryOpts.Force

	// 1. Resolve backup file
	if err := s.resolveBackupFile(&s.RestorePrimaryOpts.File, allowInteractive); err != nil {
		return fmt.Errorf("gagal resolve file backup: %w", err)
	}

	// 2. Resolve encryption key if needed
	if err := s.resolveEncryptionKey(s.RestorePrimaryOpts.File, &s.RestorePrimaryOpts.EncryptionKey, allowInteractive); err != nil {
		return fmt.Errorf("gagal resolve encryption key: %w", err)
	}

	// 3. Resolve target profile
	if err := s.resolveTargetProfile(&s.RestorePrimaryOpts.Profile, allowInteractive); err != nil {
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

	// Restore primary hanya boleh ke database primary (dbsf_nbc_{client} / dbsf_biznet_{client}).
	if err := helpers.ValidatePrimaryDatabaseName(s.RestorePrimaryOpts.TargetDB); err != nil {
		if s.RestorePrimaryOpts.Force {
			return err
		}

		// Mode interaktif: beri kesempatan user untuk input ulang sampai valid.
		for {
			retry, askErr := input.AskYesNo("Nama target database tidak valid. Input ulang?", true)
			if askErr != nil {
				return err
			}
			if !retry {
				return err
			}

			prefix := inferPrimaryPrefixFromTargetOrFile(s.RestorePrimaryOpts.TargetDB, s.RestorePrimaryOpts.File)
			defaultClientCode := strings.ToLower(strings.TrimSpace(s.RestorePrimaryOpts.TargetDB))
			if strings.HasPrefix(defaultClientCode, consts.PrimaryPrefixNBC) {
				defaultClientCode = strings.TrimPrefix(defaultClientCode, consts.PrimaryPrefixNBC)
			} else if strings.HasPrefix(defaultClientCode, consts.PrimaryPrefixBiznet) {
				defaultClientCode = strings.TrimPrefix(defaultClientCode, consts.PrimaryPrefixBiznet)
			}

			newName, inErr := input.AskString(
				"Masukkan client-code target (contoh: tes123_tes)",
				defaultClientCode,
				func(ans interface{}) error {
					str, ok := ans.(string)
					if !ok {
						return fmt.Errorf("input tidak valid")
					}
					str = strings.TrimSpace(str)
					if str == "" {
						return fmt.Errorf("client-code tidak boleh kosong")
					}

					candidate := buildPrimaryTargetDBFromClientCode(prefix, str)
					if !helpers.IsPrimaryDatabaseName(candidate) {
						return fmt.Errorf("client-code menghasilkan nama database primary yang tidak valid")
					}
					return nil
				},
			)
			if inErr != nil {
				return fmt.Errorf("gagal mendapatkan nama database: %w", inErr)
			}
			s.RestorePrimaryOpts.TargetDB = buildPrimaryTargetDBFromClientCode(prefix, newName)
			s.Log.Infof("Target database: %s", s.RestorePrimaryOpts.TargetDB)

			if vErr := helpers.ValidatePrimaryDatabaseName(s.RestorePrimaryOpts.TargetDB); vErr == nil {
				break
			} else {
				err = vErr
			}
		}
	}

	// 6. Resolve ticket number
	if err := s.resolveTicketNumber(&s.RestorePrimaryOpts.Ticket, allowInteractive); err != nil {
		return fmt.Errorf("gagal resolve ticket number: %w", err)
	}

	// 7. Interaktif: pilih backup pre-restore & drop target
	if err := s.resolveInteractiveSafetyOptions(&s.RestorePrimaryOpts.DropTarget, &s.RestorePrimaryOpts.SkipBackup, allowInteractive); err != nil {
		return err
	}

	// 8. Resolve grants file
	if err := s.resolveGrantsFile(&s.RestorePrimaryOpts.SkipGrants, &s.RestorePrimaryOpts.GrantsFile, s.RestorePrimaryOpts.File, allowInteractive, s.RestorePrimaryOpts.StopOnError); err != nil {
		return fmt.Errorf("gagal resolve grants file: %w", err)
	}

	// 9. Resolve companion (dmart) sebelum password/konfirmasi (agar setelah password tidak ada prompt lagi)
	if s.RestorePrimaryOpts.IncludeDmart {
		if err := s.DetectOrSelectCompanionFile(); err != nil {
			return fmt.Errorf("gagal deteksi companion database: %w", err)
		}
	}

	// 10. Setup backup options if not skipped
	if !s.RestorePrimaryOpts.SkipBackup {
		if s.RestorePrimaryOpts.BackupOptions == nil {
			s.RestorePrimaryOpts.BackupOptions = &types.RestoreBackupOptions{}
		}
		s.setupBackupOptions(s.RestorePrimaryOpts.BackupOptions, s.RestorePrimaryOpts.EncryptionKey, allowInteractive)
	}

	// 11. Validasi password aplikasi (setelah semua pemilihan interaktif)
	if err := s.validateApplicationPassword(); err != nil {
		return fmt.Errorf("validasi password aplikasi gagal: %w", err)
	}

	// 12. Display confirmation
	confirmOpts := map[string]string{
		"Target Profile":  filepath.Base(s.RestorePrimaryOpts.Profile.Path),
		"Database Server": fmt.Sprintf("%s:%d", s.Profile.DBInfo.Host, s.Profile.DBInfo.Port),
		"Target Database": s.RestorePrimaryOpts.TargetDB,
		"Backup File":     filepath.Base(s.RestorePrimaryOpts.File),
		"Ticket Number":   s.RestorePrimaryOpts.Ticket,
		"Drop Target":     fmt.Sprintf("%v", s.RestorePrimaryOpts.DropTarget),
		"Skip Backup":     fmt.Sprintf("%v", s.RestorePrimaryOpts.SkipBackup),
		"Skip Grants":     fmt.Sprintf("%v", s.RestorePrimaryOpts.SkipGrants),
	}

	if s.RestorePrimaryOpts.IncludeDmart {
		companionStatus := "Auto-detect"
		if s.RestorePrimaryOpts.CompanionFile != "" {
			companionStatus = filepath.Base(s.RestorePrimaryOpts.CompanionFile)
		}
		confirmOpts["Companion (dmart)"] = companionStatus
	}

	if s.RestorePrimaryOpts.SkipGrants {
		confirmOpts["Grants File"] = "Skipped"
	} else if s.RestorePrimaryOpts.GrantsFile != "" {
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

	// Interaktif: langsung minta client-code, lalu bentuk target DB primary dari pola.
	if s.RestorePrimaryOpts.Force {
		return fmt.Errorf("client-code wajib diisi (--client-code) pada mode non-interaktif (--force)")
	}

	// Default client-code (opsional) diambil dari filename (tanpa konfirmasi & tanpa set sebagai target DB).
	defaultClientCode := ""
	inferredDB := helper.ExtractDatabaseNameFromFile(filepath.Base(s.RestorePrimaryOpts.File))
	inferredLower := strings.ToLower(strings.TrimSpace(inferredDB))
	if strings.HasPrefix(inferredLower, consts.PrimaryPrefixNBC) {
		defaultClientCode = strings.TrimPrefix(inferredLower, consts.PrimaryPrefixNBC)
	} else if strings.HasPrefix(inferredLower, consts.PrimaryPrefixBiznet) {
		defaultClientCode = strings.TrimPrefix(inferredLower, consts.PrimaryPrefixBiznet)
	}

	clientCode, err := input.AskString(
		"Masukkan client-code target (contoh: tes123_tes)",
		defaultClientCode,
		func(ans interface{}) error {
			str, ok := ans.(string)
			if !ok {
				return fmt.Errorf("input tidak valid")
			}
			str = strings.TrimSpace(str)
			if str == "" {
				return fmt.Errorf("client-code tidak boleh kosong")
			}
			prefix := inferPrimaryPrefixFromTargetOrFile("", s.RestorePrimaryOpts.File)
			candidate := buildPrimaryTargetDBFromClientCode(prefix, str)
			if !helpers.IsPrimaryDatabaseName(candidate) {
				return fmt.Errorf("client-code menghasilkan nama database primary yang tidak valid")
			}
			return nil
		},
	)
	if err != nil {
		return fmt.Errorf("gagal mendapatkan client-code: %w", err)
	}

	prefix := inferPrimaryPrefixFromTargetOrFile("", s.RestorePrimaryOpts.File)
	s.RestorePrimaryOpts.TargetDB = buildPrimaryTargetDBFromClientCode(prefix, clientCode)

	s.Log.Infof("Target database: %s", s.RestorePrimaryOpts.TargetDB)
	return nil
}

// SetupRestoreAllSession melakukan setup untuk restore all databases session
func (s *Service) SetupRestoreAllSession(ctx context.Context) error {
	ui.Headers("Restore All Databases")
	allowInteractive := !s.RestoreAllOpts.Force

	// 1. Resolve backup file
	if err := s.resolveBackupFile(&s.RestoreAllOpts.File, allowInteractive); err != nil {
		return fmt.Errorf("gagal resolve file backup: %w", err)
	}

	// 2. Resolve encryption key if needed
	if err := s.resolveEncryptionKey(s.RestoreAllOpts.File, &s.RestoreAllOpts.EncryptionKey, allowInteractive); err != nil {
		return fmt.Errorf("gagal resolve encryption key: %w", err)
	}

	// 3. Resolve target profile
	if err := s.resolveTargetProfile(&s.RestoreAllOpts.Profile, allowInteractive); err != nil {
		return fmt.Errorf("gagal resolve target profile: %w", err)
	}

	// 4. Connect to target database
	if err := s.connectToTargetDatabase(ctx); err != nil {
		return fmt.Errorf("gagal koneksi ke database target: %w", err)
	}

	// 5. Resolve ticket number
	if err := s.resolveTicketNumber(&s.RestoreAllOpts.Ticket, allowInteractive); err != nil {
		return fmt.Errorf("gagal resolve ticket number: %w", err)
	}

	// 6. Interaktif: pilih backup pre-restore & drop target
	if err := s.resolveInteractiveSafetyOptions(&s.RestoreAllOpts.DropTarget, &s.RestoreAllOpts.SkipBackup, allowInteractive); err != nil {
		return err
	}

	// 7. Resolve grants file
	if err := s.resolveGrantsFile(&s.RestoreAllOpts.SkipGrants, &s.RestoreAllOpts.GrantsFile, s.RestoreAllOpts.File, allowInteractive, s.RestoreAllOpts.StopOnError); err != nil {
		return fmt.Errorf("gagal resolve grants file: %w", err)
	}

	// 8. Setup backup options if not skipped
	if !s.RestoreAllOpts.SkipBackup {
		if s.RestoreAllOpts.BackupOptions == nil {
			s.RestoreAllOpts.BackupOptions = &types.RestoreBackupOptions{}
		}
		s.setupBackupOptions(s.RestoreAllOpts.BackupOptions, s.RestoreAllOpts.EncryptionKey, allowInteractive)
	}

	// 9. Display confirmation
	if !s.RestoreAllOpts.Force && !s.RestoreAllOpts.DryRun {
		ui.PrintWarning("⚠️  PERINGATAN: Operasi ini akan restore SEMUA database dari file dump!")
		ui.PrintWarning("    Database yang sudah ada akan ditimpa (jika drop-target aktif)")
		if len(s.RestoreAllOpts.ExcludeDBs) > 0 {
			s.Log.Infof("Database yang akan di-exclude: %v", s.RestoreAllOpts.ExcludeDBs)
		}
		if s.RestoreAllOpts.SkipSystemDBs {
			s.Log.Info("System databases (mysql, sys, information_schema, performance_schema) akan di-skip")
		}
	}

	confirmOpts := map[string]string{
		"Source File":       filepath.Base(s.RestoreAllOpts.File),
		"Target Host":       fmt.Sprintf("%s:%d", s.Profile.DBInfo.Host, s.Profile.DBInfo.Port),
		"Skip System DBs":   fmt.Sprintf("%v", s.RestoreAllOpts.SkipSystemDBs),
		"Drop Target":       fmt.Sprintf("%v", s.RestoreAllOpts.DropTarget),
		"Skip Backup":       fmt.Sprintf("%v", s.RestoreAllOpts.SkipBackup),
		"Skip Grants":       fmt.Sprintf("%v", s.RestoreAllOpts.SkipGrants),
		"Dry Run":           fmt.Sprintf("%v", s.RestoreAllOpts.DryRun),
		"Continue on Error": fmt.Sprintf("%v", !s.RestoreAllOpts.StopOnError),
		"Ticket Number":     s.RestoreAllOpts.Ticket,
	}
	if s.RestoreAllOpts.SkipGrants {
		confirmOpts["Grants File"] = "Skipped"
	} else if s.RestoreAllOpts.GrantsFile != "" {
		confirmOpts["Grants File"] = filepath.Base(s.RestoreAllOpts.GrantsFile)
	} else {
		confirmOpts["Grants File"] = "Tidak ada"
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
	allowInteractive := !s.RestoreSelOpts.Force

	// 1. Resolve CSV path (interaktif jika kosong, kecuali --force)
	if s.RestoreSelOpts == nil {
		return fmt.Errorf("opsi selection tidak tersedia")
	}
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
			s.RestoreSelOpts.BackupOptions = &types.RestoreBackupOptions{}
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
