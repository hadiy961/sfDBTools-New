// File : internal/app/dbcopy/command.go
// Deskripsi : Entry point execution untuk db-copy (dipanggil dari cmd layer)
// Author : Hadiyatna Muflihun
// Tanggal : 26 Januari 2026
// Last Modified : 26 Januari 2026
package dbcopy

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	profileconn "sfdbtools/internal/app/profile/connection"
	"sfdbtools/internal/app/restore"
	restoremodel "sfdbtools/internal/app/restore/model"
	"sfdbtools/internal/cli/deps"
	"sfdbtools/internal/domain"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/naming"
	"sfdbtools/internal/shared/runtimecfg"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
)

func enforceNotCopyToSelf(srcProfile *domain.ProfileInfo, tgtProfile *domain.ProfileInfo, sourceDB string, targetDB string, mode string) error {
	if srcProfile == nil || tgtProfile == nil {
		return nil
	}
	srcDB := strings.TrimSpace(sourceDB)
	tgtDB := strings.TrimSpace(targetDB)
	if srcDB == "" || tgtDB == "" {
		return nil
	}

	// Kriteria wajib ditolak:
	// - endpoint sama (host/port/user)
	// - dan nama database sama
	srcEff := profileconn.EffectiveDBInfo(srcProfile)
	tgtEff := profileconn.EffectiveDBInfo(tgtProfile)

	srcHost := strings.ToLower(strings.TrimSpace(srcEff.Host))
	tgtHost := strings.ToLower(strings.TrimSpace(tgtEff.Host))
	srcUser := strings.TrimSpace(srcProfile.DBInfo.User)
	tgtUser := strings.TrimSpace(tgtProfile.DBInfo.User)

	sameEndpoint := srcEff.Port == tgtEff.Port && srcHost == tgtHost && srcUser == tgtUser
	if strings.EqualFold(mode, "p2p") && sameEndpoint && !strings.EqualFold(srcDB, tgtDB) {
		return fmt.Errorf("db-copy p2p ditolak: source dan target berada di server yang sama (host=%s port=%d user=%s). Untuk p2p, target harus beda server via --target-profile (atau gunakan p2s/s2s)", srcHost, srcEff.Port, srcUser)
	}
	if sameEndpoint && strings.EqualFold(srcDB, tgtDB) {
		return fmt.Errorf("db-copy ditolak: source dan target menunjuk database yang sama (host=%s port=%d user=%s db=%s)", srcHost, srcEff.Port, srcUser, srcDB)
	}
	return nil
}

func ExecuteCopyP2S(cmd *cobra.Command, d *deps.Dependencies) error {
	opts, err := parseP2S(cmd)
	if err != nil {
		return err
	}
	return executeP2S(cmd, d, opts)
}

func ExecuteCopyP2P(cmd *cobra.Command, d *deps.Dependencies) error {
	opts, err := parseP2P(cmd)
	if err != nil {
		return err
	}
	return executeP2P(cmd, d, opts)
}

func ExecuteCopyS2S(cmd *cobra.Command, d *deps.Dependencies) error {
	opts, err := parseS2S(cmd)
	if err != nil {
		return err
	}
	return executeS2S(cmd, d, opts)
}

