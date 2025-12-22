// File : internal/profile/wizard.go
// Deskripsi : Interactive wizard logic for profile creation and editing
// Author : Hadiyatna Muflihun
// Tanggal : 17 Desember 2025
// Last Modified : 17 Desember 2025

package profile

import (
	"fmt"
	"os"
	"path/filepath"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/fsops"
	"sfDBTools/pkg/helper"
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/profilehelper"
	"sfDBTools/pkg/ui"
	"sfDBTools/pkg/validation"

	"github.com/AlecAivazis/survey/v2"
)

// runWizard orchestrates the interactive profile creation/editing process
func (s *Service) runWizard(mode string) error {
	s.Log.Info("Wizard profil dimulai...")

	for {
		// Edit Mode: Select Existing Config
		if mode == "edit" {
			if err := s.promptSelectExistingConfig(); err != nil {
				return err
			}
		}

		// 1. Config Name
		if err := s.promptDBConfigName(mode); err != nil {
			return err
		}

		// 2. Profile Details
		if err := s.promptProfileInfo(); err != nil {
			return err
		}

		if s.ProfileInfo.Name == "" {
			return fmt.Errorf("nama konfigurasi tidak boleh kosong")
		}

		// 3. Review
		s.DisplayProfileDetails()

		// 4. Confirm Save
		confirmSave, err := input.AskYesNo("Apakah Anda ingin menyimpan konfigurasi ini?", true)
		if err != nil {
			return validation.HandleInputError(err)
		}

		if confirmSave {
			break
		} else {
			confirmRetry, err := input.AskYesNo("Apakah Anda ingin mengulang proses?", false)
			if err != nil {
				return validation.HandleInputError(err)
			}
			if confirmRetry {
				ui.PrintWarning("Penyimpanan dibatalkan. Memulai ulang wizard...")
				continue
			} else {
				return validation.ErrUserCancelled
			}
		}
	}

	ui.PrintSuccess("Konfirmasi diterima. Mempersiapkan enkripsi dan penyimpanan...")

	// Get Encryption Key
	key, source, err := helper.ResolveEncryptionKey(s.ProfileInfo.EncryptionKey, consts.ENV_SOURCE_PROFILE_KEY)
	if err != nil {
		return fmt.Errorf("gagal mendapatkan password enkripsi: %w", err)
	}

	s.Log.WithField("Sumber Kunci", source).Debug("Password enkripsi berhasil didapatkan.")
	s.ProfileInfo.EncryptionKey = key
	return nil
}

// promptSelectExistingConfig interactive file selection logic
func (s *Service) promptSelectExistingConfig() error {
	info, err := profilehelper.ResolveAndLoadProfile(profilehelper.ProfileLoadOptions{
		ConfigDir:         s.Config.ConfigDir.DatabaseProfile,
		ProfilePath:       "",
		AllowInteractive:  true,
		InteractivePrompt: "Select Existing Configuration File",
		RequireProfile:    true,
	})
	if err != nil {
		return err
	}

	s.ProfileInfo = info
	s.OriginalProfileName = info.Name
	s.OriginalProfileInfo = &types.ProfileInfo{
		Path:         info.Path,
		Name:         info.Name,
		DBInfo:       info.DBInfo,
		Size:         info.Size,
		LastModified: info.LastModified,
	}

	s.Log.Debug("Memuat konfigurasi dari: " + info.Path + " Name: " + info.Name)
	return nil
}

// promptDBConfigName asks for and validates configuration name
func (s *Service) promptDBConfigName(mode string) error {
	ui.PrintSubHeader("Please provide the configuration name:")

	for {
		nameValidator := input.ComposeValidators(survey.Required, input.ValidateFilename)
		configName, err := input.AskString("Configuration Name", s.ProfileInfo.Name, nameValidator)
		if err != nil {
			return validation.HandleInputError(err)
		}

		s.ProfileInfo.Name = helper.TrimProfileSuffix(configName)

		if err = s.CheckConfigurationNameUnique(mode); err != nil {
			ui.PrintError(err.Error())
			continue
		}
		break
	}

	ui.PrintInfo("Konfigurasi akan disimpan sebagai : " + buildFileName(s.ProfileInfo.Name))
	return nil
}

