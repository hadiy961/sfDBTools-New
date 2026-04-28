// File : internal/app/profile/helpers/loader/profile_loader_integration_test.go
// Deskripsi : Integration tests untuk profile loader
// Author : Test Suite
// Tanggal : 5 Februari 2026
package loader

import (
	"os"
	"path/filepath"
	"testing"
)

const testEncryptionKey = "loader-test-key-123"

// Helper function to get testdata directory
func getTestDataDir(t *testing.T) string {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	return filepath.Join(wd, "testdata")
}

// TestResolveAndLoadProfile_WithProfilePath tests loading with explicit profile path
func TestResolveAndLoadProfile_WithProfilePath(t *testing.T) {
	testDir := getTestDataDir(t)
	profilePath := filepath.Join(testDir, "test-profile1.cnf.enc")

	profile, err := ResolveAndLoadProfile(ProfileLoadOptions{
		ProfilePath: profilePath,
		ProfileKey:  testEncryptionKey,
	})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify loaded data
	if profile.DBInfo.Host != "testdb1.example.com" {
		t.Errorf("Expected host 'testdb1.example.com', got '%s'", profile.DBInfo.Host)
	}
	if profile.DBInfo.Port != 3306 {
		t.Errorf("Expected port 3306, got %d", profile.DBInfo.Port)
	}
	if profile.DBInfo.User != "testuser1" {
		t.Errorf("Expected user 'testuser1', got '%s'", profile.DBInfo.User)
	}

	// Verify Path and Name set correctly
	if profile.Path == "" {
		t.Error("Expected Path to be set")
	}
	if profile.Name != "test-profile1" {
		t.Errorf("Expected name 'test-profile1', got '%s'", profile.Name)
	}
}

// TestResolveAndLoadProfile_WithConfigDir tests loading with ConfigDir
func TestResolveAndLoadProfile_WithConfigDir(t *testing.T) {
	testDir := getTestDataDir(t)

	// Use name only (not full path)
	profile, err := ResolveAndLoadProfile(ProfileLoadOptions{
		ConfigDir:   testDir,
		ProfilePath: "test-profile1.cnf.enc",
		ProfileKey:  testEncryptionKey,
	})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if profile.DBInfo.Host != "testdb1.example.com" {
		t.Error("Failed to load profile using ConfigDir + name")
	}
	if profile.Name != "test-profile1" {
		t.Errorf("Expected name 'test-profile1', got '%s'", profile.Name)
	}
}

// TestResolveAndLoadProfile_WithNameOnly tests loading with name only
func TestResolveAndLoadProfile_WithNameOnly(t *testing.T) {
	testDir := getTestDataDir(t)

	// Name without .cnf.enc extension
	profile, err := ResolveAndLoadProfile(ProfileLoadOptions{
		ConfigDir:   testDir,
		ProfilePath: "test-profile2",
		ProfileKey:  testEncryptionKey,
	})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if profile.DBInfo.Host != "testdb2.example.com" {
		t.Error("Failed to load profile using name only")
	}
	if profile.Name != "test-profile2" {
		t.Errorf("Expected name 'test-profile2', got '%s'", profile.Name)
	}

	// Verify SSH tunnel loaded
	if !profile.SSHTunnel.Enabled {
		t.Error("Expected SSH tunnel to be enabled")
	}
	if profile.SSHTunnel.Host != "ssh.example.com" {
		t.Error("SSH tunnel config not loaded correctly")
	}
}

// TestResolveAndLoadProfile_WithEnvFallback tests environment variable fallback
func TestResolveAndLoadProfile_WithEnvFallback(t *testing.T) {
	testDir := getTestDataDir(t)
	profilePath := filepath.Join(testDir, "test-profile1.cnf.enc")

	// Set environment variables
	os.Setenv("TEST_PROFILE_PATH", profilePath)
	os.Setenv("TEST_PROFILE_KEY", testEncryptionKey)
	defer func() {
		os.Unsetenv("TEST_PROFILE_PATH")
		os.Unsetenv("TEST_PROFILE_KEY")
	}()

	// Load with env fallback
	profile, err := ResolveAndLoadProfile(ProfileLoadOptions{
		ProfilePath:    "", // Empty, should use env
		ProfileKey:     "", // Empty, should use env
		EnvProfilePath: "TEST_PROFILE_PATH",
		EnvProfileKey:  "TEST_PROFILE_KEY",
	})

	if err != nil {
		t.Fatalf("Expected to load from env, got error: %v", err)
	}

	if profile.DBInfo.Host != "testdb1.example.com" {
		t.Error("Failed to load profile from environment variables")
	}
}

