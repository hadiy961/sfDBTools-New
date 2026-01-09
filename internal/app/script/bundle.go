package script

import (
	"archive/tar"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	scriptmodel "sfdbtools/internal/app/script/model"
	"sfdbtools/internal/crypto"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/envx"
	"strings"
	"time"
)

const (
	manifestFilename = ".sftools-manifest.json"
	bundleVersion    = 1
)

type manifest struct {
	Version    int    `json:"version"`
	Entrypoint string `json:"entrypoint"`
	CreatedAt  string `json:"created_at"`
	Mode       string `json:"mode,omitempty"`
	RootDir    string `json:"root_dir,omitempty"`
}

func EncryptBundle(opts scriptmodel.ScriptEncryptOptions) error {
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

	if err := os.MkdirAll(filepath.Dir(outputPath), 0700); err != nil {
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

	outFile, err := os.OpenFile(outputPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("gagal membuat output file: %w", err)
	}
	defer outFile.Close()

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
	if err := writeTarBytes(tarWriter, manifestFilename, manifestBytes, 0600); err != nil {
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

		if err := addFileToTar(tarWriter, filePath, rel, 0600); err != nil {
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

func RunBundle(opts scriptmodel.ScriptRunOptions) error {
	bundlePath := strings.TrimSpace(opts.FilePath)
	if bundlePath == "" {
		return fmt.Errorf("--file wajib diisi")
	}
	bundlePath = envx.ExpandPath(bundlePath)

	key, _, err := crypto.ResolveKey(opts.EncryptionKey, consts.ENV_SCRIPT_KEY, true)
	if err != nil {
		return fmt.Errorf("gagal mendapatkan encryption key: %w", err)
	}

	f, err := os.Open(bundlePath)
	if err != nil {
		return fmt.Errorf("gagal membuka .sftools: %w", err)
	}
	defer f.Close()

	decReader, err := crypto.NewStreamDecryptor(f, key)
	if err != nil {
		return fmt.Errorf("gagal membuat decrypting reader: %w", err)
	}

	tmpDir, err := os.MkdirTemp("", "sfdbtools-script-*")
	if err != nil {
		return fmt.Errorf("gagal membuat temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	tr := tar.NewReader(decReader)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("gagal membaca tar: %w", err)
		}

		cleanName, err := safeTarPath(hdr.Name)
		if err != nil {
			return err
		}
		outPath := filepath.Join(tmpDir, filepath.FromSlash(cleanName))

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(outPath, 0700); err != nil {
				return fmt.Errorf("gagal membuat folder: %w", err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(outPath), 0700); err != nil {
				return fmt.Errorf("gagal membuat parent folder: %w", err)
			}
			wf, err := os.OpenFile(outPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
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

	entry := filepath.Join(tmpDir, filepath.FromSlash(m.Entrypoint))
	if _, err := os.Stat(entry); err != nil {
		return fmt.Errorf("entrypoint tidak ditemukan: %w", err)
	}

	cmdArgs := append([]string{entry}, opts.Args...)
	cmd := exec.Command("bash", cmdArgs...)
	cmd.Dir = tmpDir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("script gagal dijalankan: %w", err)
	}

	return nil
}

func defaultBundleOutputPath(entryPath string) string {
	ext := strings.ToLower(filepath.Ext(entryPath))
	if ext == ".sh" {
		return strings.TrimSuffix(entryPath, filepath.Ext(entryPath)) + ".sftools"
	}
	return entryPath + ".sftools"
}

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
