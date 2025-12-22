package backup

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sfDBTools/internal/types/types_backup"
	"sfDBTools/pkg/backuphelper"
	"sfDBTools/pkg/ui"
	"strings"
)

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
