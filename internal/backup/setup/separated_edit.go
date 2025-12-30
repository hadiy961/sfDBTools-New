// File : internal/backup/setup/separated_edit.go
// Deskripsi : Handler interaktif untuk mengubah opsi backup-separated (multi-file)
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-30
// Last Modified : 2025-12-30

package setup

import (
	"context"
	"fmt"
	"strings"
	"time"

	"sfDBTools/internal/backup/selection"
	"sfDBTools/pkg/compress"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/database"
	profilehelper "sfDBTools/pkg/helper/profile"
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/ui"
)

func (s *Setup) editBackupSeparatedOptionsInteractive(ctx context.Context, clientPtr **database.Client, customOutputDir *string) error {
	options := []string{
		"Profile",
		"Ticket number",
		"Pilih database (multi-select)",
		"Reset ke mode interaktif",
		"Include list (manual)",
		"Include file",
		"Export user grants",
		"Exclude system databases",
		"Exclude empty databases",
		"Exclude data (schema only)",
		"Backup directory",
		"Encryption",
		"Compression",
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
		return s.changeBackupSeparatedProfile(ctx, clientPtr)
	case "Ticket number":
		return s.changeBackupSeparatedTicket()
	case "Pilih database (multi-select)":
		return s.changeBackupSeparatedSelectDatabases(ctx, clientPtr)
	case "Reset ke mode interaktif":
		return s.changeBackupSeparatedResetInteractive()
	case "Include list (manual)":
		return s.changeBackupSeparatedIncludeListManual()
	case "Include file":
		return s.changeBackupSeparatedIncludeFile()
	case "Export user grants":
		return s.changeBackupSeparatedExportUserGrants()
	case "Exclude system databases":
		return s.changeBackupSeparatedExcludeSystem()
	case "Exclude empty databases":
		return s.changeBackupSeparatedExcludeEmpty()
	case "Exclude data (schema only)":
		return s.changeBackupSeparatedExcludeData()
	case "Backup directory":
		return s.changeBackupSeparatedBackupDirectory(customOutputDir)
	case "Encryption":
		return s.changeBackupSeparatedEncryption()
	case "Compression":
		return s.changeBackupSeparatedCompression()
	}

	return nil
}

func (s *Setup) changeBackupSeparatedResetInteractive() error {
	s.Options.Filter.IncludeDatabases = nil
	s.Options.Filter.IncludeFile = ""
	ui.PrintInfo("Include selection di-reset. Mode interaktif (multi-select) akan muncul lagi.")
	ui.WaitForEnter("Tekan Enter untuk lanjut...")
	return nil
}

func (s *Setup) changeBackupSeparatedProfile(ctx context.Context, clientPtr **database.Client) error {
	s.Options.Profile.Path = ""
	s.Options.Profile.EncryptionKey = ""

	// Reset include selection state (berbeda server -> db bisa berubah)
	s.Options.Filter.IncludeDatabases = nil
	s.Options.Filter.IncludeFile = ""

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
	*clientPtr = newClient

	serverHostname, err := newClient.GetServerHostname(ctx)
	if err != nil {
		s.Log.Warnf("gagal mendapatkan hostname dari server: %v, menggunakan dari config", err)
		s.Options.Profile.DBInfo.HostName = s.Options.Profile.DBInfo.Host
	} else {
		s.Options.Profile.DBInfo.HostName = serverHostname
		s.Log.Infof("menggunakan hostname dari server: %s", serverHostname)
	}

	return nil
}

func (s *Setup) changeBackupSeparatedTicket() error {
	current := strings.TrimSpace(s.Options.Ticket)
	if current == "" {
		current = fmt.Sprintf("bk-%d", time.Now().UnixNano())
	}
	val, err := input.AskString("Ticket number", current, nil)
	if err != nil {
		return fmt.Errorf("gagal mengubah ticket number: %w", err)
	}
	s.Options.Ticket = strings.TrimSpace(val)
	return nil
}

func (s *Setup) changeBackupSeparatedSelectDatabases(ctx context.Context, clientPtr **database.Client) error {
	if clientPtr == nil || *clientPtr == nil {
		return fmt.Errorf("koneksi database belum tersedia")
	}

	selected, _, err := selection.New(s.Log, s.Options).GetFilteredDatabasesWithMultiSelect(ctx, *clientPtr)
	if err != nil {
		return err
	}

	s.Options.Filter.IncludeDatabases = selected
	s.Options.Filter.IncludeFile = ""
	ui.PrintInfo(fmt.Sprintf("Dipilih %d database untuk backup", len(selected)))
	ui.WaitForEnter("Tekan Enter untuk lanjut...")
	return nil
}