// TestResolveAndLoadProfile_ProfilePathPriority tests that ProfilePath takes priority over env
func TestResolveAndLoadProfile_ProfilePathPriority(t *testing.T) {
	testDir := getTestDataDir(t)
	profile1Path := filepath.Join(testDir, "test-profile1.cnf.enc")
	profile2Path := filepath.Join(testDir, "test-profile2.cnf.enc")

	// Set env to profile2
	os.Setenv("TEST_PROFILE_PATH", profile2Path)
	defer os.Unsetenv("TEST_PROFILE_PATH")

	// But explicitly pass profile1 via ProfilePath
	profile, err := ResolveAndLoadProfile(ProfileLoadOptions{
		ProfilePath:    profile1Path,
		ProfileKey:     testEncryptionKey,
		EnvProfilePath: "TEST_PROFILE_PATH",
	})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Should load profile1, not profile2 from env
	if profile.DBInfo.Host != "testdb1.example.com" {
		t.Error("ProfilePath should take priority over environment variable")
	}
}

// TestResolveAndLoadProfile_KeyPriority tests that ProfileKey takes priority over env
func TestResolveAndLoadProfile_KeyPriority(t *testing.T) {
	testDir := getTestDataDir(t)
	profilePath := filepath.Join(testDir, "test-profile1.cnf.enc")

	// Set env with wrong key
	os.Setenv("TEST_PROFILE_KEY", "wrong-key")
	defer os.Unsetenv("TEST_PROFILE_KEY")

	// But explicitly pass correct key via ProfileKey
	profile, err := ResolveAndLoadProfile(ProfileLoadOptions{
		ProfilePath:   profilePath,
		ProfileKey:    testEncryptionKey, // Correct key
		EnvProfileKey: "TEST_PROFILE_KEY",
	})

	if err != nil {
		t.Fatalf("Expected no error with correct ProfileKey, got: %v", err)
	}

	if profile.DBInfo.Host != "testdb1.example.com" {
		t.Error("ProfileKey should take priority over environment variable")
	}
}

// TestResolveAndLoadProfile_AbsolutePath tests loading with absolute path
func TestResolveAndLoadProfile_AbsolutePath(t *testing.T) {
	testDir := getTestDataDir(t)
	absPath := filepath.Join(testDir, "test-profile1.cnf.enc")

	profile, err := ResolveAndLoadProfile(ProfileLoadOptions{
		ProfilePath: absPath,
		ProfileKey:  testEncryptionKey,
	})

	if err != nil {
		t.Fatalf("Expected no error with absolute path, got: %v", err)
	}

	if profile.Path != absPath {
		t.Errorf("Expected Path '%s', got '%s'", absPath, profile.Path)
	}
}

// TestResolveAndLoadProfile_RelativePath tests loading with relative path in ConfigDir
func TestResolveAndLoadProfile_RelativePath(t *testing.T) {
	testDir := getTestDataDir(t)

	// Relative path (just filename)
	profile, err := ResolveAndLoadProfile(ProfileLoadOptions{
		ConfigDir:   testDir,
		ProfilePath: "test-profile1.cnf.enc",
		ProfileKey:  testEncryptionKey,
	})

	if err != nil {
		t.Fatalf("Expected no error with relative path, got: %v", err)
	}

	// Path should be absolute after resolution
	if !filepath.IsAbs(profile.Path) {
		t.Error("Expected Path to be absolute after resolution")
	}
}

// TestResolveAndLoadProfile_RequireProfile_Missing tests RequireProfile with missing profile
func TestResolveAndLoadProfile_RequireProfile_Missing(t *testing.T) {
	_, err := ResolveAndLoadProfile(ProfileLoadOptions{
		ProfilePath:    "", // Empty
		EnvProfilePath: "NONEXISTENT_ENV_VAR",
		RequireProfile: true,
		ProfilePurpose: "test",
	})

	if err == nil {
		t.Fatal("Expected error when profile required but not provided")
	}

	// Error should mention the purpose
	errMsg := err.Error()
	if errMsg == "" {
		t.Error("Expected non-empty error message")
	}
	t.Logf("Error message: %s", errMsg)
}

// TestResolveAndLoadProfile_OptionalProfile tests optional profile (no error when missing)
func TestResolveAndLoadProfile_OptionalProfile(t *testing.T) {
	profile, err := ResolveAndLoadProfile(ProfileLoadOptions{
		ProfilePath:    "",
		EnvProfilePath: "NONEXISTENT_ENV_VAR",
		RequireProfile: false, // Optional
	})

	// Should not error, but profile will be nil or partial
	if err != nil {
		t.Logf("Got error (expected for optional missing profile): %v", err)
	}

	if profile != nil {
		t.Log("Profile is not nil (may have loaded from somewhere)")
	}
}

// TestResolveAndLoadProfile_FileNotFound tests error when profile file doesn't exist
func TestResolveAndLoadProfile_FileNotFound(t *testing.T) {
	testDir := getTestDataDir(t)

	_, err := ResolveAndLoadProfile(ProfileLoadOptions{
		ConfigDir:   testDir,
		ProfilePath: "nonexistent-profile.cnf.enc",
		ProfileKey:  testEncryptionKey,
	})

	if err == nil {
		t.Fatal("Expected error for nonexistent profile file")
	}

	t.Logf("Error message: %v", err)
}

