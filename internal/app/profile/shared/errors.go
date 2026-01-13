// File : internal/app/profile/shared/errors.go
// Deskripsi : Centralized error definitions untuk profile module (P1 improvement)
// Author : Hadiyatna Muflihun
// Tanggal : 14 Januari 2026
// Last Modified : 14 Januari 2026

package shared

import (
	"errors"
	"fmt"

	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/validation"
)

// =============================================================================
// Base Errors (dapat digunakan dengan errors.Is untuk pattern matching)
// =============================================================================

var (
	// Profile general errors
	ErrProfileNil           = errors.New("profile tidak boleh nil")
	ErrProfileNameEmpty     = errors.New("nama profil tidak boleh kosong")
	ErrProfileNotFound      = errors.New("profil tidak ditemukan")
	ErrProfileAlreadyExists = errors.New("profil sudah ada")
	ErrInvalidProfileMode   = errors.New("mode profile tidak valid")
	ErrNoConfigToSelect     = errors.New("tidak ada file konfigurasi untuk dipilih")
	ErrNoSnapshotToShow     = errors.New("tidak ada snapshot profil untuk ditampilkan")
	ErrProfilePathEmpty     = errors.New("path profile kosong")
	ErrConfigNameEmpty      = errors.New("nama konfigurasi kosong")

	// Database connection errors
	ErrDBHostEmpty      = errors.New("host database kosong")
	ErrDBPortInvalid    = errors.New("port database tidak valid")
	ErrDBUserEmpty      = errors.New("user database kosong")
	ErrDBInfoNil        = errors.New("db info tidak boleh nil")
	ErrConnectionFailed = errors.New("koneksi database gagal")

	// SSH tunnel errors
	ErrSSHHostEmpty           = errors.New("ssh host kosong")
	ErrSSHPortInvalid         = errors.New("ssh port tidak valid")
	ErrSSHTunnelHostEmpty     = errors.New("ssh tunnel aktif tetapi ssh host kosong")
	ErrSSHLocalPortInvalid    = errors.New("ssh local port tidak valid")
	ErrSSHAuthMethodMissing   = errors.New("metode autentikasi SSH tidak tersedia")
	ErrSSHIdentityFileInvalid = errors.New("identity file SSH tidak valid")
	ErrSSHKnownHostsInvalid   = errors.New("known_hosts tidak valid")
	ErrRemoteHostEmpty        = errors.New("remote host kosong")
	ErrRemotePortEmpty        = errors.New("remote port kosong")

	// Encryption/Decryption errors
	ErrEncryptionKeyMissing  = errors.New("kunci enkripsi tidak tersedia")
	ErrEncryptionKeyMismatch = errors.New("kunci enkripsi tidak cocok")
	ErrDecryptionFailed      = errors.New("gagal mendekripsi file")
	ErrEncryptionFailed      = errors.New("gagal mengenkripsi konfigurasi")

	// Parsing/Loading errors
	ErrParseConfigFailed   = errors.New("gagal mem-parse konfigurasi")
	ErrLoadProfileFailed   = errors.New("gagal load profile")
	ErrReadConfigDirFailed = errors.New("gagal membaca direktori konfigurasi")
	ErrINIFormatInvalid    = errors.New("format INI tidak valid")

	// File operations errors
	ErrConfigFileNotFound    = errors.New("file konfigurasi tidak ditemukan")
	ErrCreateConfigDirFailed = errors.New("gagal membuat direktori konfigurasi")
	ErrWriteConfigFailed     = errors.New("gagal menulis file konfigurasi")
	ErrReadFileFailed        = errors.New("gagal membaca file")
	ErrFileIsDirectory       = errors.New("path adalah direktori, bukan file")
	ErrFileNotAccessible     = errors.New("file tidak dapat diakses")

	// Validation errors
	ErrInputHasLeadingTrailingSpace = errors.New("input tidak boleh diawali/diakhiri spasi")
	ErrInputHasControlChars         = errors.New("input tidak boleh mengandung karakter kontrol")
	ErrInputHasSpaces               = errors.New("input tidak boleh mengandung spasi")
	ErrInputEmpty                   = errors.New("input tidak boleh kosong")
	ErrInputNotNumeric              = errors.New("input harus berupa angka")
	ErrInputOutOfRange              = errors.New("input di luar range yang diizinkan")

	// Interactive/Non-interactive mode errors
	ErrNonInteractiveMissingKey      = errors.New("mode non-interaktif membutuhkan kunci enkripsi")
	ErrNonInteractiveProfileRequired = errors.New("mode non-interaktif membutuhkan flag --profile")
	ErrCallbackUnavailable           = errors.New("callback tidak tersedia")
)

