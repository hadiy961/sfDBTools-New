package backup

import (
	"context"
	"fmt"
	"sfDBTools/internal/backup/display"
	"sfDBTools/internal/types"
	"sfDBTools/internal/types/types_backup"
	"sfDBTools/pkg/backuphelper"
	pkghelper "sfDBTools/pkg/helper"
	"sfDBTools/pkg/input"
	"strings"
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
		candidates := backuphelper.FilterCandidatesByMode(dbFiltered, mode)

		// Filter berdasarkan client-code dan instance jika diperlukan
		if mode == "primary" && s.BackupDBOptions.ClientCode != "" {
			candidates = backuphelper.FilterCandidatesByClientCode(candidates, s.BackupDBOptions.ClientCode)
			if len(candidates) == 0 {
				return nil, "", nil, fmt.Errorf("tidak ada database primary dengan client code '%s' yang ditemukan", s.BackupDBOptions.ClientCode)
			}
			// Jika hanya ada satu kandidat, langsung pilih
			if len(candidates) == 1 {
				selectedDB = candidates[0]
			}
		} else if mode == "secondary" {
			// Untuk secondary, filter berdasarkan client-code dan/atau instance
			if s.BackupDBOptions.ClientCode != "" || s.BackupDBOptions.Instance != "" {
				candidates = backuphelper.FilterSecondaryByClientCodeAndInstance(
					candidates,
					s.BackupDBOptions.ClientCode,
					s.BackupDBOptions.Instance,
				)

				if len(candidates) == 0 {
					// Error message berbeda berdasarkan apa yang di-provide
					if s.BackupDBOptions.ClientCode != "" && s.BackupDBOptions.Instance != "" {
						return nil, "", nil, fmt.Errorf("tidak ada database secondary dengan client code '%s' dan instance '%s' yang ditemukan",
							s.BackupDBOptions.ClientCode, s.BackupDBOptions.Instance)
					} else if s.BackupDBOptions.ClientCode != "" {
						return nil, "", nil, fmt.Errorf("tidak ada database secondary dengan client code '%s' yang ditemukan", s.BackupDBOptions.ClientCode)
					} else {
						return nil, "", nil, fmt.Errorf("tidak ada database secondary dengan instance '%s' yang ditemukan", s.BackupDBOptions.Instance)
					}
				}

				// Jika instance juga di-provide dan hanya ada satu kandidat, langsung pilih
				if s.BackupDBOptions.Instance != "" && len(candidates) == 1 {
					selectedDB = candidates[0]
				} else if s.BackupDBOptions.Instance != "" && s.BackupDBOptions.ClientCode != "" && len(candidates) == 0 {
					// Instance di-provide tapi tidak ada yang match - tampilkan warning
					s.Log.Warnf("Instance '%s' tidak ditemukan untuk client code '%s', menampilkan semua secondary database untuk client tersebut",
						s.BackupDBOptions.Instance, s.BackupDBOptions.ClientCode)
					// Re-filter hanya dengan client code
					candidates = backuphelper.FilterSecondaryByClientCodeAndInstance(
						backuphelper.FilterCandidatesByMode(dbFiltered, mode),
						s.BackupDBOptions.ClientCode,
						"", // tanpa instance
					)
				}
			}
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
	if mode == "primary" || mode == "secondary" {
		// Consolidated loop for all suffixes
		for suffix, enabled := range map[string]bool{
			"_dmart":   s.BackupDBOptions.IncludeDmart,
			"_temp":    s.BackupDBOptions.IncludeTemp,
			"_archive": s.BackupDBOptions.IncludeArchive,
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
}, dbFiltered []string, compressionSettings types_backup.CompressionSettings) ([]string, error) {
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
	display.DisplayFilterStats(stats, s.Log)

	s.BackupDBOptions.DBName = selectedDB
	s.BackupDBOptions.CompanionStatus = companionStatus

	// Update filename untuk database yang dipilih
	previewFilename, err := pkghelper.GenerateBackupFilename(
		selectedDB,
		s.BackupDBOptions.Mode,
		s.BackupDBOptions.Profile.DBInfo.HostName,
		compressionSettings.Type,
		s.BackupDBOptions.Encryption.Enabled,
	)
	if err != nil {
		s.Log.Warn("gagal generate filename: " + err.Error())
		previewFilename = "error_generating_filename"
	}
	s.BackupDBOptions.File.Path = previewFilename

	// Return companionDbs sebagai dbFiltered yang baru
	// Mode single: [database_yang_dipilih]
	// Mode primary/secondary: [database_yang_dipilih, companion_dmart, companion_temp, companion_archive]
	return companionDbs, nil
}

// selectMultipleDatabases menampilkan multi-select menu untuk memilih database
func (s *Service) selectMultipleDatabases(databases []string) ([]string, error) {
	if len(databases) == 0 {
		return nil, fmt.Errorf("tidak ada database yang tersedia untuk dipilih")
	}

	s.Log.Info(fmt.Sprintf("Tersedia %d database non-system", len(databases)))
	s.Log.Info("Gunakan [Space] untuk memilih/membatalkan, [Enter] untuk konfirmasi")

	// Gunakan ShowMultiSelect dari input package
	indices, err := input.ShowMultiSelect("Pilih database untuk backup:", databases)
	if err != nil {
		return nil, fmt.Errorf("pemilihan database dibatalkan: %w", err)
	}

	if len(indices) == 0 {
		return nil, fmt.Errorf("tidak ada database yang dipilih")
	}

	// Convert indices to database names
	selectedDBs := make([]string, 0, len(indices))
	for _, idx := range indices {
		if idx > 0 && idx <= len(databases) {
			selectedDBs = append(selectedDBs, databases[idx-1])
		}
	}

	s.Log.Info(fmt.Sprintf("Dipilih %d database: %s", len(selectedDBs), strings.Join(selectedDBs, ", ")))

	return selectedDBs, nil
}
