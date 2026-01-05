// File : internal/restore/modes/custom.go
// Deskripsi : Executor untuk restore custom (SFCola account detail)
// Author : Hadiyatna Muflihun
// Tanggal : 24 Desember 2025

package modes

import (
	"context"
	"fmt"
	"path/filepath"
	restoremodel "sfDBTools/internal/app/restore/model"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/ui"
	"strings"
	"time"
)

type customExecutor struct {
	svc RestoreService
}

func NewCustomExecutor(svc RestoreService) RestoreExecutor { return &customExecutor{svc: svc} }

func (e *customExecutor) Execute(ctx context.Context) (*restoremodel.RestoreResult, error) {
	opts := e.svc.GetCustomOptions()
	if opts == nil {
		return nil, fmt.Errorf("opsi custom tidak tersedia")
	}

	start := time.Now()
	logger := e.svc.GetLogger()
	client := e.svc.GetTargetClient()

	result := &restoremodel.RestoreResult{
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

	// Step 5-6: Restore main database using common flow
	mainFlow := &commonRestoreFlow{
		service:       e.svc,
		ctx:           ctx,
		dbName:        opts.Database,
		sourceFile:    opts.DatabaseFile,
		encryptionKey: opts.EncryptionKey,
		skipBackup:    opts.SkipBackup,
		dropTarget:    opts.DropTarget,
		stopOnError:   opts.StopOnError,
		backupOpts:    opts.BackupOptions,
	}

	backupMain, err := mainFlow.execute()
	if err != nil {
		if opts.StopOnError {
			return nil, err
		}
		logger.Warnf("restore main DB gagal (lanjut karena continue-on-error): %v", err)
		result.Success = false
	}
	result.BackupFile = backupMain

	// Restore DMART database using companion flow
	dmartFlow := &companionRestoreFlow{
		service:       e.svc,
		ctx:           ctx,
		primaryDB:     strings.TrimSuffix(opts.DatabaseDmart, consts.SuffixDmart),
		sourceFile:    opts.DatabaseDmartFile,
		encryptionKey: opts.EncryptionKey,
		skipBackup:    opts.SkipBackup,
		dropTarget:    opts.DropTarget,
		stopOnError:   opts.StopOnError,
		backupOpts:    opts.BackupOptions,
	}

	backupDmart, err := dmartFlow.execute()
	if err != nil && opts.StopOnError {
		return nil, err
	}
	if err != nil {
		logger.Warnf("restore DMART gagal (lanjut karena continue-on-error): %v", err)
		result.Success = false
	}
	result.CompanionBackup = backupDmart

	// Post-restore operations using helpers
	if result.Success {
		performPostRestoreOperations(ctx, e.svc, opts.Database)
		copyGrantsBetweenDatabases(ctx, e.svc, opts.Database, opts.DatabaseDmart)
	}

	ui.PrintSuccess("Restore custom selesai")
	result.Duration = time.Since(start).String()
	return result, nil
}
