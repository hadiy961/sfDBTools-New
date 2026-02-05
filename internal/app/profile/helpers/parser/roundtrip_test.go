// File : internal/app/profile/helpers/parser/roundtrip_test.go
// Deskripsi : Round-trip encryption/decryption tests
// Author : Test Suite
// Tanggal : 5 Februari 2026
package parser

import (
	"os"
	"path/filepath"
	"testing"

	"sfdbtools/internal/crypto"
)

// TestRoundTrip_EncryptDecrypt tests save → load → verify data integrity
func TestRoundTrip_EncryptDecrypt(t *testing.T) {
	// Original profile content
	originalContent := `[client]
host = roundtrip.test.com
port = 3307
user = roundtripuser
password = roundtrippass123

[ssh]
enabled = true
host = ssh.roundtrip.com
port = 2222
user = sshuser
ssh_password = sshpass456
local_port = 13307
`

	testKey := "roundtrip-test-key-xyz"
	tempDir := filepath.Join("testdata", "roundtrip_temp")
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir)

	plainFile := filepath.Join(tempDir, "original.cnf")
	encFile := filepath.Join(tempDir, "encrypted.cnf.enc")

	// Step 1: Write original content
	if err := os.WriteFile(plainFile, []byte(originalContent), 0644); err != nil {
		t.Fatalf("Failed to write original file: %v", err)
	}

	// Step 2: Encrypt
	if err := crypto.EncryptFile(plainFile, encFile, []byte(testKey)); err != nil {
		t.Fatalf("Failed to encrypt file: %v", err)
	}

	// Step 3: Load and parse encrypted file
	profile, err := LoadAndParseProfile(encFile, testKey)
	if err != nil {
		t.Fatalf("Failed to load encrypted profile: %v", err)
	}

	// Step 4: Verify all data preserved
	if profile.DBInfo.Host != "roundtrip.test.com" {
		t.Errorf("Host mismatch: expected 'roundtrip.test.com', got '%s'", profile.DBInfo.Host)
	}
	if profile.DBInfo.Port != 3307 {
		t.Errorf("Port mismatch: expected 3307, got %d", profile.DBInfo.Port)
	}
	if profile.DBInfo.User != "roundtripuser" {
		t.Errorf("User mismatch: expected 'roundtripuser', got '%s'", profile.DBInfo.User)
	}
	if profile.DBInfo.Password != "roundtrippass123" {
		t.Errorf("Password mismatch: expected 'roundtrippass123', got '%s'", profile.DBInfo.Password)
	}

	// Verify SSH data
	if !profile.SSHTunnel.Enabled {
		t.Error("SSH tunnel should be enabled")
	}
	if profile.SSHTunnel.Host != "ssh.roundtrip.com" {
		t.Errorf("SSH host mismatch: expected 'ssh.roundtrip.com', got '%s'", profile.SSHTunnel.Host)
	}
	if profile.SSHTunnel.Port != 2222 {
		t.Errorf("SSH port mismatch: expected 2222, got %d", profile.SSHTunnel.Port)
	}
	if profile.SSHTunnel.User != "sshuser" {
		t.Errorf("SSH user mismatch: expected 'sshuser', got '%s'", profile.SSHTunnel.User)
	}
	if profile.SSHTunnel.Password != "sshpass456" {
		t.Errorf("SSH password mismatch: expected 'sshpass456', got '%s'", profile.SSHTunnel.Password)
	}
	if profile.SSHTunnel.LocalPort != 13307 {
		t.Errorf("SSH local port mismatch: expected 13307, got %d", profile.SSHTunnel.LocalPort)
	}
}

// TestRoundTrip_DifferentKeys tests encryption with different keys
func TestRoundTrip_DifferentKeys(t *testing.T) {
	content := `[client]
host = localhost
port = 3306
user = testuser
password = testpass
`

	tempDir := filepath.Join("testdata", "multikey_temp")
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir)

	keys := []string{
		"key-one",
		"key-two-longer",
		"key-three-with-special-chars-!@#$%",
		"very-long-key-with-many-characters-to-test-key-derivation-process",
	}

	for i, key := range keys {
		plainFile := filepath.Join(tempDir, "profile.cnf")
		encFile := filepath.Join(tempDir, "profile_"+string(rune('a'+i))+".cnf.enc")

		// Write content
		if err := os.WriteFile(plainFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write file for key %d: %v", i, err)
		}

		// Encrypt with specific key
		if err := crypto.EncryptFile(plainFile, encFile, []byte(key)); err != nil {
			t.Fatalf("Failed to encrypt with key %d: %v", i, err)
		}

		// Load with correct key
		profile, err := LoadAndParseProfile(encFile, key)
		if err != nil {
			t.Fatalf("Failed to load with correct key %d: %v", i, err)
		}

		if profile.DBInfo.Host != "localhost" {
			t.Errorf("Key %d: Host mismatch", i)
		}

		// Try loading with wrong key
		wrongKey := "wrong-key-" + key
		_, err = LoadAndParseProfile(encFile, wrongKey)
		if err == nil {
			t.Errorf("Key %d: Should fail with wrong key", i)
		}
	}
}

