package catalog

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"sfdbtools/internal/app/backup/model/types_backup"
	applog "sfdbtools/internal/services/log"

	"github.com/dustin/go-humanize"
	"github.com/google/uuid"
)

// Service provides business logic for the catalog domain.
type Service struct {
	repo   Repository
	logger applog.Logger
}

// NewService creates a new catalog service.
func NewService(repo Repository, logger applog.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

// TryRegisterBackup wraps RegisterBackup to log errors silently, suitable for post-backup hooks.
func (s *Service) TryRegisterBackup(meta *types_backup.BackupMetadata, mode string, profile string) {
	if err := s.RegisterBackup(meta, mode, profile); err != nil {
		s.logger.Errorf("Failed to auto-register backup to catalog: %v", err)
	}
}

// RegisterBackup converts metadata into an entry and appends it to the catalog.
func (s *Service) RegisterBackup(meta *types_backup.BackupMetadata, mode string, profile string) error {
	cat, err := s.repo.Load()
	if err != nil {
		return err
	}

	checksum := ""
	if meta.Verification != nil {
		checksum = meta.Verification.ChecksumHash
	}

	backupTime := meta.BackupStartTime
	if backupTime.IsZero() {
		backupTime = time.Now()
	}

	entry := CatalogEntry{
		ID:              uuid.New().String(),
		BackupFile:      meta.BackupFile,
		MetadataFile:    meta.BackupFile + ".meta.json", // Approximation
		DatabaseNames:   meta.DatabaseNames,
		Hostname:        meta.Hostname,
		BackupType:      meta.BackupType,
		BackupMode:      mode,
		BackupStatus:    meta.BackupStatus,
		BackupTime:      backupTime,
		FileSizeBytes:   meta.FileSize,
		FileSizeHuman:   humanize.Bytes(uint64(meta.FileSize)),
		Compressed:      meta.Compressed,
		CompressionType: meta.CompressionType,
		Encrypted:       meta.Encrypted,
		Ticket:          meta.Ticket,
		GTIDInfo:        meta.GTIDInfo,
		ChecksumHash:    checksum,
		ProfileUsed:     profile,
		RegisteredAt:    time.Now(),
	}

	// Simple deduplication based on BackupFile path
	for i, existing := range cat.Entries {
		if existing.BackupFile == entry.BackupFile {
			cat.Entries[i] = entry
			return s.repo.Save(cat)
		}
	}

	cat.Entries = append(cat.Entries, entry)
	return s.repo.Save(cat)
}

// Query returns a filtered and sorted list of catalog entries.
func (s *Service) Query(opts QueryOptions) ([]CatalogEntry, error) {
	cat, err := s.repo.Load()
	if err != nil {
		return nil, err
	}

	var results []CatalogEntry
	var cutoffTime time.Time

	if opts.Since != "" {
		dur, err := time.ParseDuration(opts.Since)
		if err == nil {
			cutoffTime = time.Now().Add(-dur)
		}
	}

	for _, e := range cat.Entries {
		if opts.Database != "" {
			match := false
			for _, db := range e.DatabaseNames {
				if strings.Contains(strings.ToLower(db), strings.ToLower(opts.Database)) {
					match = true
					break
				}
			}
			if !match {
				continue
			}
		}

		if opts.Status != "" && !strings.EqualFold(e.BackupStatus, opts.Status) {
			continue
		}

		if opts.Hostname != "" && !strings.Contains(strings.ToLower(e.Hostname), strings.ToLower(opts.Hostname)) {
			continue
		}

		if !cutoffTime.IsZero() && e.BackupTime.Before(cutoffTime) {
			continue
		}

		results = append(results, e)
	}

	// Sort by BackupTime DESC
	sort.Slice(results, func(i, j int) bool {
		return results[i].BackupTime.After(results[j].BackupTime)
	})

	if opts.Limit > 0 && len(results) > opts.Limit {
		results = results[:opts.Limit]
	}

	return results, nil
}

// RebuildFromDirectory scans a directory for .meta.json files and rebuilds the catalog.
func (s *Service) RebuildFromDirectory(dir string, recursive bool) (int, error) {
	_, err := s.repo.Load()
	if err != nil {
		return 0, err
	}

	count := 0
	
	processFile := func(path string, d os.DirEntry) error {
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".meta.json") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil // Skip unreadable
		}

		var meta types_backup.BackupMetadata
		if err := json.Unmarshal(data, &meta); err != nil {
			return nil // Skip invalid
		}

		// Use metadata's own backup file path if relative
		if !filepath.IsAbs(meta.BackupFile) {
			meta.BackupFile = filepath.Join(filepath.Dir(path), meta.BackupFile)
		}

		// Mode/Profile extraction from metadata if possible, otherwise unknown
		mode := "unknown"
		profile := "unknown"
		
		s.RegisterBackup(&meta, mode, profile)
		count++
		return nil
	}

	if recursive {
		err = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil // Skip errors
			}
			return processFile(path, d)
		})
	} else {
		files, err := os.ReadDir(dir)
		if err != nil {
			return 0, fmt.Errorf("failed to read directory: %w", err)
		}
		for _, f := range files {
			processFile(filepath.Join(dir, f.Name()), f)
		}
	}

	return count, err
}

// GetPrunableEntries returns entries whose backup files no longer exist on disk.
func (s *Service) GetPrunableEntries() ([]CatalogEntry, error) {
	cat, err := s.repo.Load()
	if err != nil {
		return nil, err
	}

	var prunable []CatalogEntry
	for _, e := range cat.Entries {
		if _, err := os.Stat(e.BackupFile); os.IsNotExist(err) {
			prunable = append(prunable, e)
		}
	}

	return prunable, nil
}

// DeleteEntries removes multiple entries from the catalog by their IDs.
func (s *Service) DeleteEntries(ids []string) (int, error) {
	cat, err := s.repo.Load()
	if err != nil {
		return 0, err
	}

	idMap := make(map[string]bool)
	for _, id := range ids {
		idMap[id] = true
	}

	var newEntries []CatalogEntry
	removed := 0
	for _, e := range cat.Entries {
		if idMap[e.ID] {
			removed++
			continue
		}
		newEntries = append(newEntries, e)
	}

	if removed > 0 {
		cat.Entries = newEntries
		if err := s.repo.Save(cat); err != nil {
			return 0, err
		}
	}

	return removed, nil
}

// Prune removes entries whose backup files no longer exist on disk automatically.
func (s *Service) Prune() (int, error) {
	prunable, err := s.GetPrunableEntries()
	if err != nil {
		return 0, err
	}

	if len(prunable) == 0 {
		return 0, nil
	}

	var ids []string
	for _, e := range prunable {
		ids = append(ids, e.ID)
	}

	return s.DeleteEntries(ids)
}
