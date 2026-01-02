// File : internal/profile/executor.go
// Deskripsi : Core execution logic untuk profile CRUD operations
// Author : Hadiyatna Muflihun
// Tanggal : 16 Desember 2025
// Last Modified : 2 Januari 2026

package profile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sfDBTools/internal/types"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/encrypt"
	"sfDBTools/pkg/fsops"
	"sfDBTools/pkg/helper"
	profilehelper "sfDBTools/pkg/helper/profile"
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/ui"
	"sfDBTools/pkg/validation"
)

// CreateProfile menangani proses pembuatan profil baru
func (s *Service) CreateProfile() error {
	ui.Headers("Pembuatan Profil Baru")
	s.Log.Info("Memulai proses pembuatan profil baru...")

	for {
		if s.ProfileCreate.Interactive {
			s.Log.Info("Mode interaktif diaktifkan.")
			if err := s.runWizard("create"); err != nil {
				return err
			}
		} else {
			s.Log.Info("Mode non-interaktif diaktifkan.")
			s.Log.Info("Memvalidasi parameter...")
			if err := ValidateProfileInfo(s.ProfileInfo); err != nil {
				s.Log.Errorf("Validasi parameter gagal: %v", err)
				return err
			}
			s.Log.Info("Validasi parameter berhasil.")
			s.DisplayProfileDetails()
		}

		if err := s.CheckConfigurationNameUnique("create"); err != nil {
			ui.PrintError(err.Error())
			return err
		}

		if err := s.SaveProfile("Create"); err != nil {
			if err == validation.ErrConnectionFailedRetry {
				retryInput, askErr := input.AskYesNo("Apakah Anda ingin mengulang input konfigurasi?", true)
				if askErr != nil {
					return validation.HandleInputError(askErr)
				}
				if retryInput {
					ui.PrintWarning("Mengulang proses pembuatan profil...")
					continue
				} else {
					ui.PrintInfo("Proses pembuatan profil dibatalkan.")
					return validation.ErrUserCancelled
				}
			}
			return err
		}
		break
	}

	s.Log.Info("Profil baru berhasil dibuat.")
	return nil
}

// ShowProfile menampilkan detail profil yang ada
func (s *Service) ShowProfile() error {
	ui.Headers("Display User Profile Details")

	if s.ProfileShow == nil || strings.TrimSpace(s.ProfileShow.Path) == "" {
		var revealPassword bool
		if s.ProfileShow != nil {
			revealPassword = s.ProfileShow.RevealPassword
		}

		if err := s.promptSelectExistingConfig(); err != nil {
			return err
		}
		if s.ProfileShow == nil {
			s.ProfileShow = &types.ProfileShowOptions{} // Should not happen given NewService init logic but safe check
		}
		s.ProfileShow.Path = s.ProfileInfo.Path
		s.ProfileShow.RevealPassword = revealPassword
	} else {
		abs, name, err := helper.ResolveConfigPath(s.ProfileShow.Path)
		if err != nil {
			return err
		}
		if !fsops.PathExists(abs) {
			return fmt.Errorf("file konfigurasi tidak ditemukan: %s", abs)
		}
		s.ProfileInfo.Name = name
		if err := s.loadSnapshotFromPath(abs); err != nil {
			s.Log.Warn("Gagal memuat isi detail konfigurasi: " + err.Error())
		}
	}

	if s.OriginalProfileInfo == nil || s.OriginalProfileInfo.Path == "" {
		return fmt.Errorf("tidak ada snapshot konfigurasi untuk ditampilkan")
	}

	if !fsops.PathExists(s.OriginalProfileInfo.Path) {
		return fmt.Errorf("file konfigurasi tidak ditemukan: %s", s.OriginalProfileInfo.Path)
	}

	s.ProfileInfo.Path = s.OriginalProfileInfo.Path
	if s.OriginalProfileInfo != nil {
		s.ProfileInfo.DBInfo = s.OriginalProfileInfo.DBInfo
	}

	if c, err := profilehelper.ConnectWithProfile(s.ProfileInfo, consts.DefaultInitialDatabase); err != nil {
		// Tetap tampilkan detail profil walaupun koneksi gagal.
		ui.PrintWarning("Koneksi database gagal: " + err.Error())
	} else {
		c.Close()
	}

	s.DisplayProfileDetails()
	return nil
}