func executeP2S(_ *cobra.Command, d *deps.Dependencies, opts P2SOptions) error {
	if d == nil || d.Logger == nil || d.Config == nil {
		return fmt.Errorf("dependensi tidak tersedia")
	}

	nonInteractive := runtimecfg.IsQuiet() || opts.Common.SkipConfirm

	// Resolve workdir
	workdir, cleanup, err := ensureWorkdir(opts.Common.Workdir)
	if err != nil {
		return err
	}
	if strings.TrimSpace(opts.Common.Workdir) == "" {
		defer cleanup()
	}

	srcProfile, err := loadProfile(d.Config, opts.Common.SourceProfile, opts.Common.SourceProfileKey, consts.ENV_SOURCE_PROFILE, consts.ENV_SOURCE_PROFILE_KEY, true, !nonInteractive, "source")
	if err != nil {
		return err
	}
	targetPath, targetKey := resolveTargetProfileEffective(srcProfile, opts.Common.TargetProfile, opts.Common.TargetProfileKey)
	tgtProfile, err := loadProfile(d.Config, targetPath, targetKey, consts.ENV_TARGET_PROFILE, consts.ENV_TARGET_PROFILE_KEY, true, !nonInteractive, "target")
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Signal handling untuk cancel
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	// Connect source (untuk SSH tunnel + existence checks)
	srcClient, err := connectForOps(srcProfile)
	if err != nil {
		return err
	}
	defer srcClient.Close()

	sourceDB := opts.SourceDB
	targetDB := opts.TargetDB
	clientCode := opts.ClientCode
	instance := opts.Instance

	// Rule-based resolve bila tidak eksplisit
	if sourceDB == "" && targetDB == "" {
		primary, err := resolvePrimaryOnServer(ctx, srcClient, clientCode)
		if err != nil {
			return err
		}
		sourceDB = primary
		targetDB = secondaryDBNameFromPrimary(primary, instance)
	}

	if err := enforceNotCopyToSelf(srcProfile, tgtProfile, sourceDB, targetDB, "p2s"); err != nil {
		return err
	}

	// companion handling
	companionSource := naming.BuildCompanionDBName(sourceDB)
	hasCompanion := false
	if opts.Common.IncludeDmart {
		ok, err := databaseExists(ctx, srcClient, companionSource)
		if err != nil {
			return err
		}
		hasCompanion = ok
	}

	// Dry-run: hanya print plan
	if opts.Common.DryRun {
		d.Logger.Infof("[DRY-RUN] Source DB: %s", sourceDB)
		d.Logger.Infof("[DRY-RUN] Target DB: %s", targetDB)
		d.Logger.Infof("[DRY-RUN] Include dmart: %v (exists=%v)", opts.Common.IncludeDmart, hasCompanion)
		d.Logger.Infof("[DRY-RUN] Workdir: %s", workdir)
		return nil
	}

	encKey, err := resolveBackupEncryptionKeyForRestore(d.Config)
	if err != nil {
		return fmt.Errorf("gagal resolve backup encryption key untuk restore: %w", err)
	}

	// Backup source
	srcDump, err := backupSingleDatabase(ctx, d.Logger, d.Config, srcProfile, srcClient, sourceDB, opts.Common.Ticket, workdir, opts.Common.ExcludeData)
	if err != nil {
		return err
	}

	var dmartDump string
	if hasCompanion {
		dmartDump, err = backupSingleDatabase(ctx, d.Logger, d.Config, srcProfile, srcClient, companionSource, opts.Common.Ticket, workdir, opts.Common.ExcludeData)
		if err != nil {
			if opts.Common.ContinueOnError {
				d.Logger.Warnf("Gagal backup companion (dmart), skip karena continue-on-error: %v", err)
				dmartDump = ""
			} else {
				return err
			}
		}
	}

	// Restore ke target: jika mode eksplisit, gunakan restore single agar bisa target-db arbitrary.
	if opts.SourceDB != "" || opts.TargetDB != "" {
		if err := runRestoreSingle(ctx, d, tgtProfile, srcDump, targetDB, opts.Common, encKey); err != nil {
			return err
		}
		if dmartDump != "" {
			if err := runRestoreSingle(ctx, d, tgtProfile, dmartDump, naming.BuildCompanionDBName(targetDB), opts.Common, encKey); err != nil {
				if opts.Common.ContinueOnError {
					d.Logger.Warnf("Gagal restore companion (dmart), skip karena continue-on-error: %v", err)
					return nil
				}
				return err
			}
		}
		return nil
	}

	// Mode rule-based: gunakan restore secondary agar konsisten (include dmart ditangani restore)
	rOpts := restoremodel.RestoreSecondaryOptions{
		Profile:         *tgtProfile,
		DropTarget:      true,
		SkipBackup:      !opts.Common.PrebackupTarget,
		File:            srcDump,
		Ticket:          opts.Common.Ticket,
		IncludeDmart:    opts.Common.IncludeDmart,
		AutoDetectDmart: false,
		CompanionFile:   dmartDump,
		ClientCode:      clientCode,
		Instance:        instance,
		From:            "file",
		EncryptionKey:   encKey,
		DryRun:          false,
		Force:           nonInteractive,
		StopOnError:     !opts.Common.ContinueOnError,
	}
	return runRestoreSecondary(ctx, d, &rOpts)
}

