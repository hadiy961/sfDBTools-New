// File : internal/backup/setup/backup_option_changes.go
// Deskripsi : Kumpulan handler interaktif untuk opsi backup (ticket, toggle, filter, selection, profile)
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-31
// Last Modified : 2026-01-05

package setup

import (
	"context"
	"fmt"
	"strings"
	"time"

	"sfDBTools/internal/app/backup/selection"
	"sfDBTools/internal/ui/print"
	"sfDBTools/internal/ui/prompt"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/database"
	profilehelper "sfDBTools/pkg/helper/profile"
)

func (s *Setup) changeBackupTicketInteractive() error {
	current := strings.TrimSpace(s.Options.Ticket)
	defaultTicket := current
	if defaultTicket == "" {
		defaultTicket = fmt.Sprintf("bk-%d", time.Now().UnixNano())
	}

	val, err := prompt.AskText(
		"Ticket number",
		prompt.WithDefault(defaultTicket),
		prompt.WithValidator(func(ans interface{}) error {
			v, ok := ans.(string)
			if !ok {
				return fmt.Errorf("input tidak valid")
			}
			if strings.TrimSpace(v) == "" {
				return fmt.Errorf("ticket number tidak boleh kosong")
			}
			return nil
		}),
	)
	if err != nil {
		return fmt.Errorf("gagal mengubah ticket number: %w", err)
	}

	s.Options.Ticket = strings.TrimSpace(val)
	return nil
}

func (s *Setup) changeBackupCaptureGTIDInteractive() error {
	val, err := prompt.Confirm("Capture GTID?", s.Options.CaptureGTID)
	if err != nil {
		return fmt.Errorf("gagal mengubah opsi capture-gtid: %w", err)
	}
	s.Options.CaptureGTID = val
	return nil
}

func (s *Setup) changeBackupExportUserGrantsInteractive() error {
	val, err := prompt.Confirm("Export user grants?", !s.Options.ExcludeUser)
	if err != nil {
		return fmt.Errorf("gagal mengubah opsi export user grants: %w", err)
	}
	s.Options.ExcludeUser = !val
	return nil
}

func (s *Setup) changeBackupExcludeSystemInteractive() error {
	val, err := prompt.Confirm("Exclude system databases?", s.Options.Filter.ExcludeSystem)
	if err != nil {
		return fmt.Errorf("gagal mengubah opsi exclude-system: %w", err)
	}
	s.Options.Filter.ExcludeSystem = val
	return nil
}

func (s *Setup) changeBackupExcludeEmptyInteractive() error {
	val, err := prompt.Confirm("Exclude empty databases?", s.Options.Filter.ExcludeEmpty)
	if err != nil {
		return fmt.Errorf("gagal mengubah opsi exclude-empty: %w", err)
	}
	s.Options.Filter.ExcludeEmpty = val
	return nil
}

func (s *Setup) changeBackupExcludeDataInteractive() error {
	val, err := prompt.Confirm("Exclude data (schema only)?", s.Options.Filter.ExcludeData)
	if err != nil {
		return fmt.Errorf("gagal mengubah opsi exclude-data: %w", err)
	}
	s.Options.Filter.ExcludeData = val
	return nil
}

func (s *Setup) changeBackupIncludeDmartInteractive() error {
	val, err := prompt.Confirm("Include DMart?", s.Options.IncludeDmart)
	if err != nil {
		return fmt.Errorf("gagal mengubah opsi include-dmart: %w", err)
	}
	s.Options.IncludeDmart = val
	// Companion status tergantung flag ini.
	s.Options.CompanionStatus = nil
	return nil
}

func (s *Setup) changeBackupClientCodeInteractive() error {
	current := strings.TrimSpace(s.Options.ClientCode)
	val, err := prompt.AskText("Client code (kosongkan untuk nonaktif)", prompt.WithDefault(current))
	if err != nil {
		return fmt.Errorf("gagal mengubah client code: %w", err)
	}
	next := strings.TrimSpace(val)
	if next != current {
		// Jika filter berubah, reset pemilihan DB agar selector melakukan pemilihan ulang.
		s.Options.DBName = ""
		s.Options.CompanionStatus = nil
	}
	s.Options.ClientCode = next
	return nil
}

func (s *Setup) changeBackupDatabaseSelectionResetInteractive(resetCompanion bool) error {
	// Reset agar di loop berikutnya selector menampilkan prompt pilihan DB.
	s.Options.DBName = ""
	if resetCompanion {
		s.Options.CompanionStatus = nil
	}
	print.PrintInfo("Database selection di-reset. Anda akan diminta memilih database lagi.")
	prompt.WaitForEnter("Tekan Enter untuk lanjut...")
	return nil
}

func (s *Setup) changeBackupIncludeSelectionResetInteractive() error {
	s.Options.Filter.IncludeDatabases = nil
	s.Options.Filter.IncludeFile = ""
	print.PrintInfo("Include selection di-reset. Mode interaktif (multi-select) akan muncul lagi.")
	prompt.WaitForEnter("Tekan Enter untuk lanjut...")
	return nil
}