func (s *Setup) changeBackupSeparatedIncludeListManual() error {
	current := strings.Join(s.Options.Filter.IncludeDatabases, ",")
	val, err := input.AskString("Include list (pisahkan dengan koma, kosongkan untuk reset)", current, nil)
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

func (s *Setup) changeBackupSeparatedIncludeFile() error {
	current := strings.TrimSpace(s.Options.Filter.IncludeFile)
	val, err := input.AskString("Include file (path; kosongkan untuk reset)", current, nil)
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

func (s *Setup) changeBackupSeparatedExportUserGrants() error {
	val, err := input.AskYesNo("Export user grants?", !s.Options.ExcludeUser)
	if err != nil {
		return fmt.Errorf("gagal mengubah opsi export user grants: %w", err)
	}
	s.Options.ExcludeUser = !val
	return nil
}

func (s *Setup) changeBackupSeparatedExcludeSystem() error {
	val, err := input.AskYesNo("Exclude system databases?", s.Options.Filter.ExcludeSystem)
	if err != nil {
		return fmt.Errorf("gagal mengubah opsi exclude-system: %w", err)
	}
	s.Options.Filter.ExcludeSystem = val
	return nil
}

func (s *Setup) changeBackupSeparatedExcludeEmpty() error {
	val, err := input.AskYesNo("Exclude empty databases?", s.Options.Filter.ExcludeEmpty)
	if err != nil {
		return fmt.Errorf("gagal mengubah opsi exclude-empty: %w", err)
	}
	s.Options.Filter.ExcludeEmpty = val
	return nil
}

func (s *Setup) changeBackupSeparatedExcludeData() error {
	val, err := input.AskYesNo("Exclude data (schema only)?", s.Options.Filter.ExcludeData)
	if err != nil {
		return fmt.Errorf("gagal mengubah opsi exclude-data: %w", err)
	}
	s.Options.Filter.ExcludeData = val
	return nil
}

func (s *Setup) changeBackupSeparatedBackupDirectory(customOutputDir *string) error {
	current := strings.TrimSpace(s.Options.OutputDir)
	if current == "" {
		current = strings.TrimSpace(s.Config.Backup.Output.BaseDirectory)
		if current == "" {
			current = "."
		}
	}

	val, err := input.AskString("Backup directory", current, nil)
	if err != nil {
		return fmt.Errorf("gagal mengubah backup directory: %w", err)
	}
	val = strings.TrimSpace(val)
	if val == "" {
		val = current
	}

	s.Options.OutputDir = val
	if customOutputDir != nil {
		*customOutputDir = val
	}
	return nil
}

func (s *Setup) changeBackupSeparatedEncryption() error {
	enabled, err := input.AskYesNo("Encrypt backup file?", s.Options.Encryption.Enabled)
	if err != nil {
		return fmt.Errorf("gagal mengubah opsi encryption: %w", err)
	}

	s.Options.Encryption.Enabled = enabled
	if !enabled {
		s.Options.Encryption.Key = ""
		return nil
	}

	if strings.TrimSpace(s.Options.Encryption.Key) == "" {
		key, err := input.AskPassword("Backup Key (required)", nil)
		if err != nil {
			return fmt.Errorf("gagal mendapatkan backup key: %w", err)
		}
		if strings.TrimSpace(key) == "" {
			ui.PrintError("Backup key tidak boleh kosong saat encryption aktif.")
			return nil
		}
		s.Options.Encryption.Key = key
	}

	return nil
}

func (s *Setup) changeBackupSeparatedCompression() error {
	enabled, err := input.AskYesNo("Compress backup file?", s.Options.Compression.Enabled)
	if err != nil {
		return fmt.Errorf("gagal mengubah opsi compression: %w", err)
	}

	if !enabled {
		s.Options.Compression.Enabled = false
		s.Options.Compression.Type = consts.CompressionTypeNone
		return nil
	}

	types := []string{"zstd", "gzip", "xz", "zlib", "pgzip", "none"}
	selected, err := input.SelectSingleFromList(types, "Pilih jenis kompresi")
	if err != nil {
		return fmt.Errorf("gagal memilih compression type: %w", err)
	}

	ct, err := compress.ValidateCompressionType(selected)
	if err != nil {
		return err
	}

	if string(ct) == consts.CompressionTypeNone {
		s.Options.Compression.Enabled = false
		s.Options.Compression.Type = consts.CompressionTypeNone
		return nil
	}

	s.Options.Compression.Enabled = true
	s.Options.Compression.Type = string(ct)

	lvl, err := input.AskInt("Compression level (1-9)", s.Options.Compression.Level, nil)
	if err != nil {
		return fmt.Errorf("gagal mengubah compression level: %w", err)
	}

	if _, err := compress.ValidateCompressionLevel(lvl); err != nil {
		return err
	}
	s.Options.Compression.Level = lvl
	return nil
}
