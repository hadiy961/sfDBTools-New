package profile

import (
	"fmt"
	"sfDBTools/pkg/fsops"
	"sfDBTools/pkg/helper"
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/ui"
	"sfDBTools/pkg/validation"
	"strings"
)

// EditDatabaseConfig menjalankan logika pengeditan file konfigurasi.
func (s *Service) EditProfile() error {
	ui.Headers("Database Configuration Editing")

	// Loop untuk menangani retry jika koneksi gagal
	for {
		if s.ProfileEdit != nil && s.ProfileEdit.Interactive {
			s.Log.Info("Mode interaktif diaktifkan. Memulai wizard konfigurasi...")
			if err := s.runWizard("edit"); err != nil {
				if err == validation.ErrUserCancelled {
					s.Log.Warn("Proses pengeditan konfigurasi dibatalkan oleh pengguna.")
					return validation.ErrUserCancelled
				}
				s.Log.Warn("Proses pengeditan konfigurasi gagal: " + err.Error())
				return err
			}
		} else {
			s.Log.Info("Mode non-interaktif. Menggunakan flag yang diberikan.")

			// Non-interactive edit harus memiliki flag --file (tersimpan di OriginalProfileName saat NewService)
			if strings.TrimSpace(s.OriginalProfileName) == "" {
				return fmt.Errorf("flag --file wajib disertakan pada mode non-interaktif (contoh: --file /path/to/name atau --file name)")
			}

			// Resolve path dan nama
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

			// Preserve password if flags didn't provide one
			if s.OriginalProfileInfo != nil && s.ProfileInfo.DBInfo.Password == "" {
				s.ProfileInfo.DBInfo.Password = s.OriginalProfileInfo.DBInfo.Password
			}

			// Preserve other DB fields (host, port, user, name) when flags didn't provide them
			if s.OriginalProfileInfo != nil {
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
			}

			// validasi input non-interaktif
			s.Log.Info("Memvalidasi parameter yang diberikan...")

			// Jika user menyediakan --new-name, validasi format sederhana (tidak boleh berisi path)
			if s.ProfileEdit != nil && strings.TrimSpace(s.ProfileEdit.NewName) != "" {
				newName := helper.TrimProfileSuffix(strings.TrimSpace(s.ProfileEdit.NewName))
				if newName == "" {
					s.Log.Error("Nama baru tidak boleh kosong")
					return fmt.Errorf("nama baru tidak boleh kosong")
				}
				if strings.Contains(newName, "/") || strings.Contains(newName, "\\\\") {
					return fmt.Errorf("nama baru tidak boleh berisi path atau karakter '/' '\\\\'")
				}
				// set ke ProfileInfo supaya pengecekan keunikan nama memakai target baru
				s.ProfileInfo.Name = newName
			}

			if err := ValidateProfileInfo(s.ProfileInfo); err != nil {
				s.Log.Errorf("Validasi parameter gagal: %v", err)
				return err
			}
			s.Log.Info("Validasi parameter berhasil.")

			// Tampilkan data dari file yang dimuat ke table
			s.DisplayProfileDetails()
		}

		// 2. Lakukan validasi keunikan nama
		err := s.CheckConfigurationNameUnique("edit")
		if err != nil {
			// 3. JIKA TIDAK UNIK: Tampilkan error dan loop akan berulang
			ui.PrintError(err.Error())
			return err
		}

		// 2. Panggil fungsi untuk menyimpan file
		if err := s.SaveProfile("Edit"); err != nil {
			// Cek apakah error adalah permintaan retry karena koneksi gagal
			if err == validation.ErrConnectionFailedRetry {
				// Tanya user apakah ingin mengulang input
				retryInput, askErr := input.AskYesNo("Apakah Anda ingin mengulang input konfigurasi?", true)
				if askErr != nil {
					return validation.HandleInputError(askErr)
				}

				if retryInput {
					ui.PrintWarning("Mengulang proses pengeditan profil...")
					continue // Ulangi dari awal
				} else {
					ui.PrintInfo("Proses pengeditan profil dibatalkan.")
					return validation.ErrUserCancelled
				}
			}

			// Error lainnya
			return err
		}

		// Berhasil, keluar dari loop
		break
	}

	// Logika selanjutnya, seperti menyimpan file konfigurasi, bisa ditambahkan di sini
	s.Log.Info("Wizard interaktif selesai.")

	return nil
}
