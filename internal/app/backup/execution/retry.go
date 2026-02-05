// File : internal/backup/execution/retry.go
// Deskripsi : Error detection dan retry strategies untuk dump tool failures (mariadb-dump/mysqldump)
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-30
// Last Modified : 20 Januari 2026

package execution

import "strings"

// maxRetries membatasi jumlah retry agar tidak terjadi loop tak berujung
// jika pola error yang sama terus terdeteksi.
//
// Catatan:
// Di lapangan, perbedaan versi/flavor dump tool (MySQL vs MariaDB, distro berbeda)
// sering membuat beberapa flag tidak dikenali. Karena mekanisme retry hanya berjalan
// saat args berubah (ada opsi yang dihapus/ditambahkan), menaikkan limit ini tetap aman.
const maxRetries = 5

// IsSSLMismatchRequiredButServerNoSupport mendeteksi client/server SSL mismatch:
// client requires SSL tapi server tidak support.
//
// Contoh stderr output:
//
//	Got error: 2026: "TLS/SSL error: SSL is required, but the server does not support it"
func IsSSLMismatchRequiredButServerNoSupport(stderrOutput string) bool {
	if stderrOutput == "" {
		return false
	}
	l := strings.ToLower(stderrOutput)
	return strings.Contains(l, "tls/ssl error") &&
		strings.Contains(l, "ssl is required") &&
		strings.Contains(l, "server does not support")
}

// AddDisableSSLArgs menambahkan opsi untuk disable SSL/TLS pada dump tool.
// Menggunakan '--skip-ssl' untuk kompatibilitas MariaDB/MySQL client.
//
// Returns: (newArgs, added)
// - newArgs: args baru dengan --skip-ssl
// - added: true jika berhasil ditambahkan, false jika sudah ada
func AddDisableSSLArgs(args []string) ([]string, bool) {
	// Check apakah sudah ada SSL-related args
	for _, a := range args {
		al := strings.ToLower(strings.TrimSpace(a))
		if al == "--skip-ssl" || strings.HasPrefix(al, "--ssl-mode") || strings.HasPrefix(al, "--ssl=") {
			return args, false
		}
	}

	newArgs := make([]string, 0, len(args)+1)
	newArgs = append(newArgs, args...)
	newArgs = append(newArgs, "--skip-ssl")
	return newArgs, true
}

// RemoveUnsupportedMysqldumpOption mencoba menghapus opsi yang tidak didukung
// dari args berdasarkan stderr output.
//
// Target: mysqldump exit status 2 cases seperti:
// - "unknown option '--set-gtid-purged=OFF'"
// - "unknown variable 'set-gtid-purged=OFF'"
//
// Returns: (newArgs, removedOption, success)
// - newArgs: args setelah opsi dihapus
// - removedOption: opsi yang dihapus (untuk logging). Bisa berisi beberapa flag jika
//   opsi yang sama muncul berkali-kali (contoh: banyak --ignore-table-data=...).
// - success: true jika berhasil mendeteksi dan menghapus
func RemoveUnsupportedMysqldumpOption(args []string, stderrOutput string) ([]string, string, bool) {
	stderrLower := strings.ToLower(stderrOutput)
	if !strings.Contains(stderrLower, "unknown option") && !strings.Contains(stderrLower, "unknown variable") {
		return args, "", false
	}

	lines := strings.Split(stderrOutput, "\n")
	for _, line := range lines {
		l := strings.ToLower(line)
		if !strings.Contains(l, "unknown option") && !strings.Contains(l, "unknown variable") {
			continue
		}

		// Extract option name dari error message
		tok := extractQuotedToken(line)
		if tok == "" {
			tok = extractDashedToken(line)
		}
		tok = strings.TrimSpace(tok)
		if tok == "" {
			continue
		}

		// Coba hapus berdasarkan nama opsi (lebih robust untuk kasus banyak flag).
		// Contoh stderr: "unknown variable 'ignore-table-data=db.tbl'"
		// -> hapus SEMUA arg yang match "--ignore-table-data" (bukan hanya 1).
		if newArgs, removed, ok := removeByOptionKey(args, tok); ok {
			return newArgs, removed, true
		}

		// Fallback: hapus berdasarkan token penuh (behaviour lama).
		candidate := tok
		// mysqldump kadang menulis tanpa leading dashes (e.g., set-gtid-purged=OFF)
		if !strings.HasPrefix(candidate, "-") {
			if strings.Contains(candidate, "-") || strings.Contains(candidate, "=") {
				candidate = "--" + candidate
			} else {
				continue
			}
		}

		newArgs, removed, ok := removeFlagArg(args, candidate, tok)
		if ok {
			return newArgs, removed, true
		}
	}

	return args, "", false
}

