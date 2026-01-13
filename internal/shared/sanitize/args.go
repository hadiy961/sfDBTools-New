// File : internal/shared/sanitize/args.go
// Deskripsi : Sanitizer argumen CLI untuk logging (masking nilai sensitif)
// Author : Hadiyatna Muflihun
// Tanggal : 13 Januari 2026
// Last Modified : 13 Januari 2026

package sanitize

import "strings"

const maskedValue = "******"

// Args mengembalikan salinan args dengan nilai sensitif dimasking.
// Tujuan: argumen bisa dicatat di log tanpa membocorkan password/key/token.
func Args(args []string) []string {
	out := make([]string, 0, len(args))
	maskNext := false

	for _, a := range args {
		if maskNext {
			out = append(out, maskedValue)
			maskNext = false
			continue
		}

		trim := strings.TrimSpace(a)
		if trim == "" {
			out = append(out, a)
			continue
		}

		// Format ENV=VALUE
		if k, v, ok := strings.Cut(trim, "="); ok {
			if isSensitiveName(k) {
				out = append(out, k+"="+maskedValue)
				continue
			}
			_ = v
		}

		// Long flag: --name=value atau --name value
		if strings.HasPrefix(trim, "--") {
			nameAndMaybeVal := strings.TrimPrefix(trim, "--")
			if name, val, hasEq := strings.Cut(nameAndMaybeVal, "="); hasEq {
				if isSensitiveName(name) {
					out = append(out, "--"+name+"="+maskedValue)
					continue
				}
				_ = val
				out = append(out, a)
				continue
			}

			if isSensitiveName(nameAndMaybeVal) {
				out = append(out, a)
				maskNext = true
				continue
			}

			out = append(out, a)
			continue
		}

		// Short flag: -p value, -pVALUE
		if strings.HasPrefix(trim, "-") && !strings.HasPrefix(trim, "--") {
			s := strings.TrimPrefix(trim, "-")
			if len(s) == 1 && isSensitiveShortFlag(s) {
				out = append(out, a)
				maskNext = true
				continue
			}
			if len(s) > 1 {
				// Handle -pVALUE (attached)
				first := s[:1]
				rest := s[1:]
				if isSensitiveShortFlag(first) && strings.TrimSpace(rest) != "" {
					out = append(out, "-"+first+maskedValue)
					continue
				}
			}
		}

		out = append(out, a)
	}

	return out
}

func isSensitiveShortFlag(flag string) bool {
	// Hindari terlalu agresif; hanya yang sangat umum.
	// -p sering dipakai untuk password di banyak CLI.
	// -k sering dipakai untuk key.
	switch strings.ToLower(flag) {
	case "p", "k":
		return true
	default:
		return false
	}
}

func isSensitiveName(name string) bool {
	n := strings.ToLower(strings.TrimSpace(name))
	if n == "" {
		return false
	}

	// Pattern umum (flag name atau ENV var).
	// Contoh: --db-password, SFDB_DB_PASSWORD, --source-profile-key, --backup-encryption-key
	if strings.Contains(n, "password") || strings.Contains(n, "passwd") {
		return true
	}
	if strings.Contains(n, "passphrase") {
		return true
	}
	if strings.Contains(n, "encryption") && strings.Contains(n, "key") {
		return true
	}
	if strings.Contains(n, "_password") {
		return true
	}
	if strings.Contains(n, "_key") {
		return true
	}
	if strings.HasSuffix(n, "key") || strings.Contains(n, "-key") || strings.Contains(n, "_key") {
		return true
	}
	if strings.Contains(n, "secret") || strings.Contains(n, "token") {
		return true
	}

	return false
}
