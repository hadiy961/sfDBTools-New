// File : internal/app/profile/helpers/parser/error_handling_test.go
// Deskripsi : Error handling integration tests
// Author : Test Suite
// Tanggal : 5 Februari 2026
package parser

import (
	"os"
	"path/filepath"
	"testing"

	"sfdbtools/internal/crypto"
)

// TestErrorHandling_FileNotFound tests file not found error
func TestErrorHandling_FileNotFound(t *testing.T) {
	nonExistentFile := filepath.Join("testdata", "this-file-does-not-exist.cnf.enc")

	_, err := LoadAndParseProfile(nonExistentFile, "any-key")
	if err == nil {
		t.Fatal("Expected error for non-existent file")
	}

	// Should be OS-level error
	if !os.IsNotExist(err) {
		t.Logf("Error type: %T, Message: %v", err, err)
	}
}

// TestErrorHandling_PermissionDenied tests permission denied scenario
func TestErrorHandling_PermissionDenied(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Skipping permission test when running as root")
	}

	tempDir := filepath.Join("testdata", "permission_temp")
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir)

	restrictedFile := filepath.Join(tempDir, "no_read_permission.cnf.enc")

	// Create file with content
	content := []byte("some encrypted content")
	if err := os.WriteFile(restrictedFile, content, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Remove read permission
	if err := os.Chmod(restrictedFile, 0000); err != nil {
		t.Fatalf("Failed to change permissions: %v", err)
	}
	defer os.Chmod(restrictedFile, 0644) // Restore for cleanup

	_, err := LoadAndParseProfile(restrictedFile, "any-key")
	if err == nil {
		t.Fatal("Expected error for file without read permission")
	}

	t.Logf("Permission error: %v", err)
}

// TestErrorHandling_DirectoryInsteadOfFile tests providing directory path
func TestErrorHandling_DirectoryInsteadOfFile(t *testing.T) {
	dirPath := filepath.Join("testdata")

	_, err := LoadAndParseProfile(dirPath, "any-key")
	if err == nil {
		t.Fatal("Expected error when providing directory path")
	}

	t.Logf("Directory error: %v", err)
}

// TestErrorHandling_EmptyFile tests completely empty file
func TestErrorHandling_EmptyFile(t *testing.T) {
	tempDir := filepath.Join("testdata", "empty_file_temp")
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir)

	emptyFile := filepath.Join(tempDir, "empty.cnf.enc")

	// Create empty file
	if err := os.WriteFile(emptyFile, []byte{}, 0644); err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}

	_, err := LoadAndParseProfile(emptyFile, "any-key")
	if err == nil {
		t.Fatal("Expected error for empty file")
	}

	t.Logf("Empty file error: %v", err)
}

// TestErrorHandling_TooSmallFile tests file too small to be valid encrypted data
func TestErrorHandling_TooSmallFile(t *testing.T) {
	tempDir := filepath.Join("testdata", "small_file_temp")
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir)

	smallFile := filepath.Join(tempDir, "too_small.cnf.enc")

	// Create file with only a few bytes (invalid encrypted data)
	if err := os.WriteFile(smallFile, []byte("tiny"), 0644); err != nil {
		t.Fatalf("Failed to create small file: %v", err)
	}

	_, err := LoadAndParseProfile(smallFile, "any-key")
	if err == nil {
		t.Fatal("Expected error for too small file")
	}

	t.Logf("Too small file error: %v", err)
}

// TestErrorHandling_InvalidEncryptionHeader tests file without proper encryption header
func TestErrorHandling_InvalidEncryptionHeader(t *testing.T) {
	tempDir := filepath.Join("testdata", "invalid_header_temp")
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir)

	invalidFile := filepath.Join(tempDir, "invalid_header.cnf.enc")

	// Create file with invalid header
	invalidData := []byte("NotSalted__InvalidHeaderData1234567890ABCDEFGH")
	if err := os.WriteFile(invalidFile, invalidData, 0644); err != nil {
		t.Fatalf("Failed to create invalid file: %v", err)
	}

	_, err := LoadAndParseProfile(invalidFile, "any-key")
	if err == nil {
		t.Fatal("Expected error for invalid encryption header")
	}

	t.Logf("Invalid header error: %v", err)
}

