package backuphelper

import (
	"strconv"
	"strings"
)

// DatabaseConn is a small, package-local representation of connection info
// required to build mysqldump args. We keep this minimal to avoid importing
// internal types from other packages.
type DatabaseConn struct {
	Host     string
	Port     int
	User     string
	Password string
}

// FilterOptions mirrors the small subset of backup filter options used when
// building mysqldump args.
type FilterOptions struct {
	ExcludeData      bool
	ExcludeDatabases []string
	IncludeDatabases []string
	ExcludeSystem    bool
	ExcludeDBFile    string
	IncludeFile      string
}

// BuildMysqldumpArgs membuat argumen mysqldump dari parameter umum.
// Fungsi ini mengekstrak logika dari internal/backup sehingga bisa dipakai
// di beberapa tempat tanpa duplikasi.
func BuildMysqldumpArgs(baseDumpArgs string, conn DatabaseConn, filter FilterOptions, dbFiltered []string, singleDB string, totalDBFound int) []string {
	var args []string

	// ... (bagian koneksi host/port/user/pass tetap sama) ...
	if conn.Host != "" {
		args = append(args, "--host="+conn.Host)
	}
	if conn.Port != 0 {
		args = append(args, "--port="+strconv.Itoa(conn.Port))
	}
	if conn.User != "" {
		args = append(args, "--user="+conn.User)
	}
	if conn.Password != "" {
		args = append(args, "--password="+conn.Password)
	}

	if baseDumpArgs != "" {
		baseArgs := strings.Fields(baseDumpArgs)
		args = append(args, baseArgs...)
	}

	if filter.ExcludeData {
		args = append(args, "--no-data")
	}

	// ---------------- PERUBAHAN UTAMA DI SINI ----------------

	// KASUS 1: Single DB yang diminta secara eksplisit
	if singleDB != "" {
		// Jangan gunakan --databases. Cukup append nama DB-nya saja.
		// Output tidak akan ada 'CREATE DATABASE' atau 'USE'.
		args = append(args, singleDB)
		return args
	}

	hasFilter := len(filter.ExcludeDatabases) > 0 ||
		len(filter.IncludeDatabases) > 0 ||
		filter.ExcludeSystem ||
		filter.ExcludeDBFile != "" ||
		filter.IncludeFile != ""

	// KASUS 2: Full Backup (All Databases)
	if !hasFilter && len(dbFiltered) == totalDBFound {
		args = append(args, "--all-databases")
	} else {
		// KASUS 3: Filtered Databases

		// Cek: Jika hasil filter ternyata hanya menyisakan 1 database saja
		if len(dbFiltered) == 1 {
			// Perlakukan seperti Single DB agar bisa di-restore ke DB lain
			args = append(args, dbFiltered[0])
		} else if len(dbFiltered) > 1 {
			// Jika lebih dari 1 database, WAJIB pakai --databases.
			// Jika tidak, mysqldump akan error atau salah mengira nama db kedua sebagai nama tabel.
			// Konsekuensinya: file dump ini akan terikat pada nama DB aslinya.
			args = append(args, "--databases")
			args = append(args, dbFiltered...)
		}
	}

	return args
}

// IsFatalMysqldumpError menentukan apakah error dari mysqldump bersifat fatal.
// Implementasi ini meniru heuristik yang sebelumnya ada di internal package.
func IsFatalMysqldumpError(err error, stderrOutput string) bool {
	if err == nil {
		return false
	}
	if stderrOutput == "" {
		return true
	}
	stderrLower := strings.ToLower(stderrOutput)

	fatalPatterns := []string{
		"access denied",
		"unknown database",
		"unknown server",
		"can't connect",
		"connection refused",
		"got error:",
		"error:",
		"failed",
	}
	for _, p := range fatalPatterns {
		if strings.Contains(stderrLower, p) {
			return true
		}
	}

	// Non-fatal patterns
	nonFatalPatterns := []string{
		"couldn't read keys from table",
		"references invalid table(s) or column(s) or function(s)",
		"definer/invoker of view lack rights",
		"warning:",
	}
	for _, p := range nonFatalPatterns {
		if strings.Contains(stderrLower, p) {
			return false
		}
	}

	// Default conservative behaviour: treat as fatal
	return true
}

// MaskPasswordInArgs returns a copy of args with password values masked for logging.
func MaskPasswordInArgs(args []string) []string {
	masked := make([]string, len(args))
	copy(masked, args)

	for i := 0; i < len(masked); i++ {
		arg := masked[i]
		if strings.HasPrefix(arg, "-p") && len(arg) > 2 {
			masked[i] = "-p********"
		} else if strings.HasPrefix(arg, "--password=") {
			masked[i] = "--password=********"
		} else if arg == "-p" || arg == "--password" {
			if i+1 < len(masked) {
				masked[i+1] = "********"
			}
		}
	}

	return masked
}