func executeP2P(_ *cobra.Command, d *deps.Dependencies, opts P2POptions) error {
	if d == nil || d.Logger == nil || d.Config == nil {
		return fmt.Errorf("dependensi tidak tersedia")
	}

	nonInteractive := runtimecfg.IsQuiet() || opts.Common.SkipConfirm

	workdir, cleanup, err := ensureWorkdir(opts.Common.Workdir)
	if err != nil {
		return err
	}
	if strings.TrimSpace(opts.Common.Workdir) == "" {
		defer cleanup()
	}

	srcProfile, err := loadProfile(d.Config, opts.Common.SourceProfile, opts.Common.SourceProfileKey, consts.ENV_SOURCE_PROFILE, consts.ENV_SOURCE_PROFILE_KEY, true, !nonInteractive, "source")
	if err != nil {
		return err
	}
	targetPath, targetKey := resolveTargetProfileEffective(srcProfile, opts.Common.TargetProfile, opts.Common.TargetProfileKey)
	tgtProfile, err := loadProfile(d.Config, targetPath, targetKey, consts.ENV_TARGET_PROFILE, consts.ENV_TARGET_PROFILE_KEY, true, !nonInteractive, "target")
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	srcClient, err := connectForOps(srcProfile)
	if err != nil {
		return err
	}
	defer srcClient.Close()

	sourceDB := opts.SourceDB
	targetDB := opts.TargetDB

	if sourceDB == "" && targetDB == "" {
		primary, err := resolvePrimaryOnServer(ctx, srcClient, opts.ClientCode)
		if err != nil {
			return err
		}
		sourceDB = primary
		prefix := naming.InferPrimaryPrefix(sourceDB, "")
		targetDB = naming.BuildPrimaryDBName(prefix, opts.TargetClientCode)
	}

	if err := enforceNotCopyToSelf(srcProfile, tgtProfile, sourceDB, targetDB, "p2p"); err != nil {
		return err
	}

	companionSource := naming.BuildCompanionDBName(sourceDB)
	hasCompanion := false
	if opts.Common.IncludeDmart {
		ok, err := databaseExists(ctx, srcClient, companionSource)
		if err != nil {
			return err
		}
		hasCompanion = ok
	}

	if opts.Common.DryRun {
		d.Logger.Infof("[DRY-RUN] Source DB: %s", sourceDB)
		d.Logger.Infof("[DRY-RUN] Target DB: %s", targetDB)
		d.Logger.Infof("[DRY-RUN] Include dmart: %v (exists=%v)", opts.Common.IncludeDmart, hasCompanion)
		d.Logger.Infof("[DRY-RUN] Workdir: %s", workdir)
		return nil
	}

	encKey, err := resolveBackupEncryptionKeyForRestore(d.Config)
	if err != nil {
		return fmt.Errorf("gagal resolve backup encryption key untuk restore: %w", err)
	}

	srcDump, err := backupSingleDatabase(ctx, d.Logger, d.Config, srcProfile, srcClient, sourceDB, opts.Common.Ticket, workdir, opts.Common.ExcludeData)
	if err != nil {
		return err
	}
	var dmartDump string
	if hasCompanion {
		dmartDump, err = backupSingleDatabase(ctx, d.Logger, d.Config, srcProfile, srcClient, companionSource, opts.Common.Ticket, workdir, opts.Common.ExcludeData)
		if err != nil {
			if opts.Common.ContinueOnError {
				d.Logger.Warnf("Gagal backup companion (dmart), skip karena continue-on-error: %v", err)
				dmartDump = ""
			} else {
				return err
			}
		}
	}

	rOpts := restoremodel.RestorePrimaryOptions{
		Profile:            *tgtProfile,
		DropTarget:         true,
		SkipBackup:         !opts.Common.PrebackupTarget,
		File:               srcDump,
		CompanionFile:      dmartDump,
		Ticket:             opts.Common.Ticket,
		TargetDB:           targetDB,
		IncludeDmart:       opts.Common.IncludeDmart,
		AutoDetectDmart:    false,
		ConfirmIfNotExists: false,
		EncryptionKey:      encKey,
		DryRun:             false,
		Force:              nonInteractive,
		StopOnError:        !opts.Common.ContinueOnError,
		SkipGrants:         true,
	}

	return runRestorePrimary(ctx, d, &rOpts)
}

