// File : internal/app/profile/shared/errors.go
// Deskripsi : (DEPRECATED) Facade errors untuk profile module
// Author : Hadiyatna Muflihun
// Tanggal : 14 Januari 2026
// Last Modified : 14 Januari 2026

package shared

import (
	profileerrors "sfdbtools/internal/app/profile/errors"
)

// =============================================================================
// Base Errors (dapat digunakan dengan errors.Is untuk pattern matching)
// =============================================================================

var (
	// Profile general errors
	ErrProfileNil           = profileerrors.ErrProfileNil
	ErrProfileNameEmpty     = profileerrors.ErrProfileNameEmpty
	ErrProfileNotFound      = profileerrors.ErrProfileNotFound
	ErrProfileAlreadyExists = profileerrors.ErrProfileAlreadyExists
	ErrInvalidProfileMode   = profileerrors.ErrInvalidProfileMode
	ErrNoConfigToSelect     = profileerrors.ErrNoConfigToSelect
	ErrNoSnapshotToShow     = profileerrors.ErrNoSnapshotToShow
	ErrProfilePathEmpty     = profileerrors.ErrProfilePathEmpty
	ErrConfigNameEmpty      = profileerrors.ErrConfigNameEmpty

	// Database connection errors
	ErrDBHostEmpty      = profileerrors.ErrDBHostEmpty
	ErrDBPortInvalid    = profileerrors.ErrDBPortInvalid
	ErrDBUserEmpty      = profileerrors.ErrDBUserEmpty
	ErrDBInfoNil        = profileerrors.ErrDBInfoNil
	ErrConnectionFailed = profileerrors.ErrConnectionFailed

	// SSH tunnel errors
	ErrSSHHostEmpty           = profileerrors.ErrSSHHostEmpty
	ErrSSHPortInvalid         = profileerrors.ErrSSHPortInvalid
	ErrSSHTunnelHostEmpty     = profileerrors.ErrSSHTunnelHostEmpty
	ErrSSHLocalPortInvalid    = profileerrors.ErrSSHLocalPortInvalid
	ErrSSHAuthMethodMissing   = profileerrors.ErrSSHAuthMethodMissing
	ErrSSHIdentityFileInvalid = profileerrors.ErrSSHIdentityFileInvalid
	ErrSSHKnownHostsInvalid   = profileerrors.ErrSSHKnownHostsInvalid
	ErrRemoteHostEmpty        = profileerrors.ErrRemoteHostEmpty
	ErrRemotePortEmpty        = profileerrors.ErrRemotePortEmpty

	// Encryption/Decryption errors
	ErrEncryptionKeyMissing  = profileerrors.ErrEncryptionKeyMissing
	ErrEncryptionKeyMismatch = profileerrors.ErrEncryptionKeyMismatch
	ErrDecryptionFailed      = profileerrors.ErrDecryptionFailed
	ErrEncryptionFailed      = profileerrors.ErrEncryptionFailed

	// Parsing/Loading errors
	ErrParseConfigFailed   = profileerrors.ErrParseConfigFailed
	ErrLoadProfileFailed   = profileerrors.ErrLoadProfileFailed
	ErrReadConfigDirFailed = profileerrors.ErrReadConfigDirFailed
	ErrINIFormatInvalid    = profileerrors.ErrINIFormatInvalid

	// File operations errors
	ErrConfigFileNotFound    = profileerrors.ErrConfigFileNotFound
	ErrCreateConfigDirFailed = profileerrors.ErrCreateConfigDirFailed
	ErrWriteConfigFailed     = profileerrors.ErrWriteConfigFailed
	ErrReadFileFailed        = profileerrors.ErrReadFileFailed
	ErrFileIsDirectory       = profileerrors.ErrFileIsDirectory
	ErrFileNotAccessible     = profileerrors.ErrFileNotAccessible

	// Validation errors
	ErrInputHasLeadingTrailingSpace = profileerrors.ErrInputHasLeadingTrailingSpace
	ErrInputHasControlChars         = profileerrors.ErrInputHasControlChars
	ErrInputHasSpaces               = profileerrors.ErrInputHasSpaces
	ErrInputEmpty                   = profileerrors.ErrInputEmpty
	ErrInputNotNumeric              = profileerrors.ErrInputNotNumeric
	ErrInputOutOfRange              = profileerrors.ErrInputOutOfRange

	// Interactive/Non-interactive mode errors
	ErrNonInteractiveMissingKey      = profileerrors.ErrNonInteractiveMissingKey
	ErrNonInteractiveProfileRequired = profileerrors.ErrNonInteractiveProfileRequired
	ErrCallbackUnavailable           = profileerrors.ErrCallbackUnavailable
)

// =============================================================================
// Error Constructors (untuk error dengan context tambahan)
// =============================================================================

// ProfileNotFoundError creates error dengan nama profile yang tidak ditemukan
func ProfileNotFoundError(name string) error {
	return profileerrors.ProfileNotFoundError(name)
}

// ProfileAlreadyExistsError creates error dengan nama profile yang sudah ada
func ProfileAlreadyExistsError(name string) error {
	return profileerrors.ProfileAlreadyExistsError(name)
}

// ConfigFileNotFoundError creates error dengan path file yang tidak ditemukan
func ConfigFileNotFoundError(path string) error {
	return profileerrors.ConfigFileNotFoundError(path)
}

// DBPortInvalidError creates error dengan port value yang tidak valid
func DBPortInvalidError(port int) error {
	return profileerrors.DBPortInvalidError(port)
}

// SSHPortInvalidError creates error dengan ssh port yang tidak valid
func SSHPortInvalidError(port int) error {
	return profileerrors.SSHPortInvalidError(port)
}

// ReadConfigDirError creates error saat gagal baca direktori config
func ReadConfigDirError(dir string, err error) error {
	return profileerrors.ReadConfigDirError(dir, err)
}

// DecryptionFailedError creates error dengan hint kenapa dekripsi gagal
func DecryptionFailedError(path string, hint string) error {
	return profileerrors.DecryptionFailedError(path, hint)
}

// ParseConfigFailedError creates error saat parsing config gagal
func ParseConfigFailedError(path string) error {
	return profileerrors.ParseConfigFailedError(path)
}

// LoadProfileError wraps underlying error saat load profile
func LoadProfileError(err error) error {
	return profileerrors.LoadProfileError(err)
}

// ConnectionFailedError creates error dengan detail koneksi yang gagal
func ConnectionFailedError(user, host string, port int, err error) error {
	return profileerrors.ConnectionFailedError(user, host, port, err)
}

// SSHTunnelError creates error saat ssh tunnel gagal dibuat
func SSHTunnelError(host string, port int, err error) error {
	return profileerrors.SSHTunnelError(host, port, err)
}

// SSHIdentityFileError creates error untuk masalah identity file
func SSHIdentityFileError(path string, reason string) error {
	return profileerrors.SSHIdentityFileError(path, reason)
}

// ValidationError wraps field name dengan error message
func ValidationError(fieldName string, err error) error {
	return profileerrors.ValidationError(fieldName, err)
}

// InputRangeError creates error untuk input di luar range
func InputRangeError(fieldName string, min, max int, allowZero bool) error {
	return profileerrors.InputRangeError(fieldName, min, max, allowZero)
}

// =============================================================================
// Legacy compatibility functions (untuk backward compatibility dengan consts)
// =============================================================================

// NonInteractiveProfileKeyRequiredError adalah error standar saat mode non-interaktif butuh kunci enkripsi.
func NonInteractiveProfileKeyRequiredError() error {
	return profileerrors.NonInteractiveProfileKeyRequiredError()
}
