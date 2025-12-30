// File : internal/backup/setup/all_edit.go
// Deskripsi : Handler interaktif untuk mengubah opsi backup-all
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
)

func (s *Setup) editBackupAllOptionsInteractive(ctx context.Context, clientPtr **database.Client, customOutputDir *string) error {
	options := []string{
		"Profile",
		"Ticket number",
		"Capture GTID",
		"Export user grants",
		"Exclude system databases",
		"Exclude empty databases",
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
		return s.changeBackupAllProfile(ctx, clientPtr)
	case "Ticket number":
		return s.changeBackupAllTicket()
	case "Capture GTID":
		return s.changeBackupAllCaptureGTID()
	case "Export user grants":
		return s.changeBackupAllExportUserGrants()
	case "Exclude system databases":
		return s.changeBackupAllExcludeSystem()
	case "Exclude empty databases":
		return s.changeBackupAllExcludeEmpty()
	case "Exclude data (schema only)":
		return s.changeBackupAllExcludeData()
	case "Backup directory":
		return s.changeBackupAllBackupDirectory(customOutputDir)
	case "Filename":
		return s.changeBackupAllFilename()
	case "Encryption":
		return s.changeBackupAllEncryption()
	case "Compression":
		return s.changeBackupAllCompression()
	}

	return nil
}

func (s *Setup) changeBackupAllProfile(ctx context.Context, clientPtr **database.Client) error {
	// Paksa pemilihan ulang profile (termasuk prompt untuk memilih file profile)
	s.Options.Profile.Path = ""
	s.Options.Profile.EncryptionKey = ""

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

func (s *Setup) changeBackupAllTicket() error {
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

func (s *Setup) changeBackupAllCaptureGTID() error {
	val, err := input.AskYesNo("Capture GTID?", s.Options.CaptureGTID)
	if err != nil {
		return fmt.Errorf("gagal mengubah opsi capture-gtid: %w", err)
	}
	s.Options.CaptureGTID = val
	return nil
}

func (s *Setup) changeBackupAllExportUserGrants() error {
	val, err := input.AskYesNo("Export user grants?", !s.Options.ExcludeUser)
	if err != nil {
		return fmt.Errorf("gagal mengubah opsi export user grants: %w", err)
	}
	s.Options.ExcludeUser = !val
	return nil
}

func (s *Setup) changeBackupAllExcludeSystem() error {
	val, err := input.AskYesNo("Exclude system databases?", s.Options.Filter.ExcludeSystem)
	if err != nil {
		return fmt.Errorf("gagal mengubah opsi exclude-system: %w", err)
	}
	s.Options.Filter.ExcludeSystem = val
	return nil
}

func (s *Setup) changeBackupAllExcludeEmpty() error {
	val, err := input.AskYesNo("Exclude empty databases?", s.Options.Filter.ExcludeEmpty)
	if err != nil {
		return fmt.Errorf("gagal mengubah opsi exclude-empty: %w", err)
	}
	s.Options.Filter.ExcludeEmpty = val
	return nil
}

func (s *Setup) changeBackupAllExcludeData() error {
	val, err := input.AskYesNo("Exclude data (schema only)?", s.Options.Filter.ExcludeData)
	if err != nil {
		return fmt.Errorf("gagal mengubah opsi exclude-data: %w", err)
	}
	s.Options.Filter.ExcludeData = val
	return nil
}

func (s *Setup) changeBackupAllBackupDirectory(customOutputDir *string) error {
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

func (s *Setup) changeBackupAllFilename() error {
	val, err := input.AskString("Custom filename (tanpa ekstensi, kosongkan untuk auto)", s.Options.File.Filename, nil)
	if err != nil {
		return fmt.Errorf("gagal mengubah filename: %w", err)
	}
	s.Options.File.Filename = strings.TrimSpace(val)
	return nil
}

func (s *Setup) changeBackupAllEncryption() error {
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

func (s *Setup) changeBackupAllCompression() error {
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
		// validator survey menerima string (karena AskInt menggunakan Input string)
		v, ok := ans.(string)
		if !ok {
			return fmt.Errorf("input tidak valid")
		}
		v = strings.TrimSpace(v)
		if v == "" {
			return fmt.Errorf("compression level wajib diisi")
		}
		// Parse dilakukan oleh AskInt; kita cukup validasi range via helper
		// tapi butuh int; lakukan parse sederhana di sini.
		parsed := 0
		_, _ = fmt.Sscanf(v, "%d", &parsed)
		if _, err := compress.ValidateCompressionLevel(parsed); err != nil {
			return err
		}
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
