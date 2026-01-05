package script

import (
	"archive/tar"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"sfDBTools/internal/types"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/encrypt"
	"sfDBTools/pkg/helper"
)

type BundleInfo struct {
	Version    int      `json:"version"`
	Entrypoint string   `json:"entrypoint"`
	CreatedAt  string   `json:"created_at"`
	Mode       string   `json:"mode"`
	RootDir    string   `json:"root_dir"`
	Scripts    []string `json:"scripts"`
	FileCount  int      `json:"file_count"`
}

func GetBundleInfo(opts types.ScriptInfoOptions) (BundleInfo, error) {
	bundlePath := strings.TrimSpace(opts.FilePath)
	if bundlePath == "" {
		return BundleInfo{}, fmt.Errorf("--file wajib diisi")
	}
	bundlePath = helper.ExpandPath(bundlePath)

	key, _, err := helper.ResolveEncryptionKey(opts.EncryptionKey, consts.ENV_SCRIPT_KEY)
	if err != nil {
		return BundleInfo{}, fmt.Errorf("gagal mendapatkan encryption key: %w", err)
	}

	f, err := os.Open(bundlePath)
	if err != nil {
		return BundleInfo{}, fmt.Errorf("gagal membuka .sftools: %w", err)
	}
	defer f.Close()

	decReader, err := encrypt.NewDecryptingReader(f, key)
	if err != nil {
		return BundleInfo{}, fmt.Errorf("gagal membuat decrypting reader: %w", err)
	}

	// Scan seluruh tar: ambil manifest + list script .sh tanpa harus extract.
	type manifestLite struct {
		Version    int    `json:"version"`
		Entrypoint string `json:"entrypoint"`
		CreatedAt  string `json:"created_at"`
		Mode       string `json:"mode"`
		RootDir    string `json:"root_dir"`
	}

	var m manifestLite
	manifestFound := false
	fileCount := 0
	var regFiles []string
	var scripts []string

	tr := tar.NewReader(decReader)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return BundleInfo{}, fmt.Errorf("gagal membaca tar: %w", err)
		}
		cleanName, err := safeTarPath(hdr.Name)
		if err != nil {
			return BundleInfo{}, err
		}

		if hdr.Typeflag != tar.TypeReg {
			continue
		}

		if cleanName == manifestFilename {
			b, err := io.ReadAll(tr)
			if err != nil {
				return BundleInfo{}, fmt.Errorf("gagal membaca manifest: %w", err)
			}
			if err := json.Unmarshal(b, &m); err != nil {
				return BundleInfo{}, fmt.Errorf("manifest invalid: %w", err)
			}
			manifestFound = true
			continue
		}

		fileCount++
		regFiles = append(regFiles, cleanName)
		if strings.EqualFold(filepath.Ext(cleanName), ".sh") {
			scripts = append(scripts, cleanName)
		}
	}

	if !manifestFound {
		return BundleInfo{}, fmt.Errorf("manifest tidak ditemukan dalam bundle")
	}

	mode := strings.ToLower(strings.TrimSpace(m.Mode))
	if mode == "" {
		// Backward-compat: infer mode dari jumlah file.
		if len(regFiles) == 1 && strings.TrimSpace(m.Entrypoint) != "" && filepath.Base(regFiles[0]) == m.Entrypoint {
			mode = "single"
		} else {
			mode = "bundle"
		}
	}

	rootDir := strings.TrimSpace(m.RootDir)
	if rootDir == "" {
		// Backward-compat fallback: pakai nama bundle file.
		base := filepath.Base(bundlePath)
		rootDir = strings.TrimSuffix(base, filepath.Ext(base))
	}

	sort.Strings(scripts)

	return BundleInfo{
		Version:    m.Version,
		Entrypoint: m.Entrypoint,
		CreatedAt:  m.CreatedAt,
		Mode:       mode,
		RootDir:    rootDir,
		Scripts:    scripts,
		FileCount:  fileCount,
	}, nil
}
