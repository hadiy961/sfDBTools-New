// File : internal/app/profile/helpers/loader/load_source_test.go
// Deskripsi : Tests for LoadSourceProfile function
// Author : Test Suite
// Tanggal : 5 Februari 2026
package loader

import (
	"path/filepath"
	"testing"
)

// TestLoadSourceProfile_Success tests successful source profile loading
func TestLoadSourceProfile_Success(t *testing.T) {
	testDir := getTestDataDir(t)
	profilePath := filepath.Join(testDir, "source-db.cnf.enc")

	profile, err := LoadSourceProfile(testDir, profilePath, testEncryptionKey, false)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify it's the correct source profile
	if profile.DBInfo.Host != "source-db.local" {
		t.Errorf("Expected host 'source-db.local', got '%s'", profile.DBInfo.Host)
	}
	if profile.DBInfo.User != "sourceuser" {
		t.Errorf("Expected user 'sourceuser', got '%s'", profile.DBInfo.User)
	}

	// Verify profile name
	if profile.Name != "source-db" {
		t.Errorf("Expected name 'source-db', got '%s'", profile.Name)
	}
}

// TestLoadSourceProfile_WithName tests loading with name only (not full path)
func TestLoadSourceProfile_WithName(t *testing.T) {
	testDir := getTestDataDir(t)

	// Use name only
	profile, err := LoadSourceProfile(testDir, "source-db", testEncryptionKey, false)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if profile.DBInfo.Host != "source-db.local" {
		t.Error("Failed to load source profile with name only")
	}
}

// TestLoadSourceProfile_RequireProfile tests that profile is required
func TestLoadSourceProfile_RequireProfile(t *testing.T) {
	testDir := getTestDataDir(t)

	// Empty profile path, non-interactive
	_, err := LoadSourceProfile(testDir, "", testEncryptionKey, false)

	if err == nil {
		t.Fatal("Expected error when source profile not provided (RequireProfile=true)")
	}

	// Error should mention "source"
	errMsg := err.Error()
	t.Logf("Error message (should mention source): %s", errMsg)
}

// TestLoadSourceProfile_NotFound tests error when source profile doesn't exist
func TestLoadSourceProfile_NotFound(t *testing.T) {
	testDir := getTestDataDir(t)

	_, err := LoadSourceProfile(testDir, "nonexistent-source.cnf.enc", testEncryptionKey, false)

	if err == nil {
		t.Fatal("Expected error when source profile file not found")
	}

	t.Logf("Error message: %v", err)
}

// TestLoadSourceProfile_WrongKey tests error with wrong encryption key
func TestLoadSourceProfile_WrongKey(t *testing.T) {
	testDir := getTestDataDir(t)

	_, err := LoadSourceProfile(testDir, "source-db.cnf.enc", "wrong-key", false)

	if err == nil {
		t.Fatal("Expected error with wrong encryption key")
	}

	t.Logf("Error message: %v", err)
}

// TestLoadSourceProfile_InteractiveFlag tests allowInteractive parameter
func TestLoadSourceProfile_InteractiveFlag(t *testing.T) {
	testDir := getTestDataDir(t)

	// With explicit path, interactive flag doesn't matter
	profile, err := LoadSourceProfile(testDir, "source-db", testEncryptionKey, true)

	if err != nil {
		t.Fatalf("Expected success with explicit path, got: %v", err)
	}

	if profile.DBInfo.Host != "source-db.local" {
		t.Error("Failed to load with allowInteractive=true")
	}
}

// TestLoadSourceProfile_NonInteractiveNoPath tests non-interactive mode requires path
func TestLoadSourceProfile_NonInteractiveNoPath(t *testing.T) {
	testDir := getTestDataDir(t)

	// Empty path, non-interactive (would need interactive prompt)
	_, err := LoadSourceProfile(testDir, "", testEncryptionKey, false)

	if err == nil {
		t.Fatal("Expected error in non-interactive mode without path")
	}

	// Error should indicate profile is required
	errMsg := err.Error()
	if errMsg == "" {
		t.Error("Expected non-empty error message")
	}
	t.Logf("Non-interactive error: %s", errMsg)
}

