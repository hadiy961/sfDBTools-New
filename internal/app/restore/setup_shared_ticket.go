// File : internal/restore/setup_shared_ticket.go
// Deskripsi : Helper ticket dan opsi interaktif keamanan restore
// Author : Hadiyatna Muflihun
// Tanggal : 30 Desember 2025
// Last Modified : 26 Januari 2026

package restore

import (
	"context"
	"fmt"
	"strings"

	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/ui/prompt"
)

func (s *Service) resolveTicketNumber(ticket *string, allowInteractive bool) error {
	if strings.TrimSpace(*ticket) == "" {
		if !allowInteractive {
			return fmt.Errorf("ticket number wajib diisi (--ticket) pada mode non-interaktif (--skip-confirm/--quiet)")
		}
		result, err := prompt.AskTicket(consts.FeatureRestore)
		if err != nil {
			return fmt.Errorf("gagal mendapatkan ticket number: %w", err)
		}
		*ticket = result
	}

	s.Log.Infof("Ticket number: %s", *ticket)
	return nil
}

// resolveInteractiveSafetyOptions memberikan opsi interaktif untuk backup pre-restore dan drop target.
// Hanya aktif jika allowInteractive=true (tanpa --skip-confirm/--quiet).
func (s *Service) resolveInteractiveSafetyOptions(dropTarget *bool, skipBackup *bool, allowInteractive bool) error {
	if !allowInteractive {
		return nil
	}

	backupDefault := true
	if skipBackup != nil {
		backupDefault = !*skipBackup
	}
	shouldBackup, err := prompt.Confirm("Lakukan backup sebelum restore?", backupDefault)
	if err != nil {
		return fmt.Errorf("gagal mendapatkan pilihan backup pre-restore: %w", err)
	}
	if skipBackup != nil {
		*skipBackup = !shouldBackup
	}

	dropDefault := true
	if dropTarget != nil {
		dropDefault = *dropTarget
	}
	shouldDrop, err := prompt.Confirm("Drop target database sebelum restore?", dropDefault)
	if err != nil {
		return fmt.Errorf("gagal mendapatkan pilihan drop target: %w", err)
	}
	if dropTarget != nil {
		*dropTarget = shouldDrop
	}

	var dtVal interface{} = "<nil>"
	var sbVal interface{} = "<nil>"
	if dropTarget != nil {
		dtVal = *dropTarget
	}
	if skipBackup != nil {
		sbVal = *skipBackup
	}
	s.Log.Infof("Pilihan interaktif: drop-target=%v, skip-backup=%v", dtVal, sbVal)
	return nil
}

// resolveInteractiveSafetyOptionsForTargets melakukan resolve opsi backup/drop secara interaktif,
// tetapi akan di-skip jika semua target DB belum ada (backup/drop tidak relevan).
//
// Behavior:
// - Jika semua target DB belum ada: set skip-backup=true dan drop-target=false, lalu return tanpa prompt.
// - Jika ada minimal 1 target DB yang sudah ada: gunakan prompt interaktif biasa (jika allowInteractive=true).
func (s *Service) resolveInteractiveSafetyOptionsForTargets(
	ctx context.Context,
	targetDBs []string,
	dropTarget *bool,
	skipBackup *bool,
	allowInteractive bool,
) error {
	// Jika tidak ada target spesifik, fallback ke behavior lama.
	if len(targetDBs) == 0 {
		return s.resolveInteractiveSafetyOptions(dropTarget, skipBackup, allowInteractive)
	}
	if s.TargetClient == nil {
		return fmt.Errorf("target client belum terinisialisasi")
	}

	unique := make(map[string]struct{}, len(targetDBs))
	for _, db := range targetDBs {
		db = strings.TrimSpace(db)
		if db == "" {
			continue
		}
		unique[db] = struct{}{}
	}
	if len(unique) == 0 {
		return s.resolveInteractiveSafetyOptions(dropTarget, skipBackup, allowInteractive)
	}

	existsCount := 0
	for db := range unique {
		ok, err := s.TargetClient.CheckDatabaseExists(ctx, db)
		if err != nil {
			return fmt.Errorf("gagal mengecek database target (%s): %w", db, err)
		}
		if ok {
			existsCount++
		}
	}

	// Semua target DB belum ada: tidak perlu menawarkan backup/drop.
	if existsCount == 0 {
		if skipBackup != nil {
			*skipBackup = true
		}
		if dropTarget != nil {
			*dropTarget = false
		}
		s.Log.Infof(
			"Semua target database belum ada; skip prompt backup pre-restore & drop-target (skip-backup=true, drop-target=false)",
		)
		return nil
	}

	return s.resolveInteractiveSafetyOptions(dropTarget, skipBackup, allowInteractive)
}

func (s *Service) getBackupDirectory(allowInteractive bool) string {
	defaultDir := s.Config.Backup.Output.BaseDirectory
	if defaultDir == "" {
		defaultDir = "./backups"
	}

	if !allowInteractive {
		s.Log.Infof("Direktori backup pre-restore (non-interaktif): %s", defaultDir)
		return defaultDir
	}

	fmt.Println()
	fmt.Println("💾 Backup pre-restore akan dilakukan sebelum restore database")
	fmt.Printf("   Default directory: %s\n", defaultDir)
	fmt.Println()

	backupDir, err := prompt.AskText("Masukkan direktori untuk backup pre-restore (kosongkan untuk default)", prompt.WithDefault(defaultDir))
	if err != nil {
		s.Log.Warnf("Gagal mendapatkan input direktori backup, menggunakan default: %v", err)
		return defaultDir
	}

	backupDir = strings.TrimSpace(backupDir)
	if backupDir == "" {
		backupDir = defaultDir
	}

	s.Log.Infof("Direktori backup pre-restore: %s", backupDir)
	return backupDir
}
