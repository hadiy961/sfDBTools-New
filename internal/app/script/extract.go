package script

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	scriptmodel "sfDBTools/internal/app/script/model"
	"sfDBTools/internal/ui/prompt"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/encrypt"
	"sfDBTools/pkg/helper"
	"strings"
)

func ExtractBundle(opts scriptmodel.ScriptExtractOptions) error {
	bundlePath := strings.TrimSpace(opts.FilePath)
	if bundlePath == "" {
		return fmt.Errorf("--file wajib diisi")
	}
	bundlePath = helper.ExpandPath(bundlePath)

	outDir := strings.TrimSpace(opts.OutDir)
	if outDir == "" {
		return fmt.Errorf("--out-dir wajib diisi")
	}
	outDir = helper.ExpandPath(outDir)
	outDir = filepath.Clean(outDir)

	key, _, err := helper.ResolveEncryptionKey(opts.EncryptionKey, consts.ENV_SCRIPT_KEY)
	if err != nil {
		return fmt.Errorf("gagal mendapatkan encryption key: %w", err)
	}

	// Proteksi ganda: setelah key, wajib password aplikasi.
	password, err := prompt.AskPassword("Masukkan password aplikasi untuk melanjutkan extract:", nil)
	if err != nil {
		return fmt.Errorf("gagal membaca password aplikasi: %w", err)
	}
	if password != consts.ENV_PASSWORD_APP {
		return fmt.Errorf("password aplikasi tidak valid")
	}

	if err := ensureEmptyDir(outDir); err != nil {
		return err
	}

	f, err := os.Open(bundlePath)
	if err != nil {
		return fmt.Errorf("gagal membuka .sftools: %w", err)
	}
	defer f.Close()

	decReader, err := encrypt.NewDecryptingReader(f, key)
	if err != nil {
		return fmt.Errorf("gagal membuat decrypting reader: %w", err)
	}

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
		outPath := filepath.Join(outDir, filepath.FromSlash(cleanName))

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

	return nil
}

func ensureEmptyDir(dir string) error {
	st, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(dir, 0700); err != nil {
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