func executeS2S(_ *cobra.Command, d *deps.Dependencies, opts S2SOptions) error {
	if d == nil || d.Logger == nil || d.Config == nil {
		return fmt.Errorf("dependensi tidak tersedia")
	}

	nonInteractive := runtimecfg.IsQuiet() || opts.Common.SkipConfirm

	workdir, cleanup, err := ensureWorkdir(opts.Common.Workdir)
	if err != nil {
		return err
	}
	if strings.TrimSpace(opts.Common.Workdir) == "" {
		defer cleanup()
	}

	srcProfile, err := loadProfile(d.Config, opts.Common.SourceProfile, opts.Common.SourceProfileKey, consts.ENV_SOURCE_PROFILE, consts.ENV_SOURCE_PROFILE_KEY, true, !nonInteractive, "source")
	if err != nil {
		return err
	}
	targetPath, targetKey := resolveTargetProfileEffective(srcProfile, opts.Common.TargetProfile, opts.Common.TargetProfileKey)
	tgtProfile, err := loadProfile(d.Config, targetPath, targetKey, consts.ENV_TARGET_PROFILE, consts.ENV_TARGET_PROFILE_KEY, true, !nonInteractive, "target")
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	srcClient, err := connectForOps(srcProfile)
	if err != nil {
		return err
	}
	defer srcClient.Close()

	sourceDB := opts.SourceDB
	targetDB := opts.TargetDB
	if sourceDB == "" && targetDB == "" {
		// Source secondary name dibangun dari primary + source-instance.
		primary, err := resolvePrimaryOnServer(ctx, srcClient, opts.ClientCode)
		if err != nil {
			return err
		}
		sourceDB = naming.BuildSecondaryDBName(primary, opts.SourceInstance)
		targetDB = naming.BuildSecondaryDBName(primary, opts.TargetInstance)
	}

	if err := enforceNotCopyToSelf(srcProfile, tgtProfile, sourceDB, targetDB, "s2s"); err != nil {
		return err
	}

	companionSource := naming.BuildCompanionDBName(sourceDB)
	hasCompanion := false
	if opts.Common.IncludeDmart {
		ok, err := databaseExists(ctx, srcClient, companionSource)
		if err != nil {
			return err
		}
		hasCompanion = ok
	}

	if opts.Common.DryRun {
		d.Logger.Infof("[DRY-RUN] Source DB: %s", sourceDB)
		d.Logger.Infof("[DRY-RUN] Target DB: %s", targetDB)
		d.Logger.Infof("[DRY-RUN] Include dmart: %v (exists=%v)", opts.Common.IncludeDmart, hasCompanion)
		d.Logger.Infof("[DRY-RUN] Workdir: %s", workdir)
		return nil
	}

	encKey, err := resolveBackupEncryptionKeyForRestore(d.Config)
	if err != nil {
		return fmt.Errorf("gagal resolve backup encryption key untuk restore: %w", err)
	}

	srcDump, err := backupSingleDatabase(ctx, d.Logger, d.Config, srcProfile, srcClient, sourceDB, opts.Common.Ticket, workdir, opts.Common.ExcludeData)
	if err != nil {
		return err
	}
	var dmartDump string
	if hasCompanion {
		dmartDump, err = backupSingleDatabase(ctx, d.Logger, d.Config, srcProfile, srcClient, companionSource, opts.Common.Ticket, workdir, opts.Common.ExcludeData)
		if err != nil {
			if opts.Common.ContinueOnError {
				d.Logger.Warnf("Gagal backup companion (dmart), skip karena continue-on-error: %v", err)
				dmartDump = ""
			} else {
				return err
			}
		}
	}

	// Restore: jika mode eksplisit, restore single. Jika rule-based, tetap restore secondary.
	if opts.SourceDB != "" || opts.TargetDB != "" {
		if err := runRestoreSingle(ctx, d, tgtProfile, srcDump, targetDB, opts.Common, encKey); err != nil {
			return err
		}
		if dmartDump != "" {
			if err := runRestoreSingle(ctx, d, tgtProfile, dmartDump, naming.BuildCompanionDBName(targetDB), opts.Common, encKey); err != nil {
				if opts.Common.ContinueOnError {
					d.Logger.Warnf("Gagal restore companion (dmart), skip karena continue-on-error: %v", err)
					return nil
				}
				return err
			}
		}
		return nil
	}

	rOpts := restoremodel.RestoreSecondaryOptions{
		Profile:         *tgtProfile,
		DropTarget:      true,
		SkipBackup:      !opts.Common.PrebackupTarget,
		File:            srcDump,
		Ticket:          opts.Common.Ticket,
		IncludeDmart:    opts.Common.IncludeDmart,
		AutoDetectDmart: false,
		CompanionFile:   dmartDump,
		ClientCode:      opts.ClientCode,
		Instance:        opts.TargetInstance,
		From:            "file",
		EncryptionKey:   encKey,
		DryRun:          false,
		Force:           nonInteractive,
		StopOnError:     !opts.Common.ContinueOnError,
	}
	return runRestoreSecondary(ctx, d, &rOpts)
}

