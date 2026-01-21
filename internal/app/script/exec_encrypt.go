// File : internal/app/script/exec_encrypt.go
// Deskripsi : Execute layer untuk encrypt bundle (dengan interactive prompts)
// Author : Hadiyatna Muflihun
// Tanggal : 21 Januari 2026
// Last Modified : 21 Januari 2026

package script

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	appconfig "sfdbtools/internal/services/config"
	applog "sfdbtools/internal/services/log"
	"sfdbtools/internal/ui/prompt"
	"strings"
)

// ExecuteEncryptBundle handles encryption workflow dengan interactive prompts.
// Semua business logic dan user interaction ada di sini (BUKAN di cmd/).
func ExecuteEncryptBundle(logger applog.Logger, cfg *appconfig.Config, opts EncryptOptions) error {
	// Interaktif: pilih mode jika kosong
	if opts.Mode == "" {
		selectedMode, _, err := prompt.SelectOne("Pilih mode encrypt", []string{"bundle", "single"}, -1)
		if err != nil {
			return err
		}
		opts.Mode = selectedMode
	}

	// Interaktif: pilih file jika kosong
	if strings.TrimSpace(opts.FilePath) == "" {
		p, err := prompt.SelectFile(".", "Pilih entrypoint script (.sh) yang akan dienkripsi", []string{".sh"})
		if err != nil {
			return err
		}
		opts.FilePath = p
	}

	sourceDir := filepath.Clean(filepath.Dir(opts.FilePath))
	entryBase := filepath.Base(opts.FilePath)
	entryExt := strings.ToLower(filepath.Ext(entryBase))
	baseName := entryBase
	if entryExt != "" {
		baseName = strings.TrimSuffix(entryBase, entryExt)
	}
	defaultOutputName := baseName

	mode := strings.ToLower(strings.TrimSpace(opts.Mode))
	if mode == "" {
		mode = "bundle"
	}

	// Tentukan default output directory
	outDir := sourceDir
	outDirFromConfig := false
	if cfg != nil {
		cfgOutDir := strings.TrimSpace(cfg.Script.BundleOutputDir)
		if cfgOutDir != "" {
			outDir = filepath.Clean(cfgOutDir)
			outDirFromConfig = true
		}
	}

	// Interaktif: custom output filename jika belum di-set
	if opts.OutputPath == "" {
		outputName, err := prompt.AskText(
			"Nama output file (tanpa ekstensi, otomatis .sftools)",
			prompt.WithDefault(defaultOutputName),
			prompt.WithValidator(func(ans interface{}) error {
				s, _ := ans.(string)
				s = strings.TrimSpace(s)
				if s == "" {
					return fmt.Errorf("nama file tidak boleh kosong")
				}
				if strings.Contains(s, string(os.PathSeparator)) || strings.Contains(s, "/") || strings.Contains(s, "\\") {
					return fmt.Errorf("nama file tidak boleh mengandung path")
				}
				return nil
			}),
		)
		if err != nil {
			return err
		}
		outputName = strings.TrimSpace(outputName)
		outputFilename := outputName
		if !strings.HasSuffix(strings.ToLower(outputFilename), ".sftools") {
			outputFilename = outputFilename + ".sftools"
		}
		opts.OutputPath = filepath.Join(outDir, outputFilename)
	}

	// Interaktif: konfirmasi delete-source jika belum di-set
	deleteSource := opts.DeleteSource
	if !opts.DeleteSource {
		targetLabel := sourceDir
		if mode == "single" {
			targetLabel = opts.FilePath
		}
		wantDelete, err := prompt.Confirm(
			fmt.Sprintf("Hapus sumber setelah encrypt berhasil? (%s)", targetLabel),
			false,
		)
		if err != nil {
			return err
		}
		deleteSource = wantDelete
	}

	// Safety: jika delete-source dan output di folder yang sama, redirect ke current dir
	if deleteSource {
		if mode == "bundle" && !outDirFromConfig && filepath.Clean(filepath.Dir(opts.OutputPath)) == sourceDir {
			logger.Warn("Output berada di folder yang akan dihapus; output diarahkan ke folder kerja saat ini agar aman.")
			opts.OutputPath = filepath.Join(".", filepath.Base(opts.OutputPath))
		}

		confirm, err := prompt.AskText(
			"Ketik DELETE untuk konfirmasi penghapusan sumber",
			prompt.WithValidator(func(ans interface{}) error {
				s, _ := ans.(string)
				if strings.TrimSpace(s) != "DELETE" {
					return fmt.Errorf("konfirmasi tidak cocok")
				}
				return nil
			}),
		)
		if err != nil {
			return err
		}
		_ = confirm
	}

	// Core operation: encrypt bundle dengan context
	ctx := context.Background()
	if err := EncryptBundle(ctx, opts); err != nil {
		return err
	}

	// Post-operation: delete source if requested
	if deleteSource {
		if mode == "single" {
			if err := deleteSingleFile(opts.FilePath, logger); err != nil {
				return err
			}
		} else {
			if err := deleteSourceDir(sourceDir, opts.OutputPath, logger); err != nil {
				return err
			}
		}
	}

	logger.Infof("✓ Bundle terenkripsi dibuat dari: %s", opts.FilePath)
	if strings.TrimSpace(opts.OutputPath) != "" {
		logger.Infof("Output: %s", opts.OutputPath)
	} else {
		logger.Info("(Output: otomatis .sftools di folder yang sama)")
	}

	return nil
}

