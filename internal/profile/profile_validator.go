package profile

import (
	"fmt"
	"path/filepath"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/fsops"
	"sfDBTools/pkg/helper"
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/ui"
	"sfDBTools/pkg/validation"

	"github.com/AlecAivazis/survey/v2"
)

func (s *Service) CheckConfigurationNameUnique(mode string) error {
	// Implementasi pengecekan dengan absolute path untuk konsistensi
	// Normalisasi nama yang terlibat untuk memastikan konsistensi
	s.ProfileInfo.Name = helper.TrimProfileSuffix(s.ProfileInfo.Name)
	switch mode {
	case "create":
		abs := s.filePathInConfigDir(s.ProfileInfo.Name)
		exists := fsops.Exists(abs)
		if exists {
			return fmt.Errorf("nama konfigurasi '%s' sudah ada. Silakan pilih nama lain", s.ProfileInfo.Name)
		}
		return nil
	case "edit":
		// Untuk mode edit kita punya dua skenario:
		// - Jika user tidak merubah nama (ConfigName == OriginalConfigName): pastikan file lama ada
		// - Jika user merubah nama: pastikan target baru TIDAK ada (agar tidak menimpa)
		// Catatan: Jika --file absolute dipakai, gunakan direktori file tersebut sebagai base.
		original := helper.TrimProfileSuffix(s.OriginalProfileName)
		newName := helper.TrimProfileSuffix(s.ProfileInfo.Name)

		// Tentukan baseDir: prioritas ke absolute FilePath (saat edit), fallback ke config dir
		baseDir := s.Config.ConfigDir.DatabaseProfile
		if s.ProfileInfo.Path != "" && filepath.IsAbs(s.ProfileInfo.Path) {
			baseDir = filepath.Dir(s.ProfileInfo.Path)
		}

		if original == "" {
			// Tidak ada nama asli yang direkam -> fallback: cek keberadaan target
			targetAbs := filepath.Join(baseDir, validation.ProfileExt(newName))
			if !fsops.Exists(targetAbs) {
				return fmt.Errorf("file konfigurasi '%s' tidak ditemukan. Silakan pilih nama lain", newName)
			}
			return nil
		}

		// Jika nama tidak berubah, pastikan file lama tetap ada
		if original == newName {
			origAbs := filepath.Join(baseDir, validation.ProfileExt(original))
			if !fsops.Exists(origAbs) {
				return fmt.Errorf("file konfigurasi asli '%s' tidak ditemukan", original)
			}
			return nil
		}

		// Nama berubah: pastikan target baru tidak ada
		newAbs := filepath.Join(baseDir, validation.ProfileExt(newName))
		if fsops.Exists(newAbs) {
			return fmt.Errorf("nama konfigurasi tujuan '%s' sudah ada. Silakan pilih nama lain", newName)
		}
		// juga pastikan file original ada (agar dapat dihapus setelah rename)
		origAbs := filepath.Join(baseDir, validation.ProfileExt(original))
		if !fsops.Exists(origAbs) {
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

	// Validasi nama profil
	if p.Name == "" {
		return fmt.Errorf("nama profil tidak boleh kosong")
	}

	// Validasi informasi database
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

	// Validasi host
	if db.Host == "" {
		return fmt.Errorf("host database tidak boleh kosong")
	}

	// Validasi port
	if db.Port <= 0 || db.Port > 65535 {
		return fmt.Errorf("port database tidak valid: %d", db.Port)
	}

	// Validasi user
	if db.User == "" {
		return fmt.Errorf("user database tidak boleh kosong")
	}

	// Validasi password
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
