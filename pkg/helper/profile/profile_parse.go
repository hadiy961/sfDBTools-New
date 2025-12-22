package profile

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"sfDBTools/internal/parsing"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/encrypt"
	cryptokey "sfDBTools/pkg/helper/crypto"
	"sfDBTools/pkg/helper/profileutil"
)

// LoadAndParseProfile membaca file terenkripsi, mendapatkan kunci (jika tidak diberikan),
// mendekripsi, dan mem-parsing isi INI menjadi ProfileInfo (tanpa metadata file).
func LoadAndParseProfile(absPath string, key string) (*types.ProfileInfo, error) {
	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, err
	}

	k := strings.TrimSpace(key)
	info := &types.ProfileInfo{}
	if k == "" {
		var src string
		k, src, err = cryptokey.ResolveEncryptionKey(key, consts.ENV_SOURCE_PROFILE_KEY)
		if err != nil {
			return nil, fmt.Errorf("kunci enkripsi tidak tersedia: %w", err)
		}
		info.EncryptionSource = src
	}

	plaintext, err := encrypt.DecryptAES(data, []byte(k))
	if err != nil {
		var hint string
		if info.EncryptionSource == "env" {
			hint = "Menggunakan kunci enkripsi dari environment variable " + consts.ENV_SOURCE_PROFILE_KEY + ", pastikan sesuai dengan yang digunakan saat enkripsi"
		} else {
			hint = "Menggunakan kunci enkripsi dari prompt, pastikan sesuai dengan yang digunakan saat enkripsi"
		}
		return nil, fmt.Errorf("gagal mendekripsi file '%s' %s ", absPath, hint)
	}

	parsed := parsing.ParseINIClient(string(plaintext))
	if parsed == nil {
		return nil, fmt.Errorf("gagal mem-parse isi konfigurasi '%s': format INI bagian [client] tidak ditemukan atau rusak", absPath)
	}

	info.Name = profileutil.TrimProfileSuffix(filepath.Base(absPath))
	{
		if h, ok := parsed["host"]; ok {
			info.DBInfo.Host = h
		}
		if p, ok := parsed["port"]; ok {
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
