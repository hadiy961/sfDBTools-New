// File : internal/restore/restore_reader.go
// Deskripsi : Reader pipeline helper untuk decrypt dan decompress backup files
// Author : Hadiyatna Muflihun
// Tanggal : 2025-11-11
// Last Modified : 2025-11-14

package restore

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sfDBTools/pkg/compress"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/encrypt"
	"sfDBTools/pkg/helper"
	"sfDBTools/pkg/ui"
	"strings"
)

const (
	// readBufferSize untuk buffered reader (512KB untuk optimal read performance)
	readBufferSize = 512 * 1024
)

// ReaderPipelineResult berisi hasil setup reader pipeline
type ReaderPipelineResult struct {
	Reader          io.Reader
	File            *os.File     // Perlu di-close oleh caller
	DecompressClose func() error // Optional, perlu di-call jika ada
	IsEncrypted     bool
	CompressionType string
}

// setupReaderPipeline membuat reader pipeline: file → buffer → decrypt → decompress
// Caller bertanggung jawab untuk close file dan decompressor
func (s *Service) setupReaderPipeline(sourceFile string) (*ReaderPipelineResult, error) {
	result := &ReaderPipelineResult{}

	// Open file
	file, err := os.Open(sourceFile)
	if err != nil {
		return nil, fmt.Errorf("gagal open backup file: %w", err)
	}
	result.File = file

	// Tambahkan buffered reader untuk mengurangi syscall overhead
	var reader io.Reader = bufio.NewReaderSize(file, readBufferSize)

	// Detect encryption dan compression
	result.IsEncrypted = strings.HasSuffix(sourceFile, ".enc")
	result.CompressionType = detectCompressionType(sourceFile)

	// Decrypt if encrypted
	if result.IsEncrypted {
		encryptionKey, err := s.resolveEncryptionKey()
		if err != nil {
			file.Close()
			return nil, err
		}

		decryptReader, err := encrypt.NewDecryptingReader(reader, encryptionKey)
		if err != nil {
			file.Close()
			return nil, fmt.Errorf("gagal setup decrypt reader: %w", err)
		}
		reader = decryptReader
		s.Log.Debug("Decrypting backup file...")
	}

	// Decompress if compressed
	if result.CompressionType != "" {
		decompressReader, err := compress.NewDecompressingReader(reader, compress.CompressionType(result.CompressionType))
		if err != nil {
			file.Close()
			return nil, fmt.Errorf("gagal setup decompress reader: %w", err)
		}
		result.DecompressClose = decompressReader.Close
		reader = decompressReader
		s.Log.Debugf("Decompressing backup file (%s)...", result.CompressionType)
	}

	result.Reader = reader
	return result, nil
}

// resolveEncryptionKey mendapatkan encryption key dari options, env var, atau prompt
func (s *Service) resolveEncryptionKey() (string, error) {
	// Priority 1: Dari options
	if s.RestoreOptions.EncryptionKey != "" {
		return s.RestoreOptions.EncryptionKey, nil
	}

	// Priority 2: Dari environment variable
	encryptionKey := helper.GetEnvOrDefault(consts.ENV_BACKUP_ENCRYPTION_KEY, "")
	if encryptionKey != "" {
		s.Log.Debug("Using encryption key from environment variable")
		return encryptionKey, nil
	}

	// Priority 3: Interactive prompt (jika tidak quiet mode)
	quietMode := helper.GetEnvOrDefault(consts.ENV_QUIET, "false") == "true"
	if !quietMode {
		// Tampilkan prompt untuk memasukkan encryption key
		ui.PrintSubHeader("Encryption Key Required")
		ui.PrintInfo("File backup terenkripsi memerlukan kunci dekripsi.")

		encKey, source, err := encrypt.EncryptionPrompt(
			"Masukkan encryption key untuk decrypt backup",
			consts.ENV_BACKUP_ENCRYPTION_KEY,
		)
		if err != nil {
			return "", fmt.Errorf("gagal mendapatkan encryption key: %w", err)
		}

		s.Log.Debugf("Encryption key obtained from %s", source)
		return encKey, nil
	}

	return "", fmt.Errorf("encryption key required untuk restore encrypted backup (use --encryption-key or SFDB_BACKUP_ENCRYPTION_KEY)")
}

// closePipeline membersihkan resources dari reader pipeline
func closePipeline(result *ReaderPipelineResult) {
	if result == nil {
		return
	}

	if result.DecompressClose != nil {
		result.DecompressClose()
	}

	if result.File != nil {
		result.File.Close()
	}
}
