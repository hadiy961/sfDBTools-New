// File : internal/restore/setup_primary.go
// Deskripsi : Setup untuk restore primary database mode
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-30
// Last Modified : 2025-12-30

package restore

import (
	"context"
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