// TestRoundTrip_LargeProfileData tests with large profile content
func TestRoundTrip_LargeProfileData(t *testing.T) {
	// Create a large profile with many comments and whitespace
	content := `[client]
host = large-profile-test.example.com
port = 3306
user = user_with_very_long_username_that_exceeds_normal_length
password = password_with_many_special_characters_!@#$%^&*()_+-=[]{}|;:'"<>?,.~/`

	// Add large comments section
	for i := 0; i < 100; i++ {
		content += "\n# This is comment line number " + string(rune('0'+i%10))
	}

	content += `

[ssh]
enabled = true
host = ssh-server-with-long-hostname.example.com
port = 22
user = sshuser_with_long_name
ssh_password = ssh_password_with_many_chars_123456789
identity_file = /very/long/path/to/identity/file/that/tests/path/handling/id_rsa
local_port = 13306
`

	testKey := "large-profile-key"
	tempDir := filepath.Join("testdata", "large_temp")
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir)

	plainFile := filepath.Join(tempDir, "large.cnf")
	encFile := filepath.Join(tempDir, "large.cnf.enc")

	if err := os.WriteFile(plainFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write large file: %v", err)
	}

	if err := crypto.EncryptFile(plainFile, encFile, []byte(testKey)); err != nil {
		t.Fatalf("Failed to encrypt large file: %v", err)
	}

	profile, err := LoadAndParseProfile(encFile, testKey)
	if err != nil {
		t.Fatalf("Failed to load large profile: %v", err)
	}

	// Verify data integrity
	if profile.DBInfo.Host != "large-profile-test.example.com" {
		t.Error("Large profile host mismatch")
	}
	if profile.DBInfo.User != "user_with_very_long_username_that_exceeds_normal_length" {
		t.Error("Large profile user mismatch")
	}
	if profile.SSHTunnel.IdentityFile != "/very/long/path/to/identity/file/that/tests/path/handling/id_rsa" {
		t.Error("Large profile SSH identity file mismatch")
	}
}

// TestRoundTrip_EmptyValues tests round-trip with empty values
func TestRoundTrip_EmptyValues(t *testing.T) {
	content := `[client]
host = 
port = 3306
user = 
password = 
`

	testKey := "empty-values-key"
	tempDir := filepath.Join("testdata", "empty_temp")
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir)

	plainFile := filepath.Join(tempDir, "empty.cnf")
	encFile := filepath.Join(tempDir, "empty.cnf.enc")

	if err := os.WriteFile(plainFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	if err := crypto.EncryptFile(plainFile, encFile, []byte(testKey)); err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	profile, err := LoadAndParseProfile(encFile, testKey)
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	// Empty values should be preserved as empty strings
	if profile.DBInfo.Host != "" {
		t.Errorf("Expected empty host, got '%s'", profile.DBInfo.Host)
	}
	if profile.DBInfo.User != "" {
		t.Errorf("Expected empty user, got '%s'", profile.DBInfo.User)
	}
	if profile.DBInfo.Password != "" {
		t.Errorf("Expected empty password, got '%s'", profile.DBInfo.Password)
	}
	if profile.DBInfo.Port != 3306 {
		t.Errorf("Expected port 3306, got %d", profile.DBInfo.Port)
	}
}

// TestRoundTrip_MultipleEncryptionDecryption tests encrypting multiple times
func TestRoundTrip_MultipleEncryptionDecryption(t *testing.T) {
	content := `[client]
host = localhost
port = 3306
user = testuser
password = testpass
`

	testKey := "multi-roundtrip-key"
	tempDir := filepath.Join("testdata", "multi_temp")
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir)

	plainFile := filepath.Join(tempDir, "original.cnf")
	
	// Write original
	if err := os.WriteFile(plainFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}

	// Encrypt and decrypt multiple times
	for i := 0; i < 5; i++ {
		encFile := filepath.Join(tempDir, "encrypted_"+string(rune('0'+i))+".cnf.enc")
		
		// Encrypt
		if err := crypto.EncryptFile(plainFile, encFile, []byte(testKey)); err != nil {
			t.Fatalf("Round %d: Failed to encrypt: %v", i, err)
		}

		// Load
		profile, err := LoadAndParseProfile(encFile, testKey)
		if err != nil {
			t.Fatalf("Round %d: Failed to load: %v", i, err)
		}

		// Verify
		if profile.DBInfo.Host != "localhost" {
			t.Errorf("Round %d: Host mismatch", i)
		}
		if profile.DBInfo.User != "testuser" {
			t.Errorf("Round %d: User mismatch", i)
		}
		if profile.DBInfo.Password != "testpass" {
			t.Errorf("Round %d: Password mismatch", i)
		}
	}
}

// TestRoundTrip_UnicodeContent tests with Unicode characters
func TestRoundTrip_UnicodeContent(t *testing.T) {
	content := `[client]
host = データベース.example.com
port = 3306
user = 用户名
password = пароль123密码

[ssh]
enabled = false
`

	testKey := "unicode-key-キー"
	tempDir := filepath.Join("testdata", "unicode_temp")
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir)

	plainFile := filepath.Join(tempDir, "unicode.cnf")
	encFile := filepath.Join(tempDir, "unicode.cnf.enc")

	if err := os.WriteFile(plainFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}

	if err := crypto.EncryptFile(plainFile, encFile, []byte(testKey)); err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	profile, err := LoadAndParseProfile(encFile, testKey)
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	// Verify Unicode preserved
	if profile.DBInfo.Host != "データベース.example.com" {
		t.Errorf("Unicode host mismatch: got '%s'", profile.DBInfo.Host)
	}
	if profile.DBInfo.User != "用户名" {
		t.Errorf("Unicode user mismatch: got '%s'", profile.DBInfo.User)
	}
	if profile.DBInfo.Password != "пароль123密码" {
		t.Errorf("Unicode password mismatch: got '%s'", profile.DBInfo.Password)
	}
}
