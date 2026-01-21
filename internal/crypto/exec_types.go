// File : internal/crypto/exec_types.go
// Deskripsi : Type definitions untuk CLI executor options
// Author : Hadiyatna Muflihun
// Tanggal : 21 Januari 2026
// Last Modified : 21 Januari 2026

package crypto

// ========================
// CLI Options Structs
// ========================
// Note: Struct ini didefinisikan di sini (bukan di package terpisah) sesuai prinsip:
// "Sedikit penyalinan lebih baik daripada sedikit dependensi."
// Struct ini hanya digunakan untuk CLI operations, tidak perlu package model/ tersendiri.

// Base64EncodeOptions menyimpan opsi untuk base64 encode.
type Base64EncodeOptions struct {
	InputText  string
	OutputPath string
}

// Base64DecodeOptions menyimpan opsi untuk base64 decode.
type Base64DecodeOptions struct {
	InputData  string
	OutputPath string
}

// EncryptFileOptions menyimpan opsi untuk encrypt file.
type EncryptFileOptions struct {
	InputPath  string
	OutputPath string
	Key        string
}

// DecryptFileOptions menyimpan opsi untuk decrypt file.
type DecryptFileOptions struct {
	InputPath  string
	OutputPath string
	Key        string
}

// EncryptTextOptions menyimpan opsi untuk encrypt text.
type EncryptTextOptions struct {
	InputText  string
	OutputPath string
	Key        string
}

// DecryptTextOptions menyimpan opsi untuk decrypt text.
type DecryptTextOptions struct {
	InputData  string
	OutputPath string
	Key        string
}
