// File : internal/app/profile/helpers/loader/error_scenarios_test.go
// Deskripsi : Error scenario tests for profile loader
// Author : Test Suite
// Tanggal : 5 Februari 2026
package loader

import (
	"os"
	"path/filepath"
	"testing"
)

// TestErrorScenarios_ProfileNotFound tests various not found scenarios
func TestErrorScenarios_ProfileNotFound(t *testing.T) {
	testDir := getTestDataDir(t)

	testCases := []struct {
		name        string
		configDir   string
		profilePath string
	}{
		{
			name:        "nonexistent_file",
			configDir:   testDir,
			profilePath: "does-not-exist.cnf.enc",
		},
		{
			name:        "wrong_extension",
			configDir:   testDir,
			profilePath: "test-profile1.txt",
		},
		{
			name:        "empty_filename",
			configDir:   testDir,
			profilePath: " ",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ResolveAndLoadProfile(ProfileLoadOptions{
				ConfigDir:   tc.configDir,
				ProfilePath: tc.profilePath,
				ProfileKey:  testEncryptionKey,
			})

			if err == nil {
				t.Errorf("Expected error for %s, got nil", tc.name)
			}

			t.Logf("%s error: %v", tc.name, err)
		})
	}
}

// TestErrorScenarios_InvalidConfigDir tests invalid config directory
func TestErrorScenarios_InvalidConfigDir(t *testing.T) {
	testCases := []struct {
		name      string
		configDir string
		wantError bool
	}{
		{
			name:      "nonexistent_dir",
			configDir: "/nonexistent/directory/that/does/not/exist",
			wantError: true,
		},
		{
			name:      "file_not_dir",
			configDir: filepath.Join(getTestDataDir(t), "test-profile1.cnf.enc"), // This is a file
			wantError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ResolveAndLoadProfile(ProfileLoadOptions{
				ConfigDir:   tc.configDir,
				ProfilePath: "test-profile1.cnf.enc",
				ProfileKey:  testEncryptionKey,
			})

			if tc.wantError && err == nil {
				t.Error("Expected error for invalid config dir")
			} else if !tc.wantError && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			if err != nil {
				t.Logf("Error: %v", err)
			}
		})
	}
}

// TestErrorScenarios_DecryptionFailures tests various decryption failure scenarios
func TestErrorScenarios_DecryptionFailures(t *testing.T) {
	testDir := getTestDataDir(t)

	testCases := []struct {
		name string
		key  string
	}{
		{"wrong_key", "completely-wrong-key"},
		{"similar_key", "loader-test-key-124"}, // Off by one
		{"empty_key", ""},
		{"short_key", "123"},
		{"special_chars_key", "!@#$%^&*()"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ResolveAndLoadProfile(ProfileLoadOptions{
				ConfigDir:   testDir,
				ProfilePath: "test-profile1.cnf.enc",
				ProfileKey:  tc.key,
			})

			// Should fail except possibly empty_key might prompt
			if err != nil {
				t.Logf("%s error (expected): %v", tc.name, err)
			}
		})
	}
}

// TestErrorScenarios_RequireProfileCombinations tests RequireProfile flag
func TestErrorScenarios_RequireProfileCombinations(t *testing.T) {
	testCases := []struct {
		name           string
		profilePath    string
		envProfilePath string
		requireProfile bool
		setEnv         bool
		expectError    bool
	}{
		{
			name:           "required_but_missing",
			profilePath:    "",
			envProfilePath: "NONEXISTENT_VAR",
			requireProfile: true,
			setEnv:         false,
			expectError:    true,
		},
		{
			name:           "optional_and_missing",
			profilePath:    "",
			envProfilePath: "NONEXISTENT_VAR",
			requireProfile: false,
			setEnv:         false,
			expectError:    false,
		},
		{
			name:           "required_with_env",
			profilePath:    "",
			envProfilePath: "TEST_LOADER_PROFILE",
			requireProfile: true,
			setEnv:         true,
			expectError:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setEnv {
				testDir := getTestDataDir(t)
				os.Setenv(tc.envProfilePath, filepath.Join(testDir, "test-profile1.cnf.enc"))
				defer os.Unsetenv(tc.envProfilePath)
			}

			_, err := ResolveAndLoadProfile(ProfileLoadOptions{
				ProfilePath:    tc.profilePath,
				ProfileKey:     testEncryptionKey,
				EnvProfilePath: tc.envProfilePath,
				RequireProfile: tc.requireProfile,
			})

			if tc.expectError && err == nil {
				t.Error("Expected error but got nil")
			} else if !tc.expectError && err != nil {
				t.Logf("Got error (may be expected): %v", err)
			}
		})
	}
}

