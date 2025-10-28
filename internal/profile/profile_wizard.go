package profile

import (
	"fmt"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/helper"
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/ui"
	"sfDBTools/pkg/validation"
)

func (s *Service) runWizard(mode string) error {
	s.Log.Info("Wizard profil dimulai...")
	// Gunakan loop for tak terbatas untuk mengulang proses jika diperlukan
	for {
		// Jika mode edit, tampilkan daftar file yang ada dan biarkan user memilih
		if mode == "edit" {
			if err := s.promptSelectExistingConfig(); err != nil {
				return err
			}
		}

		// 1. Prompt untuk nama konfigurasi
		if err := s.promptDBConfigName(mode); err != nil {
			return err
		}

		// 2. Prompt untuk detail koneksi database
		if err := s.promptProfileInfo(); err != nil {
			return err
		}

		// 3. Tampilkan ringkasan konfigurasi
		ConfigName := s.ProfileInfo.Name
		if ConfigName == "" {
			return fmt.Errorf("nama konfigurasi tidak boleh kosong")
		}

		// 4. Tampilkan detail konfigurasi
		s.DisplayProfileDetails()

		// 5. Konfirmasi penyimpanan konfigurasi
		var confirmSave bool
		confirmSave, err := input.AskYesNo("Apakah Anda ingin menyimpan konfigurasi ini?", true)
		if err != nil {
			return validation.HandleInputError(err)
		}

		// Proses berdasarkan konfirmasi
		if confirmSave {
			break // Keluar dari loop dan lanjutkan proses penyimpanan
		} else {
			var confirmRetry bool
			// Tanyakan apakah user ingin mengulang atau keluar
			confirmRetry, err = input.AskYesNo("Apakah Anda ingin mengulang proses?", false)
			if err != nil {
				return validation.HandleInputError(err)
			}
			if confirmRetry {
				ui.PrintWarning("Penyimpanan konfigurasi dibatalkan oleh pengguna. Memulai ulang wizard...")
				continue // Mulai ulang wizard
			} else {
				return validation.ErrUserCancelled
			}
		}
	}

	ui.PrintSuccess("Konfirmasi diterima. Mempersiapkan enkripsi dan penyimpanan...")

	// 1. Dapatkan password enkripsi dari pengguna (atau env var)
	key, source, err := helper.ResolveEncryptionKey(s.ProfileInfo.EncryptionKey, consts.ENV_SOURCE_PROFILE_KEY)
	if err != nil {
		return fmt.Errorf("gagal mendapatkan password enkripsi: %w", err)
	}

	s.Log.WithField("Sumber Kunci", source).Debug("Password enkripsi berhasil didapatkan.")
	s.ProfileInfo.EncryptionKey = key
	return nil
}
