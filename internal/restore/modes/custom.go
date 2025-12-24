// File : internal/restore/modes/custom.go
// Deskripsi : Executor untuk restore custom (SFCola account detail)
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-24

package modes

import (
	"context"
	"fmt"
	"path/filepath"
	"sfDBTools/pkg/consts"
	"strings"
	"time"

	"sfDBTools/internal/types"
	"sfDBTools/pkg/ui"
)

type customExecutor struct {
	svc RestoreService
}

func NewCustomExecutor(svc RestoreService) RestoreExecutor { return &customExecutor{svc: svc} }

func (e *customExecutor) Execute(ctx context.Context) (*types.RestoreResult, error) {
	opts := e.svc.GetCustomOptions()
	if opts == nil {
		return nil, fmt.Errorf("opsi custom tidak tersedia")
	}

	start := time.Now()
	logger := e.svc.GetLogger()
	client := e.svc.GetTargetClient()

	result := &types.RestoreResult{
		Success:       true,
		TargetDB:      opts.Database,
		SourceFile:    filepath.Base(opts.DatabaseFile),
		CompanionDB:   opts.DatabaseDmart,
		CompanionFile: filepath.Base(opts.DatabaseDmartFile),
	}

	// For dry-run, validate files exist (already validated in setup), and stop here.
	if opts.DryRun {
		ui.PrintWarning("Dry-run: tidak ada perubahan database/user yang dilakukan")
		result.Duration = time.Since(start).String()
		return result, nil
	}

	// Step 4: Ensure users exist and have grants to DB and DMART (recreate if exists)
	// Standard host for these SFCola users is '%'
	host := "%"
	users := []struct {
		user string
		pass string
		tag  string
	}{
		{user: opts.UserAdmin, pass: opts.PassAdmin, tag: "admin"},
		{user: opts.UserFin, pass: opts.PassFin, tag: "fin"},
		{user: opts.UserUser, pass: opts.PassUser, tag: "user"},
	}

	for _, u := range users {
		if err := client.DropUserIfExists(ctx, u.user, host); err != nil {
			return nil, fmt.Errorf("drop user gagal (%s): %w", u.user, err)
		}
		if err := client.CreateUser(ctx, u.user, host, u.pass); err != nil {
			return nil, fmt.Errorf("create user gagal (%s): %w", u.user, err)
		}
		if err := client.GrantAllPrivilegesOnDatabase(ctx, u.user, host, opts.Database, true); err != nil {
			return nil, fmt.Errorf("grant gagal (%s → %s): %w", u.user, opts.Database, err)
		}
		if err := client.GrantAllPrivilegesOnDatabase(ctx, u.user, host, opts.DatabaseDmart, true); err != nil {
			return nil, fmt.Errorf("grant gagal (%s → %s): %w", u.user, opts.DatabaseDmart, err)
		}
		logger.Infof("User %s (%s) siap + grants applied", u.user, u.tag)
	}

	// Step 5-6 + rest: restore database & dmart like other restore modes
	// Main DB
	mainExists, err := client.CheckDatabaseExists(ctx, opts.Database)
	if err != nil {
		return nil, err
	}
	result.DroppedDB = mainExists && opts.DropTarget

	var backupMain string
	if !opts.SkipBackup {
		b, err := e.svc.BackupDatabaseIfNeeded(ctx, opts.Database, mainExists, opts.SkipBackup, opts.BackupOptions)
		if err != nil {
			return nil, err
		}
		backupMain = b
	}
	result.BackupFile = backupMain

	if err := e.svc.DropDatabaseIfNeeded(ctx, opts.Database, mainExists, opts.DropTarget); err != nil {
		return nil, err
	}
	if err := e.svc.CreateAndRestoreDatabase(ctx, opts.Database, opts.DatabaseFile, opts.EncryptionKey); err != nil {
		if opts.StopOnError {
			return nil, err
		}
		logger.Warnf("restore main DB gagal (lanjut karena continue-on-error): %v", err)
		result.Success = false
	}

	// DMART
	dmartExists, err := client.CheckDatabaseExists(ctx, opts.DatabaseDmart)
	if err != nil {
		return nil, err
	}
	result.DroppedCompanion = dmartExists && opts.DropTarget

	var backupDmart string
	if !opts.SkipBackup {
		b, err := e.svc.BackupDatabaseIfNeeded(ctx, opts.DatabaseDmart, dmartExists, opts.SkipBackup, opts.BackupOptions)
		if err != nil {
			if opts.StopOnError {
				return nil, err
			}
			logger.Warnf("backup DMART gagal (lanjut karena continue-on-error): %v", err)
			result.Success = false
		} else {
			backupDmart = b
		}
	}
	result.CompanionBackup = backupDmart

	if err := e.svc.DropDatabaseIfNeeded(ctx, opts.DatabaseDmart, dmartExists, opts.DropTarget); err != nil {
		if opts.StopOnError {
			return nil, err
		}
		logger.Warnf("drop DMART gagal (lanjut karena continue-on-error): %v", err)
		result.Success = false
	}
	if err := e.svc.CreateAndRestoreDatabase(ctx, opts.DatabaseDmart, opts.DatabaseDmartFile, opts.EncryptionKey); err != nil {
		if opts.StopOnError {
			return nil, err
		}
		logger.Warnf("restore DMART gagal (lanjut karena continue-on-error): %v", err)
		result.Success = false
	}

	// Post-restore (warning-only): buat <db>_temp + copy grants.
	// Hanya dijalankan jika keseluruhan restore dianggap sukses.
	if result.Success && !strings.HasSuffix(opts.Database, consts.SuffixDmart) {
		tempDB, terr := e.svc.CreateTempDatabaseIfNeeded(ctx, opts.Database)
		if terr != nil {
			logger.Warnf("Gagal membuat temp DB: %v", terr)
			ui.PrintWarning(fmt.Sprintf("⚠️  Restore berhasil, tapi gagal membuat temp DB: %v", terr))
		} else if strings.TrimSpace(tempDB) != "" {
			if gerr := e.svc.CopyDatabaseGrants(ctx, opts.Database, tempDB); gerr != nil {
				logger.Warnf("Gagal copy grants ke temp DB: %v", gerr)
				ui.PrintWarning(fmt.Sprintf("⚠️  Restore berhasil, tapi gagal copy grants ke temp DB: %v", gerr))
			}
		}
	}

	// Copy grants main -> dmart (warning-only) jika overall sukses.
	if result.Success {
		if gerr := e.svc.CopyDatabaseGrants(ctx, opts.Database, opts.DatabaseDmart); gerr != nil {
			logger.Warnf("Gagal copy grants ke DMART: %v", gerr)
			ui.PrintWarning(fmt.Sprintf("⚠️  Restore berhasil, tapi gagal copy grants ke DMART: %v", gerr))
		}
	}

	ui.PrintSuccess("Restore custom selesai")
	result.Duration = time.Since(start).String()
	return result, nil
}
