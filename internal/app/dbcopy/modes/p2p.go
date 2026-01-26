// File : internal/app/dbcopy/modes/p2p.go
// Deskripsi : P2P (Primary to Primary) executor
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

	// Validate tidak copy ke self
	if err := e.svc.ValidateNotCopyToSelf(srcProfile, tgtProfile, sourceDB, targetDB, "p2p"); err != nil {
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

	// Backup source database
	srcDump, err := e.svc.BackupSingleDB(ctx, srcProfile, srcClient, sourceDB, e.opts.Ticket, workdir, e.opts.ExcludeData)
	if err != nil {
		result.Error = err
		return result, err
	}

	// Backup companion if exists
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
		true, // dropTarget
		!e.opts.PrebackupTarget,
		true, // skipGrants
		e.opts.ContinueOnError,
		DetermineNonInteractiveMode(&e.opts.CommonCopyOptions),
	)

	if err != nil {
		result.Error = err
		return result, err
	}

	result.Success = true
	result.Message = fmt.Sprintf("P2P copy berhasil: %s → %s", sourceDB, targetDB)
	return result, nil
}

func (e *P2PExecutor) resolveDatabaseNames(ctx context.Context, srcClient interface{}) (sourceDB, targetDB string, err error) {
	// Explicit mode: gunakan langsung dari options
	if e.opts.SourceDB != "" && e.opts.TargetDB != "" {
		return e.opts.SourceDB, e.opts.TargetDB, nil
	}

	// Rule-based mode: resolve primary source, build target dengan target-client-code
	primary, err := e.svc.ResolvePrimaryDB(ctx, srcClient, e.opts.ClientCode)
	if err != nil {
		return "", "", err
	}

	prefix := naming.InferPrimaryPrefix(primary, "")
	targetDB = naming.BuildPrimaryDBName(prefix, e.opts.TargetClientCode)

	return primary, targetDB, nil
}

// DetermineNonInteractiveMode helper untuk menentukan mode non-interaktif
func DetermineNonInteractiveMode(opts *model.CommonCopyOptions) bool {
	return opts.SkipConfirm // quiet mode sudah di-handle di tempat lain
}
