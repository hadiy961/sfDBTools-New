// File : internal/app/script/exec_extract.go
// Deskripsi : Execute layer untuk extract bundle (dengan interactive prompts)
// Author : Hadiyatna Muflihun
// Tanggal : 21 Januari 2026
// Last Modified : 21 Januari 2026

package script

import (
	"context"
	"path/filepath"
	appconfig "sfdbtools/internal/services/config"
	applog "sfdbtools/internal/services/log"
	"sfdbtools/internal/ui/prompt"
	"strings"
)

// ExecuteExtractBundle handles extraction workflow dengan interactive prompts.
func ExecuteExtractBundle(logger applog.Logger, cfg *appconfig.Config, opts ExtractOptions) error {
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
	} else {
		// Normalize path jika user provide via flag
		configuredDir := ""
		if cfg != nil {
			configuredDir = strings.TrimSpace(cfg.Script.BundleOutputDir)
		}
		opts.FilePath = normalizeSFToolsFlagPath(opts.FilePath, configuredDir)
	}

	// Interaktif: tentukan output dir jika kosong
	if strings.TrimSpace(opts.OutDir) == "" {
		base := strings.TrimSuffix(filepath.Base(opts.FilePath), filepath.Ext(opts.FilePath))
		defaultOut := "./" + base + "_extracted"
		out, err := prompt.AskText("Output directory untuk hasil extract", prompt.WithDefault(defaultOut))
		if err != nil {
			return err
		}
		opts.OutDir = strings.TrimSpace(out)
	}

	// Core operation: extract bundle dengan context
	ctx := context.Background()
	if err := ExtractBundle(ctx, opts); err != nil {
		return err
	}

	absOut, _ := filepath.Abs(opts.OutDir)
	logger.Infof("✓ Extract selesai. Output: %s", absOut)
	return nil
}
