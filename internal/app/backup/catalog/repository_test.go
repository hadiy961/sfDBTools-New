package catalog

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestJSONFileRepository(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "catalog.json")

	repo := NewJSONFileRepository(filePath)

	// 1. Test Load empty
	cat, err := repo.Load()
	assert.NoError(t, err)
	assert.NotNil(t, cat)
	assert.Equal(t, "1.0", cat.Version)
	assert.Empty(t, cat.Entries)

	// 2. Test Save
	now := time.Now()
	entry := CatalogEntry{
		ID:            "123",
		BackupFile:    "/tmp/test.sql",
		DatabaseNames: []string{"db1"},
		BackupStatus:  "success",
		BackupTime:    now,
	}
	cat.Entries = append(cat.Entries, entry)
	err = repo.Save(cat)
	assert.NoError(t, err)

	// 3. Test Load existing
	cat2, err := repo.Load()
	assert.NoError(t, err)
	assert.Len(t, cat2.Entries, 1)
	assert.Equal(t, "123", cat2.Entries[0].ID)
	assert.Equal(t, "success", cat2.Entries[0].BackupStatus)
}
