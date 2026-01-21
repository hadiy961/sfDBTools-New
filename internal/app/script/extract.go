// File : internal/app/script/extract.go
// Deskripsi : Bundle extraction untuk encrypted script bundles
// Author : Hadiyatna Muflihun
// Tanggal : 21 Januari 2026
// Last Modified : 21 Januari 2026

package script

import (
	"archive/tar"
	"context"
	"crypto/subtle"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sfdbtools/internal/crypto"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/envx"
	"sfdbtools/internal/ui/prompt"
	"strings"
)

// ExtractBundle decrypts and extracts script bundle to output directory.
//
// Security:
//   - Requires encryption key + application password
//   - Output directory must be empty
//   - Path traversal protection via safeTarPath()
//
// Context dapat digunakan untuk cancellation.
func ExtractBundle(ctx context.Context, opts ExtractOptions) error {
	// Check context cancellation
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("operasi dibatalkan: %w", err)
	}
	bundlePath := strings.TrimSpace(opts.FilePath)
	if bundlePath == "" {
		return fmt.Errorf("--file wajib diisi")
	}
	bundlePath = envx.ExpandPath(bundlePath)

	// P2 #7: Validate bundle extension
	if err := validateBundleExtension(bundlePath); err != nil {
		return err
	}

	outDir := strings.TrimSpace(opts.OutDir)
	if outDir == "" {
		return fmt.Errorf("--out-dir wajib diisi")
	}
	outDir = envx.ExpandPath(outDir)
	outDir = filepath.Clean(outDir)

	key, _, err := crypto.ResolveKey(opts.EncryptionKey, consts.ENV_SCRIPT_KEY, true)
	if err != nil {
		return fmt.Errorf("gagal mendapatkan encryption key: %w", err)
	}

	// Proteksi ganda: setelah key, wajib password aplikasi.
	password, err := prompt.AskPassword("Masukkan password aplikasi untuk melanjutkan extract:", nil)
	if err != nil {
		return fmt.Errorf("gagal membaca password aplikasi: %w", err)
	}
	// P0 #2: Use constant-time comparison untuk prevent timing attacks
	expectedBytes := []byte(consts.ENV_PASSWORD_APP)
	providedBytes := []byte(password)
	if subtle.ConstantTimeCompare(expectedBytes, providedBytes) != 1 {
		return fmt.Errorf("password aplikasi tidak valid: autentikasi gagal")
	}

	if err := ensureEmptyDir(outDir); err != nil {
		return err
	}

	f, err := os.Open(bundlePath)
	if err != nil {
		return fmt.Errorf("gagal membuka .sftools: %w", err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			log.Printf("Warning: gagal menutup bundle file: %v", closeErr)
		}
	}()

	decReader, err := crypto.NewStreamDecryptor(f, key)
	if err != nil {
		return fmt.Errorf("gagal membuat decrypting reader: %w", err)
	}

	tr := tar.NewReader(decReader)
	// P2 #4: Use shared extraction logic dengan zip bomb protection
	var totalExtracted int64
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("gagal membaca tar: %w", err)
		}
		if err := extractTarEntry(hdr, tr, outDir, &totalExtracted); err != nil {
			return err
		}
	}

	return nil
}

func ensureEmptyDir(dir string) error {
	st, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(dir, SecureDirPermission); err != nil {
				return fmt.Errorf("gagal membuat out-dir: %w", err)
			}
			return nil
		}
		return fmt.Errorf("gagal membaca out-dir: %w", err)
	}
	if !st.IsDir() {
		return fmt.Errorf("out-dir bukan folder: %s", dir)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("gagal membaca out-dir: %w", err)
	}
	if len(entries) > 0 {
		return fmt.Errorf("out-dir harus kosong: %s", dir)
	}
	return nil
}
