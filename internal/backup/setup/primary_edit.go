// File : internal/backup/setup/primary_edit.go
// Deskripsi : Handler interaktif untuk mengubah opsi backup-primary
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-30
// Last Modified : 2025-12-30

package setup

import (
	"context"
	"fmt"
	"strings"
	"time"

	"sfDBTools/pkg/compress"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/database"
	profilehelper "sfDBTools/pkg/helper/profile"
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/ui"
	"sfDBTools/pkg/validation"
)

func (s *Setup) editBackupPrimaryOptionsInteractive(ctx context.Context, clientPtr **database.Client, customOutputDir *string) error {
	options := []string{
		"Profile",
		"Ticket number",
		"Client code",
		"Database selection",
		"Include DMart",
		"Export user grants",
		"Exclude data (schema only)",
		"Backup directory",
		"Filename",
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
		return s.changeBackupPrimaryProfile(ctx, clientPtr)
	case "Ticket number":
		return s.changeBackupPrimaryTicket()
	case "Client code":
		return s.changeBackupPrimaryClientCode()
	case "Database selection":
		return s.changeBackupPrimaryDatabaseSelection()
	case "Include DMart":
		return s.changeBackupPrimaryIncludeDmart()
	case "Export user grants":
		return s.changeBackupPrimaryExportUserGrants()
	case "Exclude data (schema only)":
		return s.changeBackupPrimaryExcludeData()
	case "Backup directory":
		return s.changeBackupPrimaryBackupDirectory(customOutputDir)
	case "Filename":
		return s.changeBackupPrimaryFilename()
	case "Encryption":
		return s.changeBackupPrimaryEncryption()
	case "Compression":
		return s.changeBackupPrimaryCompression()
	}

	return nil
}

func (s *Setup) changeBackupPrimaryProfile(ctx context.Context, clientPtr **database.Client) error {
	// Paksa pemilihan ulang profile (termasuk prompt untuk memilih file profile)
	s.Options.Profile.Path = ""
	s.Options.Profile.EncryptionKey = ""

	// Reset selection state (berbeda server -> db bisa berubah)
	s.Options.DBName = ""
	s.Options.CompanionStatus = nil

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

func (s *Setup) changeBackupPrimaryTicket() error {
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

func (s *Setup) changeBackupPrimaryClientCode() error {
	current := strings.TrimSpace(s.Options.ClientCode)
	val, err := input.AskString("Client code (kosongkan untuk nonaktif)", current, nil)
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

func (s *Setup) changeBackupPrimaryDatabaseSelection() error {
	// Reset saja agar di loop berikutnya selector menampilkan prompt pilihan DB.
	s.Options.DBName = ""
	s.Options.CompanionStatus = nil
	ui.PrintInfo("Database selection di-reset. Anda akan diminta memilih database lagi.")
	ui.WaitForEnter("Tekan Enter untuk lanjut...")
	return nil
}

func (s *Setup) changeBackupPrimaryIncludeDmart() error {
	val, err := input.AskYesNo("Include DMart?", s.Options.IncludeDmart)
	if err != nil {
		return fmt.Errorf("gagal mengubah opsi include-dmart: %w", err)
	}
	s.Options.IncludeDmart = val
	// Companion status tergantung flag ini.
	s.Options.CompanionStatus = nil
	return nil
}

func (s *Setup) changeBackupPrimaryExportUserGrants() error {
	val, err := input.AskYesNo("Export user grants?", !s.Options.ExcludeUser)
	if err != nil {
		return fmt.Errorf("gagal mengubah opsi export user grants: %w", err)
	}
	s.Options.ExcludeUser = !val
	return nil
}

func (s *Setup) changeBackupPrimaryExcludeData() error {
	val, err := input.AskYesNo("Exclude data (schema only)?", s.Options.Filter.ExcludeData)
	if err != nil {
		return fmt.Errorf("gagal mengubah opsi exclude-data: %w", err)
	}
	s.Options.Filter.ExcludeData = val
	return nil
}

func (s *Setup) changeBackupPrimaryBackupDirectory(customOutputDir *string) error {
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

func (s *Setup) changeBackupPrimaryFilename() error {
	val, err := input.AskString("Custom filename (tanpa ekstensi, kosongkan untuk auto)", s.Options.File.Filename, func(ans interface{}) error {
		v, ok := ans.(string)
		if !ok {
			return fmt.Errorf("input tidak valid")
		}
		return validation.ValidateCustomFilenameBase(v)
	})
	if err != nil {
		return fmt.Errorf("gagal mengubah filename: %w", err)
	}
	s.Options.File.Filename = strings.TrimSpace(val)
	return nil
}

func (s *Setup) changeBackupPrimaryEncryption() error {
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

func (s *Setup) changeBackupPrimaryCompression() error {
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

	lvl, err := input.AskInt("Compression level (1-9)", s.Options.Compression.Level, func(ans interface{}) error {
		v, ok := ans.(string)
		if !ok {
			return fmt.Errorf("input tidak valid")
		}
		v = strings.TrimSpace(v)
		if v == "" {
			return nil
		}
		// AskInt akan mem-parse int sendiri, kita validasi range lewat helper
		return nil
	})
	if err != nil {
		return fmt.Errorf("gagal mengubah compression level: %w", err)
	}

	if _, err := compress.ValidateCompressionLevel(lvl); err != nil {
		return err
	}
	s.Options.Compression.Level = lvl
	return nil
}
