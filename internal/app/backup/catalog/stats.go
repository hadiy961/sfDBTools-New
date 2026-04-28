package catalog

import (
	"os"
	"time"

	"github.com/dustin/go-humanize"
)

// CatalogStats stores general statistics about the backup catalog
type CatalogStats struct {
	TotalEntries    int       `json:"total_entries"`
	TotalSizeBytes  int64     `json:"total_size_bytes"`
	TotalSizeHuman  string    `json:"total_size_human"`
	OldestBackup    time.Time `json:"oldest_backup"`
	NewestBackup    time.Time `json:"newest_backup"`
	UniqueDatabases int       `json:"unique_databases"`
	SuccessRate     float64   `json:"success_rate"`
	CatalogFile     string    `json:"catalog_file"`
	CatalogSize     int64     `json:"catalog_size"`
}

// GetStats calculates summary metrics from all catalog entries.
func (s *Service) GetStats() (*CatalogStats, error) {
	cat, err := s.repo.Load()
	if err != nil {
		return nil, err
	}

	stats := &CatalogStats{}
	if repo, ok := s.repo.(*JSONFileRepository); ok {
		stats.CatalogFile = repo.FilePath
		if fi, err := os.Stat(repo.FilePath); err == nil {
			stats.CatalogSize = fi.Size()
		}
	}

	stats.TotalEntries = len(cat.Entries)
	if stats.TotalEntries == 0 {
		return stats, nil
	}

	successCount := 0
	dbMap := make(map[string]bool)

	for i, e := range cat.Entries {
		stats.TotalSizeBytes += e.FileSizeBytes
		if e.BackupStatus == "success" {
			successCount++
		}
		for _, db := range e.DatabaseNames {
			dbMap[db] = true
		}

		if i == 0 {
			stats.OldestBackup = e.BackupTime
			stats.NewestBackup = e.BackupTime
		} else {
			if e.BackupTime.Before(stats.OldestBackup) {
				stats.OldestBackup = e.BackupTime
			}
			if e.BackupTime.After(stats.NewestBackup) {
				stats.NewestBackup = e.BackupTime
			}
		}
	}

	stats.TotalSizeHuman = humanize.Bytes(uint64(stats.TotalSizeBytes))
	stats.UniqueDatabases = len(dbMap)
	stats.SuccessRate = float64(successCount) / float64(stats.TotalEntries) * 100

	return stats, nil
}
