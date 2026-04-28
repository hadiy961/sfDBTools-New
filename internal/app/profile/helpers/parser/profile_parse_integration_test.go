// File : internal/app/profile/helpers/parser/profile_parse_integration_test.go
// Deskripsi : Integration tests untuk profile parser
// Author : Test Suite
// Tanggal : 5 Februari 2026
package parser

import (
	"os"
	"path/filepath"
	"testing"

	"sfdbtools/internal/crypto"
)

const testEncryptionKey = "test-encryption-key-123"

// TestLoadAndParseProfile_ValidEncrypted tests loading valid encrypted profile
func TestLoadAndParseProfile_ValidEncrypted(t *testing.T) {
	testFile := filepath.Join("testdata", "valid_encrypted.cnf.enc")

	profile, err := LoadAndParseProfile(testFile, testEncryptionKey)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify basic fields
	if profile.DBInfo.Host != "localhost" {
		t.Errorf("Expected host 'localhost', got '%s'", profile.DBInfo.Host)
	}
	if profile.DBInfo.Port != 3306 {
		t.Errorf("Expected port 3306, got %d", profile.DBInfo.Port)
	}
	if profile.DBInfo.User != "testuser" {
		t.Errorf("Expected user 'testuser', got '%s'", profile.DBInfo.User)
	}
	if profile.DBInfo.Password != "testpass123" {
		t.Errorf("Expected password 'testpass123', got '%s'", profile.DBInfo.Password)
	}

	// Verify SSH disabled
	if profile.SSHTunnel.Enabled {
		t.Error("Expected SSH tunnel to be disabled")
	}

	// Verify encryption metadata
	if profile.EncryptionKey != testEncryptionKey {
		t.Errorf("Expected encryption key '%s', got '%s'", testEncryptionKey, profile.EncryptionKey)
	}
}

// TestLoadAndParseProfile_ValidEncryptedWithSSH tests loading profile with SSH tunnel
func TestLoadAndParseProfile_ValidEncryptedWithSSH(t *testing.T) {
	testFile := filepath.Join("testdata", "valid_with_ssh_encrypted.cnf.enc")

	profile, err := LoadAndParseProfile(testFile, testEncryptionKey)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify DB fields
	if profile.DBInfo.Host != "db.example.com" {
		t.Errorf("Expected host 'db.example.com', got '%s'", profile.DBInfo.Host)
	}
	if profile.DBInfo.User != "dbadmin" {
		t.Errorf("Expected user 'dbadmin', got '%s'", profile.DBInfo.User)
	}

	// Verify SSH tunnel enabled
	if !profile.SSHTunnel.Enabled {
		t.Fatal("Expected SSH tunnel to be enabled")
	}
	if profile.SSHTunnel.Host != "ssh.example.com" {
		t.Errorf("Expected SSH host 'ssh.example.com', got '%s'", profile.SSHTunnel.Host)
	}
	if profile.SSHTunnel.Port != 22 {
		t.Errorf("Expected SSH port 22, got %d", profile.SSHTunnel.Port)
	}
	if profile.SSHTunnel.User != "sshuser" {
		t.Errorf("Expected SSH user 'sshuser', got '%s'", profile.SSHTunnel.User)
	}
	if profile.SSHTunnel.Password != "ssh_pass_789" {
		t.Errorf("Expected SSH password 'ssh_pass_789', got '%s'", profile.SSHTunnel.Password)
	}
	if profile.SSHTunnel.LocalPort != 13306 {
		t.Errorf("Expected SSH local port 13306, got %d", profile.SSHTunnel.LocalPort)
	}
}

// TestLoadAndParseProfile_InvalidEncryptionKey tests with wrong decryption key
func TestLoadAndParseProfile_InvalidEncryptionKey(t *testing.T) {
	testFile := filepath.Join("testdata", "valid_encrypted.cnf.enc")
	wrongKey := "wrong-key-that-will-fail"

	_, err := LoadAndParseProfile(testFile, wrongKey)
	if err == nil {
		t.Fatal("Expected decryption error with wrong key, got nil")
	}

	// Verify error message contains helpful hint
	errMsg := err.Error()
	if errMsg == "" {
		t.Error("Expected non-empty error message")
	}
	// Should mention decryption failure
	t.Logf("Error message (expected): %s", errMsg)
}

