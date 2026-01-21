// File : internal/app/script/bundle.go
// Deskripsi : Bundle creation and execution untuk encrypted script bundles
// Author : Hadiyatna Muflihun
// Tanggal : 21 Januari 2026
// Last Modified : 21 Januari 2026

package script

import (
	"archive/tar"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sfdbtools/internal/crypto"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/envx"
	"strings"
	"time"
)

// EncryptBundle creates encrypted script bundle (.sftools) from script file(s).
//
// Modes:
//   - "bundle": Package entire directory with entrypoint
//   - "single": Package only the entrypoint file
//
// Output format: encrypted tar archive with manifest + script files
// Context dapat digunakan untuk cancellation.
func EncryptBundle(ctx context.Context, opts EncryptOptions) error {
	// Check context cancellation sebelum operasi berat
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("operasi dibatalkan: %w", err)
	}
	entryPath := strings.TrimSpace(opts.FilePath)
	if entryPath == "" {
		return fmt.Errorf("--file wajib diisi")
	}

	entryPath = envx.ExpandPath(entryPath)
	entryInfo, err := os.Stat(entryPath)
	if err != nil {
		return fmt.Errorf("gagal membaca file entrypoint: %w", err)
	}
	if entryInfo.IsDir() {
		return fmt.Errorf("--file harus file, bukan folder")
	}

	rootDir := filepath.Dir(entryPath)
	entryBase := filepath.Base(entryPath)
	outputPath := strings.TrimSpace(opts.OutputPath)
	if outputPath == "" {
		outputPath = defaultBundleOutputPath(entryPath)
	}
	outputPath = envx.ExpandPath(outputPath)

	entryAbs, _ := filepath.Abs(entryPath)
	outAbs, _ := filepath.Abs(outputPath)
	if entryAbs != "" && outAbs != "" && entryAbs == outAbs {
		return fmt.Errorf("output tidak boleh menimpa file entrypoint")
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), SecureDirPermission); err != nil {
		return fmt.Errorf("gagal membuat folder output: %w", err)
	}

	key, _, err := crypto.ResolveKey(opts.EncryptionKey, consts.ENV_SCRIPT_KEY, true)
	if err != nil {
		return fmt.Errorf("gagal mendapatkan encryption key: %w", err)
	}

	mode := strings.ToLower(strings.TrimSpace(opts.Mode))
	if mode == "" {
		mode = "bundle"
	}

	var filtered []string
	switch mode {
	case "bundle":
		// Kumpulkan list file dulu supaya output file tidak ikut kebundle.
		files, err := listFilesRecursive(rootDir)
		if err != nil {
			return err
		}
		var outAbs2 string
		outAbs2, _ = filepath.Abs(outputPath)
		for _, f := range files {
			abs, _ := filepath.Abs(f)
			if outAbs2 != "" && abs == outAbs2 {
				continue
			}
			filtered = append(filtered, f)
		}
	case "single":
		filtered = []string{entryPath}
	default:
		return fmt.Errorf("mode tidak valid: %s (pilih: bundle|single)", opts.Mode)
	}

	outFile, err := os.OpenFile(outputPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, SecureFilePermission)
	if err != nil {
		return fmt.Errorf("gagal membuat output file: %w", err)
	}
	defer func() {
		if closeErr := outFile.Close(); closeErr != nil {
			log.Printf("Warning: gagal menutup output file: %v", closeErr)
		}
	}()

	ew, err := crypto.NewStreamEncryptor(outFile, []byte(key))
	if err != nil {
		return fmt.Errorf("gagal membuat encrypting writer: %w", err)
	}

	tarWriter := tar.NewWriter(ew)

	m := manifest{
		Version:    bundleVersion,
		Entrypoint: entryBase,
		CreatedAt:  time.Now().UTC().Format(time.RFC3339),
		Mode:       mode,
		RootDir:    filepath.Base(rootDir),
	}
	manifestBytes, _ := json.MarshalIndent(m, "", "  ")
	if err := writeTarBytes(tarWriter, manifestFilename, manifestBytes, SecureFilePermission); err != nil {
		_ = tarWriter.Close()
		_ = ew.Close()
		return err
	}

	for _, filePath := range filtered {
		rel, err := filepath.Rel(rootDir, filePath)
		if err != nil {
			_ = tarWriter.Close()
			_ = ew.Close()
			return fmt.Errorf("gagal resolve relative path: %w", err)
		}
		rel = filepath.ToSlash(rel)
		if rel == "." {
			continue
		}

		info, err := os.Stat(filePath)
		if err != nil {
			_ = tarWriter.Close()
			_ = ew.Close()
			return fmt.Errorf("gagal stat file: %w", err)
		}
		if info.IsDir() {
			continue
		}

		if err := addFileToTar(tarWriter, filePath, rel, SecureFilePermission); err != nil {
			_ = tarWriter.Close()
			_ = ew.Close()
			return err
		}
	}

	if err := tarWriter.Close(); err != nil {
		_ = ew.Close()
		return fmt.Errorf("gagal menutup tar writer: %w", err)
	}
	if err := ew.Close(); err != nil {
		return fmt.Errorf("gagal menutup encrypting writer: %w", err)
	}

	return nil
}