// removeByOptionKey menghapus semua arg yang sesuai dengan "key" dari token error.
// token bisa dalam bentuk:
// - "--opt=value"
// - "opt=value"
// - "--opt"
// - "opt"
//
// Untuk form yang mengandung "=", hanya bagian sebelum "=" yang dianggap key.
// Fungsi ini akan menghapus:
// - "--key=value" (prefix match)
// - "--key" + value (two-arg form)
//
// Returns: (newArgs, removedListString, success)
func removeByOptionKey(args []string, rawToken string) ([]string, string, bool) {
	if strings.TrimSpace(rawToken) == "" {
		return args, "", false
	}

	// Normalisasi token: buang leading dashes, ambil key sebelum '='
	norm := strings.TrimLeft(strings.TrimSpace(rawToken), "-")
	if norm == "" {
		return args, "", false
	}
	key := norm
	if i := strings.Index(key, "="); i >= 0 {
		key = strings.TrimSpace(key[:i])
	}
	if key == "" {
		return args, "", false
	}

	candidate := "--" + key
	out := make([]string, 0, len(args))
	removed := make([]string, 0, 2)

	for i := 0; i < len(args); i++ {
		arg := args[i]

		// Skip non-flag args (nama database, dsb)
		if !strings.HasPrefix(arg, "-") {
			out = append(out, arg)
			continue
		}

		// Match "--key" or "--key=..."
		matched := arg == candidate || strings.HasPrefix(arg, candidate+"=")
		if !matched {
			out = append(out, arg)
			continue
		}

		removed = append(removed, arg)

		// Jika dalam format two-arg: "--key" "value", hapus value juga.
		if arg == candidate &&
			i+1 < len(args) &&
			!strings.HasPrefix(args[i+1], "-") {
			i++ // skip value
		}
	}

	if len(removed) == 0 {
		return args, "", false
	}

	return out, strings.Join(removed, ", "), true
}

// extractQuotedToken mengekstrak token dari dalam quotes ('...' atau "...")
func extractQuotedToken(s string) string {
	// Prefer single quotes
	if i := strings.Index(s, "'"); i >= 0 {
		if j := strings.Index(s[i+1:], "'"); j >= 0 {
			return s[i+1 : i+1+j]
		}
	}
	// Fallback to double quotes
	if i := strings.Index(s, "\""); i >= 0 {
		if j := strings.Index(s[i+1:], "\""); j >= 0 {
			return s[i+1 : i+1+j]
		}
	}
	return ""
}

// extractDashedToken mengekstrak token yang dimulai dengan dash (-- atau -)
func extractDashedToken(s string) string {
	idx := strings.Index(s, "--")
	if idx < 0 {
		idx = strings.Index(s, "-")
	}
	if idx < 0 {
		return ""
	}

	end := idx
	for end < len(s) {
		c := s[end]
		if c == ' ' || c == '\t' || c == '\r' || c == '\n' {
			break
		}
		end++
	}
	return s[idx:end]
}

// removeFlagArg menghapus flag dari args list.
// Juga menghapus value-nya jika flag dalam format dua-argumen (--opt value).
//
// Returns: (newArgs, removedArg, success)
func removeFlagArg(args []string, candidate string, rawToken string) ([]string, string, bool) {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "-") {
			continue // Skip database names
		}

		// Check for match
		match := arg == candidate || arg == rawToken
		if !match {
			// Common case: stderr token omits leading dashes
			match = rawToken != "" && strings.Contains(arg, rawToken)
		}
		if !match {
			continue
		}

		removed := arg

		// Jika opsi dalam format dua-arg ("--opt" "value"), hapus value juga
		removeNextValue := !strings.Contains(arg, "=") &&
			i+1 < len(args) &&
			!strings.HasPrefix(args[i+1], "-")

		out := make([]string, 0, len(args))
		out = append(out, args[:i]...)
		if removeNextValue {
			out = append(out, args[i+2:]...) // Skip both flag dan value
		} else {
			out = append(out, args[i+1:]...) // Skip hanya flag
		}

		return out, removed, true
	}

	return args, "", false
}
