// File : internal/app/dbscan/local_scan.go
// Deskripsi : Local filesystem scanning untuk database sizes
// Author : Hadiyatna Muflihun
// Tanggal : 23 Januari 2026
// Last Modified : 23 Januari 2026
package dbscan

import (
	"context"
	"os"
	"path/filepath"
)

// LocalScanSizes menghitung ukuran database dari filesystem lokal
func (s *Service) LocalScanSizes(ctx context.Context, datadir string, dbNames []string) (map[string]int64, error) {
	s.Log.Infof("Starting local size scan in: %s", datadir)
	sizes := make(map[string]int64, len(dbNames))

	for _, dbName := range dbNames {
		if ctx.Err() != nil {
			return sizes, ctx.Err()
		}

		dbPath := filepath.Join(datadir, dbName)
		size, err := dirSize(ctx, dbPath)
		if err != nil {
			s.Log.Warnf("Failed to calculate size for %s: %v", dbName, err)
			sizes[dbName] = 0
			continue
		}
		sizes[dbName] = size
	}
	return sizes, nil
}

// dirSize helper untuk menghitung ukuran direktori rekursif
func dirSize(ctx context.Context, root string) (int64, error) {
	var total int64
	info, err := os.Stat(root)
	if err != nil {
		return 0, err
	}
	if !info.IsDir() {
		return info.Size(), nil
	}

	err = filepath.WalkDir(root, func(_ string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if d.Type().IsRegular() {
			if fi, err := d.Info(); err == nil {
				total += fi.Size()
			}
		}
		return nil
	})
	return total, err
}