// TestLoadSourceProfile_ProfileData tests all profile data loaded correctly
func TestLoadSourceProfile_ProfileData(t *testing.T) {
	testDir := getTestDataDir(t)

	profile, err := LoadSourceProfile(testDir, "source-db", testEncryptionKey, false)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify all fields
	if profile.DBInfo.Host != "source-db.local" {
		t.Error("Host not loaded correctly")
	}
	if profile.DBInfo.Port != 3306 {
		t.Error("Port not loaded correctly")
	}
	if profile.DBInfo.User != "sourceuser" {
		t.Error("User not loaded correctly")
	}
	if profile.DBInfo.Password != "sourcepass" {
		t.Error("Password not loaded correctly")
	}

	// SSH should be disabled
	if profile.SSHTunnel.Enabled {
		t.Error("Expected SSH tunnel to be disabled")
	}

	// Path and Name should be set
	if profile.Path == "" {
		t.Error("Expected Path to be set")
	}
	if profile.Name == "" {
		t.Error("Expected Name to be set")
	}
}

// TestLoadSourceProfile_WithSSHTunnel tests loading source profile with SSH
func TestLoadSourceProfile_WithSSHTunnel(t *testing.T) {
	testDir := getTestDataDir(t)

	// test-profile2 has SSH tunnel enabled
	profile, err := LoadSourceProfile(testDir, "test-profile2", testEncryptionKey, false)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify SSH tunnel loaded
	if !profile.SSHTunnel.Enabled {
		t.Fatal("Expected SSH tunnel to be enabled")
	}
	if profile.SSHTunnel.Host != "ssh.example.com" {
		t.Error("SSH host not loaded correctly")
	}
	if profile.SSHTunnel.User != "sshuser" {
		t.Error("SSH user not loaded correctly")
	}
}

// TestLoadSourceProfile_AbsolutePath tests with absolute path
func TestLoadSourceProfile_AbsolutePath(t *testing.T) {
	testDir := getTestDataDir(t)
	absPath := filepath.Join(testDir, "source-db.cnf.enc")

	profile, err := LoadSourceProfile("", absPath, testEncryptionKey, false)

	if err != nil {
		t.Fatalf("Expected no error with absolute path, got: %v", err)
	}

	if profile.DBInfo.Host != "source-db.local" {
		t.Error("Failed to load with absolute path")
	}
}

// TestLoadSourceProfile_EncryptionMetadata tests encryption metadata preserved
func TestLoadSourceProfile_EncryptionMetadata(t *testing.T) {
	testDir := getTestDataDir(t)

	profile, err := LoadSourceProfile(testDir, "source-db", testEncryptionKey, false)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Encryption key should be preserved
	if profile.EncryptionKey != testEncryptionKey {
		t.Error("Encryption key not preserved")
	}

	// Encryption source should be set
	if profile.EncryptionSource == "" {
		t.Error("Encryption source not set")
	}
}

// TestLoadSourceProfile_MultipleLoads tests loading multiple source profiles
func TestLoadSourceProfile_MultipleLoads(t *testing.T) {
	testDir := getTestDataDir(t)

	sources := []struct {
		name         string
		expectedHost string
	}{
		{"source-db", "source-db.local"},
		{"test-profile1", "testdb1.example.com"},
		{"test-profile2", "testdb2.example.com"},
	}

	for _, src := range sources {
		t.Run(src.name, func(t *testing.T) {
			profile, err := LoadSourceProfile(testDir, src.name, testEncryptionKey, false)

			if err != nil {
				t.Fatalf("Failed to load %s: %v", src.name, err)
			}

			if profile.DBInfo.Host != src.expectedHost {
				t.Errorf("Expected host '%s', got '%s'", src.expectedHost, profile.DBInfo.Host)
			}
		})
	}
}