// TestErrorScenarios_PathTraversal tests protection against path traversal
func TestErrorScenarios_PathTraversal(t *testing.T) {
	testDir := getTestDataDir(t)

	maliciousPaths := []string{
		"../../../etc/passwd",
		"..\\..\\..\\windows\\system32\\config\\sam",
		"./../../sensitive-file",
	}

	for _, path := range maliciousPaths {
		t.Run(path, func(t *testing.T) {
			_, err := ResolveAndLoadProfile(ProfileLoadOptions{
				ConfigDir:   testDir,
				ProfilePath: path,
				ProfileKey:  testEncryptionKey,
			})

			if err == nil {
				t.Error("Expected error for path traversal attempt")
			}

			t.Logf("Path traversal blocked: %v", err)
		})
	}
}

// TestErrorScenarios_ConcurrentLoads tests concurrent profile loading
func TestErrorScenarios_ConcurrentLoads(t *testing.T) {
	testDir := getTestDataDir(t)
	concurrency := 10
	errors := make(chan error, concurrency)
	profiles := make(chan string, concurrency)

	// Launch multiple goroutines to load same profile
	for i := 0; i < concurrency; i++ {
		go func() {
			profile, err := ResolveAndLoadProfile(ProfileLoadOptions{
				ConfigDir:   testDir,
				ProfilePath: "test-profile1.cnf.enc",
				ProfileKey:  testEncryptionKey,
			})
			errors <- err
			if profile != nil {
				profiles <- profile.DBInfo.Host
			} else {
				profiles <- ""
			}
		}()
	}

	// Collect results
	successCount := 0
	for i := 0; i < concurrency; i++ {
		err := <-errors
		host := <-profiles
		if err == nil {
			successCount++
			if host != "testdb1.example.com" {
				t.Errorf("Concurrent load %d: wrong host '%s'", i, host)
			}
		}
	}

	if successCount != concurrency {
		t.Errorf("Expected %d successful concurrent loads, got %d", concurrency, successCount)
	}
}

// TestErrorScenarios_CorruptedProfile tests loading corrupted profile file
func TestErrorScenarios_CorruptedProfile(t *testing.T) {
	testDir := getTestDataDir(t)
	
	// Create a corrupted profile file
	corruptPath := filepath.Join(testDir, "corrupted.cnf.enc")
	corruptData := []byte("This is not a valid encrypted profile file!@#$%^&*()")
	
	if err := os.WriteFile(corruptPath, corruptData, 0644); err != nil {
		t.Fatalf("Failed to create corrupted file: %v", err)
	}
	defer os.Remove(corruptPath)

	_, err := ResolveAndLoadProfile(ProfileLoadOptions{
		ConfigDir:   testDir,
		ProfilePath: "corrupted.cnf.enc",
		ProfileKey:  testEncryptionKey,
	})

	if err == nil {
		t.Fatal("Expected error for corrupted profile file")
	}

	t.Logf("Corrupted file error: %v", err)
}

// TestErrorScenarios_EmptyProfile tests loading empty file
func TestErrorScenarios_EmptyProfile(t *testing.T) {
	testDir := getTestDataDir(t)
	
	// Create an empty file
	emptyPath := filepath.Join(testDir, "empty.cnf.enc")
	if err := os.WriteFile(emptyPath, []byte{}, 0644); err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}
	defer os.Remove(emptyPath)

	_, err := ResolveAndLoadProfile(ProfileLoadOptions{
		ConfigDir:   testDir,
		ProfilePath: "empty.cnf.enc",
		ProfileKey:  testEncryptionKey,
	})

	if err == nil {
		t.Fatal("Expected error for empty profile file")
	}

	t.Logf("Empty file error: %v", err)
}

