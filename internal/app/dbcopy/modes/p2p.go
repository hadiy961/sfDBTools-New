// File : internal/app/dbcopy/modes/p2p.go
// Deskripsi : P2P (Primary to Primary) executor
// Author : Hadiyatna Muflihun
// Tanggal : 26 Januari 2026
// Last Modified : 12 Februari 2026
package modes

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"sfdbtools/internal/app/dbcopy/helpers"
	"sfdbtools/internal/app/dbcopy/model"
	applog "sfdbtools/internal/services/log"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/database"
	"sfdbtools/internal/shared/runtimecfg"
)

// P2PExecutor implements Executor interface untuk P2P copy
type P2PExecutor struct {
	log  applog.Logger
	svc  CopyService
	opts *model.P2POptions
}

// NewP2PExecutor membuat instance baru P2P executor
func NewP2PExecutor(log applog.Logger, svc CopyService, opts *model.P2POptions) *P2PExecutor {
	return &P2PExecutor{
		log:  log,
		svc:  svc,
		opts: opts,
	}
}

// Execute menjalankan P2P copy operation
func (e *P2PExecutor) Execute(ctx context.Context) (*model.CopyResult, error) {
	result := &model.CopyResult{}

	// Informasi awal (membantu debugging di lapangan)
	e.log.Infof("P2P: ticket=%s excludeData=%v includeDmart=%v continueOnError=%v workdir=%s",
		strings.TrimSpace(e.opts.Ticket),
		e.opts.ExcludeData,
		e.opts.IncludeDmart,
		e.opts.ContinueOnError,
		strings.TrimSpace(e.opts.Workdir),
	)

	// Setup profiles
	srcProfile, tgtProfile, err := e.svc.SetupProfiles(&e.opts.CommonCopyOptions, !DetermineNonInteractiveMode(&e.opts.CommonCopyOptions))
	if err != nil {
		return nil, err
	}

	// Setup workdir
	workdir, cleanup, err := e.svc.SetupWorkdir(&e.opts.CommonCopyOptions)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	// Connect to source
	srcClient, err := e.svc.SetupConnections(srcProfile)
	if err != nil {
		return nil, err
	}
	defer srcClient.Close()

	// Resolve database names
	sourceDB, targetDB, err := e.resolveDatabaseNames(ctx, srcClient)
	if err != nil {
		return nil, err
	}

	result.SourceDB = sourceDB
	result.TargetDB = targetDB

	e.log.Infof("P2P: source-profile=%s", filepath.Base(strings.TrimSpace(srcProfile.Path)))
	e.log.Infof("P2P: target-profile=%s", filepath.Base(strings.TrimSpace(tgtProfile.Path)))
	e.log.Infof("P2P: akan copy database: %s -> %s", sourceDB, targetDB)

	// Validate tidak copy ke self
	if err := e.svc.ValidateNotCopyToSelf(srcProfile, tgtProfile, sourceDB, targetDB, string(model.ModeP2P)); err != nil {
		return nil, err
	}

	// Check companion
	companionSource, hasCompanion, err := e.svc.CheckCompanionDatabase(ctx, srcClient, sourceDB, e.opts.IncludeDmart)
	if err != nil {
		return nil, err
	}

	// Dry-run mode
	if e.opts.DryRun {
		e.log.Infof("[DRY-RUN] P2P Copy Plan:")
		e.log.Infof("  Source DB: %s", sourceDB)
		e.log.Infof("  Target DB: %s", targetDB)
		e.log.Infof("  Companion: %v (exists=%v)", e.opts.IncludeDmart, hasCompanion)
		e.log.Infof("  Workdir: %s", workdir)
		result.Success = true
		result.Message = "Dry-run completed"
		return result, nil
	}

	// Resolve encryption key
	encKey, err := e.svc.ResolveBackupEncryptionKey()
	if err != nil {
		return nil, fmt.Errorf("gagal resolve backup encryption key: %w", err)
	}

	// Backup source database (P2P backup primary database, gunakan ModePrimary untuk konsistensi logging)
	srcDump, err := e.svc.BackupDB(ctx, srcProfile, srcClient, sourceDB, e.opts.Ticket, workdir, e.opts.ExcludeData, consts.ModePrimary)
	if err != nil {
		helpers.MaybePrintBackupFailureHint(err)
		result.Error = err
		return result, fmt.Errorf("gagal backup source database %s: %w", sourceDB, err)
	}

	// Backup companion if exists (companion tetap ModeSingle)
	var dmartDump string
	if hasCompanion {
		dmartDump, err = e.svc.BackupSingleDB(ctx, srcProfile, srcClient, companionSource, e.opts.Ticket, workdir, e.opts.ExcludeData)
		if err != nil {
			if e.opts.ContinueOnError {
				e.log.Warnf("Gagal backup companion (dmart), skip karena continue-on-error: %v", err)
				dmartDump = ""
			} else {
				helpers.MaybePrintBackupFailureHint(err)
				result.Error = err
				return result, fmt.Errorf("gagal backup companion database %s: %w", companionSource, err)
			}
		} else {
			result.CompanionCopied = true
		}
	}

	// Restore to target (P2P selalu gunakan restore primary)
	err = e.svc.RestorePrimary(
		ctx,
		tgtProfile,
		srcDump,
		dmartDump,
		targetDB,
		e.opts.Ticket,
		encKey,
		e.opts.IncludeDmart,
		true,  // dropTarget
		false, // skipBackup (P2P wajib prebackup target)
		true,  // skipGrants
		e.opts.ContinueOnError,
		DetermineNonInteractiveMode(&e.opts.CommonCopyOptions),
	)

	if err != nil {
		result.Error = err
		return result, fmt.Errorf("gagal restore ke target database %s: %w", targetDB, err)
	}

	result.Success = true
	result.Message = fmt.Sprintf("P2P copy berhasil: %s → %s", sourceDB, targetDB)
	return result, nil
}

func (e *P2PExecutor) resolveDatabaseNames(ctx context.Context, srcClient *database.Client) (sourceDB, targetDB string, err error) {
	// Explicit mode
	if strings.TrimSpace(e.opts.SourceDB) != "" {
		src := strings.TrimSpace(e.opts.SourceDB)
		tgt := strings.TrimSpace(e.opts.TargetDB)
		if tgt == "" {
			tgt = src
		}
		return src, tgt, nil
	}

	// Rule-based: resolve primary dari client-code; target default = source, bisa override via --target-db.
	primary, err := e.svc.ResolvePrimaryDB(ctx, srcClient, e.opts.ClientCode)
	if err != nil {
		return "", "", err
	}

	primary = strings.TrimSpace(primary)
	if strings.TrimSpace(e.opts.TargetDB) != "" {
		return primary, strings.TrimSpace(e.opts.TargetDB), nil
	}

	return primary, primary, nil
}

// DetermineNonInteractiveMode helper untuk menentukan mode non-interaktif
func DetermineNonInteractiveMode(opts *model.CommonCopyOptions) bool {
	return runtimecfg.IsQuiet() || opts.SkipConfirm
}
