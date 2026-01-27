// File : internal/app/dbcopy/service.go
// Deskripsi : Core service untuk db-copy operations (streaming backup → restore)
// Author : Hadiyatna Muflihun
// Tanggal : 26 Januari 2026
// Last Modified : 27 Januari 2026
package dbcopy

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sfdbtools/internal/app/backup/execution"
	backuppath "sfdbtools/internal/app/backup/helpers/path"
	"sfdbtools/internal/app/backup/model/types_backup"
	profileconn "sfdbtools/internal/app/profile/connection"
	"sfdbtools/internal/app/profile/helpers/loader"
	"sfdbtools/internal/app/restore"
	restoremodel "sfdbtools/internal/app/restore/model"
	"sfdbtools/internal/crypto"
	"sfdbtools/internal/domain"
	appconfig "sfdbtools/internal/services/config"
	applog "sfdbtools/internal/services/log"
	"sfdbtools/internal/shared/compress"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/database"
	"sfdbtools/internal/shared/errorlog"
	"sfdbtools/internal/shared/naming"
)

// Service handles db-copy orchestration (backup source → restore target)
type Service struct {
	log    applog.Logger
	cfg    *appconfig.Config
	errLog *errorlog.ErrorLogger
}

func NewService(log applog.Logger, cfg *appconfig.Config) *Service {
	logDir := consts.DefaultLogDir
	if cfg != nil && cfg.Log.Output.File.Dir != "" {
		logDir = cfg.Log.Output.File.Dir
	}
	return &Service{
		log:    log,
		cfg:    cfg,
		errLog: errorlog.NewErrorLogger(log, logDir, consts.FeatureBackup),
	}
}

// ============================================================================
// Profile & Connection Helpers
// ============================================================================

func (s *Service) LoadProfile(profilePath, profileKey string, envPath, envKey string, allowInteractive bool, purpose string) (*domain.ProfileInfo, error) {
	configDir := ""
	if s.cfg != nil {
		configDir = s.cfg.ConfigDir.DatabaseProfile
	}
	return loader.ResolveAndLoadProfile(loader.ProfileLoadOptions{
		ConfigDir:        configDir,
		ProfilePath:      profilePath,
		ProfileKey:       profileKey,
		EnvProfilePath:   envPath,
		EnvProfileKey:    envKey,
		RequireProfile:   true,
		ProfilePurpose:   purpose,
		AllowInteractive: allowInteractive,
	})
}

func (s *Service) ConnectDB(profile *domain.ProfileInfo) (*database.Client, error) {
	if profile == nil {
		return nil, fmt.Errorf("profile nil")
	}
	client, err := profileconn.ConnectWithProfile(s.cfg, profile, "")
	if err != nil {
		return nil, fmt.Errorf("gagal connect ke database: %w", err)
	}
	return client, nil
}

// ============================================================================
// Database Discovery & Validation
// ============================================================================

func (s *Service) ResolvePrimaryDB(ctx context.Context, client interface{}, clientCode string) (string, error) {
	// Type assert to *database.Client
	dbClient, ok := client.(*database.Client)
	if !ok {
		return "", fmt.Errorf("client type assertion failed: expected *database.Client")
	}

	cc := strings.TrimSpace(clientCode)
	ccLower := strings.ToLower(cc)

	// Jika sudah dalam format primary, langsung return
	if strings.HasPrefix(ccLower, consts.PrimaryPrefixNBC) || strings.HasPrefix(ccLower, consts.PrimaryPrefixBiznet) {
		return cc, nil
	}

	// Auto-detect: coba NBC dulu, fallback Biznet
	nbc := naming.BuildPrimaryDBName(consts.PrimaryPrefixNBC, cc)
	biz := naming.BuildPrimaryDBName(consts.PrimaryPrefixBiznet, cc)

	nbcExists, err := dbClient.CheckDatabaseExists(ctx, nbc)
	if err != nil {
		return "", fmt.Errorf("gagal cek database primary (NBC): %w", err)
	}
	if nbcExists {
		return nbc, nil
	}

	bizExists, err := dbClient.CheckDatabaseExists(ctx, biz)
	if err != nil {
		return "", fmt.Errorf("gagal cek database primary (Biznet): %w", err)
	}
	if bizExists {
		return biz, nil
	}

	return "", fmt.Errorf("database primary tidak ditemukan untuk client-code %q (coba: %s atau %s)", cc, nbc, biz)
}

