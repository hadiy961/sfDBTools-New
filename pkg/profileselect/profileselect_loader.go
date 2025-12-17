package profileselect

import (
	"fmt"
	"os"
	"path/filepath"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/encrypt"
	"sfDBTools/pkg/helper"
	"sfDBTools/internal/parsing"
	"strconv"
	"strings"
)

// LoadAndParseProfile membaca file terenkripsi, mendapatkan kunci (jika tidak diberikan),
// mendekripsi, dan mem-parsing isi INI menjadi DBConfigInfo (tanpa metadata file).
func LoadAndParseProfile(absPath string, key string) (*types.ProfileInfo, error) {
	// Ambil kunci enkripsi dari argumen, jika kosong minta dari user
	// Gunakan nilai default dari types.DBConfigInfo jika ada
	// Baca file
	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, err
	}

	// Dapatkan kunci enkripsi (jika tidak diberikan)
	// Gunakan helper yang sudah ada untuk konsistensi
	k := strings.TrimSpace(key)
	// siapkan struct profile yang akan diisi dan dikembalikan
	info := &types.ProfileInfo{}
	if k == "" {
		var src string
		// Jika kunci tidak diberikan, minta dari env atau prompt
		k, src, err = helper.ResolveEncryptionKey(key, consts.ENV_SOURCE_PROFILE_KEY)
		if err != nil {
			return nil, fmt.Errorf("kunci enkripsi tidak tersedia: %w", err)
		}
		info.EncryptionSource = src
	}

	// Dekripsi konten
	plaintext, err := encrypt.DecryptAES(data, []byte(k))
	if err != nil {
		// Berikan konteks tambahan agar user tahu kemungkinan penyebab
		var hint string
		if info.EncryptionSource == "env" {
			hint = "Menggunakan kunci enkripsi dari environment variable " + consts.ENV_SOURCE_PROFILE_KEY + ", pastikan sesuai dengan yang digunakan saat enkripsi"
		} else {
			hint = "Menggunakan kunci enkripsi dari prompt, pastikan sesuai dengan yang digunakan saat enkripsi"
		}
		return nil, fmt.Errorf("gagal mendekripsi file '%s' %s ", absPath, hint)
	}

	// Parsing INI (gunakan helper parsing yang sudah ada)
	// Hasil parsing adalah map[string]string, kita perlu mapping ke DBConfigInfo
	// Parsing hanya bagian [client]
	parsed := parsing.ParseINIClient(string(plaintext))
	if parsed == nil {
		return nil, fmt.Errorf("gagal mem-parse isi konfigurasi '%s': format INI bagian [client] tidak ditemukan atau rusak", absPath)
	}
	// pastikan name pada info di-set (pertahankan EncryptionSource jika sudah diisi)
	info.Name = helper.TrimProfileSuffix(filepath.Base(absPath))
	// parsed sudah pasti non-nil di sini (dicek di atas), langsung gunakan
	{
		if h, ok := parsed["host"]; ok {
			info.DBInfo.Host = h
		}
		if p, ok := parsed["port"]; ok {
			// gunakan strconv agar kontrol error lebih baik
			if port, perr := strconv.Atoi(strings.TrimSpace(p)); perr == nil {
				info.DBInfo.Port = port
			}
		}
		if u, ok := parsed["user"]; ok {
			info.DBInfo.User = u
		}
		if pw, ok := parsed["password"]; ok {
			info.DBInfo.Password = pw
		}
	}
	return info, nil
}
