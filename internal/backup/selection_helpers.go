package backup

import (
	"context"
	"fmt"
	"sfDBTools/internal/backup/helpers"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/consts"
	pkghelper "sfDBTools/pkg/helper"
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/ui"
)

// =============================================================================
// Database Selection Helpers
// =============================================================================

// selectDatabaseAndBuildList menangani database selection dan companion databases logic
// Untuk mode single: return hanya database yang dipilih
// Untuk mode primary/secondary: return database yang dipilih + companion (jika enabled)
func (s *Service) selectDatabaseAndBuildList(ctx context.Context, client interface {
	GetDatabaseList(context.Context) ([]string, error)
}, selectedDBName string, dbFiltered []string, mode string) ([]string, string, map[string]bool, error) {

	allDatabases, listErr := client.GetDatabaseList(ctx)
	if listErr != nil {
		return nil, "", nil, fmt.Errorf("gagal mengambil daftar database: %w", listErr)
	}

	selectedDB := selectedDBName
	if selectedDB == "" {
		candidates := helpers.FilterCandidatesByMode(dbFiltered, mode)

		filteredCandidates, autoSelectedDB, filterErr := s.filterCandidatesByModeAndOptions(mode, candidates)
		if filterErr != nil {
			return nil, "", nil, filterErr
		}
		candidates = filteredCandidates
		if selectedDB == "" {
			selectedDB = autoSelectedDB
		}

		if len(candidates) == 0 {
			return nil, "", nil, fmt.Errorf("tidak ada database yang tersedia untuk dipilih")
		}

		// Jika selectedDB masih kosong, tampilkan menu interaktif
		if selectedDB == "" {
			choice, choiceErr := input.ShowMenu("Pilih database yang akan di-backup:", candidates)
			if choiceErr != nil {
				return nil, "", nil, choiceErr
			}
			selectedDB = candidates[choice-1]
		}
	}

	if !pkghelper.StringSliceContainsFold(allDatabases, selectedDB) {
		return nil, "", nil, fmt.Errorf("database %s tidak ditemukan di server", selectedDB)
	}

	companionDbs := []string{selectedDB}
	companionStatus := map[string]bool{selectedDB: true}

	// Add companion databases - hanya untuk mode primary dan secondary, bukan untuk single
	if mode == consts.ModePrimary || mode == consts.ModeSecondary {
		// Consolidated loop for all suffixes
		for suffix, enabled := range map[string]bool{
			consts.SuffixDmart:   s.BackupDBOptions.IncludeDmart,
			consts.SuffixTemp:    s.BackupDBOptions.IncludeTemp,
			consts.SuffixArchive: s.BackupDBOptions.IncludeArchive,
		} {
			if !enabled {
				continue
			}

			dbName := selectedDB + suffix
			exists := pkghelper.StringSliceContainsFold(allDatabases, dbName)

			if exists {
				s.Log.Infof("Menambahkan database companion: %s", dbName)
				companionDbs = append(companionDbs, dbName)
			} else {
				s.Log.Warnf("Database %s tidak ditemukan, melewati", dbName)
			}
			companionStatus[dbName] = exists
		}
	}

	return companionDbs, selectedDB, companionStatus, nil
}

// handleSingleModeSetup handle setup untuk mode single/primary/secondary
// Returns companionDbs (database yang dipilih, + companion untuk primary/secondary) sebagai dbFiltered yang baru
func (s *Service) handleSingleModeSetup(ctx context.Context, client interface {
	GetDatabaseList(context.Context) ([]string, error)
}, dbFiltered []string) ([]string, error) {
	compressionSettings := s.buildCompressionSettings()
	// Get all databases untuk menghitung statistik yang akurat
	allDatabases, err := client.GetDatabaseList(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar database: %w", err)
	}

	companionDbs, selectedDB, companionStatus, selErr := s.selectDatabaseAndBuildList(
		ctx, client, s.BackupDBOptions.DBName, dbFiltered, s.BackupDBOptions.Mode)
	if selErr != nil {
		return nil, selErr
	}

	// Tampilkan statistik filtering setelah selection
	stats := &types.FilterStats{
		TotalFound:    len(allDatabases),
		TotalIncluded: len(companionDbs),
		TotalExcluded: len(allDatabases) - len(companionDbs),
	}
	ui.DisplayFilterStats(stats, consts.FeatureBackup, s.Log)

	s.BackupDBOptions.DBName = selectedDB
	s.BackupDBOptions.CompanionStatus = companionStatus

	// Update filename untuk database yang dipilih
	previewFilename, err := pkghelper.GenerateBackupFilename(
		selectedDB,
		s.BackupDBOptions.Mode,
		s.BackupDBOptions.Profile.DBInfo.HostName,
		compressionSettings.Type,
		s.BackupDBOptions.Encryption.Enabled,
		s.BackupDBOptions.Filter.ExcludeData,
	)
	if err != nil {
		s.Log.Warn("gagal generate filename: " + err.Error())
		previewFilename = consts.FilenameGenerateErrorPlaceholder
	}
	s.BackupDBOptions.File.Path = previewFilename

	// Return companionDbs sebagai dbFiltered yang baru
	// Mode single: [database_yang_dipilih]
	// Mode primary/secondary: [database_yang_dipilih, companion_dmart, companion_temp, companion_archive]
	return companionDbs, nil
}
