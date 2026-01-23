// File : internal/services/config/appconfig_loaders.go
// Deskripsi : Fungsi untuk memuat konfigurasi dari file YAML dan variabel lingkungan
// Author : Hadiyatna Muflihun
// Tanggal : 3 Oktober 2024
// Last Modified : 5 Januari 2026
package appconfig

import (
	"errors"
	"fmt"
	"os"

	"sfdbtools/internal/shared/consts"

	"github.com/joho/godotenv" // Untuk memuat file .env
	"gopkg.in/yaml.v3"         // Untuk mem-parsing file YAML
)

// LoadConfigFromEnv memuat file .env (jika ada) dan kemudian
// membaca path konfigurasi dari variabel lingkungan SFDB_APPS_CONFIG.
// Ini mengembalikan struct Config yang terisi, atau error jika gagal.
func LoadConfigFromEnv() (*Config, error) {
	// 1. Memuat variabel lingkungan dari file .env (best practice)
	// Kita abaikan error jika file .env tidak ditemukan, ini wajar
	// di lingkungan produksi di mana variabel env sudah diset.
	_ = godotenv.Load()

	// 2. Membaca path file konfigurasi dari variabel lingkungan
	configPath := os.Getenv(consts.ENV_APPS_CONFIG)
	if configPath == "" {
		// Jika variabel lingkungan tidak ada, gunakan path default
		configPath = consts.DefaultAppConfigPath
	}

	// 2a. Auto-init config jika belum ada.
	// Jika path default (/etc/...) tidak bisa dibuat (non-root), fallback ke user config.
	if _, err := os.Stat(configPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if err := writeDefaultConfig(configPath); err != nil {
				userPath, uerr := userDefaultConfigPath()
				if uerr != nil {
					return nil, fmt.Errorf("gagal membuat config default di %s: %w", configPath, err)
				}
				if err2 := writeDefaultConfig(userPath); err2 != nil {
					return nil, fmt.Errorf("gagal membuat config default di %s: %w", userPath, err2)
				}
				configPath = userPath
			}
		} else {
			return nil, err
		}
	}

	// 3. Memuat dan mem-parsing file konfigurasi
	cfg, err := LoadConfig(configPath)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// LoadConfig membaca dan mem-parsing file YAML dari path yang diberikan
// ke dalam struct Config. Ini adalah fungsi yang reusable.
func LoadConfig(configPath string) (*Config, error) {
	// Membaca konten file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	// Inisialisasi struct Config
	cfg := &Config{}

	// Parsing konten YAML ke dalam struct
	err = yaml.Unmarshal(data, cfg)
	if err != nil {
		return nil, err
	}

	// Best practice: Anda bisa menambahkan logika validasi kustom di sini
	// setelah parsing berhasil (misalnya, cek apakah BaseDirectory tidak kosong)
	applyRuntimeDefaults(cfg, configPath)

	return cfg, nil
}

func applyRuntimeDefaults(cfg *Config, configPath string) {
	if cfg == nil {
		return
	}
	// Hardcode app identity (tidak boleh diubah lewat config)
	cfg.General.AppName = HardcodedAppName
	cfg.General.Author = HardcodedAuthor
	cfg.General.ClientCode = HardcodedClientCode
	if cfg.General.Version == "" {
		cfg.General.Version = defaultConfigForPath(configPath).General.Version
	}
	if cfg.ConfigDir.DatabaseProfile == "" {
		cfg.ConfigDir.DatabaseProfile = defaultConfigForPath(configPath).ConfigDir.DatabaseProfile
	}
}
