package copy

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"sfdbtools/internal/app/backup/execution"
	backuppath "sfdbtools/internal/app/backup/helpers/path"
	"sfdbtools/internal/app/backup/model/types_backup"
	"sfdbtools/internal/app/restore"
	restoremodel "sfdbtools/internal/app/restore/model"
	"sfdbtools/internal/domain"
	"sfdbtools/internal/shared/compress"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/database"
)

// runSafetyBackup melakukan backup cepat ke direktori default config sebelum ditimpa.
func (s *Service) runSafetyBackup(ctx context.Context, profile *domain.ProfileInfo, client *database.Client, dbName string) error {
	ticket := s.ticket
	if ticket == "" {
		ticket = "INTERNAL_SAFETY_BACKUP"
	}

	opts := &types_backup.BackupDBOptions{
		Profile: *profile,
		Mode:    consts.ModeSingle,
		DBName:  dbName,
		Ticket:  ticket,
	}

	// Gunakan config default untuk kompresi agar cepat
	opts.Compression.Enabled = true
	opts.Compression.Type = consts.CompressionTypeGzip

	// Generate path
	hostname := profile.DBInfo.Host
	filename, _ := backuppath.GenerateBackupFilename(dbName, consts.ModeSingle, hostname, compress.CompressionType(consts.CompressionTypeGzip), false, false)

	outputDir := s.cfg.Backup.Output.BaseDirectory
	if outputDir == "" {
		outputDir = filepath.Join(os.TempDir(), "sfdbtools_safety")
		_ = os.MkdirAll(outputDir, 0755)
	}
	outputPath := filepath.Join(outputDir, filename)

	s.log.Infof("Backup disimpan di: %s", outputPath)

	eng := execution.New(s.log, s.cfg, opts, s.errLog).WithDependencies(client, nil, nil, nil, nil)
	_, err := eng.ExecuteAndBuildBackup(ctx, types_backup.BackupExecutionConfig{
		DBName:       dbName,
		OutputPath:   outputPath,
		IsMultiDB:    false,
		TotalDBFound: 1,
	})
	return err
}

func (s *Service) executeDiskCopy(ctx context.Context, profile *domain.ProfileInfo, client *database.Client, sourceDB, targetDB string, schemaOnly bool) error {
	baseDir := ""
	if s.cfg != nil {
		baseDir = s.cfg.Backup.Output.BaseDirectory
	}
	if baseDir == "" {
		baseDir = os.TempDir()
	}

	workdir := filepath.Join(baseDir, fmt.Sprintf("sfdbtools_copy_%d", time.Now().Unix()))
	if err := os.MkdirAll(workdir, 0755); err != nil {
		return err
	}
	defer os.RemoveAll(workdir)

	ticket := s.ticket
	if ticket == "" {
		ticket = "INTERNAL_DB_COPY"
	}

	// A. Backup
	opts := &types_backup.BackupDBOptions{
		Profile: *profile,
		Mode:    consts.ModeSingle,
		DBName:  sourceDB,
		Ticket:  ticket,
	}
	opts.Compression.Enabled = true
	opts.Compression.Type = consts.CompressionTypeGzip
	opts.Filter.ExcludeData = schemaOnly

	filename, err := backuppath.GenerateBackupFilename(sourceDB, consts.ModeSingle, profile.DBInfo.Host, compress.CompressionType(consts.CompressionTypeGzip), false, false)
	if err != nil {
		return err
	}
	outputPath := filepath.Join(workdir, filename)

	eng := execution.New(s.log, s.cfg, opts, s.errLog).WithDependencies(client, nil, nil, nil, nil)
	_, err = eng.ExecuteAndBuildBackup(ctx, types_backup.BackupExecutionConfig{
		DBName:       sourceDB,
		OutputPath:   outputPath,
		IsMultiDB:    false,
		TotalDBFound: 1,
	})
	if err != nil {
		return fmt.Errorf("gagal backup source: %w", err)
	}

	// B. Restore
	restOpts := &restoremodel.RestoreSingleOptions{
		Profile:     *profile,
		File:        outputPath,
		TargetDB:    targetDB,
		Force:       true, // Restore engine sudah support Force flag
		StopOnError: true,
		SkipBackup:  true, // Karena ini copy, kita asumsikan target bisa ditimpa
		DropTarget:  true,
		Ticket:      ticket,
	}

	restSvc := restore.NewRestoreService(s.log, s.cfg, restOpts)
	if err := restSvc.SetupRestoreSession(ctx); err != nil {
		return fmt.Errorf("gagal setup restore: %w", err)
	}
	defer restSvc.Close()

	_, err = restSvc.ExecuteRestoreSingle(ctx)
	return err
}

func (s *Service) runSafetyTableBackup(ctx context.Context, profile *domain.ProfileInfo, client *database.Client, dbName, tableName string) error {
	ticket := s.ticket
	if ticket == "" {
		ticket = "INTERNAL_SAFETY_BACKUP_TABLE"
	}

	opts := &types_backup.BackupDBOptions{
		Profile: *profile,
		Mode:    consts.ModeSingle,
		DBName:  dbName,
		Ticket:  ticket,
	}
	opts.Compression.Enabled = true
	opts.Compression.Type = consts.CompressionTypeGzip

	hostname := profile.DBInfo.Host
	filename, _ := backuppath.GenerateBackupFilename(dbName+"_"+tableName, consts.ModeSingle, hostname, compress.CompressionType(consts.CompressionTypeGzip), false, false)

	outputDir := s.cfg.Backup.Output.BaseDirectory
	if outputDir == "" {
		outputDir = filepath.Join(os.TempDir(), "sfdbtools_safety")
		_ = os.MkdirAll(outputDir, 0755)
	}
	outputPath := filepath.Join(outputDir, filename)

	s.log.Infof("Backup tabel disimpan di: %s", outputPath)

	eng := execution.New(s.log, s.cfg, opts, s.errLog).WithDependencies(client, nil, nil, nil, nil)

	// Gunakan argumen mysqldump untuk tabel spesifik
	_, err := eng.ExecuteAndBuildBackup(ctx, types_backup.BackupExecutionConfig{
		DBName:       dbName,
		OutputPath:   outputPath,
		IsMultiDB:    false,
		TotalDBFound: 1,
	})
	return err
}
