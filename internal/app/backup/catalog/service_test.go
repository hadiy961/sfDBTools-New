package catalog

import (
	"sfdbtools/internal/app/backup/model/types_backup"
	applog "sfdbtools/internal/services/log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// MockRepository is an in-memory repository for testing business logic
type MockRepository struct {
	catalog *Catalog
}

func (m *MockRepository) Load() (*Catalog, error) {
	if m.catalog == nil {
		m.catalog = &Catalog{Version: "1.0"}
	}
	return m.catalog, nil
}

func (m *MockRepository) Save(cat *Catalog) error {
	m.catalog = cat
	return nil
}

func TestServiceRegisterBackup(t *testing.T) {
	repo := &MockRepository{}
	svc := NewService(repo, applog.NullLogger())

	meta := &types_backup.BackupMetadata{
		BackupFile:    "/tmp/test.sql",
		DatabaseNames: []string{"testdb"},
		Hostname:      "localhost",
		BackupStatus:  "success",
		FileSize:      1024,
	}

	err := svc.RegisterBackup(meta, "single", "local")
	assert.NoError(t, err)

	cat, _ := repo.Load()
	assert.Len(t, cat.Entries, 1)
	assert.Equal(t, "/tmp/test.sql", cat.Entries[0].BackupFile)
	assert.Equal(t, "testdb", cat.Entries[0].DatabaseNames[0])
}

func TestServiceQuery(t *testing.T) {
	repo := &MockRepository{}
	svc := NewService(repo, applog.NullLogger())

	now := time.Now()
	// Add dummy data
	repo.catalog = &Catalog{
		Version: "1.0",
		Entries: []CatalogEntry{
			{DatabaseNames: []string{"db1"}, BackupStatus: "success", BackupTime: now},
			{DatabaseNames: []string{"db2"}, BackupStatus: "failed", BackupTime: now.Add(-10 * time.Minute)},
			{DatabaseNames: []string{"db1_test"}, BackupStatus: "success", BackupTime: now.Add(-2 * time.Hour)},
		},
	}

	// 1. Test Filter by Database (Substring)
	res1, _ := svc.Query(QueryOptions{Database: "db1"})
	assert.Len(t, res1, 2)

	// 2. Test Filter by Status
	res2, _ := svc.Query(QueryOptions{Status: "failed"})
	assert.Len(t, res2, 1)

	// 3. Test Filter by Since
	res3, _ := svc.Query(QueryOptions{Since: "1h"})
	assert.Len(t, res3, 2)
}
