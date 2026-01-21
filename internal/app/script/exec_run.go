// File : internal/app/script/exec_run.go
// Deskripsi : Execute layer untuk run bundle (dengan interactive prompts)
// Author : Hadiyatna Muflihun
// Tanggal : 21 Januari 2026
// Last Modified : 21 Januari 2026

package script

import (
	"fmt"
	"os"
	"path/filepath"
	appconfig "sfdbtools/internal/services/config"
	applog "sfdbtools/internal/services/log"
	"sfdbtools/internal/ui/prompt"
	"sort"
	"strings"
)

// ExecuteRunBundle handles run workflow dengan interactive prompts.
func ExecuteRunBundle(logger applog.Logger, cfg *appconfig.Config, opts RunOptions) error {
	// Interaktif: pilih file jika kosong
	if strings.TrimSpace(opts.FilePath) == "" {
		configuredDir := ""
		if cfg != nil {
			configuredDir = strings.TrimSpace(cfg.Script.BundleOutputDir)
		}
		p, err := selectSFToolsFileInteractive(configuredDir)
		if err != nil {
			return err
		}
		opts.FilePath = p

		// Interaktif: jika user belum provide args via CLI (-- ...), tanyakan args opsional.
		if len(opts.Args) == 0 {
			line, err := prompt.AskText("Masukkan args untuk script (opsional, pisahkan dengan spasi; kosongkan jika tidak ada)")
			if err != nil {
				return err
			}
			line = strings.TrimSpace(line)
			if line != "" {
				opts.Args = strings.Fields(line)
			}
		}
	} else {
		// Normalize path jika user provide via flag
		configuredDir := ""
		if cfg != nil {
			configuredDir = strings.TrimSpace(cfg.Script.BundleOutputDir)
		}
		opts.FilePath = normalizeSFToolsFlagPath(opts.FilePath, configuredDir)
	}

	// Core operation: run bundle
	return RunBundle(opts)
}

// normalizeSFToolsFlagPath normalizes file path dengan auto-append .sftools dan config dir resolution.
func normalizeSFToolsFlagPath(fileArg string, configuredDir string) string {
	p := strings.TrimSpace(fileArg)
	if p == "" {
		return p
	}

	// If user only provides a name (no path separator), resolve against configured bundle dir.
	hasSep := strings.Contains(p, "/") || strings.Contains(p, "\\")
	if !hasSep && strings.TrimSpace(configuredDir) != "" {
		p = filepath.Join(filepath.Clean(configuredDir), p)
	}

	// Auto-append .sftools if extension is missing.
	if strings.TrimSpace(filepath.Ext(p)) == "" {
		p = p + ".sftools"
	}
	return p
}

// selectSFToolsFileInteractive prompts user untuk memilih .sftools file dengan UX yang baik.
func selectSFToolsFileInteractive(configuredDir string) (string, error) {
	// Default UX: tampilkan daftar file dari configured dir jika ada.
	if configuredDir != "" {
		items, err := listSFToolsFiles(configuredDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
			return prompt.SelectFile(".", "Cari file .sftools yang akan dijalankan", []string{".sftools"})
		}
		if len(items) > 0 {
			const browseLabel = "🔎 Cari manual (browse)"
			const inputLabel = "⌨️  Input path manual"

			options := append([]string{}, items...)
			options = append(options, browseLabel, inputLabel)

			selected, _, err := prompt.SelectOne(fmt.Sprintf("Pilih file .sftools (default dari config: %s)", configuredDir), options, 0)
			if err != nil {
				return "", err
			}
			switch selected {
			case browseLabel:
				return prompt.SelectFile(configuredDir, "Cari file .sftools yang akan dijalankan", []string{".sftools"})
			case inputLabel:
				for {
					p, err := prompt.AskText("Masukkan path file .sftools")
					if err != nil {
						return "", err
					}
					p = strings.TrimSpace(p)
					if p == "" {
						continue
					}
					p = normalizeSFToolsFlagPath(p, configuredDir)
					if _, err := os.Stat(p); err != nil {
						fmt.Printf("File tidak ditemukan: %s\n", p)
						continue
					}
					return p, nil
				}
			default:
				return selected, nil
			}
		}
	}

	// Fallback: browse dari current dir (tetap bisa ketik path bebas).
	return prompt.SelectFile(".", "Pilih file .sftools yang akan dijalankan", []string{".sftools"})
}

// listSFToolsFiles lists all .sftools files in directory.
func listSFToolsFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca directory bundle_output_dir (%s): %w", dir, err)
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.EqualFold(filepath.Ext(name), ".sftools") {
			files = append(files, filepath.Join(dir, name))
		}
	}
	sort.Strings(files)
	return files, nil
}
