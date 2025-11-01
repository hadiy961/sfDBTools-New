package dbscan

import (
	"context"
	"sfDBTools/pkg/database"
)

// setupScanConnections melakukan setup koneksi source dan target database
// Returns: sourceClient, targetClient, dbFiltered, cleanupFunc, error
func (s *Service) setupScanConnections(ctx context.Context, headerTitle string, showOptions bool) (*database.Client, *database.Client, []string, func(), error) {
	// Jika mode rescan, gunakan PrepareRescanSession
	if s.ScanOptions.Mode == "rescan" {
		sourceClient, targetClient, dbFiltered, err := s.PrepareRescanSession(ctx, headerTitle, showOptions)
		if err != nil {
			return nil, nil, nil, nil, err
		}

		// Force enable SaveToDB untuk rescan karena kita perlu update error_message
		s.ScanOptions.SaveToDB = true

		// Cleanup function untuk close semua connections
		cleanup := func() {
			if sourceClient != nil {
				sourceClient.Close()
			}
			if targetClient != nil {
				targetClient.Close()
			}
		}

		return sourceClient, targetClient, dbFiltered, cleanup, nil
	}

	// Setup session (koneksi database source) untuk mode normal
	sourceClient, dbFiltered, err := s.PrepareScanSession(ctx, headerTitle, showOptions)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	// Koneksi ke target database untuk menyimpan hasil scan
	var targetClient *database.Client
	if s.ScanOptions.SaveToDB {
		targetClient, err = s.ConnectToTargetDB(ctx)
		if err != nil {
			s.Logger.Warn("Gagal koneksi ke target database, hasil scan tidak akan disimpan: " + err.Error())
			s.ScanOptions.SaveToDB = false
		}
	}

	// Cleanup function untuk close semua connections
	cleanup := func() {
		if sourceClient != nil {
			sourceClient.Close()
		}
		if targetClient != nil {
			targetClient.Close()
		}
	}

	return sourceClient, targetClient, dbFiltered, cleanup, nil
}