// EditProfile menangani proses pengeditan profil
func (s *Service) EditProfile() error {
	ui.Headers("Database Configuration Editing")

	for {
		if s.ProfileEdit != nil && s.ProfileEdit.Interactive {
			s.Log.Info("Mode interaktif diaktifkan.")
			if err := s.runWizard("edit"); err != nil {
				if err == validation.ErrUserCancelled {
					s.Log.Warn("Proses pengeditan konfigurasi dibatalkan oleh pengguna.")
					return validation.ErrUserCancelled
				}
				s.Log.Warn("Proses pengeditan konfigurasi gagal: " + err.Error())
				return err
			}
		} else {
			s.Log.Info("Mode non-interaktif.")
			if strings.TrimSpace(s.OriginalProfileName) == "" {
				return fmt.Errorf("flag --file wajib disertakan pada mode non-interaktif")
			}

			absPath, name, err := helper.ResolveConfigPath(s.OriginalProfileName)
			if err != nil {
				return err
			}
			if !fsops.PathExists(absPath) {
				ui.PrintWarning(fmt.Sprintf("File konfigurasi '%s' tidak ditemukan.", absPath))
				return fmt.Errorf("file konfigurasi tidak ditemukan: %s", absPath)
			}

			s.ProfileInfo.Name = name
			s.ProfileInfo.Path = absPath
			s.OriginalProfileName = name

			s.Log.Info("File ditemukan. Mencoba memuat konten...")
			if err := s.loadSnapshotFromPath(absPath); err != nil {
				ui.PrintWarning(fmt.Sprintf("Gagal memuat isi file '%s': %v. Lanjut dengan data minimum.", absPath, err))
			}
			s.Log.Info("File konfigurasi berhasil dimuat.")

			// Preserve existing values if not provided in flags
			if s.OriginalProfileInfo != nil {
				if s.ProfileInfo.DBInfo.Password == "" {
					s.ProfileInfo.DBInfo.Password = s.OriginalProfileInfo.DBInfo.Password
				}
				if strings.TrimSpace(s.ProfileInfo.DBInfo.Host) == "" {
					s.ProfileInfo.DBInfo.Host = s.OriginalProfileInfo.DBInfo.Host
				}
				if s.ProfileInfo.DBInfo.Port == 0 {
					s.ProfileInfo.DBInfo.Port = s.OriginalProfileInfo.DBInfo.Port
				}
				if strings.TrimSpace(s.ProfileInfo.DBInfo.User) == "" {
					s.ProfileInfo.DBInfo.User = s.OriginalProfileInfo.DBInfo.User
				}
				if strings.TrimSpace(s.ProfileInfo.Name) == "" {
					s.ProfileInfo.Name = s.OriginalProfileInfo.Name
				}
				// Preserve SSH settings jika user tidak mengisi flag ssh apapun
				if !s.ProfileInfo.SSHTunnel.Enabled && strings.TrimSpace(s.ProfileInfo.SSHTunnel.Host) == "" &&
					strings.TrimSpace(s.ProfileInfo.SSHTunnel.User) == "" && strings.TrimSpace(s.ProfileInfo.SSHTunnel.Password) == "" &&
					strings.TrimSpace(s.ProfileInfo.SSHTunnel.IdentityFile) == "" &&
					s.ProfileInfo.SSHTunnel.LocalPort == 0 {
					s.ProfileInfo.SSHTunnel = s.OriginalProfileInfo.SSHTunnel
				}
			}

			s.Log.Info("Memvalidasi parameter...")
			if s.ProfileEdit != nil && strings.TrimSpace(s.ProfileEdit.NewName) != "" {
				newName := helper.TrimProfileSuffix(strings.TrimSpace(s.ProfileEdit.NewName))
				if newName == "" {
					return fmt.Errorf("nama baru tidak boleh kosong")
				}
				if strings.Contains(newName, "/") || strings.Contains(newName, "\\") {
					return fmt.Errorf("nama baru tidak boleh berisi path")
				}
				s.ProfileInfo.Name = newName
			}

			if err := ValidateProfileInfo(s.ProfileInfo); err != nil {
				s.Log.Errorf("Validasi parameter gagal: %v", err)
				return err
			}
			s.DisplayProfileDetails()
		}

		if err := s.CheckConfigurationNameUnique("edit"); err != nil {
			ui.PrintError(err.Error())
			return err
		}

		if err := s.SaveProfile("Edit"); err != nil {
			if err == validation.ErrConnectionFailedRetry {
				retryInput, askErr := input.AskYesNo("Apakah Anda ingin mengulang input konfigurasi?", true)
				if askErr != nil {
					return validation.HandleInputError(askErr)
				}
				if retryInput {
					ui.PrintWarning("Mengulang proses pengeditan profil...")
					continue
				} else {
					ui.PrintInfo("Proses pengeditan profil dibatalkan.")
					return validation.ErrUserCancelled
				}
			}
			return err
		}
		break
	}
	s.Log.Info("Wizard interaktif selesai.")
	return nil
}

