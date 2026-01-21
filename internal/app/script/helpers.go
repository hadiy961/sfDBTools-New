// File : internal/app/script/helpers.go
// Deskripsi : Helper utilities untuk script operations
// Author : Hadiyatna Muflihun
// Tanggal : 21 Januari 2026
// Last Modified : 21 Januari 2026

package script

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// defaultBundleOutputPath generates default output path for bundle.
// Converts "script.sh" → "script.sftools"
func defaultBundleOutputPath(entryPath string) string {
	ext := strings.ToLower(filepath.Ext(entryPath))
	if ext == ".sh" {
		return strings.TrimSuffix(entryPath, filepath.Ext(entryPath)) + ".sftools"
	}
	return entryPath + ".sftools"
}

// listFilesRecursive returns all files (not directories) in rootDir recursively.
func listFilesRecursive(rootDir string) ([]string, error) {
	var files []string
	walkErr := filepath.WalkDir(rootDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		files = append(files, p)
		return nil
	})
	if walkErr != nil {
		return nil, fmt.Errorf("gagal scan folder: %w", walkErr)
	}
	return files, nil
}

// writeTarBytes writes bytes to tar archive with specified mode.
func writeTarBytes(tw *tar.Writer, name string, b []byte, mode int64) error {
	h := &tar.Header{
		Name: name,
		Mode: mode,
		Size: int64(len(b)),
	}
	if err := tw.WriteHeader(h); err != nil {
		return fmt.Errorf("gagal menulis tar header: %w", err)
	}
	if _, err := tw.Write(b); err != nil {
		return fmt.Errorf("gagal menulis tar data: %w", err)
	}
	return nil
}

// addFileToTar adds a file to tar archive with specified name and mode.
func addFileToTar(tw *tar.Writer, srcPath string, tarName string, mode int64) error {
	f, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("gagal membuka file: %w", err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return fmt.Errorf("gagal stat file: %w", err)
	}

	h := &tar.Header{
		Name: tarName,
		Mode: mode,
		Size: info.Size(),
	}
	if err := tw.WriteHeader(h); err != nil {
		return fmt.Errorf("gagal menulis tar header: %w", err)
	}
	if _, err := io.Copy(tw, f); err != nil {
		return fmt.Errorf("gagal menulis tar payload: %w", err)
	}
	return nil
}

// safeTarPath validates and cleans tar entry path to prevent path traversal attacks.
//
// Rejects:
//   - Absolute paths ("/foo")
//   - Parent directory references ("../foo")
//   - Empty or invalid paths
//
// Returns cleaned path or error if validation fails.
func safeTarPath(name string) (string, error) {
	// Tar paths menggunakan '/', gunakan path.Clean.
	clean := path.Clean(name)
	clean = strings.TrimPrefix(clean, "./")

	if clean == "." || clean == "" {
		return "", errors.New("invalid tar entry name")
	}
	if strings.HasPrefix(clean, "../") || clean == ".." {
		return "", fmt.Errorf("invalid tar entry path: %s", name)
	}
	if strings.HasPrefix(clean, "/") {
		return "", fmt.Errorf("invalid tar entry absolute path: %s", name)
	}
	return clean, nil
}
