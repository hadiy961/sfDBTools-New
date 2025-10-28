package profile

import (
	"fmt"
	"os"
	"path/filepath"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/encrypt"
	"sfDBTools/pkg/fsops"
	"sfDBTools/pkg/helper"
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/ui"
	"sfDBTools/pkg/validation"
	"strings"
)

func (s *Service) SaveProfile(mode string) error {
	s.Log.Info("Memulai proses penyimpanan file konfigurasi...")

	// Deteksi apakah mode interactive atau tidak
	isInteractive := s.isInteractiveMode()

	// 1. Tentukan base directory tujuan penyimpanan.
	//    - Untuk mode Edit dengan --file absolute: simpan di direktori file asli (in-place)
	//    - Selain itu: gunakan direktori konfigurasi aplikasi.
	var baseDir string
	var originalAbsPath string
	if mode == "Edit" && s.ProfileInfo.Path != "" && filepath.IsAbs(s.ProfileInfo.Path) {
		originalAbsPath = s.ProfileInfo.Path
		baseDir = filepath.Dir(s.ProfileInfo.Path)
	} else {
		baseDir = s.Config.ConfigDir.DatabaseProfile
	}

	// 2. Pastikan base directory ada, jika tidak buat (berlaku untuk config dir maupun absolut)
	dir, err := fsops.CheckDirExists(baseDir)
	if err != nil {
		return fmt.Errorf("gagal memastikan direktori konfigurasi ada: %w", err)
	}
	if !dir {
		s.Log.Info("Direktori konfigurasi tidak ada. Mencoba membuat: " + baseDir)
		if err := fsops.CreateDirIfNotExist(baseDir); err != nil {
			return fmt.Errorf("gagal membuat direktori konfigurasi: %w", err)
		}
		s.Log.Info("Direktori konfigurasi berhasil dibuat: " + baseDir)
	}

	// 3. Pastikan koneksi database valid sebelum menyimpan
	if err := database.ConnectionTest(&s.ProfileInfo.DBInfo, s.Log); err != nil {
		// Di mode non-interaktif, langsung gagal tanpa prompt
		if !isInteractive {
			return err
		}

		// Mode interaktif: Tanya user apakah tetap ingin menyimpan
		continueAnyway, askErr := input.AskYesNo("\nKoneksi database gagal. Apakah Anda tetap ingin menyimpan konfigurasi ini?", false)
		if askErr != nil {
			return validation.HandleInputError(askErr)
		}

		if !continueAnyway {
			// User memilih untuk tidak menyimpan, return special error untuk retry
			return validation.ErrConnectionFailedRetry
		}

		// User memilih untuk tetap menyimpan meski koneksi gagal
		ui.PrintWarning("⚠️  PERINGATAN: Menyimpan konfigurasi dengan koneksi database yang tidak valid.")
		s.Log.Warn("User memilih untuk menyimpan konfigurasi meskipun koneksi database gagal.")
	} else {
		s.Log.Info("Koneksi database valid.")
	}

	// 4. Ubah konfigurasi ke format INI
	iniContent := s.formatConfigToINI()

	// 5. Resolusi kunci enkripsi dan enkripsi konten
	key, _, err := helper.ResolveEncryptionKey(s.ProfileInfo.EncryptionKey, consts.ENV_SOURCE_PROFILE_KEY)
	if err != nil {
		return fmt.Errorf("kunci enkripsi tidak tersedia: %w", err)
	}
	// Enkripsi konten INI
	encryptedContent, err := encrypt.EncryptAES([]byte(iniContent), []byte(key))
	if err != nil {
		return fmt.Errorf("gagal mengenkripsi konten konfigurasi: %w", err)
	}

	// 6. Tentukan nama dan path file tujuan (normalisasi nama agar tidak double suffix)
	// Jika user memberikan NewName via flags pada mode edit, gunakan itu sebagai nama baru
	if mode == "Edit" && s.ProfileEdit != nil && strings.TrimSpace(s.ProfileEdit.NewName) != "" {
		s.ProfileInfo.Name = s.ProfileEdit.NewName
	}

	// Normalisasi nama simpanan: trim suffix jika ada lalu pastikan .cnf.enc
	s.ProfileInfo.Name = helper.TrimProfileSuffix(s.ProfileInfo.Name)
	newFileName := buildFileName(s.ProfileInfo.Name)
	newFilePath := filepath.Join(baseDir, newFileName)

	// If mode is Edit and original name exists and different -> perform rename flow
	// Rename flow: if edit mode and new name differs from original, perform rename
	if mode == "Edit" && s.OriginalProfileName != "" && s.OriginalProfileName != s.ProfileInfo.Name {
		// 1) Write new file
		if err := fsops.WriteFile(newFilePath, encryptedContent); err != nil {
			return fmt.Errorf("gagal menyimpan file konfigurasi baru: %w", err)
		}
		// 2) Delete old file
		oldFilePath := originalAbsPath
		if oldFilePath == "" && s.OriginalProfileInfo != nil && s.OriginalProfileInfo.Path != "" {
			oldFilePath = s.OriginalProfileInfo.Path
		}
		if oldFilePath == "" {
			// fallback ke baseDir + nama lama jika metadata tidak tersedia
			oldFilePath = filepath.Join(baseDir, buildFileName(s.OriginalProfileName))
		}
		if err := os.Remove(oldFilePath); err != nil {
			// Jika gagal menghapus, laporkan tapi lanjutkan (user mungkin akan membersihkan manual)
			ui.PrintWarning(fmt.Sprintf("Berhasil menyimpan '%s' tetapi gagal menghapus file lama '%s': %v", newFileName, oldFilePath, err))
		}

		ui.PrintSuccess(fmt.Sprintf("File konfigurasi berhasil disimpan sebagai '%s' (rename dari '%s').", newFileName, buildFileName(s.OriginalProfileName)))
		ui.PrintInfo("File konfigurasi tersimpan di : " + newFilePath)

		return nil
	}

	// Default: write/overwrite target file (create or update without removing anything)
	if err := fsops.WriteFile(newFilePath, encryptedContent); err != nil {
		return fmt.Errorf("gagal menyimpan file konfigurasi: %w", err)
	}

	ui.PrintSuccess(fmt.Sprintf("File konfigurasi '%s' berhasil disimpan dengan aman.", newFileName))
	ui.PrintInfo("File konfigurasi tersimpan di : " + newFilePath)

	return nil
}