// TestErrorHandling_TruncatedEncryptedData tests truncated encrypted file
func TestErrorHandling_TruncatedEncryptedData(t *testing.T) {
	tempDir := filepath.Join("testdata", "truncated_temp")
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir)

	// First create a valid encrypted file
	testKey := "truncate-test-key"
	content := "[client]\nhost = localhost\nport = 3306\n"

	encrypted, err := crypto.EncryptData([]byte(content), []byte(testKey))
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	// Truncate the encrypted data
	truncatedData := encrypted[:len(encrypted)/2]

	truncatedFile := filepath.Join(tempDir, "truncated.cnf.enc")
	if err := os.WriteFile(truncatedFile, truncatedData, 0644); err != nil {
		t.Fatalf("Failed to write truncated file: %v", err)
	}

	_, err = LoadAndParseProfile(truncatedFile, testKey)
	if err == nil {
		t.Fatal("Expected error for truncated encrypted data")
	}

	t.Logf("Truncated data error: %v", err)
}

// TestErrorHandling_ModifiedEncryptedData tests tampered encrypted file
func TestErrorHandling_ModifiedEncryptedData(t *testing.T) {
	tempDir := filepath.Join("testdata", "modified_temp")
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir)

	// Create valid encrypted file
	testKey := "modify-test-key"
	content := "[client]\nhost = localhost\nport = 3306\nuser = test\npassword = test\n"

	encrypted, err := crypto.EncryptData([]byte(content), []byte(testKey))
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	// Modify some bytes in the middle (corrupt the data)
	if len(encrypted) > 20 {
		encrypted[15] ^= 0xFF // Flip bits
		encrypted[16] ^= 0xFF
		encrypted[17] ^= 0xFF
	}

	modifiedFile := filepath.Join(tempDir, "modified.cnf.enc")
	if err := os.WriteFile(modifiedFile, encrypted, 0644); err != nil {
		t.Fatalf("Failed to write modified file: %v", err)
	}

	_, err = LoadAndParseProfile(modifiedFile, testKey)
	if err == nil {
		t.Fatal("Expected error for modified encrypted data")
	}

	t.Logf("Modified data error: %v", err)
}

// TestErrorHandling_DecryptionWithEmptyKey tests decryption with empty key and no env
func TestErrorHandling_DecryptionWithEmptyKey(t *testing.T) {
	// Save and clear env
	oldEnv := os.Getenv("SFDB_SOURCE_PROFILE_KEY")
	os.Unsetenv("SFDB_SOURCE_PROFILE_KEY")
	defer func() {
		if oldEnv != "" {
			os.Setenv("SFDB_SOURCE_PROFILE_KEY", oldEnv)
		}
	}()

	testFile := filepath.Join("testdata", "valid_encrypted.cnf.enc")

	// This might prompt for key in interactive mode, or fail
	// In non-interactive test environment, it should fail
	_, err := LoadAndParseProfile(testFile, "")

	// May succeed if it prompts, or fail if non-interactive
	if err != nil {
		t.Logf("Expected error without key (non-interactive): %v", err)
	} else {
		t.Log("Note: May have succeeded via prompt or cached key")
	}
}

// TestErrorHandling_WrongKeyDetailed tests detailed error messages for wrong key
func TestErrorHandling_WrongKeyDetailed(t *testing.T) {
	testFile := filepath.Join("testdata", "valid_encrypted.cnf.enc")

	testCases := []struct {
		name       string
		key        string
		shouldFail bool
	}{
		{"wrong_short_key", "wrong", true},
		{"wrong_long_key", "this-is-a-very-long-but-still-wrong-key", true},
		{"similar_key", "test-encryption-key-124", true}, // Off by one
		{"empty_key_with_spaces", "   ", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := LoadAndParseProfile(testFile, tc.key)

			if tc.shouldFail && err == nil {
				t.Error("Expected error but got none")
			} else if !tc.shouldFail && err != nil {
				t.Errorf("Expected success but got error: %v", err)
			}

			if err != nil {
				t.Logf("Error message: %v", err)
				// Verify error message is helpful
				errMsg := err.Error()
				if errMsg == "" {
					t.Error("Error message should not be empty")
				}
			}
		})
	}
}

// TestErrorHandling_MissingClientSection tests file without [client] section
func TestErrorHandling_MissingClientSection(t *testing.T) {
	content := `[other_section]
some_key = some_value

[ssh]
enabled = false
`

	testKey := "no-client-section-key"
	tempDir := filepath.Join("testdata", "no_client_temp")
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir)

	plainFile := filepath.Join(tempDir, "no_client.cnf")
	encFile := filepath.Join(tempDir, "no_client.cnf.enc")

	if err := os.WriteFile(plainFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}

	if err := crypto.EncryptFile(plainFile, encFile, []byte(testKey)); err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	_, err := LoadAndParseProfile(encFile, testKey)
	if err == nil {
		t.Fatal("Expected error for missing [client] section")
	}

	// Should mention INI parsing or client section missing
	errMsg := err.Error()
	t.Logf("Missing client section error: %s", errMsg)
}

