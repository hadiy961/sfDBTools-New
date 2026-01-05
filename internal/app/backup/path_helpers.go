package backup

import (
	"context"
	"path/filepath"
	"sfDBTools/internal/app/backup/modes"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/database"
	pkghelper "sfDBTools/pkg/helper"
	"sfDBTools/pkg/ui"
	"strings"
)

// =============================================================================
// Path Generation Helpers
// =============================================================================

// GenerateFullBackupPath membuat full path untuk backup file
func (s *Service) GenerateFullBackupPath(dbName string, mode string) (string, error) {
	compressionSettings := s.buildCompressionSettings()

	// Konsisten: selalu gunakan hostname (DBInfo.HostName) untuk semua mode.
	// Fallback ke DBInfo.Host hanya jika HostName kosong.
	hostIdentifier := s.BackupDBOptions.Profile.DBInfo.HostName
	if hostIdentifier == "" {
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

// GenerateBackupPaths generates output directory and filename for backup
// Returns updated dbFiltered for single/primary/secondary mode (selected database + companions)
func (s *Service) GenerateBackupPaths(ctx context.Context, client *database.Client, dbFiltered []string) ([]string, error) {
	dbHostname := s.BackupDBOptions.Profile.DBInfo.HostName
	if dbHostname == "" {
		dbHostname = s.BackupDBOptions.Profile.DBInfo.Host
	}
	compressionSettings := s.buildCompressionSettings()

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
	if s.BackupDBOptions.Mode == consts.ModeSeparated ||
		modes.IsSingleModeVariant(s.BackupDBOptions.Mode) {
		exampleDBName = "database_name"
	} else if s.BackupDBOptions.Mode == consts.ModeCombined || s.BackupDBOptions.Mode == consts.ModeAll {
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
		stats := &types.FilterStats{
			TotalFound:        len(allDatabases),
			TotalIncluded:     len(dbFiltered),
			TotalExcluded:     len(allDatabases) - len(dbFiltered),
			ExcludedDatabases: s.excludedDatabases,
		}
		ui.DisplayFilterStats(stats, consts.FeatureBackup, s.Log)
	}

	// Jika user set custom filename untuk mode all/combined, treat sebagai base name tanpa ekstensi.
	// Ekstensi mengikuti default generated filename (mis. .sql.gz.enc / .sql / .sql.enc).
	if (s.BackupDBOptions.Mode == consts.ModeAll || s.BackupDBOptions.Mode == consts.ModeCombined) && s.BackupDBOptions.File.Filename != "" {
		defaultName := s.BackupDBOptions.File.Path
		customBase := s.BackupDBOptions.File.Filename

		// Jika user sudah memasukkan .sql (atau full filename), biarkan apa adanya.
		if !strings.Contains(customBase, ".sql") {
			ext := ""
			if defaultName != "" && defaultName != consts.FilenameGenerateErrorPlaceholder {
				if idx := strings.Index(defaultName, ".sql"); idx >= 0 {
					ext = defaultName[idx:]
				} else {
					ext = filepath.Ext(defaultName)
				}
			}
			s.BackupDBOptions.File.Path = customBase + ext
		} else {
			s.BackupDBOptions.File.Path = customBase
		}
	}

	return dbFiltered, nil
}
