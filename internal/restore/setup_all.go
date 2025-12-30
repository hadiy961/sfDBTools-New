// File : internal/restore/setup_all.go
// Deskripsi : Setup untuk restore all databases mode
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-30
// Last Modified : 2025-12-30

package restore

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/ui"
	"strings"
)

// SetupRestoreAllSession melakukan setup untuk restore all databases session
func (s *Service) SetupRestoreAllSession(ctx context.Context) error {
	ui.Headers("Restore All Databases")
	allowInteractive := !s.RestoreAllOpts.Force

	// 1. Pilih profile (target)
	if err := s.resolveTargetProfile(&s.RestoreAllOpts.Profile, allowInteractive); err != nil {
		return fmt.Errorf("gagal resolve target profile: %w", err)
	}

	// 2. Koneksi ke database target
	if err := s.connectToTargetDatabase(ctx); err != nil {
		return fmt.Errorf("gagal koneksi ke database target: %w", err)
	}

	// 3. Pilih file backup
	if err := s.resolveBackupFile(&s.RestoreAllOpts.File, allowInteractive); err != nil {
		return fmt.Errorf("gagal resolve file backup: %w", err)
	}

	// 4. Masukkan backup key jika file terenkripsi
	if err := s.resolveEncryptionKey(s.RestoreAllOpts.File, &s.RestoreAllOpts.EncryptionKey, allowInteractive); err != nil {
		return fmt.Errorf("gagal resolve encryption key: %w", err)
	}

	// 5. Opsi skip grants + pilih grants (jika tidak skip)
	if allowInteractive {
		defaultSkip := s.RestoreAllOpts.SkipGrants
		skip, err := input.AskYesNo("Skip restore user grants?", defaultSkip)
		if err != nil {
			return fmt.Errorf("gagal resolve opsi skip-grants: %w", err)
		}
		s.RestoreAllOpts.SkipGrants = skip
	}

	if !s.RestoreAllOpts.SkipGrants {
		if err := s.resolveGrantsFile(&s.RestoreAllOpts.SkipGrants, &s.RestoreAllOpts.GrantsFile, s.RestoreAllOpts.File, allowInteractive, s.RestoreAllOpts.StopOnError); err != nil {
			return fmt.Errorf("gagal resolve grants file: %w", err)
		}
	} else {
		s.Log.Info("Skip restore user grants (skip-grants=true)")
	}

	// 6. Opsi skip-backup + backup-dir + drop-target
	if err := s.resolveInteractiveSafetyOptions(&s.RestoreAllOpts.DropTarget, &s.RestoreAllOpts.SkipBackup, allowInteractive); err != nil {
		return err
	}

	// Jika tidak skip-backup, tentukan backup dir
	if !s.RestoreAllOpts.SkipBackup {
		if s.RestoreAllOpts.BackupOptions == nil {
			s.RestoreAllOpts.BackupOptions = &types.RestoreBackupOptions{}
		}
		s.setupBackupOptions(s.RestoreAllOpts.BackupOptions, s.RestoreAllOpts.EncryptionKey, allowInteractive)
	}

	// 7. Opsi continue-on-error
	if allowInteractive {
		defaultContinue := !s.RestoreAllOpts.StopOnError
		cont, err := input.AskYesNo("Lanjutkan meski ada error? (continue-on-error)", defaultContinue)
		if err != nil {
			return fmt.Errorf("gagal resolve opsi continue-on-error: %w", err)
		}
		s.RestoreAllOpts.StopOnError = !cont
	}

	// 8. Ticket number
	if err := s.resolveTicketNumber(&s.RestoreAllOpts.Ticket, allowInteractive); err != nil {
		return fmt.Errorf("gagal resolve ticket number: %w", err)
	}

	// Peringatan umum sebelum konfirmasi
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

	// Mode --force: lewati konfirmasi interaktif
	if s.RestoreAllOpts.Force {
		return nil
	}

	// 9. Loop Konfirmasi Restore (Lanjutkan / Ubah opsi / Batalkan)
	for {
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

		ui.PrintSubHeader("Konfirmasi Restore")
		ui.FormatTable([]string{"Parameter", "Value"}, rows)

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
			return errors.New("restore dibatalkan oleh user")
		case "Ubah opsi":
			if err := s.editRestoreAllOptionsInteractive(); err != nil {
				return err
			}
		}
	}
}

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
		// Reset agar user bisa memilih profile baru, bukan hanya me-reload profile lama.
		s.RestoreAllOpts.Profile.Path = ""
		s.RestoreAllOpts.Profile.EncryptionKey = ""
		if err := s.resolveTargetProfile(&s.RestoreAllOpts.Profile, true); err != nil {
			return fmt.Errorf("gagal mengubah profile target: %w", err)
		}
		// Tutup koneksi lama (jika ada) lalu buat koneksi ke target baru.
		_ = s.Close()
		if err := s.connectToTargetDatabase(context.Background()); err != nil {
			return fmt.Errorf("gagal koneksi ke database target baru: %w", err)
		}

	case "File Backup":
		// Reset path agar user selalu diberi kesempatan memilih file baru, bukan hanya validasi file lama.
		s.RestoreAllOpts.File = ""
		if err := s.resolveBackupFile(&s.RestoreAllOpts.File, true); err != nil {
			return fmt.Errorf("gagal mengubah file backup: %w", err)
		}
		if err := s.resolveEncryptionKey(s.RestoreAllOpts.File, &s.RestoreAllOpts.EncryptionKey, true); err != nil {
			return fmt.Errorf("gagal resolve encryption key: %w", err)
		}
		if !s.RestoreAllOpts.SkipGrants {
			if err := s.resolveGrantsFile(&s.RestoreAllOpts.SkipGrants, &s.RestoreAllOpts.GrantsFile, s.RestoreAllOpts.File, true, s.RestoreAllOpts.StopOnError); err != nil {
				return fmt.Errorf("gagal resolve grants file: %w", err)
			}
		}

	case "File Grants":
		s.RestoreAllOpts.SkipGrants = false
		if err := s.resolveGrantsFile(&s.RestoreAllOpts.SkipGrants, &s.RestoreAllOpts.GrantsFile, s.RestoreAllOpts.File, true, s.RestoreAllOpts.StopOnError); err != nil {
			return fmt.Errorf("gagal mengubah file grants: %w", err)
		}

	case "Nomor Ticket":
		if err := s.resolveTicketNumber(&s.RestoreAllOpts.Ticket, true); err != nil {
			return fmt.Errorf("gagal mengubah ticket number: %w", err)
		}

	case "Drop Target":
		val, err := input.AskYesNo("Drop semua database non-sistem sebelum restore?", s.RestoreAllOpts.DropTarget)
		if err != nil {
			return fmt.Errorf("gagal mengubah opsi drop-target: %w", err)
		}
		s.RestoreAllOpts.DropTarget = val

	case "Continue on Error":
		defaultContinue := !s.RestoreAllOpts.StopOnError
		cont, err := input.AskYesNo("Lanjutkan meski ada error? (continue-on-error)", defaultContinue)
		if err != nil {
			return fmt.Errorf("gagal mengubah opsi continue-on-error: %w", err)
		}
		s.RestoreAllOpts.StopOnError = !cont

	case "Skip Grants":
		skip, err := input.AskYesNo("Skip restore user grants?", s.RestoreAllOpts.SkipGrants)
		if err != nil {
			return fmt.Errorf("gagal mengubah opsi skip-grants: %w", err)
		}
		s.RestoreAllOpts.SkipGrants = skip
		if skip {
			s.RestoreAllOpts.GrantsFile = ""
		} else {
			if err := s.resolveGrantsFile(&s.RestoreAllOpts.SkipGrants, &s.RestoreAllOpts.GrantsFile, s.RestoreAllOpts.File, true, s.RestoreAllOpts.StopOnError); err != nil {
				return fmt.Errorf("gagal resolve grants file: %w", err)
			}
		}

	case "Skip Backup":
		skip, err := input.AskYesNo("Skip backup sebelum restore?", s.RestoreAllOpts.SkipBackup)
		if err != nil {
			return fmt.Errorf("gagal mengubah opsi skip-backup: %w", err)
		}
		s.RestoreAllOpts.SkipBackup = skip
		if !skip {
			if s.RestoreAllOpts.BackupOptions == nil {
				s.RestoreAllOpts.BackupOptions = &types.RestoreBackupOptions{}
			}
			s.setupBackupOptions(s.RestoreAllOpts.BackupOptions, s.RestoreAllOpts.EncryptionKey, true)
		}

	case "Pre-backup Directory":
		if s.RestoreAllOpts.BackupOptions == nil {
			s.RestoreAllOpts.BackupOptions = &types.RestoreBackupOptions{}
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
	}

	return nil
}