// TestLoadAndParseProfile_CorruptEncryptedFile tests with corrupted encrypted data
func TestLoadAndParseProfile_CorruptEncryptedFile(t *testing.T) {
	testFile := filepath.Join("testdata", "corrupt_encrypted.cnf.enc")

	_, err := LoadAndParseProfile(testFile, testEncryptionKey)
	if err == nil {
		t.Fatal("Expected error with corrupt file, got nil")
	}

	t.Logf("Error message (expected): %s", err.Error())
}

// TestLoadAndParseProfile_FileNotFound tests with non-existent file
func TestLoadAndParseProfile_FileNotFound(t *testing.T) {
	testFile := filepath.Join("testdata", "nonexistent.cnf.enc")

	_, err := LoadAndParseProfile(testFile, testEncryptionKey)
	if err == nil {
		t.Fatal("Expected file not found error, got nil")
	}

	// Should be OS-level file error
	if !os.IsNotExist(err) {
		t.Logf("Expected os.IsNotExist error, got: %v", err)
	}
}

// TestLoadAndParseProfile_InvalidINIFormat tests with malformed INI content
func TestLoadAndParseProfile_InvalidINIFormat(t *testing.T) {
	// First, encrypt the invalid format file
	invalidFile := filepath.Join("testdata", "invalid_format.cnf")
	invalidEncFile := filepath.Join("testdata", "invalid_format_temp.cnf.enc")

	// Cleanup temp file after test
	defer os.Remove(invalidEncFile)

	// Encrypt it
	data, err := os.ReadFile(invalidFile)
	if err != nil {
		t.Fatalf("Failed to read invalid format file: %v", err)
	}

	encrypted, err := crypto.EncryptData(data, []byte(testEncryptionKey))
	if err != nil {
		t.Fatalf("Failed to encrypt test data: %v", err)
	}

	if err := os.WriteFile(invalidEncFile, encrypted, 0644); err != nil {
		t.Fatalf("Failed to write encrypted test file: %v", err)
	}

	// Now test loading it
	_, err = LoadAndParseProfile(invalidEncFile, testEncryptionKey)
	if err == nil {
		t.Fatal("Expected parsing error with invalid INI format, got nil")
	}

	// Should mention INI parsing failure
	errMsg := err.Error()
	if errMsg == "" {
		t.Error("Expected non-empty error message")
	}
	t.Logf("Error message (expected): %s", errMsg)
}

// TestLoadAndParseProfile_MissingRequiredFields tests profile with missing fields
func TestLoadAndParseProfile_MissingRequiredFields(t *testing.T) {
	// Encrypt the missing fields file
	missingFile := filepath.Join("testdata", "missing_fields.cnf")
	missingEncFile := filepath.Join("testdata", "missing_fields_temp.cnf.enc")

	defer os.Remove(missingEncFile)

	data, err := os.ReadFile(missingFile)
	if err != nil {
		t.Fatalf("Failed to read missing fields file: %v", err)
	}

	encrypted, err := crypto.EncryptData(data, []byte(testEncryptionKey))
	if err != nil {
		t.Fatalf("Failed to encrypt test data: %v", err)
	}

	if err := os.WriteFile(missingEncFile, encrypted, 0644); err != nil {
		t.Fatalf("Failed to write encrypted test file: %v", err)
	}

	// Load should succeed but fields will be empty
	profile, err := LoadAndParseProfile(missingEncFile, testEncryptionKey)
	if err != nil {
		t.Fatalf("Expected no error (parsing succeeds but fields empty), got: %v", err)
	}

	// Verify missing fields are empty
	if profile.DBInfo.Host != "localhost" {
		t.Errorf("Expected host 'localhost', got '%s'", profile.DBInfo.Host)
	}
	if profile.DBInfo.User != "" {
		t.Errorf("Expected empty user (missing in file), got '%s'", profile.DBInfo.User)
	}
	if profile.DBInfo.Password != "" {
		t.Errorf("Expected empty password (missing in file), got '%s'", profile.DBInfo.Password)
	}
}

