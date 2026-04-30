package sync

import (
	"context"
	"database/sql"
	"sfdbtools/internal/shared/database"
)

// Service is a backward-compatible wrapper for SyncManager.
type Service struct {
	manager *SyncManager
}

func NewService(localDB *sql.DB, remoteDB *database.Client, clientID string) *Service {
	provider := NewSQLRemoteProvider(remoteDB)
	// For simplicity, we assume masterKey is handled outside or injected.
	// In sfDBTools, the master key is usually resolved from env or key file.
	manager := NewSyncManager(localDB, provider, clientID, "")
	return &Service{manager: manager}
}

func (s *Service) MigrateRemoteHub(ctx context.Context) error {
	return s.manager.remote.Migrate(ctx)
}

func (s *Service) SendHeartbeat(ctx context.Context, backupPath string) error {
	return s.manager.SendHeartbeat(ctx)
}

func (s *Service) RunSync(ctx context.Context, mode string) error {
	return s.manager.RunSync(ctx, mode)
}

// Additional legacy methods can be mapped to manager...
func (s *Service) PushSettings(ctx context.Context) error {
	return s.manager.PushAll(ctx)
}

func (s *Service) PullSettings(ctx context.Context) error {
	return s.manager.PullAll(ctx)
}

func (s *Service) PushProfiles(ctx context.Context) error {
	// Implementation in manager needed
	return nil
}

func (s *Service) PullProfiles(ctx context.Context) error {
	// Implementation in manager needed
	return nil
}

func (s *Service) PushJobs(ctx context.Context) error {
	// Implementation in manager needed
	return nil
}

func (s *Service) PullJobs(ctx context.Context) error {
	// Implementation in manager needed
	return nil
}
