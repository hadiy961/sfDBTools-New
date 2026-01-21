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
	"os/exec"
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
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			// Log warning tapi tidak return error karena data sudah tercopy
			// (ini defer, sudah terlambat untuk propagate error)
		}
	}()

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

// extractTarEntry extracts single tar entry dengan zip bomb protection.
// Digunakan oleh RunBundle dan ExtractBundle untuk DRY.
func extractTarEntry(hdr *tar.Header, tr *tar.Reader, baseDir string, totalExtracted *int64) error {
	// P2 #4: Centralized zip bomb protection
	if hdr.Size > MaxFileSize {
		return fmt.Errorf("file terlalu besar: %s (%d bytes, max %d)", hdr.Name, hdr.Size, MaxFileSize)
	}
	*totalExtracted += hdr.Size
	if *totalExtracted > MaxExtractedSize {
		return fmt.Errorf("bundle terlalu besar: %d bytes (max %d, zip bomb protection)", *totalExtracted, MaxExtractedSize)
	}

	cleanName, err := safeTarPath(hdr.Name)
	if err != nil {
		return err
	}
	outPath := filepath.Join(baseDir, filepath.FromSlash(cleanName))

	switch hdr.Typeflag {
	case tar.TypeDir:
		if err := os.MkdirAll(outPath, SecureDirPermission); err != nil {
			return fmt.Errorf("gagal membuat folder: %w", err)
		}
	case tar.TypeReg:
		if err := os.MkdirAll(filepath.Dir(outPath), SecureDirPermission); err != nil {
			return fmt.Errorf("gagal membuat parent folder: %w", err)
		}
		wf, err := os.OpenFile(outPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, SecureFilePermission)
		if err != nil {
			return fmt.Errorf("gagal membuat file extract: %w", err)
		}
		if _, err := io.Copy(wf, tr); err != nil {
			_ = wf.Close()
			return fmt.Errorf("gagal menulis file extract: %w", err)
		}
		if err := wf.Close(); err != nil {
			return fmt.Errorf("gagal menutup file extract: %w", err)
		}
	default:
		return fmt.Errorf("tipe tar tidak didukung: %v", hdr.Typeflag)
	}
	return nil
}

// safeTarPath validates and cleans tar entry path to prevent path traversal attacks.
//
// Rejects:
//   - Absolute paths ("/foo")
//   - Parent directory references ("../foo")
//   - Empty or invalid paths
//   - Paths exceeding MaxPathLength (prevent buffer overflow)
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
	// P1 #5: Path length validation (prevent overflow attacks)
	if len(clean) > MaxPathLength {
		return "", fmt.Errorf("path too long: %d chars (max %d)", len(clean), MaxPathLength)
	}
	return clean, nil
}

// validateBundleExtension checks if file has valid .sftools extension.
func validateBundleExtension(filePath string) error {
	if !strings.HasSuffix(strings.ToLower(filePath), BundleExtension) {
		return fmt.Errorf("file bukan bundle .sftools: %s", filePath)
	}
	return nil
}

// detectShell returns available shell binary (bash preferred, fallback to sh).
func detectShell() string {
	// P2 #5: Configurable shell dengan fallback
	for _, shell := range []string{"bash", "sh"} {
		if _, err := exec.LookPath(shell); err == nil {
			return shell
		}
	}
	return DefaultShell // fallback ke hardcoded default
}