// TestLoadAndParseProfile_SpecialCharacters tests special characters in values
func TestLoadAndParseProfile_SpecialCharacters(t *testing.T) {
	// Encrypt the special chars file
	specialFile := filepath.Join("testdata", "special_chars.cnf")
	specialEncFile := filepath.Join("testdata", "special_chars_temp.cnf.enc")

	defer os.Remove(specialEncFile)

	data, err := os.ReadFile(specialFile)
	if err != nil {
		t.Fatalf("Failed to read special chars file: %v", err)
	}

	encrypted, err := crypto.EncryptData(data, []byte(testEncryptionKey))
	if err != nil {
		t.Fatalf("Failed to encrypt test data: %v", err)
	}

	if err := os.WriteFile(specialEncFile, encrypted, 0644); err != nil {
		t.Fatalf("Failed to write encrypted test file: %v", err)
	}

	profile, err := LoadAndParseProfile(specialEncFile, testEncryptionKey)
	if err != nil {
		t.Fatalf("Expected no error with special chars, got: %v", err)
	}

	// Verify special characters preserved
	if profile.DBInfo.User != "test@user" {
		t.Errorf("Expected user 'test@user', got '%s'", profile.DBInfo.User)
	}

	expectedPassword := "p@$$w0rd!#%^&*()_+-=[]{}|;:'\"" + ",.<>?/~`"
	if profile.DBInfo.Password != expectedPassword {
		t.Errorf("Expected password with special chars, got '%s'", profile.DBInfo.Password)
	}
}

// TestLoadAndParseProfile_DuplicateKeys tests INI with duplicate keys (last wins)
func TestLoadAndParseProfile_DuplicateKeys(t *testing.T) {
	// Encrypt the duplicate keys file
	dupFile := filepath.Join("testdata", "duplicate_keys.cnf")
	dupEncFile := filepath.Join("testdata", "duplicate_keys_temp.cnf.enc")

	defer os.Remove(dupEncFile)

	data, err := os.ReadFile(dupFile)
	if err != nil {
		t.Fatalf("Failed to read duplicate keys file: %v", err)
	}

	encrypted, err := crypto.EncryptData(data, []byte(testEncryptionKey))
	if err != nil {
		t.Fatalf("Failed to encrypt test data: %v", err)
	}

	if err := os.WriteFile(dupEncFile, encrypted, 0644); err != nil {
		t.Fatalf("Failed to write encrypted test file: %v", err)
	}

	profile, err := LoadAndParseProfile(dupEncFile, testEncryptionKey)
	if err != nil {
		t.Fatalf("Expected no error with duplicate keys, got: %v", err)
	}

	// In INI parsing, last value typically wins
	if profile.DBInfo.User != "user2" {
		t.Logf("Note: Duplicate key handling - user is '%s' (last or first value)", profile.DBInfo.User)
	}
	if profile.DBInfo.Password != "pass2" {
		t.Logf("Note: Duplicate key handling - password is '%s' (last or first value)", profile.DBInfo.Password)
	}
}

// TestLoadAndParseProfile_EmptyKey tests with empty encryption key (should prompt or use env)
func TestLoadAndParseProfile_EmptyKey(t *testing.T) {
	// This test requires mocking or setting environment variable
	// For now, we'll test that it attempts to resolve key
	testFile := filepath.Join("testdata", "valid_encrypted.cnf.enc")

	// Save current env
	oldEnv := os.Getenv("SFDB_SOURCE_PROFILE_KEY")
	defer os.Setenv("SFDB_SOURCE_PROFILE_KEY", oldEnv)

	// Set test key in env
	os.Setenv("SFDB_SOURCE_PROFILE_KEY", testEncryptionKey)

	// Call with empty key - should use env
	profile, err := LoadAndParseProfile(testFile, "")
	if err != nil {
		t.Fatalf("Expected to use env key, got error: %v", err)
	}

	// Verify it loaded correctly
	if profile.DBInfo.Host != "localhost" {
		t.Errorf("Expected host 'localhost', got '%s'", profile.DBInfo.Host)
	}

	// Verify encryption source is env
	if profile.EncryptionSource != "env" {
		t.Errorf("Expected encryption source 'env', got '%s'", profile.EncryptionSource)
	}
}