// RunBundle decrypts and executes script bundle in temporary directory.
//
// Steps:
//  1. Decrypt bundle to temp dir
//  2. Extract tar archive
//  3. Read manifest for entrypoint
//  4. Execute entrypoint with bash
//  5. Cleanup temp dir
//
// Context dapat digunakan untuk cancellation.
func RunBundle(ctx context.Context, opts RunOptions) error {
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

	key, _, err := crypto.ResolveKey(opts.EncryptionKey, consts.ENV_SCRIPT_KEY, true)
	if err != nil {
		return fmt.Errorf("gagal mendapatkan encryption key: %w", err)
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

	tmpDir, err := os.MkdirTemp("", "sfdbtools-script-*")
	if err != nil {
		return fmt.Errorf("gagal membuat temp dir: %w", err)
	}
	defer func() {
		if rmErr := os.RemoveAll(tmpDir); rmErr != nil {
			log.Printf("Warning: gagal cleanup temp dir %s: %v", tmpDir, rmErr)
		}
	}()

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
		if err := extractTarEntry(hdr, tr, tmpDir, &totalExtracted); err != nil {
			return err
		}
	}

	manifestPath := filepath.Join(tmpDir, manifestFilename)
	mb, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("manifest tidak ditemukan dalam bundle")
	}

	var m manifest
	if err := json.Unmarshal(mb, &m); err != nil {
		return fmt.Errorf("manifest invalid: %w", err)
	}
	if m.Version != bundleVersion {
		return fmt.Errorf("versi bundle tidak didukung: %d", m.Version)
	}
	if strings.TrimSpace(m.Entrypoint) == "" {
		return fmt.Errorf("manifest entrypoint kosong")
	}
	// P1 #6: Validasi manifest fields untuk prevent path traversal
	if strings.Contains(m.Entrypoint, "..") || strings.HasPrefix(m.Entrypoint, "/") {
		return fmt.Errorf("manifest entrypoint tidak valid: %s", m.Entrypoint)
	}
	if strings.Contains(m.RootDir, "..") || strings.HasPrefix(m.RootDir, "/") {
		return fmt.Errorf("manifest root_dir tidak valid: %s", m.RootDir)
	}

	entry := filepath.Join(tmpDir, filepath.FromSlash(m.Entrypoint))
	if _, err := os.Stat(entry); err != nil {
		return fmt.Errorf("entrypoint tidak ditemukan: %w", err)
	}

	// P0 #1: Sanitize args untuk prevent command injection
	for i, arg := range opts.Args {
		if strings.ContainsAny(arg, ";|&$`><\n") {
			return fmt.Errorf("argument %d mengandung karakter berbahaya: %q", i+1, arg)
		}
	}

	cmdArgs := append([]string{entry}, opts.Args...)
	// P2 #5: Use detected shell (bash preferred, fallback to sh)
	shell := detectShell()
	cmd := exec.CommandContext(ctx, shell, cmdArgs...)
	cmd.Dir = tmpDir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("script dibatalkan: %w", ctx.Err())
		}
		return fmt.Errorf("script gagal dijalankan: %w", err)
	}

	return nil
}