// =============================================================================
// Error Constructors (untuk error dengan context tambahan)
// =============================================================================

// ProfileNotFoundError creates error dengan nama profile yang tidak ditemukan
func ProfileNotFoundError(name string) error {
	return fmt.Errorf("%w: %s", ErrProfileNotFound, name)
}

// ProfileAlreadyExistsError creates error dengan nama profile yang sudah ada
func ProfileAlreadyExistsError(name string) error {
	return fmt.Errorf("%w: %s", ErrProfileAlreadyExists, name)
}

// ConfigFileNotFoundError creates error dengan path file yang tidak ditemukan
func ConfigFileNotFoundError(path string) error {
	return fmt.Errorf("%w: %s", ErrConfigFileNotFound, path)
}

// DBPortInvalidError creates error dengan port value yang tidak valid
func DBPortInvalidError(port int) error {
	return fmt.Errorf("%w: %d", ErrDBPortInvalid, port)
}

// SSHPortInvalidError creates error dengan ssh port yang tidak valid
func SSHPortInvalidError(port int) error {
	return fmt.Errorf("%w: %d", ErrSSHPortInvalid, port)
}

// ReadConfigDirError creates error saat gagal baca direktori config
func ReadConfigDirError(dir string, err error) error {
	return fmt.Errorf("%w '%s': %v", ErrReadConfigDirFailed, dir, err)
}

// DecryptionFailedError creates error dengan hint kenapa dekripsi gagal
func DecryptionFailedError(path string, hint string) error {
	return fmt.Errorf("%w '%s': %s", ErrDecryptionFailed, path, hint)
}

// ParseConfigFailedError creates error saat parsing config gagal
func ParseConfigFailedError(path string) error {
	return fmt.Errorf("%w '%s': format INI bagian [client] tidak ditemukan atau rusak", ErrParseConfigFailed, path)
}

// LoadProfileError wraps underlying error saat load profile
func LoadProfileError(err error) error {
	return fmt.Errorf("%w: %v", ErrLoadProfileFailed, err)
}

// ConnectionFailedError creates error dengan detail koneksi yang gagal
func ConnectionFailedError(user, host string, port int, err error) error {
	return fmt.Errorf("%w ke %s@%s:%d: %v", ErrConnectionFailed, user, host, port, err)
}

// SSHTunnelError creates error saat ssh tunnel gagal dibuat
func SSHTunnelError(host string, port int, err error) error {
	return fmt.Errorf("gagal membuat SSH tunnel ke %s:%d: %w", host, port, err)
}

// SSHIdentityFileError creates error untuk masalah identity file
func SSHIdentityFileError(path string, reason string) error {
	return fmt.Errorf("%w: %s (%s)", ErrSSHIdentityFileInvalid, path, reason)
}

// ValidationError wraps field name dengan error message
func ValidationError(fieldName string, err error) error {
	return fmt.Errorf("%s: %w", fieldName, err)
}

// InputRangeError creates error untuk input di luar range
func InputRangeError(fieldName string, min, max int, allowZero bool) error {
	if allowZero {
		return fmt.Errorf("%s harus 0 (otomatis) atau %d-%d", fieldName, min, max)
	}
	return fmt.Errorf("%s harus %d-%d", fieldName, min, max)
}

// =============================================================================
// Legacy compatibility functions (untuk backward compatibility dengan consts)
// =============================================================================

// NonInteractiveProfileKeyRequiredError adalah error standar saat mode non-interaktif butuh kunci enkripsi.
func NonInteractiveProfileKeyRequiredError() error {
	return fmt.Errorf(
		consts.ProfileErrNonInteractiveProfileKeyRequiredFmt,
		consts.ENV_TARGET_PROFILE_KEY,
		consts.ENV_SOURCE_PROFILE_KEY,
		validation.ErrNonInteractive,
	)
}