// deleteSingleFile menghapus satu file dengan safety checks.
func deleteSingleFile(filePath string, logger applog.Logger) error {
	absEntry, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("gagal resolve absolute path sumber: %w", err)
	}
	absEntry = filepath.Clean(absEntry)
	if absEntry == "/" || absEntry == "." || absEntry == "" {
		return fmt.Errorf("refuse delete: sumber tidak aman")
	}

	// P0 #3: Use Lstat untuk prevent TOCTOU race condition
	// Lstat tidak follow symlinks, jadi detect symlink attacks
	info, err := os.Lstat(absEntry)
	if err != nil {
		return fmt.Errorf("gagal membaca file info: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("refuse delete: file adalah symlink (security risk)")
	}

	if err := os.Remove(absEntry); err != nil {
		return fmt.Errorf("gagal menghapus sumber: %w", err)
	}
	logger.Infof("✓ Sumber dihapus: %s", absEntry)
	return nil
}

// deleteSourceDir menghapus direktori sumber dengan safety checks.
func deleteSourceDir(sourceDir string, outputPath string, logger applog.Logger) error {
	absSource, err := filepath.Abs(sourceDir)
	if err != nil {
		return fmt.Errorf("gagal resolve absolute path sumber: %w", err)
	}
	absSource = filepath.Clean(absSource)
	if absSource == "/" || absSource == "." || absSource == "" {
		return fmt.Errorf("refuse delete: sumber tidak aman")
	}

	// P0 #3: Use Lstat untuk prevent TOCTOU race condition
	info, err := os.Lstat(absSource)
	if err != nil {
		return fmt.Errorf("gagal membaca directory info: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("refuse delete: directory adalah symlink (security risk)")
	}
	if !info.IsDir() {
		return fmt.Errorf("refuse delete: bukan directory")
	}

	absOut, _ := filepath.Abs(outputPath)
	absOut = filepath.Clean(absOut)
	if absOut == absSource || strings.HasPrefix(absOut, absSource+string(os.PathSeparator)) {
		return fmt.Errorf("refuse delete: output berada di dalam folder sumber")
	}

	if err := os.RemoveAll(absSource); err != nil {
		return fmt.Errorf("gagal menghapus sumber: %w", err)
	}
	logger.Infof("✓ Sumber dihapus: %s", absSource)
	return nil
}
