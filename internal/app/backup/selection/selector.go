package selection

import (
	"context"
	"fmt"
	"strings"

	backuppath "sfdbtools/internal/app/backup/helpers/path"
	"sfdbtools/internal/app/backup/model/types_backup"
	"sfdbtools/internal/domain"
	applog "sfdbtools/internal/services/log"
	"sfdbtools/internal/shared/listx"
	"sfdbtools/internal/ui/print"
	"sfdbtools/internal/ui/prompt"
	"sfdbtools/pkg/backuphelper"
	"sfdbtools/pkg/consts"
)

// DatabaseLister is the minimal interface needed to list databases.
// *database.Client satisfies this.
type DatabaseLister interface {
	GetDatabaseList(context.Context) ([]string, error)
}

type Selector struct {
	Log     applog.Logger
	Options *types_backup.BackupDBOptions
}

func New(log applog.Logger, opts *types_backup.BackupDBOptions) *Selector {
	return &Selector{Log: log, Options: opts}
}

// buildCompressionSettings delegates ke shared helper untuk avoid duplication
func (s *Selector) buildCompressionSettings() types_backup.CompressionSettings {
	return backuphelper.BuildCompressionSettings(s.Options)
}

// GetFilteredDatabasesWithMultiSelect shows interactive multi-select for databases.
func (s *Selector) GetFilteredDatabasesWithMultiSelect(ctx context.Context, client DatabaseLister) ([]string, *domain.FilterStats, error) {
	allDatabases, err := client.GetDatabaseList(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("gagal mengambil daftar database: %w", err)
	}

	stats := &domain.FilterStats{
		TotalFound:    len(allDatabases),
		TotalIncluded: 0,
		TotalExcluded: 0,
	}

	if len(allDatabases) == 0 {
		return nil, stats, fmt.Errorf("tidak ada database yang ditemukan di server")
	}

	nonSystemDBs := make([]string, 0, len(allDatabases))
	for _, db := range allDatabases {
		dbLower := strings.ToLower(db)
		if _, isSystem := domain.SystemDatabases[dbLower]; isSystem {
			continue
		}
		// Backup tidak lagi mendukung database *_temp dan *_archive.
		if strings.HasSuffix(dbLower, consts.SuffixTemp) || strings.HasSuffix(dbLower, consts.SuffixArchive) {
			continue
		}
		nonSystemDBs = append(nonSystemDBs, db)
	}

	if len(nonSystemDBs) == 0 {
		return nil, stats, fmt.Errorf("tidak ada database non-system yang tersedia untuk dipilih")
	}

	print.PrintSubHeader("Pilih Database untuk Backup")
	selectedDBs, err := s.selectMultipleDatabases(nonSystemDBs)
	if err != nil {
		return nil, stats, err
	}

	if len(selectedDBs) == 0 {
		return nil, stats, fmt.Errorf("tidak ada database yang dipilih")
	}

	stats.TotalIncluded = len(selectedDBs)
	stats.TotalExcluded = len(allDatabases) - len(selectedDBs)

	// Persist pilihan sebagai include list agar flow tidak meminta multi-select lagi
	// ketika user mengubah opsi (mis: filename/encryption/compression) dan session loop re-run.
	if s.Options != nil {
		s.Options.Filter.IncludeDatabases = selectedDBs
		s.Options.Filter.IncludeFile = ""
	}

	return selectedDBs, stats, nil
}

func (s *Selector) selectMultipleDatabases(databases []string) ([]string, error) {
	if len(databases) == 0 {
		return nil, fmt.Errorf("tidak ada database yang tersedia untuk dipilih")
	}

	s.Log.Info(fmt.Sprintf("Tersedia %d database non-system", len(databases)))
	s.Log.Info("Gunakan [Space] untuk memilih/membatalkan, [Enter] untuk konfirmasi")

	selectedDBs, _, err := prompt.SelectMany("Pilih database untuk backup:", databases, nil)
	if err != nil {
		return nil, fmt.Errorf("pemilihan database dibatalkan: %w", err)
	}

	if len(selectedDBs) == 0 {
		return nil, fmt.Errorf("tidak ada database yang dipilih")
	}

	s.Log.Info(fmt.Sprintf("Dipilih %d database: %s", len(selectedDBs), strings.Join(selectedDBs, ", ")))
	return selectedDBs, nil
}

func (s *Selector) filterCandidatesByModeAndOptions(mode string, candidates []string) ([]string, string, error) {
	switch mode {
	case consts.ModePrimary:
		return s.filterPrimaryCandidates(candidates)
	case consts.ModeSecondary:
		return s.filterSecondaryCandidates(candidates)
	default:
		return candidates, "", nil
	}
}

func (s *Selector) filterPrimaryCandidates(candidates []string) ([]string, string, error) {
	if s.Options.ClientCode == "" {
		return candidates, "", nil
	}

	filtered := FilterCandidatesByClientCode(candidates, s.Options.ClientCode)
	if len(filtered) == 0 {
		return nil, "", fmt.Errorf("tidak ada database primary dengan client code '%s' yang ditemukan", s.Options.ClientCode)
	}

	if len(filtered) == 1 {
		return filtered, filtered[0], nil
	}

	return filtered, "", nil
}

