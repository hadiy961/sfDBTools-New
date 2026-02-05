// File : internal/app/profile/connection/effective_dbinfo_test.go
// Deskripsi : Tests for EffectiveDBInfo function
// Author : Test Suite
// Tanggal : 5 Februari 2026
package connection

import (
	"testing"

	"sfdbtools/internal/domain"
)

// TestEffectiveDBInfo_NilProfile tests nil profile returns empty DBInfo
func TestEffectiveDBInfo_NilProfile(t *testing.T) {
	info := EffectiveDBInfo(nil)
	
	if info.Host != "" || info.Port != 0 || info.User != "" || info.Password != "" {
		t.Error("Expected empty DBInfo for nil profile")
	}
}

// TestEffectiveDBInfo_DirectConnection tests direct connection (no SSH tunnel)
func TestEffectiveDBInfo_DirectConnection(t *testing.T) {
	profile := &domain.ProfileInfo{
		DBInfo: domain.DBInfo{
			Host:     "db.example.com",
			Port:     3306,
			User:     "dbuser",
			Password: "dbpass",
		},
		SSHTunnel: domain.SSHTunnelConfig{
			Enabled: false,
		},
	}

	info := EffectiveDBInfo(profile)

	// Should return original DB info
	if info.Host != "db.example.com" {
		t.Errorf("Expected host 'db.example.com', got '%s'", info.Host)
	}
	if info.Port != 3306 {
		t.Errorf("Expected port 3306, got %d", info.Port)
	}
	if info.User != "dbuser" {
		t.Errorf("Expected user 'dbuser', got '%s'", info.User)
	}
	if info.Password != "dbpass" {
		t.Errorf("Expected password 'dbpass', got '%s'", info.Password)
	}
}

// TestEffectiveDBInfo_SSHTunnelNotResolved tests SSH tunnel enabled but not resolved yet
func TestEffectiveDBInfo_SSHTunnelNotResolved(t *testing.T) {
	profile := &domain.ProfileInfo{
		DBInfo: domain.DBInfo{
			Host:     "db.example.com",
			Port:     3306,
			User:     "dbuser",
			Password: "dbpass",
		},
		SSHTunnel: domain.SSHTunnelConfig{
			Enabled:           true,
			Host:              "ssh.example.com",
			Port:              22,
			ResolvedLocalPort: 0, // Not resolved yet
		},
	}

	info := EffectiveDBInfo(profile)

	// Should still return original DB info (tunnel not ready)
	if info.Host != "db.example.com" {
		t.Errorf("Expected original host, got '%s'", info.Host)
	}
	if info.Port != 3306 {
		t.Errorf("Expected original port, got %d", info.Port)
	}
}

// TestEffectiveDBInfo_SSHTunnelResolved tests SSH tunnel with resolved local port
func TestEffectiveDBInfo_SSHTunnelResolved(t *testing.T) {
	profile := &domain.ProfileInfo{
		DBInfo: domain.DBInfo{
			Host:     "db.example.com",
			Port:     3306,
			User:     "dbuser",
			Password: "dbpass",
		},
		SSHTunnel: domain.SSHTunnelConfig{
			Enabled:           true,
			Host:              "ssh.example.com",
			Port:              22,
			LocalPort:         13306,
			ResolvedLocalPort: 13306, // Resolved!
		},
	}

	info := EffectiveDBInfo(profile)

	// Should redirect to localhost with resolved port
	if info.Host != "127.0.0.1" {
		t.Errorf("Expected host '127.0.0.1' (localhost), got '%s'", info.Host)
	}
	if info.Port != 13306 {
		t.Errorf("Expected port 13306 (resolved local port), got %d", info.Port)
	}
	
	// User and password should remain the same
	if info.User != "dbuser" {
		t.Errorf("Expected user 'dbuser', got '%s'", info.User)
	}
	if info.Password != "dbpass" {
		t.Errorf("Expected password 'dbpass', got '%s'", info.Password)
	}
}

// TestEffectiveDBInfo_SSHTunnelAutoPort tests SSH tunnel with auto-assigned port
func TestEffectiveDBInfo_SSHTunnelAutoPort(t *testing.T) {
	profile := &domain.ProfileInfo{
		DBInfo: domain.DBInfo{
			Host:     "db.example.com",
			Port:     3306,
			User:     "dbuser",
			Password: "dbpass",
		},
		SSHTunnel: domain.SSHTunnelConfig{
			Enabled:           true,
			Host:              "ssh.example.com",
			Port:              22,
			LocalPort:         0,     // Auto-assign
			ResolvedLocalPort: 54321, // OS-assigned port
		},
	}

	info := EffectiveDBInfo(profile)

	// Should use resolved port, not LocalPort
	if info.Host != "127.0.0.1" {
		t.Errorf("Expected host '127.0.0.1', got '%s'", info.Host)
	}
	if info.Port != 54321 {
		t.Errorf("Expected resolved port 54321, got %d", info.Port)
	}
}