// PromptDeleteProfile menangani penghapusan profil
func (s *Service) PromptDeleteProfile() error {
	ui.Headers("Delete Database Configurations")

	// 1. Jika profile path disediakan via flag --profile (support multiple)
	if s.ProfileDelete != nil && len(s.ProfileDelete.Profiles) > 0 {
		var validPaths []string
		var displayNames []string

		// Validasi semua profil input
		for _, p := range s.ProfileDelete.Profiles {
			if p == "" {
				continue
			}

			absPath, name, err := helper.ResolveConfigPath(p)
			if err != nil {
				return err
			}

			if !fsops.PathExists(absPath) {
				return fmt.Errorf("file konfigurasi tidak ditemukan: %s (name: %s)", absPath, name)
			}
			validPaths = append(validPaths, absPath)
			displayNames = append(displayNames, fmt.Sprintf("%s (%s)", name, absPath))
		}

		if len(validPaths) == 0 {
			ui.PrintInfo("Tidak ada profil valid yang ditemukan untuk dihapus.")
			return nil
		}

		// Jika force flag aktif, hapus langsung
		if s.ProfileDelete.Force {
			for _, path := range validPaths {
				if err := fsops.RemoveFile(path); err != nil {
					s.Log.Errorf("Gagal menghapus file %s: %v", path, err)
					// Continue deleting others? Yes, usually force means best effort.
					// But returning error at end might be good.
				} else {
					s.Log.Info(fmt.Sprintf("Berhasil menghapus (force): %s", path))
					ui.PrintSuccess(fmt.Sprintf("Berhasil menghapus: %s", path))
				}
			}
			return nil
		}

		// Jika tidak force, minta konfirmasi untuk batch
		ui.PrintWarning("Akan menghapus profil berikut:")
		for _, d := range displayNames {
			ui.PrintWarning(" - " + d)
		}

		ok, err := input.AskYesNo(fmt.Sprintf("Anda yakin ingin menghapus %d profil ini?", len(validPaths)), false)
		if err != nil {
			return validation.HandleInputError(err)
		}
		if !ok {
			ui.PrintInfo("Penghapusan dibatalkan oleh pengguna.")
			return nil
		}

		for _, path := range validPaths {
			if err := fsops.RemoveFile(path); err != nil {
				s.Log.Errorf("Gagal menghapus file %s: %v", path, err)
				ui.PrintError(fmt.Sprintf("Gagal menghapus: %s (%v)", path, err))
			} else {
				s.Log.Info(fmt.Sprintf("Berhasil menghapus: %s", path))
				ui.PrintSuccess(fmt.Sprintf("Berhasil menghapus: %s", path))
			}
		}
		return nil
	}

	// 2. Interactive Selection (Existing Logic)
	configDir := s.Config.ConfigDir.DatabaseProfile
	files, err := fsops.ReadDirFiles(configDir)
	if err != nil {
		return fmt.Errorf("gagal membaca direktori konfigurasi: %w", err)
	}

	filtered := make([]string, 0, len(files))
	for _, f := range files {
		if validation.ProfileExt(f) == f {
			filtered = append(filtered, f)
		}
	}
	if len(filtered) == 0 {
		ui.PrintInfo("Tidak ada file konfigurasi untuk dihapus.")
		return nil
	}

	idxs, err := input.ShowMultiSelect("Pilih file konfigurasi yang akan dihapus:", filtered)
	if err != nil {
		return validation.HandleInputError(err)
	}

	selected := make([]string, 0, len(idxs))
	for _, i := range idxs {
		if i >= 1 && i <= len(filtered) {
			selected = append(selected, filepath.Join(configDir, filtered[i-1]))
		}
	}

	if len(selected) == 0 {
		ui.PrintInfo("Tidak ada file terpilih untuk dihapus.")
		return nil
	}

	// Cek force flag untuk multi-select (opsional, tapi konsisten)
	force := false
	if s.ProfileDelete != nil {
		force = s.ProfileDelete.Force
	}

	if !force {
		ok, err := input.AskYesNo(fmt.Sprintf("Anda yakin ingin menghapus %d file?", len(selected)), false)
		if err != nil {
			return validation.HandleInputError(err)
		}
		if !ok {
			ui.PrintInfo("Penghapusan dibatalkan oleh pengguna.")
			return nil
		}
	}

	for _, p := range selected {
		if err := fsops.RemoveFile(p); err != nil {
			s.Log.Error(fmt.Sprintf("Gagal menghapus file %s: %v", p, err))
			ui.PrintError(fmt.Sprintf("Gagal menghapus file %s: %v", p, err))
		} else {
			s.Log.Info(fmt.Sprintf("Berhasil menghapus: %s", p))
			ui.PrintSuccess(fmt.Sprintf("Berhasil menghapus: %s", p))
		}
	}

	return nil
}