func (s *Selector) filterSecondaryCandidates(candidates []string) ([]string, string, error) {
	if s.Options.ClientCode == "" && s.Options.Instance == "" {
		return candidates, "", nil
	}

	filtered := FilterSecondaryByClientCodeAndInstance(
		candidates,
		s.Options.ClientCode,
		s.Options.Instance,
	)

	if len(filtered) == 0 {
		if s.Options.ClientCode != "" && s.Options.Instance != "" {
			return nil, "", fmt.Errorf(
				"tidak ada database secondary dengan client code '%s' dan instance '%s' yang ditemukan",
				s.Options.ClientCode,
				s.Options.Instance,
			)
		}
		if s.Options.ClientCode != "" {
			return nil, "", fmt.Errorf("tidak ada database secondary dengan client code '%s' yang ditemukan", s.Options.ClientCode)
		}
		return nil, "", fmt.Errorf("tidak ada database secondary dengan instance '%s' yang ditemukan", s.Options.Instance)
	}

	if s.Options.Instance != "" && len(filtered) == 1 {
		return filtered, filtered[0], nil
	}

	return filtered, "", nil
}

// SelectDatabaseAndBuildList selects a database (interactive when needed) and expands companion DBs.
func (s *Selector) SelectDatabaseAndBuildList(ctx context.Context, client DatabaseLister, selectedDBName string, dbFiltered []string, mode string) ([]string, string, map[string]bool, error) {
	allDatabases, listErr := client.GetDatabaseList(ctx)
	if listErr != nil {
		return nil, "", nil, fmt.Errorf("gagal mengambil daftar database: %w", listErr)
	}

	selectedDB := selectedDBName
	if selectedDB == "" {
		candidates := FilterCandidatesByMode(dbFiltered, mode)

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

		if selectedDB == "" {
			_, idx, choiceErr := prompt.SelectOne("Pilih database yang akan di-backup:", candidates, -1)
			if choiceErr != nil {
				return nil, "", nil, choiceErr
			}
			if idx < 0 || idx >= len(candidates) {
				return nil, "", nil, fmt.Errorf("pemilihan database dibatalkan")
			}
			selectedDB = candidates[idx]
		}
	}

	// Backup tidak mendukung database *_temp dan *_archive.
	selectedLower := strings.ToLower(selectedDB)
	if strings.HasSuffix(selectedLower, consts.SuffixTemp) || strings.HasSuffix(selectedLower, consts.SuffixArchive) {
		return nil, "", nil, fmt.Errorf("backup tidak mendukung database dengan suffix '%s' atau '%s'", consts.SuffixTemp, consts.SuffixArchive)
	}

	if !listx.StringSliceContainsFold(allDatabases, selectedDB) {
		return nil, "", nil, fmt.Errorf("database %s tidak ditemukan di server", selectedDB)
	}

	companionDbs := []string{selectedDB}
	companionStatus := map[string]bool{selectedDB: true}

	if mode == consts.ModePrimary || mode == consts.ModeSecondary {
		for suffix, enabled := range map[string]bool{
			consts.SuffixDmart: s.Options.IncludeDmart,
		} {
			if !enabled {
				continue
			}

			dbName := selectedDB + suffix
			exists := listx.StringSliceContainsFold(allDatabases, dbName)
			s.Log.Infof("Memeriksa keberadaan database companion: %s ...", dbName)
			if exists {
				s.Log.Infof("Database %s ditemukan, menambahkan sebagai database companion", dbName)
				companionDbs = append(companionDbs, dbName)
			} else {
				s.Log.Warnf("Database %s tidak ditemukan, melewati", dbName)
			}
			companionStatus[dbName] = exists
		}
	}

	return companionDbs, selectedDB, companionStatus, nil
}

// HandleSingleModeSetup updates options based on selected DB and companions for single/primary/secondary modes.
func (s *Selector) HandleSingleModeSetup(ctx context.Context, client DatabaseLister, dbFiltered []string) ([]string, error) {
	compressionSettings := s.buildCompressionSettings()

	allDatabases, err := client.GetDatabaseList(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar database: %w", err)
	}

	companionDbs, selectedDB, companionStatus, selErr := s.SelectDatabaseAndBuildList(
		ctx,
		client,
		s.Options.DBName,
		dbFiltered,
		s.Options.Mode,
	)
	if selErr != nil {
		return nil, selErr
	}

	stats := &domain.FilterStats{
		TotalFound:    len(allDatabases),
		TotalIncluded: len(companionDbs),
		TotalExcluded: len(allDatabases) - len(companionDbs),
	}
	print.PrintFilterStats(stats, consts.FeatureBackup, s.Log)

	s.Options.DBName = selectedDB
	s.Options.CompanionStatus = companionStatus

	previewFilename, err := backuppath.GenerateBackupFilename(
		selectedDB,
		s.Options.Mode,
		s.Options.Profile.DBInfo.HostName,
		compressionSettings.Type,
		s.Options.Encryption.Enabled,
		s.Options.Filter.ExcludeData,
	)
	if err != nil {
		s.Log.Warn("gagal generate filename: " + err.Error())
		previewFilename = consts.FilenameGenerateErrorPlaceholder
	}
	s.Options.File.Path = previewFilename

	return companionDbs, nil
}