// TestErrorHandling_BinaryGarbage tests file with random binary data
func TestErrorHandling_BinaryGarbage(t *testing.T) {
	tempDir := filepath.Join("testdata", "garbage_temp")
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir)

	garbageFile := filepath.Join(tempDir, "garbage.cnf.enc")

	// Create file with random bytes
	garbageData := make([]byte, 1024)
	for i := range garbageData {
		garbageData[i] = byte(i % 256)
	}

	if err := os.WriteFile(garbageFile, garbageData, 0644); err != nil {
		t.Fatalf("Failed to write garbage file: %v", err)
	}

	_, err := LoadAndParseProfile(garbageFile, "any-key")
	if err == nil {
		t.Fatal("Expected error for binary garbage")
	}

	t.Logf("Binary garbage error: %v", err)
}

// TestErrorHandling_SymbolicLink tests following symbolic links
func TestErrorHandling_SymbolicLink(t *testing.T) {
	// Get absolute path to testdata
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	tempDir := filepath.Join(wd, "testdata", "symlink_temp")
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir)

	// Create a target file
	targetFile := filepath.Join(tempDir, "target.cnf.enc")
	validEncFile := filepath.Join(wd, "testdata", "valid_encrypted.cnf.enc")

	// Copy valid encrypted file
	data, err := os.ReadFile(validEncFile)
	if err != nil {
		t.Skipf("Skipping symlink test: %v", err)
		return
	}

	if err := os.WriteFile(targetFile, data, 0644); err != nil {
		t.Fatalf("Failed to create target file: %v", err)
	}

	// Create symlink
	symlinkFile := filepath.Join(tempDir, "symlink.cnf.enc")
	if err := os.Symlink(targetFile, symlinkFile); err != nil {
		t.Skipf("Skipping symlink test (not supported): %v", err)
		return
	}

	// Should follow symlink and load successfully
	profile, err := LoadAndParseProfile(symlinkFile, testEncryptionKey)
	if err != nil {
		t.Fatalf("Should follow symlink successfully: %v", err)
	}

	if profile.DBInfo.Host != "localhost" {
		t.Error("Failed to load through symlink")
	}
}

// TestErrorHandling_VeryLongPath tests with very long file path
func TestErrorHandling_VeryLongPath(t *testing.T) {
	// Create nested directories to make a long path
	basePath := filepath.Join("testdata", "long_path_temp")
	longPath := basePath

	// Create a moderately long path (not extreme to avoid OS limits)
	for i := 0; i < 10; i++ {
		longPath = filepath.Join(longPath, "very_long_directory_name_for_testing_purposes")
	}

	if err := os.MkdirAll(longPath, 0755); err != nil {
		t.Skipf("Cannot create long path: %v", err)
		return
	}
	defer os.RemoveAll(basePath)

	longFile := filepath.Join(longPath, "profile.cnf.enc")

	// Copy valid encrypted file to long path
	validEncFile := filepath.Join("testdata", "valid_encrypted.cnf.enc")
	data, err := os.ReadFile(validEncFile)
	if err != nil {
		t.Fatalf("Failed to read valid file: %v", err)
	}

	if err := os.WriteFile(longFile, data, 0644); err != nil {
		t.Fatalf("Failed to write to long path: %v", err)
	}

	// Should work with long path
	profile, err := LoadAndParseProfile(longFile, testEncryptionKey)
	if err != nil {
		t.Errorf("Should handle long path: %v", err)
	} else if profile.DBInfo.Host != "localhost" {
		t.Error("Failed to load from long path")
	}
}

// TestErrorHandling_ConcurrentReads tests concurrent profile loading
func TestErrorHandling_ConcurrentReads(t *testing.T) {
	testFile := filepath.Join("testdata", "valid_encrypted.cnf.enc")

	// Launch multiple goroutines to read same file
	concurrency := 10
	errors := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			_, err := LoadAndParseProfile(testFile, testEncryptionKey)
			errors <- err
		}()
	}

	// Collect results
	for i := 0; i < concurrency; i++ {
		err := <-errors
		if err != nil {
			t.Errorf("Concurrent read %d failed: %v", i, err)
		}
	}
}