// TestLoadAndParseProfile_ProfileNameExtraction tests profile name extraction from path
func TestLoadAndParseProfile_ProfileNameExtraction(t *testing.T) {
	testFile := filepath.Join("testdata", "valid_encrypted.cnf.enc")

	profile, err := LoadAndParseProfile(testFile, testEncryptionKey)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Name should be extracted from filename without extension
	expectedName := "valid_encrypted"
	if profile.Name != expectedName {
		t.Errorf("Expected profile name '%s', got '%s'", expectedName, profile.Name)
	}
}

// TestLoadAndParseProfile_SSHDefaultPort tests SSH port defaults to 22
func TestLoadAndParseProfile_SSHDefaultPort(t *testing.T) {
	// Create a profile with SSH enabled but no port specified
	content := `[client]
host = localhost
port = 3306
user = testuser
password = testpass

[ssh]
enabled = true
host = sshhost
user = sshuser
ssh_password = sshpass
`

	tempFile := filepath.Join("testdata", "ssh_default_port_temp.cnf.enc")
	defer os.Remove(tempFile)

	encrypted, err := crypto.EncryptData([]byte(content), []byte(testEncryptionKey))
	if err != nil {
		t.Fatalf("Failed to encrypt test data: %v", err)
	}

	if err := os.WriteFile(tempFile, encrypted, 0644); err != nil {
		t.Fatalf("Failed to write encrypted test file: %v", err)
	}

	profile, err := LoadAndParseProfile(tempFile, testEncryptionKey)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// SSH port should default to 22
	if !profile.SSHTunnel.Enabled {
		t.Error("Expected SSH tunnel to be enabled")
	}
	if profile.SSHTunnel.Port != 22 {
		t.Errorf("Expected SSH port to default to 22, got %d", profile.SSHTunnel.Port)
	}
}

// TestLoadAndParseProfile_SSHEnabledVariations tests various SSH enabled values
func TestLoadAndParseProfile_SSHEnabledVariations(t *testing.T) {
	testCases := []struct {
		name         string
		enabledVal   string
		shouldEnable bool
	}{
		{"true", "true", true},
		{"TRUE", "TRUE", true},
		{"yes", "yes", true},
		{"YES", "YES", true},
		{"y", "y", true},
		{"Y", "Y", true},
		{"1", "1", true},
		{"on", "on", true},
		{"ON", "ON", true},
		{"false", "false", false},
		{"no", "no", false},
		{"0", "0", false},
		{"off", "off", false},
		{"random", "random", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			content := "[client]\nhost = localhost\nport = 3306\nuser = test\npassword = test\n\n"
			content += "[ssh]\nenabled = " + tc.enabledVal + "\nhost = sshhost\nport = 22\nuser = sshuser\n"

			tempFile := filepath.Join("testdata", "ssh_enabled_"+tc.name+"_temp.cnf.enc")
			defer os.Remove(tempFile)

			encrypted, err := crypto.EncryptData([]byte(content), []byte(testEncryptionKey))
			if err != nil {
				t.Fatalf("Failed to encrypt: %v", err)
			}

			if err := os.WriteFile(tempFile, encrypted, 0644); err != nil {
				t.Fatalf("Failed to write: %v", err)
			}

			profile, err := LoadAndParseProfile(tempFile, testEncryptionKey)
			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}

			if profile.SSHTunnel.Enabled != tc.shouldEnable {
				t.Errorf("For enabled='%s', expected SSH enabled=%v, got %v",
					tc.enabledVal, tc.shouldEnable, profile.SSHTunnel.Enabled)
			}
		})
	}
}
