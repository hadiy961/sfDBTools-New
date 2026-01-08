// File : internal/restore/setup_shared_grants.go
// Deskripsi : Helper pemilihan file user grants untuk restore
// Author : Hadiyatna Muflihun
// Tanggal : 30 Desember 2025
// Last Modified : 5 Januari 2026

package restore

import (
	"fmt"
	"path/filepath"
	"strings"

	backupfile "sfdbtools/internal/app/backup/helpers/file"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/ui/print"
	"sfdbtools/internal/ui/prompt"
)

func (s *Service) resolveGrantsFile(skipGrants *bool, grantsFile *string, backupFile string, allowInteractive bool, stopOnError bool) error {
	if skipGrants != nil && *skipGrants {
		s.Log.Info("Skip restore user grants (--skip-grants)")
		return nil
	}

	if trimWhitespace(backupFile) == "" {
		s.Log.Info("Grants file: source file kosong, skip pencarian grants")
		return nil
	}

	s.Log.Infof("Mencari file user grants untuk source: %s", filepath.Base(backupFile))

	if *grantsFile != "" {
		if err := validateGrantsFilePath(*grantsFile, stopOnError); err != nil {
			if !allowInteractive {
				s.Log.Warnf("Mode non-interaktif: skip restore user grants (%v)", err)
				return nil
			}
			print.PrintWarning(fmt.Sprintf("⚠️  %v", err))
			*grantsFile = ""
		} else {
			s.Log.Infof("File grants (user-specified): %s", *grantsFile)
			return nil
		}
	}

	if autoGrantsFile := backupfile.AutoDetectGrantsFile(backupFile); autoGrantsFile != "" {
		s.Log.Infof("✓ Grants file ditemukan: %s", filepath.Base(autoGrantsFile))

		if !allowInteractive {
			*grantsFile = autoGrantsFile
			return nil
		}

		if confirmed, _ := s.confirmGrantsFileSelection(autoGrantsFile, skipGrants, grantsFile); confirmed {
			return nil
		}
	}

	if !allowInteractive {
		s.Log.Info("Mode non-interaktif: user grants tidak ditemukan, skip restore user grants")
		return nil
	}

	return s.selectGrantsFileInteractive(skipGrants, grantsFile, filepath.Dir(backupFile))
}

func (s *Service) confirmGrantsFileSelection(autoGrantsFile string, skipGrants *bool, grantsFile *string) (bool, error) {
	options := []string{
		fmt.Sprintf("✅ [ Pakai grants file terdeteksi: %s ]", filepath.Base(autoGrantsFile)),
		"📁 [ Browse / pilih file grants lain ]",
		"⏭️  [ Skip restore user grants ]",
	}
	selected, _, err := prompt.SelectOne("Grants file ditemukan. Gunakan file ini atau pilih yang lain?", options, 0)
	if err == nil && selected == options[0] {
		*grantsFile = autoGrantsFile
		s.Log.Infof("Grants file dipakai (auto-detect): %s", filepath.Base(autoGrantsFile))
		return true, nil
	}
	if err == nil && selected == options[2] {
		if skipGrants != nil {
			*skipGrants = true
		}
		*grantsFile = ""
		s.Log.Info("User grants tidak akan di-restore (user memilih skip)")
		return true, nil
	}
	return false, nil
}

func (s *Service) selectGrantsFileInteractive(skipGrants *bool, grantsFile *string, backupDir string) error {
	s.Log.Infof("Mencari file user grants di folder: %s", backupDir)
	matches, err := filepath.Glob(filepath.Join(backupDir, "*"+consts.UsersSQLSuffix))
	if err == nil && len(matches) > 0 {
		s.Log.Infof("Ditemukan %d file user grants di folder: %s", len(matches), backupDir)
		print.PrintInfo(fmt.Sprintf("📁 Ditemukan %d file user grants di folder yang sama", len(matches)))

		options := []string{"⏭️  [ Skip restore user grants ]", "📁 [ Browse file grants secara manual ]"}
		options = append(options, s.toBaseNames(matches)...)

		selected, _, err := prompt.SelectOne("Pilih file user grants untuk di-restore", options, 0)
		if err != nil {
			s.Log.Warnf("Gagal memilih file grants: %v, skip restore grants", err)
			return nil
		}

		if selected == "⏭️  [ Skip restore user grants ]" {
			if skipGrants != nil {
				*skipGrants = true
			}
			s.Log.Info("User grants tidak akan di-restore (user memilih skip)")
			return nil
		}

		if selected != "📁 [ Browse file grants secara manual ]" {
			for _, match := range matches {
				if filepath.Base(match) == selected {
					*grantsFile = match
					s.Log.Infof("File grants dipilih: %s", match)
					return nil
				}
			}
		}
	}

	return s.browseGrantsFileManually(skipGrants, grantsFile, backupDir)
}

func (s *Service) browseGrantsFileManually(skipGrants *bool, grantsFile *string, backupDir string) error {
	s.Log.Infof("Tidak ada file user grants terdeteksi di folder: %s", backupDir)
	print.PrintInfo("💡 File user grants tidak ditemukan atau Anda ingin pilih file lain")

	confirmed, err := prompt.PromptConfirm("Apakah Anda ingin memilih file user grants secara manual?")
	if err != nil || !confirmed {
		if skipGrants != nil {
			*skipGrants = true
		}
		s.Log.Info("Skip restore user grants (tidak ada file grants)")
		return nil
	}

	selectedFile, err := prompt.SelectFile(backupDir, "Masukkan path directory atau file user grants", []string{consts.ExtSQL})
	if err != nil {
		s.Log.Warnf("Gagal memilih file grants: %v, skip restore grants", err)
		return nil
	}

	*grantsFile = selectedFile
	s.Log.Infof("File grants dipilih secara manual: %s", selectedFile)
	return nil
}

func (s *Service) toBaseNames(paths []string) []string {
	result := make([]string, len(paths))
	for i, p := range paths {
		result[i] = filepath.Base(p)
	}
	return result
}

func trimWhitespace(val string) string {
	return strings.TrimSpace(val)
}
