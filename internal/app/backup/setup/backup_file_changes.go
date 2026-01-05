// File : internal/backup/setup/backup_file_changes.go
// Deskripsi : Kumpulan handler interaktif terkait output file backup (dir, filename, encryption, compression)
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-31
// Last Modified : 2026-01-05

package setup

import (
	"fmt"
	"strconv"
	"strings"

	"sfDBTools/internal/ui/print"
	"sfDBTools/internal/ui/prompt"
	"sfDBTools/pkg/compress"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/validation"
)

func (s *Setup) changeBackupOutputDirInteractive(customOutputDir *string) error {
	current := strings.TrimSpace(s.Options.OutputDir)
	if current == "" {
		current = strings.TrimSpace(s.Config.Backup.Output.BaseDirectory)
		if current == "" {
			current = "."
		}
	}

	val, err := prompt.AskText("Backup directory", prompt.WithDefault(current))
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

func (s *Setup) changeBackupFilenameInteractive() error {
	val, err := prompt.AskText(
		"Custom filename (tanpa ekstensi, kosongkan untuk auto)",
		prompt.WithDefault(s.Options.File.Filename),
		prompt.WithValidator(func(ans interface{}) error {
			v, ok := ans.(string)
			if !ok {
				return fmt.Errorf("input tidak valid")
			}
			return validation.ValidateCustomFilenameBase(v)
		}),
	)
	if err != nil {
		return fmt.Errorf("gagal mengubah filename: %w", err)
	}
	s.Options.File.Filename = strings.TrimSpace(val)
	return nil
}

func (s *Setup) changeBackupEncryptionInteractive() error {
	enabled, err := prompt.Confirm("Encrypt backup file?", s.Options.Encryption.Enabled)
	if err != nil {
		return fmt.Errorf("gagal mengubah opsi encryption: %w", err)
	}

	s.Options.Encryption.Enabled = enabled
	if !enabled {
		s.Options.Encryption.Key = ""
		return nil
	}

	if strings.TrimSpace(s.Options.Encryption.Key) == "" {
		key, err := prompt.AskPassword("Backup Key (required)", nil)
		if err != nil {
			return fmt.Errorf("gagal mendapatkan backup key: %w", err)
		}
		if strings.TrimSpace(key) == "" {
			print.PrintError("Backup key tidak boleh kosong saat encryption aktif.")
			// Fail-safe: jangan biarkan encryption aktif tanpa key.
			s.Options.Encryption.Enabled = false
			s.Options.Encryption.Key = ""
			return nil
		}
		s.Options.Encryption.Key = key
	}

	return nil
}

func (s *Setup) changeBackupCompressionInteractive() error {
	enabled, err := prompt.Confirm("Compress backup file?", s.Options.Compression.Enabled)
	if err != nil {
		return fmt.Errorf("gagal mengubah opsi compression: %w", err)
	}

	if !enabled {
		s.Options.Compression.Enabled = false
		s.Options.Compression.Type = consts.CompressionTypeNone
		return nil
	}

	types := []string{"zstd", "gzip", "xz", "zlib", "pgzip", "none"}
	selected, _, err := prompt.SelectOne("Pilih jenis kompresi", types, -1)
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

	lvl, err := prompt.AskInt("Compression level (1-9)", s.Options.Compression.Level, func(ans interface{}) error {
		v, ok := ans.(string)
		if !ok {
			return fmt.Errorf("input tidak valid")
		}
		v = strings.TrimSpace(v)
		if v == "" {
			return fmt.Errorf("compression level wajib diisi")
		}
		parsed, convErr := strconv.Atoi(v)
		if convErr != nil {
			return fmt.Errorf("compression level harus berupa angka")
		}
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
	// Pastikan level hanya diset jika valid.
	s.Options.Compression.Level = lvl
	return nil
}
