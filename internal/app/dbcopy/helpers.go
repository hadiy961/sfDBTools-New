// File : internal/app/dbcopy/helpers.go
// Deskripsi : Helper untuk orkestrasi db-copy
// Author : Hadiyatna Muflihun
// Tanggal : 26 Januari 2026
// Last Modified : 26 Januari 2026
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

func ensureWorkdir(base string) (string, func(), error) {
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
	cleanup := func() { _ = os.RemoveAll(wd) }
	return wd, cleanup, nil
}

func loadProfile(cfg *appconfig.Config, profilePath, profileKey string, envPath, envKey string, require bool, allowInteractive bool, purpose string) (*domain.ProfileInfo, error) {
	configDir := ""
	if cfg != nil {
		configDir = cfg.ConfigDir.DatabaseProfile
	}
	return loader.ResolveAndLoadProfile(loader.ProfileLoadOptions{
		ConfigDir:        configDir,
		ProfilePath:      profilePath,
		ProfileKey:       profileKey,
		EnvProfilePath:   envPath,
		EnvProfileKey:    envKey,
		RequireProfile:   require,
		ProfilePurpose:   purpose,
		AllowInteractive: allowInteractive,
	})
}

func resolveTargetProfileEffective(source *domain.ProfileInfo, targetPath, targetKey string) (path string, key string) {
	if strings.TrimSpace(targetPath) == "" {
		return strings.TrimSpace(source.Path), strings.TrimSpace(source.EncryptionKey)
	}
	return strings.TrimSpace(targetPath), strings.TrimSpace(targetKey)
}

func resolvePrimaryOnServer(ctx context.Context, client *database.Client, clientCode string) (string, error) {
	cc := strings.TrimSpace(clientCode)
	ccLower := strings.ToLower(cc)
	if strings.HasPrefix(ccLower, consts.PrimaryPrefixNBC) || strings.HasPrefix(ccLower, consts.PrimaryPrefixBiznet) {
		return cc, nil
	}

	nbc := naming.BuildPrimaryDBName(consts.PrimaryPrefixNBC, cc)
	biz := naming.BuildPrimaryDBName(consts.PrimaryPrefixBiznet, cc)

	nbcExists, nbcErr := client.CheckDatabaseExists(ctx, nbc)
	if nbcErr != nil {
		return "", fmt.Errorf("gagal mengecek database primary (NBC): %w", nbcErr)
	}
	bizExists, bizErr := client.CheckDatabaseExists(ctx, biz)
	if bizErr != nil {
		return "", fmt.Errorf("gagal mengecek database primary (Biznet): %w", bizErr)
	}

	if nbcExists {
		if bizExists {
			// prefer NBC untuk deterministik, sama seperti pola di restore.
			return nbc, nil
		}
		return nbc, nil
	}
	if bizExists {
		return biz, nil
	}

	return "", fmt.Errorf("database primary tidak ditemukan untuk client-code %q (coba: %s atau %s)", cc, nbc, biz)
}

func secondaryDBNameFromPrimary(primary string, instance string) string {
	return naming.BuildSecondaryDBName(primary, instance)
}

func buildCompressionType(opts *types_backup.BackupDBOptions) compress.CompressionType {
	if opts == nil {
		return compress.CompressionType("")
	}
	if !opts.Compression.Enabled {
		return compress.CompressionType("")
	}
	return compress.CompressionType(opts.Compression.Type)
}

func resolveBackupEncryptionKeyForRestore(cfg *appconfig.Config) (string, error) {
	if cfg == nil {
		return "", fmt.Errorf("config tidak tersedia")
	}
	if !cfg.Backup.Encryption.Enabled {
		return "", nil
	}
	key, _, err := crypto.ResolveKey(cfg.Backup.Encryption.Key, consts.ENV_BACKUP_ENCRYPTION_KEY, false)
	if err != nil {
		return "", err
	}
	return key, nil
}

func backupSingleDatabase(ctx context.Context, log applog.Logger, cfg *appconfig.Config, prof *domain.ProfileInfo, dbClient *database.Client, dbName string, ticket string, workdir string, excludeData bool) (string, error) {
	if cfg == nil {
		return "", fmt.Errorf("config tidak tersedia")
	}

	opts := &types_backup.BackupDBOptions{}
	// Minimal: isi profile + ticket + mode
	opts.Profile = *prof
	opts.Ticket = ticket
	opts.Mode = consts.ModeSingle
	opts.DBName = dbName

	// Filter defaults dari config, bisa di-override flag
	opts.Filter.ExcludeData = cfg.Backup.Exclude.Data
	if excludeData {
		opts.Filter.ExcludeData = true
	}

	// Default compression/encryption mengikuti config
	opts.Compression.Enabled = cfg.Backup.Compression.Enabled
	opts.Compression.Type = cfg.Backup.Compression.Type
	opts.Compression.Level = cfg.Backup.Compression.Level

	opts.Encryption.Enabled = cfg.Backup.Encryption.Enabled
	opts.Encryption.Key = cfg.Backup.Encryption.Key

	hostname := prof.DBInfo.HostName
	if hostname == "" {
		hostname = prof.DBInfo.Host
	}
	filename, err := backuppath.GenerateBackupFilename(dbName, consts.ModeSingle, hostname, buildCompressionType(opts), opts.Encryption.Enabled, false)
	if err != nil {
		return "", err
	}
	outputPath := filepath.Join(workdir, filename)

	logDir := cfg.Log.Output.File.Dir
	if logDir == "" {
		logDir = consts.DefaultLogDir
	}
	errLog := errorlog.NewErrorLogger(log, logDir, consts.FeatureBackup)

	eng := execution.New(log, cfg, opts, errLog).WithDependencies(
		dbClient,
		nil,
		nil,
		nil,
		nil,
	)

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

func databaseExists(ctx context.Context, client *database.Client, dbName string) (bool, error) {
	if client == nil {
		return false, fmt.Errorf("client nil")
	}
	return client.CheckDatabaseExists(ctx, dbName)
}

func connectForOps(profile *domain.ProfileInfo) (*database.Client, error) {
	return profileconn.ConnectWithProfile(nil, profile, consts.DefaultInitialDatabase)
}
