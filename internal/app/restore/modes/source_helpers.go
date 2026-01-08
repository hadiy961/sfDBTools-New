// File : internal/restore/modes/source_helpers.go
// Deskripsi : Helper functions untuk source file resolution
// Author : Hadiyatna Muflihun
// Tanggal : 30 Desember 2025
// Last Modified : 5 Januari 2026
package modes

import (
	"context"
	"fmt"
	restoremodel "sfdbtools/internal/app/restore/model"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/ui/print"
	"strings"
)

// sourceFileResolver menentukan source file untuk operasi restore
type sourceFileResolver struct {
	service     RestoreService
	ctx         context.Context
	from        string // "primary" atau "file"
	primaryDB   string
	file        string
	backupOpts  *restoremodel.RestoreBackupOptions
	stopOnError bool
}

// resolveMainSource mengembalikan file source utama (membuat backup dari primary jika diperlukan)
func (r *sourceFileResolver) resolveMainSource() (sourceFile string, err error) {
	if r.from == "primary" {
		return r.backupFromPrimary(r.primaryDB)
	}
	return r.file, nil
}

// resolveCompanionSource mengembalikan companion source file (membuat backup dari primary companion jika diperlukan)
func (r *sourceFileResolver) resolveCompanionSource(companionFile string) (sourceFile string, err error) {
	if r.from == "primary" {
		companionDB := r.primaryDB + consts.SuffixDmart
		return r.backupFromPrimary(companionDB)
	}
	return companionFile, nil
}

// backupFromPrimary membuat backup dari primary database yang ditentukan
func (r *sourceFileResolver) backupFromPrimary(dbName string) (backupFile string, err error) {
	logger := r.service.GetLogger()
	client := r.service.GetTargetClient()

	// Cek apakah source database ada
	exists, err := client.CheckDatabaseExists(r.ctx, dbName)
	if err != nil {
		return "", fmt.Errorf("gagal mengecek database %s: %w", dbName, err)
	}
	if !exists {
		return "", fmt.Errorf("database %s tidak ditemukan", dbName)
	}

	// Create backup (always, since this is the source)
	logger.Infof("Backup database %s sebagai sumber restore...", dbName)
	backupFile, err = r.service.BackupDatabaseIfNeeded(r.ctx, dbName, true, false, r.backupOpts)
	if err != nil {
		return "", fmt.Errorf("gagal backup database %s: %w", dbName, err)
	}

	return backupFile, nil
}

// backupCompanionIfExists mencoba backup companion database jika ada
func (r *sourceFileResolver) backupCompanionIfExists() (companionSourceFile string, err error) {
	logger := r.service.GetLogger()
	client := r.service.GetTargetClient()
	companionDB := r.primaryDB + consts.SuffixDmart

	exists, err := client.CheckDatabaseExists(r.ctx, companionDB)
	if err != nil {
		if r.stopOnError {
			return "", fmt.Errorf("gagal mengecek companion %s: %w", companionDB, err)
		}
		logger.Warnf("Gagal cek companion %s: %v (skip)", companionDB, err)
		return "", nil
	}

	if !exists {
		print.PrintWarning(fmt.Sprintf("⚠️  Companion %s tidak ditemukan (skip)", companionDB))
		return "", nil
	}

	// Backup companion
	backupFile, err := r.service.BackupDatabaseIfNeeded(r.ctx, companionDB, true, false, r.backupOpts)
	if err != nil {
		if r.stopOnError {
			return "", fmt.Errorf("gagal backup companion %s: %w", companionDB, err)
		}
		print.PrintWarning(fmt.Sprintf("⚠️  Gagal backup companion %s: %v", companionDB, err))
		return "", nil
	}

	return backupFile, nil
}

// newSourceFileResolver membuat source file resolver baru
func newSourceFileResolver(service RestoreService, ctx context.Context, from, primaryDB, file string, backupOpts *restoremodel.RestoreBackupOptions, stopOnError bool) *sourceFileResolver {
	return &sourceFileResolver{
		service:     service,
		ctx:         ctx,
		from:        from,
		primaryDB:   primaryDB,
		file:        file,
		backupOpts:  backupOpts,
		stopOnError: stopOnError,
	}
}

// validateSourceFile memastikan source file tidak kosong
func validateSourceFile(sourceFile, fromMode string) error {
	if strings.TrimSpace(sourceFile) == "" {
		return fmt.Errorf("source file kosong (from=%s)", fromMode)
	}
	return nil
}
