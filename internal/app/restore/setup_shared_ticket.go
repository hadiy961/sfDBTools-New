// File : internal/restore/setup_shared_ticket.go
// Deskripsi : Helper ticket dan opsi interaktif keamanan restore
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-30
// Last Modified : 2025-12-30

package restore

import (
	"fmt"
	"strings"

	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/input"
)

func (s *Service) resolveTicketNumber(ticket *string, allowInteractive bool) error {
	if strings.TrimSpace(*ticket) == "" {
		if !allowInteractive {
			return fmt.Errorf("ticket number wajib diisi (--ticket) pada mode non-interaktif (--force)")
		}
		result, err := input.AskTicket(consts.FeatureRestore)
		if err != nil {
			return fmt.Errorf("gagal mendapatkan ticket number: %w", err)
		}
		*ticket = result
	}

	s.Log.Infof("Ticket number: %s", *ticket)
	return nil
}

// resolveInteractiveSafetyOptions memberikan opsi interaktif untuk backup pre-restore dan drop target.
// Hanya aktif jika allowInteractive=true (tanpa --force).
func (s *Service) resolveInteractiveSafetyOptions(dropTarget *bool, skipBackup *bool, allowInteractive bool) error {
	if !allowInteractive {
		return nil
	}

	backupDefault := true
	if skipBackup != nil {
		backupDefault = !*skipBackup
	}
	shouldBackup, err := input.AskYesNo("Lakukan backup sebelum restore?", backupDefault)
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
	shouldDrop, err := input.AskYesNo("Drop target database sebelum restore?", dropDefault)
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

	backupDir, err := input.AskString("Masukkan direktori untuk backup pre-restore (kosongkan untuk default)", defaultDir, nil)
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
