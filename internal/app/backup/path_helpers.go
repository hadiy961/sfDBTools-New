package backup

import (
	"context"
	"path/filepath"
	"sfdbtools/internal/app/backup/helpers/compression"
	backuppath "sfdbtools/internal/app/backup/helpers/path"
	"sfdbtools/internal/app/backup/modes"
	"sfdbtools/internal/domain"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/database"
	"sfdbtools/internal/ui/print"
)

// =============================================================================
// Path Generation Helpers
// =============================================================================

// GenerateFullBackupPath membuat full path untuk backup file
func (s *Service) GenerateFullBackupPath(dbName string, mode string) (string, error) {
	compressionSettings := compression.BuildCompressionSettings(s.BackupDBOptions)

	// Konsisten: selalu gunakan hostname (DBInfo.HostName) untuk semua mode.
	// Fallback ke DBInfo.Host hanya jika HostName kosong.
	hostIdentifier := s.BackupDBOptions.Profile.DBInfo.HostName
	if hostIdentifier == "" {
		hostIdentifier = s.BackupDBOptions.Profile.DBInfo.Host
	}

	filename, err := backuppath.GenerateBackupFilename(
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

// GenerateBackupPaths generates output directory and filename for backup
// Returns updated dbFiltered for single/primary/secondary mode (selected database + companions)
// state parameter provides access to ExcludedDatabases for statistics display
func (s *Service) GenerateBackupPaths(ctx context.Context, state *BackupExecutionState, client *database.Client, dbFiltered []string) ([]string, error) {
	dbHostname := s.BackupDBOptions.Profile.DBInfo.HostName
	if dbHostname == "" {
		dbHostname = s.BackupDBOptions.Profile.DBInfo.Host
	}
	compressionSettings := compression.BuildCompressionSettings(s.BackupDBOptions)

	// Generate output directory
	var err error
	s.BackupDBOptions.OutputDir, err = backuppath.GenerateBackupDirectory(
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
	if s.BackupDBOptions.Mode == consts.ModeSeparated ||
		modes.IsSingleModeVariant(s.BackupDBOptions.Mode) {
		exampleDBName = "database_name"
	} else if s.BackupDBOptions.Mode == consts.ModeCombined || s.BackupDBOptions.Mode == consts.ModeAll {
		// Untuk combined/all, gunakan jumlah database yang akan di-backup
		dbCount = len(dbFiltered)
		// exampleDBName dibiarkan kosong, akan di-generate oleh GenerateBackupFilenameWithCount
		// dengan prefix sesuai mode ('all' atau 'combined')
	}

	s.BackupDBOptions.File.Path, err = backuppath.GenerateBackupFilenameWithCount(
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
		s.BackupDBOptions.File.Path = consts.FilenameGenerateErrorPlaceholder
	}

	// Handle single/primary/secondary mode dengan database selection
	if modes.IsSingleModeVariant(s.BackupDBOptions.Mode) {
		return s.handleSingleModeSetup(ctx, client, dbFiltered)
	}

	// Untuk mode non-single (all, filter, combined), tampilkan statistik di sini
	allDatabases, err := client.GetDatabaseList(ctx)
	if err != nil {
		s.Log.Warnf("gagal mengambil daftar database untuk statistik: %v", err)
	} else {
		stats := &domain.FilterStats{
			TotalFound:        len(allDatabases),
			TotalIncluded:     len(dbFiltered),
			TotalExcluded:     len(allDatabases) - len(dbFiltered),
			ExcludedDatabases: state.ExcludedDatabases,
		}
		print.PrintFilterStats(stats, consts.FeatureBackup, s.Log)
	}

	// Jika user set custom filename untuk mode all/combined, treat sebagai base name tanpa ekstensi.
	// Ekstensi mengikuti default generated filename (mis. .sql.gz.enc / .sql / .sql.enc).
	if (s.BackupDBOptions.Mode == consts.ModeAll || s.BackupDBOptions.Mode == consts.ModeCombined) && s.BackupDBOptions.File.Filename != "" {
		s.BackupDBOptions.File.Path = backuppath.ApplyCustomBaseFilename(s.BackupDBOptions.File.Path, s.BackupDBOptions.File.Filename)
	}

	return dbFiltered, nil
}
