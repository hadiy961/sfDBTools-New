// File : internal/restore/modes/all_processing.go
// Deskripsi : Processing functions untuk AllExecutor (streaming, dry-run)
// Author : Hadiyatna Muflihun
// Tanggal : 30 Desember 2025
// Last Modified : 5 Januari 2026
package modes

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"sfDBTools/internal/app/restore/helpers"
	restoremodel "sfDBTools/internal/app/restore/model"
	"sfDBTools/internal/ui/print"
	"strings"
	"time"
)

// restoreStats menyimpan statistik restore
type restoreStats struct {
	RestoredCount int
}

// executeDryRun melakukan analisis file dump tanpa restore
func (e *AllExecutor) executeDryRun(ctx context.Context, opts *restoremodel.RestoreAllOptions, result *restoremodel.RestoreResult) (*restoremodel.RestoreResult, error) {
	logger := e.service.GetLogger()
	start := time.Now()
	logger.Info("Membuka dan menganalisis file dump...")

	reader, closers, err := helpers.OpenAndPrepareReader(opts.File, opts.EncryptionKey)
	if err != nil {
		result.Error = err
		return result, err
	}
	defer helpers.CloseReaders(closers)

	scanner := bufio.NewScanner(reader)
	configureScanner(scanner)

	dbStats := make(map[string]int)
	var currentDB string

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "USE `") {
			currentDB = extractDBName(line)
			if currentDB != "" {
				skipReason := shouldSkipDatabase(currentDB, opts)
				if skipReason == "" {
					dbStats[currentDB] = 0
				}
			}
		}

		if currentDB != "" && dbStats[currentDB] >= 0 {
			dbStats[currentDB]++
		}
	}

	if err := scanner.Err(); err != nil {
		result.Error = fmt.Errorf("error membaca file dump: %w", err)
		return result, result.Error
	}

	// Print hasil analisis
	print.PrintSuccess("\n📊 Hasil Analisis File Dump:")
	print.PrintInfo(fmt.Sprintf("Total database yang akan di-restore: %d", len(dbStats)))

	if len(dbStats) > 0 {
		print.PrintInfo("\nDatabase yang akan di-restore:")
		for db, lines := range dbStats {
			print.PrintInfo(fmt.Sprintf("  • %s (%d baris)", db, lines))
		}
	}

	result.Success = true
	result.Duration = time.Since(start).Round(time.Second).String()
	return result, nil
}

// processStreamWithFiltering membaca file, filter, dan tulis ke MySQL stdin
func (e *AllExecutor) processStreamWithFiltering(ctx context.Context, opts *restoremodel.RestoreAllOptions, output io.Writer, progressCh chan<- string) (*restoreStats, error) {
	reader, closers, err := helpers.OpenAndPrepareReader(opts.File, opts.EncryptionKey)
	if err != nil {
		return nil, err
	}
	defer helpers.CloseReaders(closers)

	scanner := bufio.NewScanner(reader)
	configureScanner(scanner)

	var currentDB string
	skipCurrentDB := false
	restoredDBs := make(map[string]bool)

	// Buffer writes to reduce syscall overhead and speed up piping to mysql
	bufWriter := bufio.NewWriterSize(output, 4*1024*1024) // 4MB buffer
	defer bufWriter.Flush()

	for scanner.Scan() {
		line := scanner.Text()

		// Deteksi statement USE `db_name`;
		if strings.HasPrefix(line, "USE `") {
			newDB := extractDBName(line)
			if newDB != "" {
				currentDB = newDB

				// Check apakah DB harus di-skip
				skipReason := shouldSkipDatabase(currentDB, opts)
				skipCurrentDB = (skipReason != "")

				if !skipCurrentDB {
					if !restoredDBs[currentDB] {
						restoredDBs[currentDB] = true
						// Kirim update progress DB ke spinner
						progressCh <- currentDB
					}
				}
			}
		}

		// Skip baris jika database sedang di-skip
		if skipCurrentDB {
			continue
		}

		// Tulis baris ke MySQL stdin (buffered) dengan newline
		if _, err := bufWriter.WriteString(line + "\n"); err != nil {
			return nil, fmt.Errorf("gagal menulis ke mysql stdin: %w", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error membaca file dump: %w", err)
	}

	return &restoreStats{
		RestoredCount: len(restoredDBs),
	}, nil
}