// TestEffectiveDBInfo_MultipleProfiles tests multiple profiles
func TestEffectiveDBInfo_MultipleProfiles(t *testing.T) {
	testCases := []struct {
		name         string
		profile      *domain.ProfileInfo
		expectedHost string
		expectedPort int
	}{
		{
			name: "direct_connection",
			profile: &domain.ProfileInfo{
				DBInfo: domain.DBInfo{
					Host: "direct.db.com",
					Port: 3307,
					User: "user1",
				},
				SSHTunnel: domain.SSHTunnelConfig{Enabled: false},
			},
			expectedHost: "direct.db.com",
			expectedPort: 3307,
		},
		{
			name: "ssh_resolved",
			profile: &domain.ProfileInfo{
				DBInfo: domain.DBInfo{
					Host: "remote.db.com",
					Port: 3306,
					User: "user2",
				},
				SSHTunnel: domain.SSHTunnelConfig{
					Enabled:           true,
					ResolvedLocalPort: 14306,
				},
			},
			expectedHost: "127.0.0.1",
			expectedPort: 14306,
		},
		{
			name: "ssh_not_resolved",
			profile: &domain.ProfileInfo{
				DBInfo: domain.DBInfo{
					Host: "remote2.db.com",
					Port: 3308,
					User: "user3",
				},
				SSHTunnel: domain.SSHTunnelConfig{
					Enabled:           true,
					ResolvedLocalPort: 0, // Not resolved
				},
			},
			expectedHost: "remote2.db.com",
			expectedPort: 3308,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			info := EffectiveDBInfo(tc.profile)

			if info.Host != tc.expectedHost {
				t.Errorf("Expected host '%s', got '%s'", tc.expectedHost, info.Host)
			}
			if info.Port != tc.expectedPort {
				t.Errorf("Expected port %d, got %d", tc.expectedPort, info.Port)
			}
		})
	}
}

// TestEffectiveDBInfo_DoesNotModifyOriginal tests that original profile is not modified
func TestEffectiveDBInfo_DoesNotModifyOriginal(t *testing.T) {
	profile := &domain.ProfileInfo{
		DBInfo: domain.DBInfo{
			Host:     "original.db.com",
			Port:     3306,
			User:     "originaluser",
			Password: "originalpass",
		},
		SSHTunnel: domain.SSHTunnelConfig{
			Enabled:           true,
			ResolvedLocalPort: 13306,
		},
	}

	// Get effective info
	info := EffectiveDBInfo(profile)

	// Verify effective info is modified
	if info.Host != "127.0.0.1" || info.Port != 13306 {
		t.Error("Effective info should be modified")
	}

	// Verify original profile is NOT modified
	if profile.DBInfo.Host != "original.db.com" {
		t.Error("Original profile Host should not be modified")
	}
	if profile.DBInfo.Port != 3306 {
		t.Error("Original profile Port should not be modified")
	}
}

// TestEffectiveDBInfo_SSHTunnelDisabledWithResolvedPort tests disabled tunnel with resolved port
func TestEffectiveDBInfo_SSHTunnelDisabledWithResolvedPort(t *testing.T) {
	// Edge case: SSH tunnel disabled but has resolved port (should ignore)
	profile := &domain.ProfileInfo{
		DBInfo: domain.DBInfo{
			Host:     "db.example.com",
			Port:     3306,
			User:     "dbuser",
			Password: "dbpass",
		},
		SSHTunnel: domain.SSHTunnelConfig{
			Enabled:           false, // Disabled
			ResolvedLocalPort: 13306, // But has resolved port
		},
	}

	info := EffectiveDBInfo(profile)

	// Should use original DB info (tunnel disabled)
	if info.Host != "db.example.com" {
		t.Errorf("Expected original host (tunnel disabled), got '%s'", info.Host)
	}
	if info.Port != 3306 {
		t.Errorf("Expected original port (tunnel disabled), got %d", info.Port)
	}
}

// TestEffectiveDBInfo_PreservesCredentials tests that credentials are always preserved
func TestEffectiveDBInfo_PreservesCredentials(t *testing.T) {
	testCases := []struct {
		name           string
		sshEnabled     bool
		resolvedPort   int
		expectedHost   string
		expectedPort   int
	}{
		{"direct", false, 0, "db.example.com", 3306},
		{"ssh_resolved", true, 13306, "127.0.0.1", 13306},
		{"ssh_not_resolved", true, 0, "db.example.com", 3306},
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
					Enabled:           tc.sshEnabled,
					ResolvedLocalPort: tc.resolvedPort,
				},
			}

			info := EffectiveDBInfo(profile)

			// Credentials should always be preserved
			if info.User != "testuser" {
				t.Error("User should be preserved")
			}
			if info.Password != "testpass" {
				t.Error("Password should be preserved")
			}

			// Host/Port should match expected
			if info.Host != tc.expectedHost {
				t.Errorf("Expected host '%s', got '%s'", tc.expectedHost, info.Host)
			}
			if info.Port != tc.expectedPort {
				t.Errorf("Expected port %d, got %d", tc.expectedPort, info.Port)
			}
		})
	}
}
