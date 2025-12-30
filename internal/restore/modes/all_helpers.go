// File : internal/restore/modes/all_helpers.go
// Deskripsi : Helper functions untuk AllExecutor
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-30
// Last Modified : 2025-12-30

package modes

import (
	"bufio"
	"context"
	"fmt"
	"sfDBTools/internal/restore/helpers"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/database"
	"strings"
)

// configureScanner mengkonfigurasi scanner dengan buffer besar untuk handle INSERT panjang
func configureScanner(scanner *bufio.Scanner) {
	const maxCapacity = 100 * 1024 * 1024 // 100MB max per baris
	buf := make([]byte, 0, 4*1024*1024)   // Awal 4MB untuk kurangi resize
	scanner.Buffer(buf, maxCapacity)
}

// extractDBName mengekstrak nama database dari statement USE `db_name`;
func extractDBName(line string) string {
	start := strings.Index(line, "`")
	if start == -1 {
		return ""
	}
	start++ // lewati backtick pertama

	end := strings.Index(line[start:], "`")
	if end != -1 {
		return line[start : start+end]
	}
	return ""
}

// shouldSkipDatabase mengecek apakah database harus di-skip (mengembalikan alasan string atau kosong)
func shouldSkipDatabase(dbName string, opts *types.RestoreAllOptions) string {
	for _, excluded := range opts.ExcludeDBs {
		if dbName == excluded {
			return "excluded by user"
		}
	}
	if opts.SkipSystemDBs && database.IsSystemDatabase(dbName) {
		return "system database"
	}
	return ""
}

// collectDatabasesToRestore melakukan pass awal untuk mengumpulkan daftar DB target dari dump
func collectDatabasesToRestore(ctx context.Context, opts *types.RestoreAllOptions) (map[string]struct{}, error) {
	reader, closers, err := helpers.OpenAndPrepareReader(opts.File, opts.EncryptionKey)
	if err != nil {
		return nil, err
	}
	defer helpers.CloseReaders(closers)

	scanner := bufio.NewScanner(reader)
	configureScanner(scanner)

	targets := make(map[string]struct{})
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "USE `") {
			db := extractDBName(line)
			if db == "" {
				continue
			}
			if reason := shouldSkipDatabase(db, opts); reason == "" {
				targets[db] = struct{}{}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error membaca file dump saat mengumpulkan DB: %w", err)
	}
	return targets, nil
}
