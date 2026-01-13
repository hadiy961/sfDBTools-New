// File : internal/shared/fsops/resolver.go
// Deskripsi : Helper resolve path file (validasi + prompt interaktif) untuk menghindari duplikasi
// Author : Hadiyatna Muflihun
// Tanggal : 14 Januari 2026
// Last Modified : 14 Januari 2026

package fsops

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sfdbtools/internal/ui/print"
	"sfdbtools/internal/ui/prompt"
)

// FileResolverOptions konfigurasi untuk resolve file.
//
// Catatan:
// - Untuk mode non-interaktif, set AllowInteractive=false dan pastikan FilePath sudah diisi.
// - ValidExtensions berupa suffix (contoh: []string{".sql", ".sql.gz", ".enc"}).
// - DefaultDir dipakai saat FilePath kosong dan perlu prompt.
//
// Example:
//
//	var filePath string
//	err := fsops.ResolveFileWithPrompt(fsops.FileResolverOptions{
//	    FilePath:          &filePath,
//	    AllowInteractive:  true,
//	    ValidExtensions:   []string{".sql", ".sql.gz", ".sql.gz.enc"},
//	    Purpose:           "file backup",
//	    PromptLabel:       "Masukkan path directory atau file backup",
//	    DefaultDir:        "/var/backups",
//	})
//	if err != nil { /* handle */ }
//
//	// filePath sudah absolute jika berhasil.
type FileResolverOptions struct {
	FilePath         *string  // Pointer ke variable yang akan diisi
	AllowInteractive bool     // Allow user prompt
	ValidExtensions  []string // Extensions yang valid (e.g., []string{".sql", ".gz"})
	Purpose          string   // Deskripsi tujuan file (untuk error message)
	PromptLabel      string   // Label untuk prompt user
	DefaultDir       string   // Default directory untuk file picker
}

// ResolveFileWithPrompt resolve file path dengan validasi dan prompt interaktif.
//
// Perilaku:
// - Jika FilePath kosong:
//   - Non-interaktif: return error (purpose wajib diisi)
//   - Interaktif: prompt user via prompt.SelectFile(defaultDir,...)
//
// - Jika path adalah direktori:
//   - Non-interaktif: return error
//   - Interaktif: prompt pilih file di direktori tsb
//
// - Jika ekstensi tidak valid:
//   - Non-interaktif: return error
//   - Interaktif: warning lalu prompt ulang dari parent dir
//
// - Jika berhasil: path diubah menjadi absolute.
func ResolveFileWithPrompt(opts FileResolverOptions) error {
	if opts.FilePath == nil {
		return fmt.Errorf("filePath pointer tidak boleh nil")
	}

	purpose := strings.TrimSpace(opts.Purpose)
	if purpose == "" {
		purpose = "file"
	}

	promptLabel := strings.TrimSpace(opts.PromptLabel)
	if promptLabel == "" {
		promptLabel = "Masukkan path file"
	}

	defaultDir := strings.TrimSpace(opts.DefaultDir)
	if defaultDir == "" {
		defaultDir = "."
	}

	if strings.TrimSpace(*opts.FilePath) == "" {
		if !opts.AllowInteractive {
			return fmt.Errorf("%s wajib diisi pada mode non-interaktif (--force)", purpose)
		}

		selectedFile, err := prompt.SelectFile(defaultDir, promptLabel, opts.ValidExtensions)
		if err != nil {
			return fmt.Errorf("gagal memilih %s: %w", purpose, err)
		}
		*opts.FilePath = selectedFile
	}

	// Validasi file exists.
	if err := ValidateFileExists(*opts.FilePath); err != nil {
		return fmt.Errorf("%s: %w", purpose, err)
	}

	fi, err := os.Stat(*opts.FilePath)
	if err != nil {
		return fmt.Errorf("gagal membaca %s: %w", purpose, err)
	}

	if fi.IsDir() {
		if !opts.AllowInteractive {
			return fmt.Errorf("%s tidak valid (path adalah direktori): %s", purpose, *opts.FilePath)
		}

		selectedFile, selErr := prompt.SelectFile(*opts.FilePath, promptLabel, opts.ValidExtensions)
		if selErr != nil {
			return fmt.Errorf("gagal memilih %s: %w", purpose, selErr)
		}
		*opts.FilePath = selectedFile
	}

	if err := ValidateFileExtension(*opts.FilePath, opts.ValidExtensions, purpose); err != nil {
		if !opts.AllowInteractive {
			return err
		}
		print.PrintWarning(fmt.Sprintf("⚠️  %s", err.Error()))

		selectedFile, selErr := prompt.SelectFile(filepath.Dir(*opts.FilePath), promptLabel, opts.ValidExtensions)
		if selErr != nil {
			return fmt.Errorf("gagal memilih %s: %w", purpose, selErr)
		}
		*opts.FilePath = selectedFile
	}

	absPath, err := filepath.Abs(*opts.FilePath)
	if err != nil {
		return fmt.Errorf("gagal mendapatkan absolute path: %w", err)
	}
	*opts.FilePath = absPath

	return nil
}

// ValidateFileExists memvalidasi apakah file ada.
func ValidateFileExists(path string) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("path file tidak boleh kosong")
	}

	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file tidak ditemukan: %s", path)
		}
		return fmt.Errorf("gagal mengakses file: %w", err)
	}
	return nil
}

// ValidateFileExtension memvalidasi extension file.
// validExts diperlakukan sebagai suffix (bisa multi-ext seperti ".sql.gz.enc").
func ValidateFileExtension(path string, validExts []string, purpose string) error {
	if len(validExts) == 0 {
		return nil // no validation needed
	}

	if HasAnySuffix(path, validExts) {
		return nil
	}

	p := strings.TrimSpace(purpose)
	if p == "" {
		p = "file"
	}
	return fmt.Errorf("%s tidak valid, expected extensions: %v", p, validExts)
}

// HasAnySuffix cek apakah path memiliki salah satu suffix.
func HasAnySuffix(path string, suffixes []string) bool {
	lower := strings.ToLower(strings.TrimSpace(path))
	for _, s := range suffixes {
		if strings.HasSuffix(lower, strings.ToLower(strings.TrimSpace(s))) {
			return true
		}
	}
	return false
}
