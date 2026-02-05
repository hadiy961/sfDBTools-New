// File : internal/app/profile/connection/preflight_test.go
// Deskripsi : Unit tests for connection preflight validation
// Author : Test Suite
// Tanggal : 5 Februari 2026
package connection

import (
	"os"
	"path/filepath"
	"testing"

	"sfdbtools/internal/domain"
)

// TestValidateConnectPreflight_NilProfile tests nil profile error
func TestValidateConnectPreflight_NilProfile(t *testing.T) {
	err := ValidateConnectPreflight(nil)
	if err == nil {
		t.Fatal("Expected error for nil profile")
	}
	t.Logf("Nil profile error: %v", err)
}

// TestValidateConnectPreflight_EmptyHost tests empty host error
func TestValidateConnectPreflight_EmptyHost(t *testing.T) {
	profile := &domain.ProfileInfo{
		DBInfo: domain.DBInfo{
			Host:     "", // Empty
			Port:     3306,
			User:     "testuser",
			Password: "testpass",
		},
	}

	err := ValidateConnectPreflight(profile)
	if err == nil {
		t.Fatal("Expected error for empty host")
	}
	t.Logf("Empty host error: %v", err)
}

// TestValidateConnectPreflight_EmptyUser tests empty user error
func TestValidateConnectPreflight_EmptyUser(t *testing.T) {
	profile := &domain.ProfileInfo{
		DBInfo: domain.DBInfo{
			Host:     "localhost",
			Port:     3306,
			User:     "", // Empty
			Password: "testpass",
		},
	}

	err := ValidateConnectPreflight(profile)
	if err == nil {
		t.Fatal("Expected error for empty user")
	}
	t.Logf("Empty user error: %v", err)
}

// TestValidateConnectPreflight_InvalidPortRanges tests invalid port values
func TestValidateConnectPreflight_InvalidPortRanges(t *testing.T) {
	testCases := []struct {
		name string
		port int
	}{
		{"zero_port", 0},
		{"negative_port", -1},
		{"port_too_large", 65536},
		{"port_way_too_large", 100000},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			profile := &domain.ProfileInfo{
				DBInfo: domain.DBInfo{
					Host:     "localhost",
					Port:     tc.port,
					User:     "testuser",
					Password: "testpass",
				},
			}

			err := ValidateConnectPreflight(profile)
			if err == nil {
				t.Errorf("Expected error for port %d", tc.port)
			}
			t.Logf("Port %d error: %v", tc.port, err)
		})
	}
}

// TestValidateConnectPreflight_ValidPortRanges tests valid port values
func TestValidateConnectPreflight_ValidPortRanges(t *testing.T) {
	testCases := []struct {
		name string
		port int
	}{
		{"port_1", 1},
		{"port_80", 80},
		{"port_3306", 3306},
		{"port_8080", 8080},
		{"port_65535", 65535},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			profile := &domain.ProfileInfo{
				DBInfo: domain.DBInfo{
					Host:     "localhost",
					Port:     tc.port,
					User:     "testuser",
					Password: "testpass",
				},
			}

			err := ValidateConnectPreflight(profile)
			if err != nil {
				t.Errorf("Expected no error for valid port %d, got: %v", tc.port, err)
			}
		})
	}
}

// TestValidateConnectPreflight_ValidDirectConnection tests valid direct connection
func TestValidateConnectPreflight_ValidDirectConnection(t *testing.T) {
	profile := &domain.ProfileInfo{
		DBInfo: domain.DBInfo{
			Host:     "localhost",
			Port:     3306,
			User:     "testuser",
			Password: "testpass",
		},
		SSHTunnel: domain.SSHTunnelConfig{
			Enabled: false,
		},
	}

	err := ValidateConnectPreflight(profile)
	if err != nil {
		t.Fatalf("Expected no error for valid direct connection, got: %v", err)
	}
}

// TestValidateConnectPreflight_SSHTunnelEmptyHost tests SSH tunnel with empty host
func TestValidateConnectPreflight_SSHTunnelEmptyHost(t *testing.T) {
	profile := &domain.ProfileInfo{
		DBInfo: domain.DBInfo{
			Host:     "db.example.com",
			Port:     3306,
			User:     "testuser",
			Password: "testpass",
		},
		SSHTunnel: domain.SSHTunnelConfig{
			Enabled: true,
			Host:    "", // Empty
			Port:    22,
			User:    "sshuser",
		},
	}

	err := ValidateConnectPreflight(profile)
	if err == nil {
		t.Fatal("Expected error for SSH tunnel with empty host")
	}
	t.Logf("SSH empty host error: %v", err)
}

