package catalog

import (
	"time"

	"github.com/dustin/go-humanize"
)

// GenerateReport creates a summary report based on the catalog.
func (s *Service) GenerateReport(period string) (*ReportSummary, error) {
	cat, err := s.repo.Load()
	if err != nil {
		return nil, err
	}

	report := &ReportSummary{
		Period:           period,
		DatabaseCoverage: []DatabaseCoverage{},
		StorageTrend:     []StorageTrendPoint{},
	}

	var cutoffTime time.Time
	now := time.Now()

	switch period {
	case "daily":
		cutoffTime = now.Add(-24 * time.Hour)
	case "weekly":
		cutoffTime = now.Add(-7 * 24 * time.Hour)
	case "monthly":
		cutoffTime = now.Add(-30 * 24 * time.Hour)
	default:
		// all time
	}

	dbMap := make(map[string]*DatabaseCoverage)

	for _, e := range cat.Entries {
		if !cutoffTime.IsZero() && e.BackupTime.Before(cutoffTime) {
			continue
		}

		report.TotalBackups++
		report.TotalSizeBytes += e.FileSizeBytes

		if e.BackupStatus == "success" {
			report.SuccessCount++
		} else {
			report.FailedCount++
		}

		for _, dbName := range e.DatabaseNames {
			if _, exists := dbMap[dbName]; !exists {
				dbMap[dbName] = &DatabaseCoverage{
					DatabaseName: dbName,
				}
			}
			dbCov := dbMap[dbName]
			dbCov.BackupCount++
			dbCov.TotalSize += e.FileSizeBytes
			if e.BackupTime.After(dbCov.LastBackup) {
				dbCov.LastBackup = e.BackupTime
			}
		}
	}

	report.TotalSizeHuman = humanize.Bytes(uint64(report.TotalSizeBytes))
	if report.TotalBackups > 0 {
		report.SuccessRate = float64(report.SuccessCount) / float64(report.TotalBackups) * 100
	}

	for _, cov := range dbMap {
		report.DatabaseCoverage = append(report.DatabaseCoverage, *cov)
	}

	// Calculate a simple storage trend (mocked for now based on available entries)
	trendMap := make(map[string]int64)
	for _, e := range cat.Entries {
		dateStr := e.BackupTime.Format("2006-01-02")
		trendMap[dateStr] += e.FileSizeBytes
	}

	for date, size := range trendMap {
		report.StorageTrend = append(report.StorageTrend, StorageTrendPoint{
			Date:      date,
			SizeBytes: size,
			SizeHuman: humanize.Bytes(uint64(size)),
		})
	}

	return report, nil
}