// promptProfileInfo asks for host, port, user, and password
func (s *Service) promptProfileInfo() error {
	ui.PrintSubHeader("Please provide the following database profile details:")
	var err error

	s.ProfileInfo.DBInfo.Host, err = input.AskString("Database Host", s.ProfileInfo.DBInfo.Host, survey.Required)
	if err != nil {
		return validation.HandleInputError(err)
	}

	s.ProfileInfo.DBInfo.Port, err = input.AskInt("Database Port", s.ProfileInfo.DBInfo.Port, survey.Required)
	if err != nil {
		return validation.HandleInputError(err)
	}

	s.ProfileInfo.DBInfo.User, err = input.AskString("Database User", s.ProfileInfo.DBInfo.User, survey.Required)
	if err != nil {
		return validation.HandleInputError(err)
	}

	isEditFlow := s.OriginalProfileInfo != nil || s.OriginalProfileName != ""
	var existingPassword string
	if isEditFlow && s.ProfileInfo != nil {
		existingPassword = s.ProfileInfo.DBInfo.Password
	}

	envPassword := os.Getenv(consts.ENV_TARGET_DB_PASSWORD)
	if !isEditFlow && envPassword != "" {
		s.ProfileInfo.DBInfo.Password = envPassword
	} else if !isEditFlow {
		ui.PrintWarning("Environment variable SFDB_DB_PASSWORD tidak ditemukan atau kosong. Silakan atur SFDB_DB_PASSWORD atau ketik password.")
	}

	if isEditFlow && existingPassword != "" {
		ui.PrintInfo("💡 Password saat ini dimuat dari file. Tekan Enter untuk mempertahankan, atau ketik password baru untuk mengupdate.")
	}

	pw, err := input.AskPassword("Database Password", nil)
	if err != nil {
		return validation.HandleInputError(err)
	}

	if pw == "" {
		if isEditFlow {
			if existingPassword != "" {
				s.ProfileInfo.DBInfo.Password = existingPassword
			}
		} else {
			pw, err = input.AskPassword("Database Password", survey.Required)
			if err != nil {
				return validation.HandleInputError(err)
			}
			s.ProfileInfo.DBInfo.Password = pw
		}
	} else {
		s.ProfileInfo.DBInfo.Password = pw
	}

	return nil
}

// CheckConfigurationNameUnique validates name uniqueness
func (s *Service) CheckConfigurationNameUnique(mode string) error {
	s.ProfileInfo.Name = helper.TrimProfileSuffix(s.ProfileInfo.Name)
	switch mode {
	case "create":
		abs := s.filePathInConfigDir(s.ProfileInfo.Name)
		if fsops.PathExists(abs) {
			return fmt.Errorf("nama konfigurasi '%s' sudah ada. Silakan pilih nama lain", s.ProfileInfo.Name)
		}
		return nil
	case "edit":
		original := helper.TrimProfileSuffix(s.OriginalProfileName)
		newName := helper.TrimProfileSuffix(s.ProfileInfo.Name)

		baseDir := s.Config.ConfigDir.DatabaseProfile
		if s.ProfileInfo.Path != "" && filepath.IsAbs(s.ProfileInfo.Path) {
			baseDir = filepath.Dir(s.ProfileInfo.Path)
		}

		if original == "" {
			targetAbs := filepath.Join(baseDir, validation.ProfileExt(newName))
			if !fsops.PathExists(targetAbs) {
				return fmt.Errorf("file konfigurasi '%s' tidak ditemukan. Silakan pilih nama lain", newName)
			}
			return nil
		}

		if original == newName {
			origAbs := filepath.Join(baseDir, validation.ProfileExt(original))
			if !fsops.PathExists(origAbs) {
				return fmt.Errorf("file konfigurasi asli '%s' tidak ditemukan", original)
			}
			return nil
		}

		newAbs := filepath.Join(baseDir, validation.ProfileExt(newName))
		if fsops.PathExists(newAbs) {
			return fmt.Errorf("nama konfigurasi tujuan '%s' sudah ada. Silakan pilih nama lain", newName)
		}
		origAbs := filepath.Join(baseDir, validation.ProfileExt(original))
		if !fsops.PathExists(origAbs) {
			return fmt.Errorf("file konfigurasi asli '%s' tidak ditemukan", original)
		}
		return nil
	}
	return nil
}

// ValidateProfileInfo validates the given ProfileInfo struct
func ValidateProfileInfo(p *types.ProfileInfo) error {
	if p == nil {
		return fmt.Errorf("informasi profil tidak boleh kosong")
	}
	if p.Name == "" {
		return fmt.Errorf("nama profil tidak boleh kosong")
	}
	if err := ValidateDBInfo(&p.DBInfo); err != nil {
		return fmt.Errorf("validasi informasi database gagal: %w", err)
	}
	return nil
}

// ValidateDBInfo validates the given DBInfo struct
func ValidateDBInfo(db *types.DBInfo) error {
	if db == nil {
		return fmt.Errorf("informasi database tidak boleh kosong")
	}
	if db.Host == "" {
		return fmt.Errorf("host database tidak boleh kosong")
	}
	if db.Port <= 0 || db.Port > 65535 {
		return fmt.Errorf("port database tidak valid: %d", db.Port)
	}
	if db.User == "" {
		return fmt.Errorf("user database tidak boleh kosong")
	}
	if db.Password == "" {
		ui.PrintWarning("Password database tidak diberikan lewat flag; meminta input password...")
		pw, err := input.AskPassword("Password untuk user ("+db.User+") : ", survey.Required)
		if err != nil {
			return validation.HandleInputError(err)
		}
		db.Password = pw
	}
	return nil
}