// TestValidateConnectPreflight_SSHTunnelInvalidPort tests invalid SSH port
func TestValidateConnectPreflight_SSHTunnelInvalidPort(t *testing.T) {
	testCases := []struct {
		name    string
		sshPort int
	}{
		{"negative_port", -1},
		{"port_too_large", 65536},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			profile := &domain.ProfileInfo{
				DBInfo: domain.DBInfo{
					Host:     "db.example.com",
					Port:     3306,
					User:     "testuser",
					Password: "testpass",
				},
				SSHTunnel: domain.SSHTunnelConfig{
					Enabled: true,
					Host:    "ssh.example.com",
					Port:    tc.sshPort,
					User:    "sshuser",
				},
			}

			err := ValidateConnectPreflight(profile)
			if err == nil {
				t.Errorf("Expected error for SSH port %d", tc.sshPort)
			}
			t.Logf("SSH port %d error: %v", tc.sshPort, err)
		})
	}
}

// TestValidateConnectPreflight_SSHTunnelDefaultPort tests SSH port defaults to 22
func TestValidateConnectPreflight_SSHTunnelDefaultPort(t *testing.T) {
	profile := &domain.ProfileInfo{
		DBInfo: domain.DBInfo{
			Host:     "db.example.com",
			Port:     3306,
			User:     "testuser",
			Password: "testpass",
		},
		SSHTunnel: domain.SSHTunnelConfig{
			Enabled: true,
			Host:    "ssh.example.com",
			Port:    0, // Should default to 22
			User:    "sshuser",
		},
	}

	err := ValidateConnectPreflight(profile)
	if err != nil {
		t.Fatalf("Expected no error (port 0 defaults to 22), got: %v", err)
	}
}

// TestValidateConnectPreflight_SSHTunnelInvalidLocalPort tests invalid local port
func TestValidateConnectPreflight_SSHTunnelInvalidLocalPort(t *testing.T) {
	testCases := []struct {
		name      string
		localPort int
	}{
		{"negative_port", -1},
		{"port_too_large", 65536},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			profile := &domain.ProfileInfo{
				DBInfo: domain.DBInfo{
					Host:     "db.example.com",
					Port:     3306,
					User:     "testuser",
					Password: "testpass",
				},
				SSHTunnel: domain.SSHTunnelConfig{
					Enabled:   true,
					Host:      "ssh.example.com",
					Port:      22,
					User:      "sshuser",
					LocalPort: tc.localPort,
				},
			}

			err := ValidateConnectPreflight(profile)
			if err == nil {
				t.Errorf("Expected error for local port %d", tc.localPort)
			}
			t.Logf("Local port %d error: %v", tc.localPort, err)
		})
	}
}

// TestValidateConnectPreflight_SSHTunnelValidLocalPort tests valid local port (0 is ok)
func TestValidateConnectPreflight_SSHTunnelValidLocalPort(t *testing.T) {
	testCases := []int{0, 1, 1024, 13306, 65535}

	for _, localPort := range testCases {
		t.Run(string(rune(localPort)), func(t *testing.T) {
			profile := &domain.ProfileInfo{
				DBInfo: domain.DBInfo{
					Host:     "db.example.com",
					Port:     3306,
					User:     "testuser",
					Password: "testpass",
				},
				SSHTunnel: domain.SSHTunnelConfig{
					Enabled:   true,
					Host:      "ssh.example.com",
					Port:      22,
					User:      "sshuser",
					LocalPort: localPort,
				},
			}

			err := ValidateConnectPreflight(profile)
			if err != nil {
				t.Errorf("Expected no error for valid local port %d, got: %v", localPort, err)
			}
		})
	}
}

// TestValidateConnectPreflight_SSHTunnelIdentityFileNotFound tests missing identity file
func TestValidateConnectPreflight_SSHTunnelIdentityFileNotFound(t *testing.T) {
	profile := &domain.ProfileInfo{
		DBInfo: domain.DBInfo{
			Host:     "db.example.com",
			Port:     3306,
			User:     "testuser",
			Password: "testpass",
		},
		SSHTunnel: domain.SSHTunnelConfig{
			Enabled:      true,
			Host:         "ssh.example.com",
			Port:         22,
			User:         "sshuser",
			IdentityFile: "/nonexistent/path/to/id_rsa",
		},
	}

	err := ValidateConnectPreflight(profile)
	if err == nil {
		t.Fatal("Expected error for nonexistent identity file")
	}
	t.Logf("Identity file not found error: %v", err)
}

// TestValidateConnectPreflight_SSHTunnelIdentityFileIsDirectory tests identity file that's actually a directory
func TestValidateConnectPreflight_SSHTunnelIdentityFileIsDirectory(t *testing.T) {
	// Create temp directory
	tempDir := t.TempDir()

	profile := &domain.ProfileInfo{
		DBInfo: domain.DBInfo{
			Host:     "db.example.com",
			Port:     3306,
			User:     "testuser",
			Password: "testpass",
		},
		SSHTunnel: domain.SSHTunnelConfig{
			Enabled:      true,
			Host:         "ssh.example.com",
			Port:         22,
			User:         "sshuser",
			IdentityFile: tempDir, // Directory, not file
		},
	}

	err := ValidateConnectPreflight(profile)
	if err == nil {
		t.Fatal("Expected error when identity file is a directory")
	}
	t.Logf("Identity file is directory error: %v", err)
}

