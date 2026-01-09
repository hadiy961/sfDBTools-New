// File : internal/crypto/model/options.go
// Deskripsi : Model options untuk perintah crypto (CLI)
// Author : Hadiyatna Muflihun
// Tanggal : 9 Januari 2026
// Last Modified : 9 Januari 2026

package model

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
