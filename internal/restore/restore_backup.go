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
	"sfDBTools/internal/types/types_backup"
	"sfDBTools/pkg/helper"
	"sfDBTools/pkg/profilehelper"
)

// executePreBackup melakukan backup database sebelum restore sebagai safety net
// Menggunakan existing backup infrastructure (backup.Service)
// Returns: backup file path dan error
func (s *Service) executePreBackup(ctx context.Context, targetDB string) (string, error) {
	s.Log.Infof("Membuat safety backup untuk database '%s'...", targetDB)

	timer := helper.NewTimer()

	// Build BackupDBOptions untuk backup service
	// Menggunakan struktur yang sama dengan backup command
	backupOpts := types_backup.BackupDBOptions{
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
		Entry: types_backup.BackupEntryConfig{
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

	// Setup connection untuk pre-backup menggunakan profilehelper
	// PENTING: Koneksi ini terpisah dari koneksi restore service
	sourceClient, err := profilehelper.ConnectWithProfile(s.TargetProfile, "mysql")
	if err != nil {
		return "", fmt.Errorf("gagal setup backup connection: %w", err)
	}
	// JANGAN close koneksi di sini karena backup service mungkin masih butuh koneksi lebih lama
	// Close akan dipanggil otomatis saat backup service selesai
	// defer sourceClient.Close() // REMOVED: menyebabkan invalid connection setelah backup

	// Filter database hanya untuk target database
	dbFiltered := []string{targetDB}

	// Execute backup dan dapatkan result
	result, err := backupSvc.ExecuteBackup(ctx, sourceClient, dbFiltered, "separated")

	// Close backup connection setelah backup selesai
	if closeErr := sourceClient.Close(); closeErr != nil {
		s.Log.Warnf("Gagal close backup connection: %v", closeErr)
	}

	if err != nil {
		return "", fmt.Errorf("pre-backup gagal: %w", err)
	}

	duration := timer.Elapsed()
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