func runRestorePrimary(ctx context.Context, d *deps.Dependencies, opts *restoremodel.RestorePrimaryOptions) error {
	svc := restore.NewRestoreService(d.Logger, d.Config, opts)
	if err := svc.SetupRestorePrimarySession(ctx); err != nil {
		return err
	}
	_, err := svc.ExecuteRestorePrimary(ctx)
	_ = svc.Close()
	return err
}

func runRestoreSecondary(ctx context.Context, d *deps.Dependencies, opts *restoremodel.RestoreSecondaryOptions) error {
	svc := restore.NewRestoreService(d.Logger, d.Config, opts)
	if err := svc.SetupRestoreSecondarySession(ctx); err != nil {
		return err
	}
	_, err := svc.ExecuteRestoreSecondary(ctx)
	_ = svc.Close()
	return err
}

func runRestoreSingle(ctx context.Context, d *deps.Dependencies, profile *domain.ProfileInfo, file string, targetDB string, common CommonOptions, encryptionKey string) error {
	nonInteractive := runtimecfg.IsQuiet() || common.SkipConfirm
	opts := &restoremodel.RestoreSingleOptions{
		Profile:       *profile,
		DropTarget:    true,
		SkipBackup:    !common.PrebackupTarget,
		File:          filepath.Clean(file),
		Ticket:        common.Ticket,
		TargetDB:      targetDB,
		EncryptionKey: encryptionKey,
		SkipGrants:    true,
		DryRun:        false,
		Force:         nonInteractive,
		StopOnError:   !common.ContinueOnError,
	}
	svc := restore.NewRestoreService(d.Logger, d.Config, opts)
	if err := svc.SetupRestoreSession(ctx); err != nil {
		return err
	}
	_, err := svc.ExecuteRestoreSingle(ctx)
	_ = svc.Close()
	return err
}
