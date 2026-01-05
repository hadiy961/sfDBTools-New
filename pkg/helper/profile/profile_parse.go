// File : pkg/helper/profile/profile_parse.go
// Deskripsi : Utility untuk load dan parse profil terenkripsi
// Author : Hadiyatna Muflihun
// Tanggal : 5 Desember 2025
// Last Modified : 5 Januari 2026
package profile

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"sfDBTools/internal/cli/parsing"
	"sfDBTools/internal/domain"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/encrypt"
	cryptokey "sfDBTools/pkg/helper/crypto"
	"sfDBTools/pkg/helper/profileutil"
)

// LoadAndParseProfile membaca file terenkripsi, mendapatkan kunci (jika tidak diberikan),
// mendekripsi, dan mem-parsing isi INI menjadi ProfileInfo (tanpa metadata file).
func LoadAndParseProfile(absPath string, key string) (*domain.ProfileInfo, error) {
	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, err
	}

	k := strings.TrimSpace(key)
	info := &domain.ProfileInfo{}
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
		switch info.EncryptionSource {
		case "env":
			hint = "Kunci enkripsi diambil dari environment variable " + consts.ENV_SOURCE_PROFILE_KEY + ". Pastikan nilainya sama dengan saat enkripsi atau gunakan --profile-key."
		case "flag/state":
			hint = "Kunci enkripsi berasal dari flag/argumen (--profile-key). Pastikan nilainya sesuai dengan saat enkripsi atau set env " + consts.ENV_SOURCE_PROFILE_KEY + "."
		case "prompt":
			hint = "Kunci enkripsi dimasukkan manual. Pastikan sesuai dengan yang digunakan saat enkripsi atau set via --profile-key/" + consts.ENV_SOURCE_PROFILE_KEY + "."
		default:
			hint = "Pastikan kunci enkripsi yang diberikan cocok dengan yang digunakan saat enkripsi (flag --profile-key atau env " + consts.ENV_SOURCE_PROFILE_KEY + ")."
		}
		return nil, fmt.Errorf("gagal mendekripsi file '%s': %s", absPath, hint)
	}

	parsed := parsing.ParseINIClient(string(plaintext))
	if parsed == nil {
		return nil, fmt.Errorf("gagal mem-parse isi konfigurasi '%s': format INI bagian [client] tidak ditemukan atau rusak", absPath)
	}

	sshParsed := parsing.ParseINISection(string(plaintext), "ssh")

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

	if sshParsed != nil {
		ssh := domain.SSHTunnelConfig{}
		if v, ok := sshParsed["enabled"]; ok {
			switch strings.ToLower(strings.TrimSpace(v)) {
			case "1", "true", "yes", "y", "on":
				ssh.Enabled = true
			}
		}
		if v, ok := sshParsed["host"]; ok {
			ssh.Host = strings.TrimSpace(v)
		}
		if v, ok := sshParsed["port"]; ok {
			if p, perr := strconv.Atoi(strings.TrimSpace(v)); perr == nil {
				ssh.Port = p
			}
		}
		if ssh.Port == 0 {
			ssh.Port = 22
		}
		if v, ok := sshParsed["user"]; ok {
			ssh.User = strings.TrimSpace(v)
		}
		if v, ok := sshParsed["ssh_password"]; ok {
			ssh.Password = strings.TrimSpace(v)
		}
		if v, ok := sshParsed["identity_file"]; ok {
			ssh.IdentityFile = strings.TrimSpace(v)
		}
		if v, ok := sshParsed["local_port"]; ok {
			if lp, lperr := strconv.Atoi(strings.TrimSpace(v)); lperr == nil {
				ssh.LocalPort = lp
			}
		}
		info.SSHTunnel = ssh
	}
	return info, nil
}
