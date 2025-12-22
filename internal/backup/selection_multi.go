package backup

import (
	"context"
	"fmt"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/ui"
	"strings"
)

// getFilteredDatabasesWithMultiSelect menampilkan multi-select untuk memilih database
func (s *Service) getFilteredDatabasesWithMultiSelect(ctx context.Context, client *database.Client) ([]string, *types.FilterStats, error) {
	// Get all databases from server
	allDatabases, err := client.GetDatabaseList(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("gagal mengambil daftar database: %w", err)
	}

	stats := &types.FilterStats{
		TotalFound:    len(allDatabases),
		TotalIncluded: 0,
		TotalExcluded: 0,
	}

	if len(allDatabases) == 0 {
		return nil, stats, fmt.Errorf("tidak ada database yang ditemukan di server")
	}

	// Filter system databases untuk pilihan
	nonSystemDBs := make([]string, 0, len(allDatabases))

	for _, db := range allDatabases {
		if _, isSystem := types.SystemDatabases[strings.ToLower(db)]; !isSystem {
			nonSystemDBs = append(nonSystemDBs, db)
		}
	}

	if len(nonSystemDBs) == 0 {
		return nil, stats, fmt.Errorf("tidak ada database non-system yang tersedia untuk dipilih")
	}

	// Tampilkan multi-select
	ui.PrintSubHeader("Pilih Database untuk Backup")
	selectedDBs, err := s.selectMultipleDatabases(nonSystemDBs)
	if err != nil {
		return nil, stats, err
	}

	if len(selectedDBs) == 0 {
		return nil, stats, fmt.Errorf("tidak ada database yang dipilih")
	}

	stats.TotalIncluded = len(selectedDBs)
	stats.TotalExcluded = len(allDatabases) - len(selectedDBs)

	return selectedDBs, stats, nil
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
