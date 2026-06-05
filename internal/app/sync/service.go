package sync

import (
	"context"
	"database/sql"
	"fmt"
	"sfdbtools/internal/shared/connectivity"
	"sfdbtools/internal/shared/database"
	"sfdbtools/internal/ui/progress"
	"time"

	"github.com/fatih/color"
)

// Service is a backward-compatible wrapper for SyncManager.
type Service struct {
	manager *SyncManager
}

func NewService(localDB *sql.DB, remoteDB *database.Client, clientID string) *Service {
	provider := NewSQLRemoteProvider(remoteDB)
	
	// Resolve sync_key from settings for E2E encryption
	var syncKey string
	_ = localDB.QueryRow("SELECT value FROM app_settings WHERE key = 'sync_key'").Scan(&syncKey)
	
	manager := NewSyncManager(localDB, provider, clientID, syncKey)
	return &Service{manager: manager}
}

func (s *Service) RunSync(ctx context.Context, mode string) error {
	return s.manager.RunSync(ctx, mode)
}

// PerformFullSync menjalankan alur sinkronisasi lengkap dengan feedback UI yang konsisten.
func PerformFullSync(ctx context.Context, clientCode string, isManual bool) error {
	sp := progress.NewSpinner("Sinkronisasi data")
	if isManual {
		fmt.Println(color.CyanString("\nMemulai sinkronisasi terpadu (Manual Sync)..."))
	}
	sp.Start()
	defer sp.Stop()

	// 1. Connectivity Check
	sp.Update("Memeriksa koneksi internet")
	if !connectivity.IsInternetAvailable(ctx, 3*time.Second) {
		return fmt.Errorf("tidak ada koneksi internet")
	}

	// 2. Remote Hub Connection
	sp.Update("Menghubungkan ke Remote Hub")
	remote, err := database.ConnectToRemoteHub()
	if err != nil {
		return fmt.Errorf("gagal terhubung ke remote hub: %w", err)
	}
	defer remote.Close()

	local, err := database.GetSQLite()
	if err != nil {
		return fmt.Errorf("gagal mengakses database lokal: %w", err)
	}

	// 3. Execution
	sp.Update("Sinkronisasi data (Settings, Profiles, Jobs, Heartbeat)")
	syncSvc := NewService(local, remote, clientCode)
	
	mode := database.GetSetting("sync_mode")
	if err := syncSvc.RunSync(ctx, mode); err != nil {
		return fmt.Errorf("sinkronisasi gagal: %w", err)
	}

	sp.Stop()
	if isManual {
		fmt.Println(color.GreenString("[SUCCESS] Sinkronisasi selesai dengan sukses!"))
	}
	return nil
}
