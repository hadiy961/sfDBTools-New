package profile

import (
	"os"
	"sfDBTools/internal/profileselect"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/helper"
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/ui"
	"sfDBTools/pkg/validation"

	"github.com/AlecAivazis/survey/v2"
)

// promptSelectExistingConfig menampilkan daftar file konfigurasi dari direktori
// konfigurasi aplikasi dan meminta pengguna memilih salah satu.
func (s *Service) promptSelectExistingConfig() error {
	info, err := profileselect.SelectExistingDBConfig("Select Existing Configuration File")
	if err != nil {
		return err
	}

	// Muat data ke struct ProfileInfo
	s.ProfileInfo = &types.ProfileInfo{
		Path:         info.Path,
		Name:         info.Name,
		DBInfo:       info.DBInfo,
		Size:         info.Size,
		LastModified: info.LastModified,
	}

	// Set OriginalProfileName untuk validasi edit
	s.OriginalProfileName = info.Name

	// Setelah berhasil memuat isi file, simpan snapshot data asli agar dapat
	// dibandingkan dengan perubahan yang dilakukan user. Sertakan metadata file.
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

// PERBAIKAN: Fungsi ini juga sekarang hanya mengembalikan error.
func (s *Service) promptDBConfigName(mode string) error {
	if s.ProfileInfo.Name == "" {
		s.ProfileInfo.Name = "my_database_config" // Set default jika kosong
	}
	ui.PrintSubHeader("Please provide the configuration name:")

	// Mulai loop untuk meminta input sampai valid
	for {
		// 1. Minta input dari pengguna
		nameValidator := input.ComposeValidators(survey.Required, input.ValidateFilename)
		configName, err := input.AskString("Configuration Name", s.ProfileInfo.Name, nameValidator)
		if err != nil {
			return validation.HandleInputError(err) // Keluar jika pengguna membatalkan (misal: Ctrl+C)
		}

		// Selalu update nama di struct dengan versi dinormalisasi (trim suffix)
		s.ProfileInfo.Name = helper.TrimProfileSuffix(configName)

		// 2. Lakukan validasi keunikan nama
		err = s.CheckConfigurationNameUnique(mode)
		if err != nil {
			// 3. JIKA TIDAK UNIK: Tampilkan error dan loop akan berulang
			ui.PrintError(err.Error())
			continue // Lanjutkan ke iterasi loop berikutnya
		}

		// 4. JIKA UNIK: Keluar dari loop
		break
	}

	// Tampilkan informasi akhir
	ui.PrintInfo("Konfigurasi akan disimpan sebagai : " + buildFileName(s.ProfileInfo.Name))
	return nil
}

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

	// Detect edit flow if we already have an OriginalProfileInfo or OriginalProfileName.
	isEditFlow := s.OriginalProfileInfo != nil || s.OriginalProfileName != ""

	// Store existing password from loaded config before checking environment
	var existingPassword string
	if isEditFlow && s.ProfileInfo != nil {
		existingPassword = s.ProfileInfo.DBInfo.Password
	}

	//Cek password dari env SFDB_DB_PASSWORD
	//Jika tidak ada, beri tahu user untuk mengaturnya di env
	//Untuk keamanan, jangan minta input password di prompt
	//Namun, jika ingin minta input, uncomment kode di bawah dan comment kode pengecekan env
	// s.Logger.Debug("Mengecek environment variable SFDB_DB_PASSWORD untuk password database...")
	envPassword := os.Getenv(consts.ENV_TARGET_DB_PASSWORD)

	// PENTING: Jangan auto-overwrite password dengan env password di edit mode
	// Biarkan user yang menentukan apakah mau pakai env password atau tetap pakai yang lama
	if !isEditFlow && envPassword != "" {
		// Hanya auto-set di create mode
		s.ProfileInfo.DBInfo.Password = envPassword
	} else if !isEditFlow {
		// Only show warning for create flow when env password not available
		ui.PrintWarning("Environment variable SFDB_DB_PASSWORD tidak ditemukan atau kosong. Silakan atur SFDB_DB_PASSWORD atau ketik password.")
	}

	// Allow empty password in edit flow to mean "keep existing password".

	// Tampilkan hint untuk mode edit agar user tahu opsi yang tersedia
	if isEditFlow && existingPassword != "" {
		ui.PrintInfo("💡 Password saat ini dimuat dari file. Tekan Enter untuk mempertahankan, atau ketik password baru untuk mengupdate.")
	}

	// First try accepting empty input (validator nil). If this is create flow and
	// user entered empty, we enforce non-empty by asking again with Required.
	pw, err := input.AskPassword("Database Password", nil)
	if err != nil {
		return validation.HandleInputError(err)
	}

	if pw == "" {
		if isEditFlow {
			// keep existing password (from loaded file, env password is NOT auto-applied in edit mode)
			if existingPassword != "" {
				s.ProfileInfo.DBInfo.Password = existingPassword
			}
			// If existing password is empty, keep it empty
		} else {
			// create flow: password required
			pw, err = input.AskPassword("Database Password", survey.Required)
			if err != nil {
				return validation.HandleInputError(err)
			}
			s.ProfileInfo.DBInfo.Password = pw
		}
	} else {
		// user provided a new password -> overwrite
		s.ProfileInfo.DBInfo.Password = pw
	}

	return nil
}
