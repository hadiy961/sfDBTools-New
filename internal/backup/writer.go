// File : internal/backup/writer.go
// Deskripsi : Writer dan mysqldump execution logic
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-05
// Last Modified : 2025-12-05

package backup

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sfDBTools/internal/types/types_backup"
	"sfDBTools/pkg/backuphelper"
	"sfDBTools/pkg/compress"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/encrypt"
	"sfDBTools/pkg/helper"
	"sfDBTools/pkg/ui"
	"strings"
	"time"
)

const bufferSize = 256 * 1024 // 256KB buffer untuk buffered I/O

// createBufferedOutputFile membuat output file dengan buffered writer
func (s *Service) createBufferedOutputFile(outputPath string) (*os.File, *bufio.Writer, error) {
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return nil, nil, fmt.Errorf("gagal membuat file output: %w", err)
	}
	bufWriter := bufio.NewWriterSize(outputFile, bufferSize)
	return outputFile, bufWriter, nil
}

// resolveEncryptionKeyIfNeeded resolve encryption key jika encryption enabled
func (s *Service) resolveEncryptionKeyIfNeeded() (string, error) {
	if !s.BackupDBOptions.Encryption.Enabled {
		return "", nil
	}

	resolvedKey, source, err := helper.ResolveEncryptionKey(
		s.BackupDBOptions.Encryption.Key,
		consts.ENV_BACKUP_ENCRYPTION_KEY,
	)
	if err != nil {
		return "", fmt.Errorf("gagal mendapatkan kunci enkripsi: %w", err)
	}

	s.Log.Debugf("Kunci enkripsi didapat dari: %s", source)
	return resolvedKey, nil
}

// executeMysqldumpWithPipe menjalankan mysqldump dengan pipe untuk kompresi dan enkripsi
func (s *Service) executeMysqldumpWithPipe(ctx context.Context, mysqldumpArgs []string, outputPath string, compressionRequired bool, compressionType string) (*types_backup.BackupWriteResult, error) {
	// Resolve encryption key SEBELUM spinner dimulai
	encryptionKey, err := s.resolveEncryptionKeyIfNeeded()
	if err != nil {
		return nil, err
	}

	// Start spinner dengan elapsed time
	spin := ui.NewSpinnerWithElapsed("Memproses backup database")
	spin.Start()
	defer spin.Stop()

	// Setup output file with buffered writer
	outputFile, bufWriter, err := s.createBufferedOutputFile(outputPath)
	if err != nil {
		return nil, err
	}
	defer outputFile.Close()
	defer func() {
		if flushErr := bufWriter.Flush(); flushErr != nil {
			s.Log.Errorf("Error flushing buffer: %v", flushErr)
		}
	}()

	// Setup writer chain: mysqldump → Compression → Encryption → Buffer → File
	writer, closers, err := s.createWriterPipeline(bufWriter, compressionRequired, compressionType, encryptionKey)
	if err != nil {
		return nil, err
	}

	// Cleanup writers
	defer func() {
		for i := len(closers) - 1; i >= 0; i-- {
			if err := closers[i].Close(); err != nil {
				s.Log.Errorf("Error closing writer: %v", err)
			}
		}
	}()

	// Execute mysqldump command
	cmd := exec.CommandContext(ctx, "mysqldump", mysqldumpArgs...)
	
	// Wrap writer with monitor for progress tracking
	monitor := newDatabaseMonitorWriter(writer, spin)
	cmd.Stdout = monitor

	// Capture stderr
	var stderrBuf strings.Builder
	cmd.Stderr = &stderrBuf

	// Run command
	runErr := cmd.Run()
	
	// Finish monitor (report last DB)
	monitor.Finish(runErr == nil)

	if runErr != nil {
		stderrOutput := stderrBuf.String()

		// Log ke error log file
		logFile := s.ErrorLog.LogWithOutput(map[string]interface{}{
			"type": "mysqldump_backup",
			"file": outputPath,
		}, stderrOutput, runErr)
		_ = logFile

		// Check fatal error
		if s.isFatalMysqldumpError(runErr, stderrOutput) {
			// Return result with stderr captured so upper layer can display it
			result := &types_backup.BackupWriteResult{
				StderrOutput: stderrOutput,
				FileSize:     0,
			}
			return result, fmt.Errorf("mysqldump gagal: %w", runErr)
		}
		s.Log.Warn("mysqldump exit with non-fatal error, treated as warning")
	}

	// Get file info
	fileInfo, statErr := os.Stat(outputPath)
	var fileSize int64
	if statErr == nil {
		fileSize = fileInfo.Size()
	}

	// Create result
	result := &types_backup.BackupWriteResult{
		StderrOutput: stderrBuf.String(),
		FileSize:     fileSize,
	}

	return result, nil
}

