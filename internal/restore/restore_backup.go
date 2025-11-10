// File : internal/restore/restore_backup.go
// Deskripsi : Pre-backup functionality sebelum restore untuk safety
// Author : Hadiyatna Muflihun
// Tanggal : 2025-11-05
// Last Modified : 2025-11-05

package restore

import (
	"context"
	"fmt"
	"sfDBTools/internal/backup"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/helper"
	"time"
)

// executePreBackup melakukan backup database sebelum restore sebagai safety net
// Menggunakan existing backup infrastructure (backup.Service)
// Returns: backup file path dan error
func (s *Service) executePreBackup(ctx context.Context, targetDB string) (string, error) {
	s.Log.Infof("Membuat backup database '%s' sebelum restore...", targetDB)

	startTime := time.Now()

	// Build BackupDBOptions untuk backup service
	// Menggunakan struktur yang sama dengan backup command
	backupOpts := types.BackupDBOptions{
		Filter: types.FilterOptions{
			IncludeDatabases: []string{targetDB}, // Backup hanya database yang akan di-restore
			ExcludeDatabases: []string{},
		},
		Profile: types.ProfileInfo{
			Path:   s.TargetProfile.Path,
			Name:   s.TargetProfile.Name,
			DBInfo: s.TargetProfile.DBInfo,
		},
		DryRun: false,
		Force:  false,
		Mode:   "separated", // Gunakan separated mode untuk single database
		Entry: types.BackupEntryConfig{
			HeaderTitle: fmt.Sprintf("Pre-Restore Backup - %s", targetDB),
			Force:       false,
			SuccessMsg:  "",
			LogPrefix:   "[Pre-Restore Backup]",
			BackupMode:  "separated",
		},
	}

	// Setup output directory untuk pre-backup
	outputDir, err := helper.GenerateBackupDirectory(
		s.Config.Backup.Output.BaseDirectory,
		s.Config.Backup.Output.Structure.Pattern,
		s.TargetProfile.DBInfo.Host,
	)
	if err != nil {
		s.Log.Warnf("gagal generate output directory, menggunakan default: %v", err)
		outputDir = s.Config.Backup.Output.BaseDirectory
	}
	backupOpts.OutputDir = outputDir

	// Setup compression jika diaktifkan di config
	if s.Config.Backup.Compression.Enabled {
		backupOpts.Compression = types.CompressionOptions{
			Enabled: true,
			Type:    s.Config.Backup.Compression.Type,
			Level:   s.Config.Backup.Compression.Level,
		}
	}

	// Setup encryption jika diaktifkan di config
	if s.Config.Backup.Encryption.Enabled {
		backupOpts.Encryption = types.EncryptionOptions{
			Enabled: true,
			Key:     "", // Key akan di-prompt jika diperlukan saat backup
		}
	}

	// Inisialisasi backup service menggunakan existing constructor
	backupSvc := backup.NewBackupService(s.Log, s.Config, &backupOpts)

	// Set cancel function untuk graceful shutdown (tanpa timeout)
	// Backup bisa memakan waktu lama tergantung ukuran database
	backupCancel := func() {}
	backupSvc.SetCancelFunc(backupCancel)

	// Direct connect ke database target tanpa meminta user memilih profile
	// Karena kami sudah mempunyai target profile yang ter-load
	sourceClient, err := setupBackupConnection(s)
	if err != nil {
		return "", fmt.Errorf("gagal setup backup connection: %w", err)
	}
	defer sourceClient.Close()

	// Filter database hanya untuk target database
	dbFiltered := []string{targetDB}

	// Execute backup dan dapatkan result
	result, err := backupSvc.ExecuteBackup(ctx, sourceClient, dbFiltered, "separated")
	if err != nil {
		return "", fmt.Errorf("pre-backup gagal: %w", err)
	}

	duration := time.Since(startTime)
	s.Log.Infof("✓ Pre-backup selesai dalam %s", duration)

	// Ambil backup file path dari result
	backupFilePath := ""
	if result != nil && len(result.BackupInfo) > 0 {
		// Ambil path file dari backup info pertama (single database backup)
		backupFilePath = result.BackupInfo[0].OutputFile
	}

	if backupFilePath == "" {
		s.Log.Warn("Backup berhasil tapi tidak dapat menentukan path file backup")
		return "", nil // Tidak error, tapi tidak ada rollback capability
	}

	return backupFilePath, nil
}

// setupBackupConnection membuat koneksi ke database target untuk pre-backup
// Menggunakan target profile yang sudah ter-load dari restore service
func setupBackupConnection(s *Service) (*database.Client, error) {
	creds := types.SourceDBConnection{
		DBInfo: types.DBInfo{
			Host:     s.TargetProfile.DBInfo.Host,
			Port:     s.TargetProfile.DBInfo.Port,
			User:     s.TargetProfile.DBInfo.User,
			Password: s.TargetProfile.DBInfo.Password,
			HostName: s.TargetProfile.DBInfo.HostName,
		},
		Database: "mysql", // gunakan schema sistem untuk koneksi awal
	}

	client, err := database.ConnectToSourceDatabase(creds)
	if err != nil {
		return nil, fmt.Errorf("gagal koneksi ke target database untuk pre-backup: %w", err)
	}

	s.Log.Debugf("✓ Connected to target database for pre-backup: %s@%s:%d",
		creds.DBInfo.User, creds.DBInfo.Host, creds.DBInfo.Port)

	return client, nil
}