// TestValidateConnectPreflight_SSHTunnelValidIdentityFile tests with valid identity file
func TestValidateConnectPreflight_SSHTunnelValidIdentityFile(t *testing.T) {
	// Create temp file
	tempDir := t.TempDir()
	identityFile := filepath.Join(tempDir, "id_rsa")
	
	if err := os.WriteFile(identityFile, []byte("fake ssh key content"), 0600); err != nil {
		t.Fatalf("Failed to create temp identity file: %v", err)
	}

	profile := &domain.ProfileInfo{
		DBInfo: domain.DBInfo{
			Host:     "db.example.com",
			Port:     3306,
			User:     "testuser",
			Password: "testpass",
		},
		SSHTunnel: domain.SSHTunnelConfig{
			Enabled:      true,
			Host:         "ssh.example.com",
			Port:         22,
			User:         "sshuser",
			IdentityFile: identityFile,
		},
	}

	err := ValidateConnectPreflight(profile)
	if err != nil {
		t.Fatalf("Expected no error for valid identity file, got: %v", err)
	}
}

// TestValidateConnectPreflight_SSHTunnelRelativeIdentityFile tests relative path identity file
func TestValidateConnectPreflight_SSHTunnelRelativeIdentityFile(t *testing.T) {
	// Create temp file in current directory
	tempFile := "temp_test_id_rsa"
	if err := os.WriteFile(tempFile, []byte("fake key"), 0600); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile)

	profile := &domain.ProfileInfo{
		DBInfo: domain.DBInfo{
			Host:     "db.example.com",
			Port:     3306,
			User:     "testuser",
			Password: "testpass",
		},
		SSHTunnel: domain.SSHTunnelConfig{
			Enabled:      true,
			Host:         "ssh.example.com",
			Port:         22,
			User:         "sshuser",
			IdentityFile: tempFile, // Relative path
		},
	}

	err := ValidateConnectPreflight(profile)
	if err != nil {
		t.Fatalf("Expected no error for relative path identity file, got: %v", err)
	}
}

// TestValidateConnectPreflight_SSHTunnelEmptyIdentityFile tests empty identity file (should be ok)
func TestValidateConnectPreflight_SSHTunnelEmptyIdentityFile(t *testing.T) {
	profile := &domain.ProfileInfo{
		DBInfo: domain.DBInfo{
			Host:     "db.example.com",
			Port:     3306,
			User:     "testuser",
			Password: "testpass",
		},
		SSHTunnel: domain.SSHTunnelConfig{
			Enabled:      true,
			Host:         "ssh.example.com",
			Port:         22,
			User:         "sshuser",
			IdentityFile: "", // Empty = use password auth
		},
	}

	err := ValidateConnectPreflight(profile)
	if err != nil {
		t.Fatalf("Expected no error when identity file empty (password auth), got: %v", err)
	}
}

// TestValidateConnectPreflight_WhitespaceHandling tests whitespace trimming
func TestValidateConnectPreflight_WhitespaceHandling(t *testing.T) {
	profile := &domain.ProfileInfo{
		DBInfo: domain.DBInfo{
			Host:     "  localhost  ", // Whitespace
			Port:     3306,
			User:     "  testuser  ", // Whitespace
			Password: "testpass",
		},
		SSHTunnel: domain.SSHTunnelConfig{
			Enabled: true,
			Host:    "  ssh.example.com  ", // Whitespace
			Port:    22,
			User:    "sshuser",
		},
	}

	err := ValidateConnectPreflight(profile)
	if err != nil {
		t.Fatalf("Expected no error (whitespace should be trimmed), got: %v", err)
	}
}

// TestValidateConnectPreflight_CompleteValidProfile tests fully valid profile with SSH
func TestValidateConnectPreflight_CompleteValidProfile(t *testing.T) {
	// Create temp identity file
	tempDir := t.TempDir()
	identityFile := filepath.Join(tempDir, "id_rsa")
	if err := os.WriteFile(identityFile, []byte("fake key"), 0600); err != nil {
		t.Fatalf("Failed to create identity file: %v", err)
	}

	profile := &domain.ProfileInfo{
		DBInfo: domain.DBInfo{
			Host:     "db.example.com",
			Port:     3306,
			User:     "dbuser",
			Password: "dbpass",
		},
		SSHTunnel: domain.SSHTunnelConfig{
			Enabled:      true,
			Host:         "ssh.example.com",
			Port:         22,
			User:         "sshuser",
			Password:     "sshpass",
			IdentityFile: identityFile,
			LocalPort:    13306,
		},
	}

	err := ValidateConnectPreflight(profile)
	if err != nil {
		t.Fatalf("Expected no error for complete valid profile, got: %v", err)
	}
}
