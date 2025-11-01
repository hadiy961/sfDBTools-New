package types

import (
	"sfDBTools/internal/appconfig"
	"sfDBTools/internal/applog"
)

// Dependencies adalah struct yang menyimpan semua dependensi global aplikasi.
type Dependencies struct {
	Config *appconfig.Config
	Logger applog.Logger
}

// Global variable untuk menyimpan dependensi yang di-inject
var Deps *Dependencies

// CompressionOptions menyimpan opsi kompresi untuk backup.
type CompressionOptions struct {
	CompressOutput bool
	CompressType   string
	CompressLevel  int
}

// EncryptionOptions menyimpan opsi enkripsi untuk backup.
type EncryptionOptions struct {
	EncryptOutput bool
	EncryptKey    string
}
