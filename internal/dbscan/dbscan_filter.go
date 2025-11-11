package dbscan

import (
	"context"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/database"
)

// GetFilteredDatabases mengambil dan memfilter daftar database sesuai aturan.
// Menggunakan general database filtering system dari pkg/database
func (s *Service) GetFilteredDatabases(ctx context.Context, client *database.Client) ([]string, *types.DatabaseFilterStats, error) {
	// Gunakan helper untuk menghindari duplikasi logic
	return database.FilterFromScanOptions(ctx, client, &s.ScanOptions)
}