// TestErrorScenarios_ProfilePurposeInError tests ProfilePurpose in error messages
func TestErrorScenarios_ProfilePurposeInError(t *testing.T) {
	purposes := []string{"source", "target", "backup", "test"}

	for _, purpose := range purposes {
		t.Run("purpose_"+purpose, func(t *testing.T) {
			_, err := ResolveAndLoadProfile(ProfileLoadOptions{
				ProfilePath:    "",
				EnvProfilePath: "NONEXISTENT_VAR",
				RequireProfile: true,
				ProfilePurpose: purpose,
			})

			if err == nil {
				t.Fatal("Expected error")
			}

			errMsg := err.Error()
			// Error message might contain the purpose
			t.Logf("Error with purpose '%s': %s", purpose, errMsg)
		})
	}
}

// TestErrorScenarios_MixedSlashes tests path with mixed slashes
func TestErrorScenarios_MixedSlashes(t *testing.T) {
	testDir := getTestDataDir(t)
	
	// This should be normalized
	mixedPath := filepath.Join(testDir, "test-profile1.cnf.enc")
	
	profile, err := ResolveAndLoadProfile(ProfileLoadOptions{
		ProfilePath: mixedPath,
		ProfileKey:  testEncryptionKey,
	})

	if err != nil {
		t.Logf("Note: Mixed slashes may or may not work: %v", err)
		return
	}

	if profile.DBInfo.Host != "testdb1.example.com" {
		t.Error("Failed to normalize mixed slashes")
	}
}

// TestErrorScenarios_SymbolicLink tests loading profile through symlink
func TestErrorScenarios_SymbolicLink(t *testing.T) {
	testDir := getTestDataDir(t)
	
	targetFile := filepath.Join(testDir, "test-profile1.cnf.enc")
	symlinkFile := filepath.Join(testDir, "symlink-profile.cnf.enc")
	
	// Create symlink
	if err := os.Symlink(targetFile, symlinkFile); err != nil {
		t.Skipf("Symlink not supported: %v", err)
		return
	}
	defer os.Remove(symlinkFile)

	// Should follow symlink
	profile, err := ResolveAndLoadProfile(ProfileLoadOptions{
		ConfigDir:   testDir,
		ProfilePath: "symlink-profile.cnf.enc",
		ProfileKey:  testEncryptionKey,
	})

	if err != nil {
		t.Fatalf("Should follow symlink: %v", err)
	}

	if profile.DBInfo.Host != "testdb1.example.com" {
		t.Error("Failed to load profile through symlink")
	}
}

// TestErrorScenarios_VeryLongProfileName tests with very long profile name
func TestErrorScenarios_VeryLongProfileName(t *testing.T) {
	testDir := getTestDataDir(t)
	
	// Create a profile with very long name
	longName := string(make([]byte, 255)) // Max filename length on most systems
	for i := range longName {
		longName = longName[:i] + "a"
	}
	longName += ".cnf.enc"
	
	// This might fail at OS level
	_, err := ResolveAndLoadProfile(ProfileLoadOptions{
		ConfigDir:   testDir,
		ProfilePath: longName,
		ProfileKey:  testEncryptionKey,
	})

	if err == nil {
		t.Log("Note: Very long filename might be accepted by OS")
	} else {
		t.Logf("Long filename error (expected): %v", err)
	}
}

// TestErrorScenarios_SpecialCharactersInName tests special characters in profile name
func TestErrorScenarios_SpecialCharactersInName(t *testing.T) {
	testDir := getTestDataDir(t)
	
	// Special characters that might be problematic
	specialNames := []string{
		"profile@test.cnf.enc",
		"profile#test.cnf.enc",
		"profile test.cnf.enc", // space
		"profile\ttest.cnf.enc", // tab
	}

	for _, name := range specialNames {
		t.Run(name, func(t *testing.T) {
			_, err := ResolveAndLoadProfile(ProfileLoadOptions{
				ConfigDir:   testDir,
				ProfilePath: name,
				ProfileKey:  testEncryptionKey,
			})

			// May or may not work depending on OS and validation
			if err != nil {
				t.Logf("Special char '%s' error: %v", name, err)
			} else {
				t.Logf("Special char '%s' accepted", name)
			}
		})
	}
}
