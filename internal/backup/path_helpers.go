package backup

import (
	"context"
	"path/filepath"
	"sfDBTools/internal/backup/display"
	"sfDBTools/internal/types"
	"sfDBTools/internal/types/types_backup"
	"sfDBTools/pkg/backuphelper"
	"sfDBTools/pkg/compress"
	"sfDBTools/pkg/database"
	pkghelper "sfDBTools/pkg/helper"
)

// =============================================================================
// Path Generation Helpers
// =============================================================================

// GenerateFullBackupPath membuat full path untuk backup file
func (s *Service) GenerateFullBackupPath(dbName string, mode string) (string, error) {
	// Build compression settings inline
	compressionType := s.BackupDBOptions.Compression.Type
	if !s.BackupDBOptions.Compression.Enabled {
		compressionType = ""
	}
	compressionSettings := types_backup.CompressionSettings{
		Type:    compress.CompressionType(compressionType),
		Enabled: s.BackupDBOptions.Compression.Enabled,
		Level:   s.BackupDBOptions.Compression.Level,
	}

	// Untuk mode separated, gunakan IP address instead of hostname
	hostIdentifier := s.BackupDBOptions.Profile.DBInfo.HostName
	if mode == "separated" || mode == "separate" {
		hostIdentifier = s.BackupDBOptions.Profile.DBInfo.Host
	}

	filename, err := pkghelper.GenerateBackupFilename(
		dbName,
		mode,
		hostIdentifier,
		compressionSettings.Type,
		s.BackupDBOptions.Encryption.Enabled,
		s.BackupDBOptions.Filter.ExcludeData,
	)
	if err != nil {
		return "", err
	}

	return filepath.Join(s.BackupDBOptions.OutputDir, filename), nil
}

// generateBackupPaths generate output directory dan filename untuk backup
// Returns updated dbFiltered untuk mode single/primary/secondary (database yang dipilih + companion)
func (s *Service) generateBackupPaths(ctx context.Context, client *database.Client, dbFiltered []string) ([]string, error) {
	dbHostname := s.BackupDBOptions.Profile.DBInfo.Host

	// Build compression settings inline
	compressionType := s.BackupDBOptions.Compression.Type
	if !s.BackupDBOptions.Compression.Enabled {
		compressionType = ""
	}
	compressionSettings := types_backup.CompressionSettings{
		Type:    compress.CompressionType(compressionType),
		Enabled: s.BackupDBOptions.Compression.Enabled,
		Level:   s.BackupDBOptions.Compression.Level,
	}

	// Generate output directory
	var err error
	s.BackupDBOptions.OutputDir, err = pkghelper.GenerateBackupDirectory(
		s.Config.Backup.Output.BaseDirectory,
		s.Config.Backup.Output.Structure.Pattern,
		dbHostname,
	)
	if err != nil {
		s.Log.Warn("gagal generate output directory: " + err.Error())
	}

	// Generate filename berdasarkan mode
	exampleDBName := ""
	dbCount := 0
	if s.BackupDBOptions.Mode == "separated" || s.BackupDBOptions.Mode == "separate" ||
		backuphelper.IsSingleModeVariant(s.BackupDBOptions.Mode) {
		exampleDBName = "database_name"
	} else if s.BackupDBOptions.Mode == "combined" || s.BackupDBOptions.Mode == "all" {
		// Untuk combined/all, gunakan jumlah database yang akan di-backup
		dbCount = len(dbFiltered)
		// exampleDBName dibiarkan kosong, akan di-generate oleh GenerateBackupFilenameWithCount
		// dengan prefix sesuai mode ('all' atau 'combined')
	}

	s.BackupDBOptions.File.Path, err = pkghelper.GenerateBackupFilenameWithCount(
		exampleDBName,
		s.BackupDBOptions.Mode,
		dbHostname,
		compressionSettings.Type,
		s.BackupDBOptions.Encryption.Enabled,
		dbCount,
		s.BackupDBOptions.Filter.ExcludeData,
	)
	if err != nil {
		s.Log.Warn("gagal generate filename preview: " + err.Error())
		s.BackupDBOptions.File.Path = "error_generating_filename"
	}

	// Handle single/primary/secondary mode dengan database selection
	if backuphelper.IsSingleModeVariant(s.BackupDBOptions.Mode) {
		return s.handleSingleModeSetup(ctx, client, dbFiltered, compressionSettings)
	}

	// Untuk mode non-single (all, filter, combined), tampilkan statistik di sini
	allDatabases, err := client.GetDatabaseList(ctx)
	if err != nil {
		s.Log.Warnf("gagal mengambil daftar database untuk statistik: %v", err)
	} else {
		stats := &types.FilterStats{
			TotalFound:        len(allDatabases),
			TotalIncluded:     len(dbFiltered),
			TotalExcluded:     len(allDatabases) - len(dbFiltered),
			ExcludedDatabases: s.excludedDatabases,
		}
		display.DisplayFilterStats(stats, s.Log)
	}

	return dbFiltered, nil
}
