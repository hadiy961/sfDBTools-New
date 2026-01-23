// File : internal/crypto/core/constants.go
// Deskripsi : Konstanta untuk crypto operations
// Author : Hadiyatna Muflihun
// Tanggal : 8 Januari 2026
// Last Modified : 8 Januari 2026
package core

// OpenSSL compatibility constants
const (
	// OpenSSLSaltedHeader adalah header standar OpenSSL untuk format salted payload
	OpenSSLSaltedHeader = "Salted__"

	// SaltSizeBytes adalah ukuran salt dalam bytes (OpenSSL compatible)
	SaltSizeBytes = 8

	// PBKDF2Iterations adalah jumlah iterasi untuk PBKDF2 key derivation
	// Updated to 600,000 per OWASP recommendation (2023)
	PBKDF2Iterations = 600000
)

// Streaming encryption constants
const (
	// StreamChunkSize adalah ukuran chunk untuk streaming encryption (64KB)
	StreamChunkSize = 64 * 1024
)

// File security constants
const (
	// SecureFilePermission adalah permission default untuk crypto file output.
	// 0600 = owner read/write only, no group/other access (secure by default)
	SecureFilePermission = 0600
)