// createWriterPipeline membuat writer pipeline untuk compression dan encryption
func (s *Service) createWriterPipeline(baseWriter io.Writer, compressionRequired bool, compressionType string, encryptionKey string) (io.Writer, []io.Closer, error) {
	var writer io.Writer = baseWriter
	var closers []io.Closer

	// Layer 1: Encryption (paling dekat dengan file)
	if s.BackupDBOptions.Encryption.Enabled {
		encryptingWriter, err := encrypt.NewEncryptingWriter(writer, []byte(encryptionKey))
		if err != nil {
			return nil, nil, fmt.Errorf("gagal membuat encrypting writer: %w", err)
		}
		closers = append(closers, encryptingWriter)
		writer = encryptingWriter
	}

	// Layer 2: Compression
	if compressionRequired {
		compressionConfig := compress.CompressionConfig{
			Type:  compress.CompressionType(compressionType),
			Level: compress.CompressionLevel(s.BackupDBOptions.Compression.Level),
		}
		compressingWriter, err := compress.NewCompressingWriter(writer, compressionConfig)
		if err != nil {
			return nil, nil, fmt.Errorf("gagal membuat compressing writer: %w", err)
		}
		closers = append(closers, compressingWriter)
		writer = compressingWriter
	}

	return writer, closers, nil
}

// isFatalMysqldumpError menentukan apakah error dari mysqldump adalah fatal
func (s *Service) isFatalMysqldumpError(err error, stderrOutput string) bool {
	if err == nil {
		return false
	}

	if stderrOutput == "" {
		s.Log.Debug("mysqldump error with empty stderr, treating as fatal")
		return true
	}

	fatal := backuphelper.IsFatalMysqldumpError(err, stderrOutput)

	if !fatal {
		s.Log.Debugf("mysqldump treated as non-fatal by helper: %s", stderrOutput)
	}

	return fatal
}

// =============================================================================
// Database Monitor Writer
// =============================================================================

// databaseMonitorWriter membungkus io.Writer untuk memantau progress per database
type databaseMonitorWriter struct {
	target    io.Writer
	spinner   *ui.SpinnerWithElapsed
	currentDB string
	startTime time.Time
	
	// Buffer untuk menangani split chunk (sederhana)
	// Kita hanya perlu cukup byte untuk mendeteksi marker "-- Current Database: "
	// Namun implementasi paling sederhana adalah scanning per chunk.
	// Risiko miss kecil jika chunk size cukup besar (default io.Copy buffer 32KB)
}

func newDatabaseMonitorWriter(target io.Writer, spinner *ui.SpinnerWithElapsed) *databaseMonitorWriter {
	return &databaseMonitorWriter{
		target:  target,
		spinner: spinner,
	}
}

var dbMarker = []byte("-- Current Database: ")

func (w *databaseMonitorWriter) Write(p []byte) (n int, err error) {
	// Scan marker
	if idx := strings.Index(string(p), string(dbMarker)); idx != -1 {
		// Marker found, try to extract DB name
		// Format: -- Current Database: `dbname`
		remainder := p[idx+len(dbMarker):]
		
		// Cari backtick pembuka
		if startTick := strings.IndexByte(string(remainder), '`'); startTick != -1 {
			remainder = remainder[startTick+1:]
			// Cari backtick penutup
			if endTick := strings.IndexByte(string(remainder), '`'); endTick != -1 {
				dbName := string(remainder[:endTick])
				
				if dbName != w.currentDB {
					w.onDatabaseSwitch(dbName)
				}
			}
		}
	}
	
	return w.target.Write(p)
}

func (w *databaseMonitorWriter) onDatabaseSwitch(newDB string) {
	now := time.Now()
	
	// Report previous DB success
	if w.currentDB != "" {
		duration := now.Sub(w.startTime)
		msg := fmt.Sprintf("✓ Database %s (%s)", w.currentDB, duration.Round(time.Millisecond))
		
		// Gunakan SuspendAndRun untuk print ke stdout tanpa ganggu spinner
		w.spinner.SuspendAndRun(func() {
			fmt.Println(msg)
		})
	}
	
	// Update state
	w.currentDB = newDB
	w.startTime = now
	
	// Update spinner
	w.spinner.UpdateMessage(fmt.Sprintf("Processing %s", newDB))
}

func (w *databaseMonitorWriter) Finish(success bool) {
	// Report last DB
	if w.currentDB != "" && success {
		duration := time.Since(w.startTime)
		msg := fmt.Sprintf("✓ Database %s (%s)", w.currentDB, duration.Round(time.Millisecond))
		
		w.spinner.SuspendAndRun(func() {
			fmt.Println(msg)
		})
	}
}