func (s *Service) CheckDatabaseExists(ctx context.Context, client *database.Client, dbName string) (bool, error) {
	if client == nil {
		return false, fmt.Errorf("client nil")
	}
	return client.CheckDatabaseExists(ctx, dbName)
}

func (s *Service) ValidateNotCopyToSelf(srcProfile, tgtProfile *domain.ProfileInfo, sourceDB, targetDB string, mode string) error {
	if srcProfile == nil || tgtProfile == nil {
		return nil
	}

	srcDB := strings.TrimSpace(sourceDB)
	tgtDB := strings.TrimSpace(targetDB)
	if srcDB == "" || tgtDB == "" {
		return nil
	}

	srcEff := profileconn.EffectiveDBInfo(srcProfile)
	tgtEff := profileconn.EffectiveDBInfo(tgtProfile)

	srcHost := strings.ToLower(strings.TrimSpace(srcEff.Host))
	tgtHost := strings.ToLower(strings.TrimSpace(tgtEff.Host))
	srcUser := strings.TrimSpace(srcProfile.DBInfo.User)
	tgtUser := strings.TrimSpace(tgtProfile.DBInfo.User)

	sameEndpoint := srcEff.Port == tgtEff.Port && srcHost == tgtHost && srcUser == tgtUser

	// P2P khusus: harus beda server
	if strings.EqualFold(mode, "p2p") && sameEndpoint && !strings.EqualFold(srcDB, tgtDB) {
		return fmt.Errorf("db-copy p2p ditolak: source dan target berada di server yang sama (host=%s port=%d user=%s). Untuk p2p, target harus beda server via --target-profile (atau gunakan p2s/s2s)", srcHost, srcEff.Port, srcUser)
	}

	// Semua mode: tolak jika endpoint + database sama
	if sameEndpoint && strings.EqualFold(srcDB, tgtDB) {
		return fmt.Errorf("db-copy ditolak: source dan target menunjuk database yang sama (host=%s port=%d user=%s db=%s)", srcHost, srcEff.Port, srcUser, srcDB)
	}

	return nil
}

// ============================================================================
// Workdir Management
// ============================================================================

func (s *Service) EnsureWorkdir(base string) (workdir string, cleanup func(), err error) {
	if strings.TrimSpace(base) != "" {
		if err := os.MkdirAll(base, 0o755); err != nil {
			return "", nil, err
		}
		return base, func() {}, nil
	}

	wd, err := os.MkdirTemp("", "sfdbtools-db-copy-")
	if err != nil {
		return "", nil, err
	}
	cleanup = func() { _ = os.RemoveAll(wd) }
	return wd, cleanup, nil
}

// ============================================================================
// Backup Operations
// ============================================================================

func (s *Service) BackupSingleDB(ctx context.Context, profile *domain.ProfileInfo, client *database.Client, dbName, ticket, workdir string, excludeData bool) (string, error) {
	if s.cfg == nil {
		return "", fmt.Errorf("config tidak tersedia")
	}

	opts := &types_backup.BackupDBOptions{
		Profile: *profile,
		Ticket:  ticket,
		Mode:    consts.ModeSingle,
		DBName:  dbName,
	}

	// Config defaults
	opts.Filter.ExcludeData = s.cfg.Backup.Exclude.Data || excludeData
	opts.Compression.Enabled = s.cfg.Backup.Compression.Enabled
	opts.Compression.Type = s.cfg.Backup.Compression.Type
	opts.Compression.Level = s.cfg.Backup.Compression.Level
	opts.Encryption.Enabled = s.cfg.Backup.Encryption.Enabled
	opts.Encryption.Key = s.cfg.Backup.Encryption.Key

	hostname := profile.DBInfo.HostName
	if hostname == "" {
		hostname = profile.DBInfo.Host
	}

	compressionType := compress.CompressionType("")
	if opts.Compression.Enabled {
		compressionType = compress.CompressionType(opts.Compression.Type)
	}

	filename, err := backuppath.GenerateBackupFilename(dbName, consts.ModeSingle, hostname, compressionType, opts.Encryption.Enabled, false)
	if err != nil {
		return "", err
	}
	outputPath := filepath.Join(workdir, filename)

	eng := execution.New(s.log, s.cfg, opts, s.errLog).WithDependencies(client, nil, nil, nil, nil)

	_, err = eng.ExecuteAndBuildBackup(ctx, types_backup.BackupExecutionConfig{
		DBName:       dbName,
		OutputPath:   outputPath,
		IsMultiDB:    false,
		TotalDBFound: 1,
	})
	if err != nil {
		return "", err
	}

	return outputPath, nil
}

