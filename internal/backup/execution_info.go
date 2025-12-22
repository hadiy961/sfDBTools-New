package backup

import (
	"context"
	"fmt"
	"sfDBTools/internal/backup/metadata"
	"sfDBTools/internal/types"
	"sfDBTools/internal/types/types_backup"
	"sfDBTools/pkg/consts"
	pkghelper "sfDBTools/pkg/helper"
	"strings"
	"time"
)

// buildDryRunBackupInfo membangun dummy DatabaseBackupInfo untuk dry-run mode
func (s *Service) buildDryRunBackupInfo(cfg types_backup.BackupExecutionConfig, mysqldumpArgs []string, timer *pkghelper.Timer, startTime time.Time) types.DatabaseBackupInfo {
	backupDuration := timer.Elapsed()
	endTime := time.Now()

	// Log dry-run info
	if cfg.IsMultiDB {
		s.Log.Info("[DRY-RUN] Akan backup database: " + strings.Join(cfg.DBList, ", "))
	} else {
		s.Log.Infof("[DRY-RUN] Akan backup database: %s", cfg.DBName)
	}
	s.Log.Info("[DRY-RUN] Output file: " + cfg.OutputPath)
	s.Log.Debug("[DRY-RUN] Mysqldump command: mysqldump " + strings.Join(mysqldumpArgs, " "))
	s.Log.Info("[DRY-RUN] Backup tidak dijalankan (dry-run mode aktif)")

	// Format display name
	displayName := formatBackupDisplayName(cfg)

	return types.DatabaseBackupInfo{
		DatabaseName:  displayName,
		OutputFile:    cfg.OutputPath,
		FileSize:      0,
		FileSizeHuman: "0 B (dry-run)",
		Duration:      backupDuration.String(),
		Status:        consts.BackupStatusDryRun,
		Warnings:      "Backup tidak dijalankan - mode dry-run aktif",
		StartTime:     startTime,
		EndTime:       endTime,
		ManifestFile:  "",
	}
}

// buildBackupInfoFromResult membangun DatabaseBackupInfo dari result
func (s *Service) buildBackupInfoFromResult(ctx context.Context, cfg types_backup.BackupExecutionConfig, writeResult *types_backup.BackupWriteResult, timer *pkghelper.Timer, startTime time.Time, dbVersion string) types.DatabaseBackupInfo {
	fileSize := writeResult.FileSize
	stderrOutput := writeResult.StderrOutput
	backupDuration := timer.Elapsed()
	endTime := time.Now()

	// Determine status
	status := consts.BackupStatusSuccess
	if stderrOutput != "" {
		status = consts.BackupStatusSuccessWithWarnings
		if !cfg.IsMultiDB {
			s.Log.Warningf("Database %s backup dengan warning: %s", cfg.DBName, stderrOutput)
		}
	} else if !cfg.IsMultiDB {
		s.Log.Infof("✓ Database %s berhasil di-backup", cfg.DBName)
	}

	// Get metadata
	meta := s.generateBackupMetadata(ctx, cfg, writeResult, backupDuration, startTime, endTime, status, dbVersion)

	// Save and get manifest path
	manifestPath := ""
	if s.Config.Backup.Output.SaveBackupInfo {
		manifestPath = metadata.TrySaveBackupMetadata(meta, s.Log)
	}

	// Format display name
	displayName := formatBackupDisplayName(cfg)

	return (&metadata.DatabaseBackupInfoBuilder{
		DatabaseName: displayName,
		OutputFile:   cfg.OutputPath,
		FileSize:     fileSize,
		Duration:     backupDuration,
		Status:       status,
		Warnings:     stderrOutput,
		StartTime:    startTime,
		EndTime:      endTime,
		ManifestFile: manifestPath,
	}).Build()
}

// formatBackupDisplayName menyatukan logika penamaan display untuk single/combined
func formatBackupDisplayName(cfg types_backup.BackupExecutionConfig) string {
	if cfg.IsMultiDB {
		return fmt.Sprintf("Combined backup (%d databases)", len(cfg.DBList))
	}
	return cfg.DBName
}
