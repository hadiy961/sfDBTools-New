// File : internal/backup/setup/single_edit.go
// Deskripsi : Handler interaktif untuk mengubah opsi backup-single
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

func (s *Setup) editBackupSingleOptionsInteractive(ctx context.Context, clientPtr **database.Client, customOutputDir *string) error {
	options := []string{
		"Profile",
		"Ticket number",
		"Database selection",
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
		return s.changeBackupSingleProfile(ctx, clientPtr)
	case "Ticket number":
		return s.changeBackupSingleTicket()
	case "Database selection":
		return s.changeBackupSingleDatabaseSelection()
	case "Export user grants":
		return s.changeBackupSingleExportUserGrants()
	case "Exclude data (schema only)":
		return s.changeBackupSingleExcludeData()
	case "Backup directory":
		return s.changeBackupSingleBackupDirectory(customOutputDir)
	case "Filename":
		return s.changeBackupSingleFilename()
	case "Encryption":
		return s.changeBackupSingleEncryption()
	case "Compression":
		return s.changeBackupSingleCompression()
	}

	return nil
}

func (s *Setup) changeBackupSingleProfile(ctx context.Context, clientPtr **database.Client) error {
	// Paksa pemilihan ulang profile (termasuk prompt untuk memilih file profile)
	s.Options.Profile.Path = ""
	s.Options.Profile.EncryptionKey = ""

	// Reset selection state (berbeda server -> db bisa berubah)
	s.Options.DBName = ""

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

func (s *Setup) changeBackupSingleTicket() error {
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

func (s *Setup) changeBackupSingleDatabaseSelection() error {
	// Reset agar di loop berikutnya selector menampilkan prompt pilihan DB.
	s.Options.DBName = ""
	ui.PrintInfo("Database selection di-reset. Anda akan diminta memilih database lagi.")
	ui.WaitForEnter("Tekan Enter untuk lanjut...")
	return nil
}

func (s *Setup) changeBackupSingleExportUserGrants() error {
	val, err := input.AskYesNo("Export user grants?", !s.Options.ExcludeUser)
	if err != nil {
		return fmt.Errorf("gagal mengubah opsi export user grants: %w", err)
	}
	s.Options.ExcludeUser = !val
	return nil
}

func (s *Setup) changeBackupSingleExcludeData() error {
	val, err := input.AskYesNo("Exclude data (schema only)?", s.Options.Filter.ExcludeData)
	if err != nil {
		return fmt.Errorf("gagal mengubah opsi exclude-data: %w", err)
	}
	s.Options.Filter.ExcludeData = val
	return nil
}

func (s *Setup) changeBackupSingleBackupDirectory(customOutputDir *string) error {
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

func (s *Setup) changeBackupSingleFilename() error {
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

func (s *Setup) changeBackupSingleEncryption() error {
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

func (s *Setup) changeBackupSingleCompression() error {
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
