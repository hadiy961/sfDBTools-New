// File : internal/restore/restore_helpers.go
// Deskripsi : Shared helper functions untuk restore executors
// Author : Hadiyatna Muflihun
// Tanggal : 17 Desember 2025
// Last Modified : 5 Januari 2026
package restore

import (
	"context"
	"fmt"
	"strings"

	"sfdbtools/internal/app/restore/helpers"
	restoremodel "sfdbtools/internal/app/restore/model"
	"sfdbtools/internal/ui/prompt"
)

// BackupDatabaseIfNeeded melakukan backup database jika diperlukan
func (s *Service) BackupDatabaseIfNeeded(ctx context.Context, dbName string, dbExists bool, skipBackup bool, backupOpts *restoremodel.RestoreBackupOptions) (string, error) {
	if skipBackup {
		return "", nil
	}

	if !dbExists {
		s.Log.Infof("Database %s belum ada, skip backup pre-restore", dbName)
		return "", nil
	}

	s.Log.Infof("Database %s sudah ada, melakukan backup pre-restore...", dbName)
	backupFile, err := s.BackupTargetDatabase(ctx, dbName, backupOpts)
	if err != nil {
		return "", fmt.Errorf("gagal backup database target: %w", err)
	}

	s.Log.Infof("Backup database berhasil: %s", backupFile)
	return backupFile, nil
}

// DropDatabaseIfNeeded melakukan drop database jika diperlukan
func (s *Service) DropDatabaseIfNeeded(ctx context.Context, dbName string, dbExists bool, shouldDrop bool) error {
	if !shouldDrop || !dbExists {
		return nil
	}

	if err := s.TargetClient.DropDatabase(ctx, dbName); err != nil {
		return fmt.Errorf("gagal drop database target: %w", err)
	}

	s.Log.Infof("Database %s berhasil di-drop", dbName)
	return nil
}

// CreateAndRestoreDatabase membuat database dan restore dari file
func (s *Service) CreateAndRestoreDatabase(ctx context.Context, dbName string, filePath string, encryptionKey string) error {
	// Create database if not exists
	if err := s.TargetClient.CreateDatabaseIfNotExists(ctx, dbName); err != nil {
		return fmt.Errorf("gagal membuat database: %w", err)
	}

	// Restore from file with retry loop
	for {
		sum, err := helpers.RestoreFromFile(ctx, filePath, dbName, s.Profile, encryptionKey, s.Log)
		if sum != nil {
			s.AddSQLIssueCounters(sum.SQLErrors, sum.SQLWarnings)
			if len(sum.ErrLines) > 0 {
				logPath := s.ErrorLog.LogWithOutput(map[string]interface{}{
					"action":   fmt.Sprintf("restore_database_%s", s.GetCurrentMode()),
					"database": dbName,
					"file":     filePath,
				}, strings.Join(sum.ErrLines, "\n"), fmt.Errorf("terdeteksi %d SQL error selama restore database", sum.SQLErrors))
				
				if logPath != "" {
					s.Log.Warnf("Terdeteksi %d SQL error (dan %d peringatan) selama restore. Detail log tersimpan di: %s", sum.SQLErrors, sum.SQLWarnings, logPath)
				}
			}
		}
		if err != nil {
			if s.IsInteractive() {
				s.Log.Warnf("Restore database gagal: %v", err)
				retry, promptErr := prompt.Confirm("Proses restore database gagal. Ingin mencoba ulang?", true)
				if promptErr == nil && retry {
					s.Log.Info("Mencoba ulang proses restore database...")
					continue
				}
			}
			return fmt.Errorf("gagal restore database: %w", err)
		}
		break // Success
	}

	return nil
}

// RestoreUserGrantsIfAvailable restore user grants jika file tersedia
func (s *Service) RestoreUserGrantsIfAvailable(ctx context.Context, grantsFile string) (bool, error) {
	if grantsFile == "" {
		return false, nil
	}

	for {
		sum, err := helpers.RestoreUserGrants(ctx, grantsFile, s.Profile, s.Log)
		if sum != nil {
			s.AddSQLIssueCounters(sum.SQLErrors, sum.SQLWarnings)
			if len(sum.ErrLines) > 0 {
				logPath := s.ErrorLog.LogWithOutput(map[string]interface{}{
					"action": fmt.Sprintf("restore_user_grants_%s", s.GetCurrentMode()),
					"file":   grantsFile,
				}, strings.Join(sum.ErrLines, "\n"), fmt.Errorf("terdeteksi %d SQL error selama restore user grants", sum.SQLErrors))
				
				if logPath != "" {
					s.Log.Warnf("Terdeteksi %d SQL error (dan %d peringatan) pada user grants. Detail log tersimpan di: %s", sum.SQLErrors, sum.SQLWarnings, logPath)
				}
			}
		}
		if err != nil {
			if s.IsInteractive() {
				s.Log.Warnf("Restore user grants gagal: %v", err)
				retry, promptErr := prompt.Confirm("Proses restore user grants gagal. Ingin mencoba ulang?", true)
				if promptErr == nil && retry {
					s.Log.Info("Mencoba ulang proses restore user grants...")
					continue
				}
			}
			return false, err
		}
		break // Success
	}

	s.Log.Info("User grants berhasil di-restore")
	return true, nil
}
