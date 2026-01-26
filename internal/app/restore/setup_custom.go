// File : internal/restore/setup_custom.go
// Deskripsi : Setup session untuk restore custom
// Author : Hadiyatna Muflihun
// Tanggal : 24 Desember 2025
// Last Modified : 26 Januari 2026
package restore

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sfdbtools/internal/shared/runtimecfg"
	"strings"

	backupfile "sfdbtools/internal/app/backup/helpers/file"
	"sfdbtools/internal/app/restore/display"
	restoremodel "sfdbtools/internal/app/restore/model"
	"sfdbtools/internal/ui/print"
	"sfdbtools/internal/ui/prompt"
)

func (s *Service) SetupRestoreCustomSession(ctx context.Context) error {
	print.PrintAppHeader("Restore Custom")

	opts := s.GetCustomOptions()
	if opts == nil {
		return fmt.Errorf("opsi custom tidak tersedia")
	}

	// custom tetap butuh input interaktif untuk paste dan pilih file.
	// --skip-confirm/--quiet hanya bypass confirmation.
	allowConfirm := !(opts.Force || runtimecfg.IsQuiet())

	// 1. Resolve target profile
	if err := s.resolveTargetProfile(&opts.Profile, true); err != nil {
		return fmt.Errorf("gagal resolve target profile: %w", err)
	}

	// 2. Connect to target database
	if err := s.connectToTargetDatabase(ctx); err != nil {
		return fmt.Errorf("gagal koneksi ke database target: %w", err)
	}

	// 3. Resolve ticket
	if err := s.resolveTicketNumber(&opts.Ticket, true); err != nil {
		return fmt.Errorf("gagal resolve ticket number: %w", err)
	}

	// 4. Interaktif: pilih backup pre-restore & drop target
	if err := s.resolveInteractiveSafetyOptions(&opts.DropTarget, &opts.SkipBackup, true); err != nil {
		return err
	}

	// 5. Setup backup options if not skipped
	if !opts.SkipBackup {
		if opts.BackupOptions == nil {
			opts.BackupOptions = &restoremodel.RestoreBackupOptions{}
		}
		s.setupBackupOptions(opts.BackupOptions, s.Profile.EncryptionKey, true)
	}

	// 6. Prompt paste account detail
	print.PrintSubHeader("Paste Account Detail")
	fmt.Println("Paste account detail dari SFCola, lalu Enter dan tekan Ctrl+D")

	pasted, err := readMultilineUntilEndToken("END")
	if err != nil {
		return err
	}

	extracted, err := parseSFColaAccountDetail(pasted)
	if err != nil {
		return err
	}

	// Store extracted (jangan log password)
	opts.Database = extracted.Database
	opts.DatabaseDmart = extracted.DatabaseDmart
	opts.UserAdmin = extracted.UserAdmin
	opts.PassAdmin = extracted.PassAdmin
	opts.UserFin = extracted.UserFin
	opts.PassFin = extracted.PassFin
	opts.UserUser = extracted.UserUser
	opts.PassUser = extracted.PassUser

	s.Log.Infof("Custom target db: %s, dmart: %s", opts.Database, opts.DatabaseDmart)
	s.Log.Infof("Custom users: admin=%s fin=%s user=%s", opts.UserAdmin, opts.UserFin, opts.UserUser)

	// 7. Pilih file backup database
	defaultDir := s.Config.Backup.Output.BaseDirectory
	if defaultDir == "" {
		defaultDir = "."
	}
	validExtensions := backupfile.ValidBackupFileExtensionsForSelection()

	dbFile, err := prompt.SelectFile(defaultDir, "Pilih file backup untuk DATABASE", validExtensions)
	if err != nil {
		return fmt.Errorf("gagal memilih file backup database: %w", err)
	}
	if _, err := os.Stat(dbFile); err != nil {
		return fmt.Errorf("file backup database tidak ditemukan: %s", dbFile)
	}
	absDB, err := filepath.Abs(dbFile)
	if err != nil {
		return fmt.Errorf("gagal mendapatkan absolute path file backup database: %w", err)
	}
	opts.DatabaseFile = absDB

	// 8. Pilih file backup dmart
	dmartFile, err := prompt.SelectFile(defaultDir, "Pilih file backup untuk DATABASE DMART", validExtensions)
	if err != nil {
		return fmt.Errorf("gagal memilih file backup dmart: %w", err)
	}
	if _, err := os.Stat(dmartFile); err != nil {
		return fmt.Errorf("file backup dmart tidak ditemukan: %s", dmartFile)
	}
	absDmart, err := filepath.Abs(dmartFile)
	if err != nil {
		return fmt.Errorf("gagal mendapatkan absolute path file backup dmart: %w", err)
	}
	opts.DatabaseDmartFile = absDmart

	// 9. Resolve encryption key jika file terenkripsi (gunakan 1 key untuk keduanya)
	if err := s.resolveEncryptionKey(opts.DatabaseFile, &opts.EncryptionKey, true); err != nil {
		return err
	}
	if err := s.resolveEncryptionKey(opts.DatabaseDmartFile, &opts.EncryptionKey, true); err != nil {
		return err
	}

	// 10. Confirmation (ringkas)
	confirmOpts := map[string]string{
		"Target Host":       fmt.Sprintf("%s:%d", s.Profile.DBInfo.Host, s.Profile.DBInfo.Port),
		"Database":          opts.Database,
		"Database DMART":    opts.DatabaseDmart,
		"DB File":           filepath.Base(opts.DatabaseFile),
		"DMART File":        filepath.Base(opts.DatabaseDmartFile),
		"Drop Target":       fmt.Sprintf("%v", opts.DropTarget),
		"Skip Backup":       fmt.Sprintf("%v", opts.SkipBackup),
		"Dry Run":           fmt.Sprintf("%v", opts.DryRun),
		"Continue on Error": fmt.Sprintf("%v", !opts.StopOnError),
		"Ticket Number":     opts.Ticket,
	}
	if !opts.SkipBackup && opts.BackupOptions != nil {
		confirmOpts["Backup Directory"] = opts.BackupOptions.OutputDir
	}
	if allowConfirm {
		if err := display.DisplayConfirmation(confirmOpts); err != nil {
			return err
		}
	}

	return nil
}

func readMultilineUntilEndToken(endToken string) (string, error) {
	endToken = strings.TrimSpace(endToken)
	r := bufio.NewScanner(os.Stdin)
	var lines []string
	for r.Scan() {
		line := r.Text()
		if strings.TrimSpace(line) == endToken {
			break
		}
		lines = append(lines, line)
	}
	if err := r.Err(); err != nil {
		return "", err
	}
	return strings.Join(lines, "\n"), nil
}
