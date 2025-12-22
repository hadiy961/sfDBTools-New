package backup

import (
	"context"
	"fmt"
	"sfDBTools/internal/types/types_backup"
	"sfDBTools/pkg/consts"
)

// ExecuteBackupLoop menjalankan backup untuk multiple databases dengan pattern yang sama
func (s *Service) ExecuteBackupLoop(ctx context.Context, databases []string, config types_backup.BackupLoopConfig, outputPathFunc func(dbName string) (string, error)) types_backup.BackupLoopResult {
	result := types_backup.BackupLoopResult{
		BackupInfos: make([]types_backup.DatabaseBackupInfo, 0),
		FailedDBs:   make([]types_backup.FailedDatabaseInfo, 0),
		Errors:      make([]string, 0),
	}

	if len(databases) == 0 {
		s.Log.Warn("Tidak ada database yang dipilih untuk backup")
		result.Errors = append(result.Errors, "tidak ada database yang dipilih")
		return result
	}

	for idx, dbName := range databases {
		if ctx.Err() != nil {
			s.Log.Warn("Proses backup dibatalkan")
			result.Errors = append(result.Errors, "Backup dibatalkan oleh user")
			break
		}

		s.Log.Infof("[%d/%d] Backup database: %s", idx+1, len(databases), dbName)

		// Generate output path
		outputPath, err := outputPathFunc(dbName)
		if err != nil {
			msg := fmt.Sprintf("gagal generate path untuk %s: %v", dbName, err)
			s.Log.Error(msg)
			result.FailedDBs = append(result.FailedDBs, types_backup.FailedDatabaseInfo{DatabaseName: dbName, Error: msg})
			result.Failed++
			continue
		}

		// Execute backup
		backupInfo, err := s.ExecuteAndBuildBackup(ctx, types_backup.BackupExecutionConfig{
			DBName:       dbName,
			OutputPath:   outputPath,
			BackupType:   config.BackupType,
			TotalDBFound: config.TotalDBs,
			IsMultiDB:    false,
		})

		if err != nil {
			result.FailedDBs = append(result.FailedDBs, types_backup.FailedDatabaseInfo{DatabaseName: dbName, Error: err.Error()})
			result.Failed++
			continue
		}

		result.BackupInfos = append(result.BackupInfos, backupInfo)
		result.Success++

		// Export user grants for separated/single modes
		if config.Mode == consts.ModeSeparated || config.Mode == consts.ModeSeparate || config.Mode == consts.ModeSingle {
			path := s.ExportUserGrantsIfNeeded(ctx, outputPath, []string{dbName})
			if s.Config.Backup.Output.SaveBackupInfo {
				s.UpdateMetadataUserGrantsPath(outputPath, path)
			}
		}
	}

	return result
}
