// File : internal/app/dbcopy/modes/p2s.go
// Deskripsi : P2S (Primary to Secondary) executor
// Author : Hadiyatna Muflihun
// Tanggal : 26 Januari 2026
// Last Modified : 26 Januari 2026
package modes

import (
	"context"
	"fmt"

	"sfdbtools/internal/app/dbcopy/model"
	applog "sfdbtools/internal/services/log"
	"sfdbtools/internal/shared/naming"
)

// P2SExecutor implements Executor interface untuk P2S copy
type P2SExecutor struct {
	log  applog.Logger
	svc  CopyService
	opts *model.P2SOptions
}

// NewP2SExecutor membuat instance baru P2S executor
func NewP2SExecutor(log applog.Logger, svc CopyService, opts *model.P2SOptions) *P2SExecutor {
	return &P2SExecutor{
		log:  log,
		svc:  svc,
		opts: opts,
	}
}

// Execute menjalankan P2S copy operation
func (e *P2SExecutor) Execute(ctx context.Context) (*model.CopyResult, error) {
	result := &model.CopyResult{}

	srcProfile, tgtProfile, err := e.svc.SetupProfiles(&e.opts.CommonCopyOptions, !DetermineNonInteractiveMode(&e.opts.CommonCopyOptions))
	if err != nil {
		return nil, err
	}

	workdir, cleanup, err := e.svc.SetupWorkdir(&e.opts.CommonCopyOptions)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	srcClient, err := e.svc.SetupConnections(srcProfile)
	if err != nil {
		return nil, err
	}
	defer srcClient.Close()

	sourceDB, targetDB, err := e.resolveDatabaseNames(ctx, srcClient)
	if err != nil {
		return nil, err
	}

	result.SourceDB = sourceDB
	result.TargetDB = targetDB

	if err := e.svc.ValidateNotCopyToSelf(srcProfile, tgtProfile, sourceDB, targetDB, "p2s"); err != nil {
		return nil, err
	}

	companionSource, hasCompanion, err := e.svc.CheckCompanionDatabase(ctx, srcClient, sourceDB, e.opts.IncludeDmart)
	if err != nil {
		return nil, err
	}

	if e.opts.DryRun {
		e.log.Infof("[DRY-RUN] P2S Copy Plan:")
		e.log.Infof("  Source DB: %s", sourceDB)
		e.log.Infof("  Target DB: %s", targetDB)
		e.log.Infof("  Companion: %v (exists=%v)", e.opts.IncludeDmart, hasCompanion)
		e.log.Infof("  Workdir: %s", workdir)
		result.Success = true
		result.Message = "Dry-run completed"
		return result, nil
	}

	encKey, err := e.svc.ResolveBackupEncryptionKey()
	if err != nil {
		return nil, fmt.Errorf("gagal resolve backup encryption key: %w", err)
	}

	srcDump, err := e.svc.BackupSingleDB(ctx, srcProfile, srcClient, sourceDB, e.opts.Ticket, workdir, e.opts.ExcludeData)
	if err != nil {
		result.Error = err
		return result, err
	}

	var dmartDump string
	if hasCompanion {
		dmartDump, err = e.svc.BackupSingleDB(ctx, srcProfile, srcClient, companionSource, e.opts.Ticket, workdir, e.opts.ExcludeData)
		if err != nil {
			if e.opts.ContinueOnError {
				e.log.Warnf("Gagal backup companion (dmart), skip karena continue-on-error: %v", err)
				dmartDump = ""
			} else {
				result.Error = err
				return result, err
			}
		} else {
			result.CompanionCopied = true
		}
	}

	// Jika eksplisit mode, gunakan restore single untuk flexibility
	if e.opts.SourceDB != "" && e.opts.TargetDB != "" {
		if err := e.svc.RestoreSingle(ctx, tgtProfile, srcDump, targetDB, e.opts.Ticket, encKey, true, !e.opts.PrebackupTarget, true, e.opts.ContinueOnError, DetermineNonInteractiveMode(&e.opts.CommonCopyOptions)); err != nil {
			result.Error = err
			return result, err
		}

		if dmartDump != "" {
			companionTarget := naming.BuildCompanionDBName(targetDB)
			if err := e.svc.RestoreSingle(ctx, tgtProfile, dmartDump, companionTarget, e.opts.Ticket, encKey, true, !e.opts.PrebackupTarget, true, e.opts.ContinueOnError, DetermineNonInteractiveMode(&e.opts.CommonCopyOptions)); err != nil {
				if e.opts.ContinueOnError {
					e.log.Warnf("Gagal restore companion (dmart), skip karena continue-on-error: %v", err)
				} else {
					result.Error = err
					return result, err
				}
			}
		}
	} else {
		// Rule-based mode: gunakan restore secondary
		err = e.svc.RestoreSecondary(ctx, tgtProfile, srcDump, dmartDump, e.opts.Ticket, e.opts.ClientCode, e.opts.Instance, encKey, e.opts.IncludeDmart, true, !e.opts.PrebackupTarget, e.opts.ContinueOnError, DetermineNonInteractiveMode(&e.opts.CommonCopyOptions))
		if err != nil {
			result.Error = err
			return result, err
		}
	}

	result.Success = true
	result.Message = fmt.Sprintf("P2S copy berhasil: %s → %s", sourceDB, targetDB)
	return result, nil
}

func (e *P2SExecutor) resolveDatabaseNames(ctx context.Context, srcClient interface{}) (sourceDB, targetDB string, err error) {
	if e.opts.SourceDB != "" && e.opts.TargetDB != "" {
		return e.opts.SourceDB, e.opts.TargetDB, nil
	}

	primary, err := e.svc.ResolvePrimaryDB(ctx, srcClient, e.opts.ClientCode)
	if err != nil {
		return "", "", err
	}

	targetDB = naming.BuildSecondaryDBName(primary, e.opts.Instance)
	return primary, targetDB, nil
}