// SaveProfile menyimpan profil ke file
func (s *Service) SaveProfile(mode string) error {
	s.Log.Info("Memulai proses penyimpanan file konfigurasi...")
	isInteractive := s.isInteractiveMode()

	var baseDir string
	var originalAbsPath string
	if mode == "Edit" && s.ProfileInfo.Path != "" && filepath.IsAbs(s.ProfileInfo.Path) {
		originalAbsPath = s.ProfileInfo.Path
		baseDir = filepath.Dir(s.ProfileInfo.Path)
	} else {
		baseDir = s.Config.ConfigDir.DatabaseProfile
	}

	if !fsops.DirExists(baseDir) {
		if err := fsops.CreateDirIfNotExist(baseDir); err != nil {
			return fmt.Errorf("gagal membuat direktori konfigurasi: %w", err)
		}
	}

	if c, err := profilehelper.ConnectWithProfile(s.ProfileInfo, consts.DefaultInitialDatabase); err != nil {
		if !isInteractive {
			return err
		}
		continueAnyway, askErr := input.AskYesNo("\nKoneksi database gagal. Apakah Anda tetap ingin menyimpan konfigurasi ini?", false)
		if askErr != nil {
			return validation.HandleInputError(askErr)
		}
		if !continueAnyway {
			return validation.ErrConnectionFailedRetry
		}
		ui.PrintWarning("⚠️  PERINGATAN: Menyimpan konfigurasi dengan koneksi database yang tidak valid.")
	} else {
		c.Close()
		s.Log.Info("Koneksi database valid.")
	}

	iniContent := s.formatConfigToINI()
	key, _, err := helper.ResolveEncryptionKey(s.ProfileInfo.EncryptionKey, consts.ENV_SOURCE_PROFILE_KEY)
	if err != nil {
		return fmt.Errorf("kunci enkripsi tidak tersedia: %w", err)
	}
	encryptedContent, err := encrypt.EncryptAES([]byte(iniContent), []byte(key))
	if err != nil {
		return fmt.Errorf("gagal mengenkripsi konten konfigurasi: %w", err)
	}

	if mode == "Edit" && s.ProfileEdit != nil && strings.TrimSpace(s.ProfileEdit.NewName) != "" {
		s.ProfileInfo.Name = s.ProfileEdit.NewName
	}

	s.ProfileInfo.Name = helper.TrimProfileSuffix(s.ProfileInfo.Name)
	newFileName := buildFileName(s.ProfileInfo.Name)
	newFilePath := filepath.Join(baseDir, newFileName)

	if mode == "Edit" && s.OriginalProfileName != "" && s.OriginalProfileName != s.ProfileInfo.Name {
		if err := fsops.WriteFile(newFilePath, encryptedContent); err != nil {
			return fmt.Errorf("gagal menyimpan file konfigurasi baru: %w", err)
		}
		oldFilePath := originalAbsPath
		if oldFilePath == "" && s.OriginalProfileInfo != nil && s.OriginalProfileInfo.Path != "" {
			oldFilePath = s.OriginalProfileInfo.Path
		}
		if oldFilePath == "" {
			oldFilePath = filepath.Join(baseDir, buildFileName(s.OriginalProfileName))
		}
		if err := os.Remove(oldFilePath); err != nil {
			ui.PrintWarning(fmt.Sprintf("Berhasil menyimpan '%s' tetapi gagal menghapus file lama '%s': %v", newFileName, oldFilePath, err))
		}
		ui.PrintSuccess(fmt.Sprintf("File konfigurasi berhasil disimpan sebagai '%s' (rename dari '%s').", newFileName, buildFileName(s.OriginalProfileName)))
		ui.PrintInfo("File konfigurasi tersimpan di : " + newFilePath)
		return nil
	}

	if err := fsops.WriteFile(newFilePath, encryptedContent); err != nil {
		return fmt.Errorf("gagal menyimpan file konfigurasi: %w", err)
	}

	ui.PrintSuccess(fmt.Sprintf("File konfigurasi '%s' berhasil disimpan dengan aman.", newFileName))
	ui.PrintInfo("File konfigurasi tersimpan di : " + newFilePath)
	return nil
}
