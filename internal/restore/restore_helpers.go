// File : internal/restore/restore_helpers.go
// Deskripsi : Shared helper functions untuk restore executors
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-17
// Last Modified : 2025-12-17

package restore

import (
	"context"
	"fmt"
	"sfDBTools/internal/restore/helpers"
	"sfDBTools/internal/types"
)

// BackupDatabaseIfNeeded melakukan backup database jika diperlukan
func (s *Service) BackupDatabaseIfNeeded(ctx context.Context, dbName string, dbExists bool, skipBackup bool, backupOpts *types.RestoreBackupOptions) (string, error) {
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

	s.Log.Infof("Backup pre-restore berhasil: %s", backupFile)
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

	// Restore from file
	if err := helpers.RestoreFromFile(ctx, filePath, dbName, s.Profile, encryptionKey); err != nil {
		return fmt.Errorf("gagal restore database: %w", err)
	}

	return nil
}

// RestoreUserGrantsIfAvailable restore user grants jika file tersedia
func (s *Service) RestoreUserGrantsIfAvailable(ctx context.Context, grantsFile string) (bool, error) {
	if grantsFile == "" {
		return false, nil
	}

	if err := helpers.RestoreUserGrants(ctx, grantsFile, s.Profile); err != nil {
		return false, err
	}

	s.Log.Info("User grants berhasil di-restore")
	return true, nil
}
