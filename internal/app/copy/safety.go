package copy

import (
	"context"
	"os"
	"path/filepath"

	"sfdbtools/internal/app/backup/execution"
	backuppath "sfdbtools/internal/app/backup/helpers/path"
	"sfdbtools/internal/app/backup/model/types_backup"
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
