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
	return s.executeSafetyBackup(ctx, profile, client, dbName, "", "INTERNAL_SAFETY_BACKUP")
}

func (s *Service) runSafetyTableBackup(ctx context.Context, profile *domain.ProfileInfo, client *database.Client, dbName, tableName string) error {
	return s.executeSafetyBackup(ctx, profile, client, dbName, tableName, "INTERNAL_SAFETY_BACKUP_TABLE")
}

func (s *Service) executeSafetyBackup(ctx context.Context, profile *domain.ProfileInfo, client *database.Client, dbName, tableName, ticketPrefix string) error {
	ticket := s.ticket
	if ticket == "" {
		ticket = ticketPrefix
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
	label := dbName
	if tableName != "" {
		label = dbName + "_" + tableName
	}
	filename, _ := backuppath.GenerateBackupFilename(label, consts.ModeSingle, hostname, compress.CompressionType(consts.CompressionTypeGzip), false, false)

	outputDir := s.cfg.Backup.Output.BaseDirectory
	if outputDir == "" {
		outputDir = filepath.Join(os.TempDir(), "sfdbtools_safety")
		_ = os.MkdirAll(outputDir, 0755)
	}
	outputPath := filepath.Join(outputDir, filename)

	if tableName != "" {
		s.log.Infof("Backup tabel disimpan di: %s", outputPath)
	} else {
		s.log.Infof("Backup disimpan di: %s", outputPath)
	}

	eng := execution.New(s.log, s.cfg, opts, s.errLog).WithDependencies(client, nil, nil, nil, nil)
	_, err := eng.ExecuteAndBuildBackup(ctx, types_backup.BackupExecutionConfig{
		DBName:       dbName,
		OutputPath:   outputPath,
		IsMultiDB:    false,
		TotalDBFound: 1,
	})
	return err
}
