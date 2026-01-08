package consts

// OpenSSLSaltedHeader adalah header standar OpenSSL untuk format salted payload.
const OpenSSLSaltedHeader = "Salted__"

// PBKDF2 parameters (must match OpenSSL-compatible settings used by this tool).
const (
	PBKDF2Iterations = 100000
	SaltSizeBytes    = 8
)

// EncryptStreamChunkSize adalah ukuran chunk untuk enkripsi streaming.
const EncryptStreamChunkSize = 64 * 1024
