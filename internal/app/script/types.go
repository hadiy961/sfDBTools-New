// File : internal/app/script/types.go
// Deskripsi : Type definitions untuk script operations
// Author : Hadiyatna Muflihun
// Tanggal : 21 Januari 2026
// Last Modified : 21 Januari 2026

package script

// ========================
// CLI Options Structs
// ========================
// Note: Struct ini didefinisikan di sini (bukan di package terpisah) sesuai prinsip:
// "Sedikit penyalinan lebih baik daripada sedikit dependensi."

// EncryptOptions menyimpan opsi untuk encrypt bundle.
type EncryptOptions struct {
	FilePath      string
	EncryptionKey string
	Mode          string
	OutputPath    string
	DeleteSource  bool
}

// RunOptions menyimpan opsi untuk run bundle.
type RunOptions struct {
	FilePath      string
	EncryptionKey string
	Args          []string
}

// ExtractOptions menyimpan opsi untuk extract bundle.
type ExtractOptions struct {
	FilePath      string
	EncryptionKey string
	OutDir        string
}

// InfoOptions menyimpan opsi untuk info bundle.
type InfoOptions struct {
	FilePath      string
	EncryptionKey string
}

// ========================
// Internal Types
// ========================

// manifest represents bundle metadata stored in .sftools-manifest.json
type manifest struct {
	Version    int    `json:"version"`
	Entrypoint string `json:"entrypoint"`
	CreatedAt  string `json:"created_at"`
	Mode       string `json:"mode,omitempty"`
	RootDir    string `json:"root_dir,omitempty"`
}

// BundleInfo represents metadata information about a script bundle.
type BundleInfo struct {
	Version    int      `json:"version"`
	Entrypoint string   `json:"entrypoint"`
	CreatedAt  string   `json:"created_at"`
	Mode       string   `json:"mode"`
	RootDir    string   `json:"root_dir"`
	Scripts    []string `json:"scripts"`
	FileCount  int      `json:"file_count"`
}

// ========================
// Constants
// ========================

const (
	// manifestFilename adalah nama file manifest dalam bundle
	manifestFilename = ".sftools-manifest.json"

	// bundleVersion adalah versi format bundle
	bundleVersion = 1
)