func (s *Setup) changeBackupIncludeSelectionSelectDatabasesInteractive(ctx context.Context, clientPtr **database.Client) error {
	if clientPtr == nil || *clientPtr == nil {
		return fmt.Errorf("koneksi database belum tersedia")
	}

	selected, _, err := selection.New(s.Log, s.Options).GetFilteredDatabasesWithMultiSelect(ctx, *clientPtr)
	if err != nil {
		return err
	}

	s.Options.Filter.IncludeDatabases = selected
	s.Options.Filter.IncludeFile = ""
	print.PrintInfo(fmt.Sprintf("Dipilih %d database untuk backup", len(selected)))
	prompt.WaitForEnter("Tekan Enter untuk lanjut...")
	return nil
}

func (s *Setup) changeBackupIncludeSelectionIncludeListManualInteractive() error {
	current := strings.Join(s.Options.Filter.IncludeDatabases, ",")
	val, err := prompt.AskText("Include list (pisahkan dengan koma, kosongkan untuk reset)", prompt.WithDefault(current))
	if err != nil {
		return fmt.Errorf("gagal mengubah include list: %w", err)
	}
	val = strings.TrimSpace(val)
	if val == "" {
		s.Options.Filter.IncludeDatabases = nil
		return nil
	}

	parts := strings.Split(val, ",")
	include := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			include = append(include, p)
		}
	}
	s.Options.Filter.IncludeDatabases = include
	s.Options.Filter.IncludeFile = ""
	return nil
}

func (s *Setup) changeBackupIncludeSelectionIncludeFileInteractive() error {
	current := strings.TrimSpace(s.Options.Filter.IncludeFile)
	val, err := prompt.AskText("Include file (path; kosongkan untuk reset)", prompt.WithDefault(current))
	if err != nil {
		return fmt.Errorf("gagal mengubah include file: %w", err)
	}
	val = strings.TrimSpace(val)
	s.Options.Filter.IncludeFile = val
	if val != "" {
		s.Options.Filter.IncludeDatabases = nil
	}
	return nil
}

func (s *Setup) changeBackupSecondaryInstance() error {
	current := strings.TrimSpace(s.Options.Instance)
	val, err := prompt.AskText("Instance (contoh: 1, 2, 3; kosongkan untuk nonaktif)", prompt.WithDefault(current))
	if err != nil {
		return fmt.Errorf("gagal mengubah instance: %w", err)
	}
	next := strings.TrimSpace(val)
	if next != current {
		// Jika filter berubah, reset pemilihan DB agar selector melakukan pemilihan ulang.
		s.Options.DBName = ""
		s.Options.CompanionStatus = nil
	}
	s.Options.Instance = next
	return nil
}

// changeBackupProfileAndReconnect mengganti source profile secara interaktif,
// menutup koneksi lama, reconnect, dan memastikan HostName tersinkron.
func (s *Setup) changeBackupProfileAndReconnect(ctx context.Context, clientPtr **database.Client) error {
	// Paksa pemilihan ulang profile (termasuk prompt untuk memilih file profile)
	s.Options.Profile.Path = ""
	s.Options.Profile.EncryptionKey = ""

	// Reset state yang bergantung pada server/profile.
	s.resetBackupStateOnProfileChange()

	if err := s.CheckAndSelectConfigFile(); err != nil {
		return fmt.Errorf("gagal mengubah profile source: %w", err)
	}

	if clientPtr != nil && *clientPtr != nil {
		(*clientPtr).Close()
		*clientPtr = nil
	}

	newClient, err := profilehelper.ConnectWithProfile(&s.Options.Profile, consts.DefaultInitialDatabase)
	if err != nil {
		return fmt.Errorf("gagal koneksi ke database source dengan profile baru: %w", err)
	}
	if clientPtr != nil {
		*clientPtr = newClient
	}

	serverHostname, err := newClient.GetServerHostname(ctx)
	if err != nil {
		s.Log.Warnf("gagal mendapatkan hostname dari server: %v, menggunakan dari config", err)
		s.Options.Profile.DBInfo.HostName = s.Options.Profile.DBInfo.Host
		return nil
	}

	s.Options.Profile.DBInfo.HostName = serverHostname
	s.Log.Infof("menggunakan hostname dari server: %s", serverHostname)
	return nil
}

func (s *Setup) resetBackupStateOnProfileChange() {
	switch s.Options.Mode {
	case consts.ModeSingle:
		s.Options.DBName = ""
	case consts.ModePrimary, consts.ModeSecondary:
		s.Options.DBName = ""
		s.Options.CompanionStatus = nil
	case consts.ModeCombined, consts.ModeSeparated:
		s.Options.Filter.IncludeDatabases = nil
		s.Options.Filter.IncludeFile = ""
	case consts.ModeAll:
		// Tidak ada state khusus untuk di-reset.
	default:
		// Defensive: mode lain tidak dipakai untuk edit interaktif.
	}
}
