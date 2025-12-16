package profileselect

import (
	"fmt"
	"os"
	"path/filepath"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/fsops"
	"sfDBTools/pkg/helper"
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/ui"
	"sfDBTools/pkg/validation"
)

func SelectExistingDBConfig(configDir, purpose string) (types.ProfileInfo, error) {
	// Tujuan: menampilkan daftar file konfigurasi yang ada dan membiarkan user memilih satu
	// Kembalikan DBConfigInfo yang berisi metadata file dan detail koneksi (jika berhasil dimuat)
	// Jika tidak ada file, kembalikan error
	// Jika user membatalkan, kembalikan error khusus
	ui.PrintSubHeader(purpose)

	// Baca isi direktori konfigurasi
	ProfileInfo := types.ProfileInfo{}

	// Bacaan direktori
	files, err := fsops.ReadDirFiles(configDir)
	if err != nil {
		return ProfileInfo, fmt.Errorf("gagal membaca direktori konfigurasi '%s': %w", configDir, err)
	}

	// filter hanya *.cnf.enc (gunakan helper EnsureConfigExt untuk konsistensi)
	filtered := make([]string, 0, len(files))
	for _, f := range files {
		if validation.ProfileExt(f) == f { // artinya sudah ber-ekstensi .cnf.enc
			filtered = append(filtered, f)
		}
	}

	// Jika tidak ada file, kembalikan error
	if len(filtered) == 0 {
		ui.PrintWarning("Tidak ditemukan file konfigurasi di direktori: " + configDir)
		ui.PrintInfo("Silakan buat file konfigurasi baru terlebih dahulu dengan perintah 'profile create'.")
		return ProfileInfo, fmt.Errorf("tidak ada file konfigurasi untuk dipilih")
	}

	// Buat opsi dengan nama file
	options := make([]string, 0, len(filtered))
	options = append(options, filtered...)

	// Tampilkan menu dan dapatkan pilihan
	idx, err := input.ShowMenu("Pilih file konfigurasi :", options)
	if err != nil {
		return ProfileInfo, validation.HandleInputError(err)
	}

	// index adalah 1-based
	selected := options[idx-1]
	name := helper.TrimProfileSuffix(selected)

	// Coba baca file yang dipilih
	filePath := filepath.Join(configDir, selected)
	// set metadata dasar lebih awal supaya dapat dikembalikan walau load/parsing gagal
	ProfileInfo.Path = filePath
	ProfileInfo.Name = name

	// Coba muat file yang dipilih (read+decrypt+parse ditangani helper)
	// Gunakan helper untuk memuat dan mem-parsing isi
	info, err := LoadAndParseProfile(filePath, ProfileInfo.EncryptionKey)
	if err != nil {
		// kembalikan metadata file untuk membantu debug/menampilkan informasi
		return ProfileInfo, err
	}
	if info != nil {
		ProfileInfo.DBInfo = info.DBInfo
		ProfileInfo.EncryptionSource = info.EncryptionSource
	}

	// Ambil metadata file (size dan last modified)
	var fileSizeStr string
	var lastModTime = ProfileInfo.LastModified
	if fi, err := os.Stat(filePath); err == nil {
		fileSizeStr = fmt.Sprintf("%d bytes", fi.Size())
		lastModTime = fi.ModTime()
	}

	// Setelah berhasil memuat isi file, simpan snapshot data asli agar dapat
	// dibandingkan dengan perubahan yang dilakukan user. Sertakan metadata file.
	// ProfileInfo.Path and Name already set above.
	ProfileInfo.DBInfo = info.DBInfo
	ProfileInfo.Size = fileSizeStr
	ProfileInfo.LastModified = lastModTime

	return ProfileInfo, nil
}