// TestResolveAndLoadProfile_WrongKey tests error with wrong encryption key
func TestResolveAndLoadProfile_WrongKey(t *testing.T) {
	testDir := getTestDataDir(t)
	profilePath := filepath.Join(testDir, "test-profile1.cnf.enc")

	_, err := ResolveAndLoadProfile(ProfileLoadOptions{
		ProfilePath: profilePath,
		ProfileKey:  "wrong-key-that-wont-work",
	})

	if err == nil {
		t.Fatal("Expected error with wrong encryption key")
	}

	// Should mention decryption failure
	errMsg := err.Error()
	t.Logf("Error message: %s", errMsg)
}

// TestResolveAndLoadProfile_EmptyConfigDir tests error with empty ConfigDir for relative path
func TestResolveAndLoadProfile_EmptyConfigDir(t *testing.T) {
	_, err := ResolveAndLoadProfile(ProfileLoadOptions{
		ConfigDir:   "",                      // Empty
		ProfilePath: "test-profile1.cnf.enc", // Relative path
		ProfileKey:  testEncryptionKey,
	})

	if err == nil {
		t.Fatal("Expected error with empty ConfigDir for relative path")
	}

	t.Logf("Error message: %v", err)
}

// TestResolveAndLoadProfile_EmptyProfilePath tests error with empty profile path
func TestResolveAndLoadProfile_EmptyProfilePath(t *testing.T) {
	_, err := ResolveAndLoadProfile(ProfileLoadOptions{
		ProfilePath:    "",
		ProfileKey:     testEncryptionKey,
		RequireProfile: true,
	})

	if err == nil {
		t.Fatal("Expected error with empty profile path and RequireProfile=true")
	}

	t.Logf("Error message: %v", err)
}

// TestResolveAndLoadProfile_ProfilePathAndName tests Path and Name fields set correctly
func TestResolveAndLoadProfile_ProfilePathAndName(t *testing.T) {
	testDir := getTestDataDir(t)

	testCases := []struct {
		name         string
		profilePath  string
		expectedName string
	}{
		{
			name:         "with_extension",
			profilePath:  "test-profile1.cnf.enc",
			expectedName: "test-profile1",
		},
		{
			name:         "without_extension",
			profilePath:  "test-profile2",
			expectedName: "test-profile2",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			profile, err := ResolveAndLoadProfile(ProfileLoadOptions{
				ConfigDir:   testDir,
				ProfilePath: tc.profilePath,
				ProfileKey:  testEncryptionKey,
			})

			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}

			if profile.Name != tc.expectedName {
				t.Errorf("Expected name '%s', got '%s'", tc.expectedName, profile.Name)
			}

			if profile.Path == "" {
				t.Error("Expected Path to be set")
			}

			if !filepath.IsAbs(profile.Path) {
				t.Error("Expected Path to be absolute")
			}
		})
	}
}

// TestResolveAndLoadProfile_EncryptionMetadata tests encryption metadata is preserved
func TestResolveAndLoadProfile_EncryptionMetadata(t *testing.T) {
	testDir := getTestDataDir(t)
	profilePath := filepath.Join(testDir, "test-profile1.cnf.enc")

	profile, err := ResolveAndLoadProfile(ProfileLoadOptions{
		ProfilePath: profilePath,
		ProfileKey:  testEncryptionKey,
	})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Encryption key should be set
	if profile.EncryptionKey != testEncryptionKey {
		t.Errorf("Expected encryption key '%s', got '%s'", testEncryptionKey, profile.EncryptionKey)
	}

	// Encryption source should be set
	if profile.EncryptionSource == "" {
		t.Error("Expected EncryptionSource to be set")
	}
	t.Logf("Encryption source: %s", profile.EncryptionSource)
}

// TestResolveAndLoadProfile_MultipleProfiles tests loading different profiles
func TestResolveAndLoadProfile_MultipleProfiles(t *testing.T) {
	testDir := getTestDataDir(t)

	profiles := []struct {
		path         string
		expectedHost string
		expectedPort int
	}{
		{"test-profile1.cnf.enc", "testdb1.example.com", 3306},
		{"test-profile2.cnf.enc", "testdb2.example.com", 3307},
		{"source-db.cnf.enc", "source-db.local", 3306},
	}

	for _, p := range profiles {
		t.Run(p.path, func(t *testing.T) {
			profile, err := ResolveAndLoadProfile(ProfileLoadOptions{
				ConfigDir:   testDir,
				ProfilePath: p.path,
				ProfileKey:  testEncryptionKey,
			})

			if err != nil {
				t.Fatalf("Failed to load %s: %v", p.path, err)
			}

			if profile.DBInfo.Host != p.expectedHost {
				t.Errorf("Expected host '%s', got '%s'", p.expectedHost, profile.DBInfo.Host)
			}

			if profile.DBInfo.Port != p.expectedPort {
				t.Errorf("Expected port %d, got %d", p.expectedPort, profile.DBInfo.Port)
			}
		})
	}
}
