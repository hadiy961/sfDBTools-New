// File : internal/restore/modes/companion_helpers.go
// Deskripsi : Helper functions untuk companion database handling
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-30
// Last Modified : 2025-12-30

package modes

import (
	"context"
	"fmt"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/ui"
	"strings"
)

// companionRestoreFlow menangani backup, drop, dan restore untuk companion (_dmart) databases
type companionRestoreFlow struct {
	service       RestoreService
	ctx           context.Context
	primaryDB     string // Nama database utama (tanpa suffix)
	sourceFile    string
	encryptionKey string
	skipBackup    bool
	dropTarget    bool
	stopOnError   bool
	backupOpts    *types.RestoreBackupOptions
}

// execute menjalankan alur restore companion dan mengembalikan file backup, error
func (f *companionRestoreFlow) execute() (backupFile string, err error) {
	logger := f.service.GetLogger()
	client := f.service.GetTargetClient()
	companionDB := f.primaryDB + consts.SuffixDmart

	// 1. Cek keberadaan companion
	companionExists, err := client.CheckDatabaseExists(f.ctx, companionDB)
	if err != nil {
		if f.stopOnError {
			return "", fmt.Errorf("gagal mengecek companion database: %w", err)
		}
		logger.Warnf("Gagal cek companion DB %s: %v (lanjut)", companionDB, err)
		return "", nil
	}

	// 2. Backup companion jika diperlukan
	if !f.skipBackup && companionExists {
		backupFile, err = f.service.BackupDatabaseIfNeeded(f.ctx, companionDB, true, false, f.backupOpts)
		if err != nil {
			if f.stopOnError {
				return "", fmt.Errorf("gagal backup companion database: %w", err)
			}
			logger.Warnf("Gagal backup companion %s: %v (lanjut)", companionDB, err)
		}
	}

	// 3. Drop companion if needed
	if f.dropTarget && companionExists {
		if err := f.service.DropDatabaseIfNeeded(f.ctx, companionDB, true, true); err != nil {
			if f.stopOnError {
				return backupFile, fmt.Errorf("gagal drop companion database: %w", err)
			}
			logger.Warnf("Gagal drop companion %s: %v (lanjut)", companionDB, err)
		}
	}

	// 4. Restore companion database
	if err := f.service.CreateAndRestoreDatabase(f.ctx, companionDB, f.sourceFile, f.encryptionKey); err != nil {
		if f.stopOnError {
			return backupFile, fmt.Errorf("gagal restore companion database: %w", err)
		}
		ui.PrintWarning(fmt.Sprintf("⚠️  Gagal restore companion %s: %v", companionDB, err))
		return backupFile, err
	}

	logger.Infof("Companion database %s berhasil di-restore", companionDB)
	return backupFile, nil
}

// backupPrimaryCompanionIfNeeded membuat backup dari companion database milik primary
func backupPrimaryCompanionIfNeeded(ctx context.Context, service RestoreService, primaryDB string, skipBackup bool, stopOnError bool, backupOpts *types.RestoreBackupOptions) (backupFile string, err error) {
	if skipBackup {
		return "", nil
	}

	logger := service.GetLogger()
	client := service.GetTargetClient()
	companionDB := primaryDB + consts.SuffixDmart

	// Check if companion exists
	exists, err := client.CheckDatabaseExists(ctx, companionDB)
	if err != nil {
		if stopOnError {
			return "", fmt.Errorf("gagal mengecek companion database %s: %w", companionDB, err)
		}
		logger.Warnf("Gagal cek companion %s: %v (skip backup)", companionDB, err)
		return "", nil
	}

	if !exists {
		ui.PrintWarning(fmt.Sprintf("⚠️  Companion %s tidak ditemukan (skip backup)", companionDB))
		return "", nil
	}

	// Backup companion tersebut
	backupFile, err = service.BackupDatabaseIfNeeded(ctx, companionDB, true, false, backupOpts)
	if err != nil {
		if stopOnError {
			return "", fmt.Errorf("gagal backup companion %s: %w", companionDB, err)
		}
		ui.PrintWarning(fmt.Sprintf("⚠️  Gagal backup companion %s: %v", companionDB, err))
		return "", nil
	}

	return backupFile, nil
}

// resolveCompanionSourceFile determines companion source file from primary source
// Returns empty string if companion not found or not needed
func resolveCompanionSourceFile(primarySourceFile, companionFile string, includeDmart bool) string {
	if !includeDmart {
		return ""
	}

	if strings.TrimSpace(companionFile) != "" {
		return companionFile
	}

	return "" // Tidak ada file companion tersedia
}
