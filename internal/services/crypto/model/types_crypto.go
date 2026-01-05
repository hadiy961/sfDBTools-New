package types

// Base64EncodeOptions menyimpan opsi untuk base64 encode
type Base64EncodeOptions struct {
	InputText  string
	OutputPath string
}

// Base64DecodeOptions menyimpan opsi untuk base64 decode
type Base64DecodeOptions struct {
	InputData  string
	OutputPath string
}

// EncryptFileOptions menyimpan opsi untuk enkripsi file
type EncryptFileOptions struct {
	InputPath  string
	OutputPath string
	Key        string
}

// DecryptFileOptions menyimpan opsi untuk dekripsi file
type DecryptFileOptions struct {
	InputPath  string
	OutputPath string
	Key        string
}

// EncryptTextOptions menyimpan opsi untuk enkripsi teks
type EncryptTextOptions struct {
	InputText  string
	OutputPath string
	Key        string
}

// DecryptTextOptions menyimpan opsi untuk dekripsi teks
type DecryptTextOptions struct {
	InputData  string
	OutputPath string
	Key        string
}
