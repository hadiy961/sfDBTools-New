// File : internal/app/script/exec_info.go
// Deskripsi : Execute layer untuk bundle info (dengan interactive prompts)
// Author : Hadiyatna Muflihun
// Tanggal : 21 Januari 2026
// Last Modified : 21 Januari 2026

package script

import (
	"fmt"
	appconfig "sfdbtools/internal/services/config"
	applog "sfdbtools/internal/services/log"
	"strings"
)

// ExecuteGetBundleInfo handles info retrieval workflow dengan interactive prompts.
func ExecuteGetBundleInfo(logger applog.Logger, cfg *appconfig.Config, opts InfoOptions) error {
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

	// Core operation: get bundle info
	info, err := GetBundleInfo(opts)
	if err != nil {
		return err
	}

	// Format output (user-friendly, bukan JSON)
	fmt.Printf("File      : %s\n", opts.FilePath)
	fmt.Printf("Version   : %d\n", info.Version)
	fmt.Printf("Mode      : %s\n", info.Mode)
	fmt.Printf("Entrypoint: %s\n", info.Entrypoint)
	if strings.TrimSpace(info.RootDir) != "" {
		fmt.Printf("Folder    : %s\n", info.RootDir)
	}
	fmt.Printf("CreatedAt : %s\n", info.CreatedAt)
	fmt.Printf("Files     : %d\n", info.FileCount)

	if info.Mode == "bundle" {
		fmt.Println("Scripts   :")
		if len(info.Scripts) == 0 {
			fmt.Println("  (tidak ada file .sh ditemukan)")
		} else {
			for _, s := range info.Scripts {
				if strings.TrimSpace(info.RootDir) != "" {
					fmt.Printf("  - %s/%s\n", info.RootDir, s)
				} else {
					fmt.Printf("  - %s\n", s)
				}
			}
		}
	}

	return nil
}
