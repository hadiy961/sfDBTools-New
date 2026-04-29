package catalog

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"sfdbtools/internal/shared/fsops"
)

// Repository is the interface for catalog storage operations.
type Repository interface {
	Load() (*Catalog, error)
	Save(catalog *Catalog) error
}

// JSONFileRepository is a file-based implementation of Repository.
type JSONFileRepository struct {
	FilePath string
}

// NewJSONFileRepository creates a new JSONFileRepository instance.
func NewJSONFileRepository(filePath string) *JSONFileRepository {
	return &JSONFileRepository{FilePath: filePath}
}

// Load reads the catalog from disk. It creates an empty catalog if the file doesn't exist.
func (r *JSONFileRepository) Load() (*Catalog, error) {
	if _, err := os.Stat(r.FilePath); os.IsNotExist(err) {
		return &Catalog{
			Version:   "1.0",
			UpdatedAt: time.Now(),
			Entries:   []CatalogEntry{},
		}, nil
	}

	data, err := os.ReadFile(r.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read catalog file: %w", err)
	}

	var catalog Catalog
	if err := json.Unmarshal(data, &catalog); err != nil {
		return nil, fmt.Errorf("failed to unmarshal catalog: %w", err)
	}

	return &catalog, nil
}

// Save writes the catalog to disk using an atomic operation.
func (r *JSONFileRepository) Save(catalog *Catalog) error {
	catalog.UpdatedAt = time.Now()

	data, err := json.MarshalIndent(catalog, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal catalog: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(r.FilePath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("failed to create catalog directory: %w", err)
	}

	if err := fsops.WriteFile(r.FilePath, data); err != nil {
		return fmt.Errorf("failed to write catalog file: %w", err)
	}

	return nil
}