// ============================================================================
// Restore Operations
// ============================================================================

func (s *Service) ResolveBackupEncryptionKey() (string, error) {
	if s.cfg == nil {
		return "", fmt.Errorf("config tidak tersedia")
	}
	if !s.cfg.Backup.Encryption.Enabled {
		return "", nil
	}
	key, _, err := crypto.ResolveKey(s.cfg.Backup.Encryption.Key, consts.ENV_BACKUP_ENCRYPTION_KEY, false)
	return key, err
}

func (s *Service) RestoreSingle(ctx context.Context, profile *domain.ProfileInfo, file, targetDB, ticket, encryptionKey string, dropTarget, skipBackup, skipGrants, continueOnError, nonInteractive bool) error {
	opts := &restoremodel.RestoreSingleOptions{
		Profile:       *profile,
		DropTarget:    dropTarget,
		SkipBackup:    skipBackup,
		File:          filepath.Clean(file),
		Ticket:        ticket,
		TargetDB:      targetDB,
		EncryptionKey: encryptionKey,
		SkipGrants:    skipGrants,
		DryRun:        false,
		// db-copy adalah orchestrator: hindari prompt restore (backup/drop/grants, dll).
		// Semua opsi penting sudah ditentukan di layer db-copy.
		Force:       true,
		StopOnError: !continueOnError,
	}

	svc := restore.NewRestoreService(s.log, s.cfg, opts)
	if err := svc.SetupRestoreSession(ctx); err != nil {
		return err
	}
	_, err := svc.ExecuteRestoreSingle(ctx)
	_ = svc.Close()
	return err
}

func (s *Service) RestorePrimary(ctx context.Context, profile *domain.ProfileInfo, file, companionFile, targetDB, ticket, encryptionKey string, includeDmart, dropTarget, skipBackup, skipGrants, continueOnError, nonInteractive bool) error {
	opts := &restoremodel.RestorePrimaryOptions{
		Profile:            *profile,
		DropTarget:         dropTarget,
		SkipBackup:         skipBackup,
		File:               file,
		CompanionFile:      companionFile,
		Ticket:             ticket,
		TargetDB:           targetDB,
		IncludeDmart:       includeDmart,
		AutoDetectDmart:    false,
		ConfirmIfNotExists: false,
		EncryptionKey:      encryptionKey,
		DryRun:             false,
		// db-copy adalah orchestrator: hindari prompt restore (backup/drop/grants, dll).
		Force:       true,
		StopOnError: !continueOnError,
		SkipGrants:  skipGrants,
	}

	svc := restore.NewRestoreService(s.log, s.cfg, opts)
	if err := svc.SetupRestorePrimarySession(ctx); err != nil {
		return err
	}
	_, err := svc.ExecuteRestorePrimary(ctx)
	_ = svc.Close()
	return err
}

func (s *Service) RestoreSecondary(ctx context.Context, profile *domain.ProfileInfo, file, companionFile, ticket, clientCode, instance, encryptionKey string, includeDmart, dropTarget, skipBackup, continueOnError, nonInteractive bool) error {
	opts := &restoremodel.RestoreSecondaryOptions{
		Profile:         *profile,
		DropTarget:      dropTarget,
		SkipBackup:      skipBackup,
		File:            file,
		Ticket:          ticket,
		IncludeDmart:    includeDmart,
		AutoDetectDmart: false,
		CompanionFile:   companionFile,
		ClientCode:      clientCode,
		Instance:        instance,
		From:            "file",
		EncryptionKey:   encryptionKey,
		DryRun:          false,
		// db-copy adalah orchestrator: hindari prompt restore (backup/drop/grants, dll).
		Force:       true,
		StopOnError: !continueOnError,
	}

	svc := restore.NewRestoreService(s.log, s.cfg, opts)
	if err := svc.SetupRestoreSecondarySession(ctx); err != nil {
		return err
	}
	_, err := svc.ExecuteRestoreSecondary(ctx)
	_ = svc.Close()
	return err
}
