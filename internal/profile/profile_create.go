package profile

import (
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/ui"
	"sfDBTools/pkg/validation"
)

func (s *Service) CreateProfile() error {
	// ui.Headers("Pembuatan Profil Baru")
	ui.Headers("Pembuatan Profil Baru")
	s.Log.Info("Memulai proses pembuatan profil baru...")

	// Loop untuk menangani retry jika koneksi gagal
	for {
		if s.ProfileCreate.Interactive {
			s.Log.Info("Mode interaktif diaktifkan. Meminta input dari pengguna...")
			if err := s.runWizard("create"); err != nil {
				return err
			}
		} else {
			s.Log.Info("Mode non-interaktif diaktifkan. Memeriksa input dari parameter flags...")

			// validasi input non-interaktif
			s.Log.Info("Memvalidasi parameter yang diberikan...")
			if err := ValidateProfileInfo(s.ProfileInfo); err != nil {
				s.Log.Errorf("Validasi parameter gagal: %v", err)
				return err
			}
			s.Log.Info("Validasi parameter berhasil.")
			// Tampilkan data yang diberikan ke table
			s.DisplayProfileDetails()
		}

		// 2. Lakukan validasi keunikan nama
		err := s.CheckConfigurationNameUnique("create")
		if err != nil {
			// 3. JIKA TIDAK UNIK: Tampilkan error dan loop akan berulang
			ui.PrintError(err.Error())
			return err
		}

		// Simpan profil ke file atau database sesuai kebutuhan
		if err := s.SaveProfile("Create"); err != nil {
			// Cek apakah error adalah permintaan retry karena koneksi gagal
			if err == validation.ErrConnectionFailedRetry {
				// Tanya user apakah ingin mengulang input
				retryInput, askErr := input.AskYesNo("Apakah Anda ingin mengulang input konfigurasi?", true)
				if askErr != nil {
					return validation.HandleInputError(askErr)
				}

				if retryInput {
					ui.PrintWarning("Mengulang proses pembuatan profil...")
					continue // Ulangi dari awal
				} else {
					ui.PrintInfo("Proses pembuatan profil dibatalkan.")
					return validation.ErrUserCancelled
				}
			}

			// Error lainnya
			return err
		}

		// Berhasil, keluar dari loop
		break
	}

	s.Log.Info("Profil baru berhasil dibuat.")
	return nil
}
