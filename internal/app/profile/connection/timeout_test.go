// File : internal/app/profile/connection/timeout_test.go
// Deskripsi : Tests for ProfileConnectTimeout function
// Author : Test Suite
// Tanggal : 5 Februari 2026
package connection

import (
	"os"
	"testing"
	"time"
)

// MockConfig implements the config interface for timeout
type MockConfig struct {
	timeout string
}

func (m *MockConfig) GetProfileConnectionTimeout() string {
	return m.timeout
}

// TestProfileConnectTimeout_Default tests default timeout
func TestProfileConnectTimeout_Default(t *testing.T) {
	// Clear any env var
	os.Unsetenv("SFDB_PROFILE_CONNECT_TIMEOUT")

	timeout := ProfileConnectTimeout(nil)

	expectedDefault := 15 * time.Second
	if timeout != expectedDefault {
		t.Errorf("Expected default timeout %v, got %v", expectedDefault, timeout)
	}
}

// TestProfileConnectTimeout_FromConfig tests timeout from config
func TestProfileConnectTimeout_FromConfig(t *testing.T) {
	os.Unsetenv("SFDB_PROFILE_CONNECT_TIMEOUT")

	testCases := []struct {
		name           string
		configTimeout  string
		expectedResult time.Duration
	}{
		{"10_seconds", "10s", 10 * time.Second},
		{"30_seconds", "30s", 30 * time.Second},
		{"1_minute", "1m", 1 * time.Minute},
		{"2_minutes", "2m", 2 * time.Minute},
		{"invalid", "invalid", 15 * time.Second}, // Fallback to default
		{"empty", "", 15 * time.Second},          // Fallback to default
		{"negative", "-5s", 15 * time.Second},    // Fallback to default (negative not allowed)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &MockConfig{timeout: tc.configTimeout}
			timeout := ProfileConnectTimeout(cfg)

			if timeout != tc.expectedResult {
				t.Errorf("Expected %v, got %v", tc.expectedResult, timeout)
			}
		})
	}
}

// TestProfileConnectTimeout_FromEnv tests timeout from environment variable
func TestProfileConnectTimeout_FromEnv(t *testing.T) {
	testCases := []struct {
		name           string
		envValue       string
		expectedResult time.Duration
	}{
		{"20_seconds", "20s", 20 * time.Second},
		{"45_seconds", "45s", 45 * time.Second},
		{"3_minutes", "3m", 3 * time.Minute},
		{"invalid", "notaduration", 15 * time.Second}, // Fallback
		{"empty", "", 15 * time.Second},               // Fallback
		{"negative", "-10s", 15 * time.Second},        // Fallback
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.envValue != "" {
				os.Setenv("SFDB_PROFILE_CONNECT_TIMEOUT", tc.envValue)
			} else {
				os.Unsetenv("SFDB_PROFILE_CONNECT_TIMEOUT")
			}
			defer os.Unsetenv("SFDB_PROFILE_CONNECT_TIMEOUT")

			timeout := ProfileConnectTimeout(nil)

			if timeout != tc.expectedResult {
				t.Errorf("Expected %v, got %v", tc.expectedResult, timeout)
			}
		})
	}
}

// TestProfileConnectTimeout_Priority tests priority order (Config > Env > Default)
func TestProfileConnectTimeout_Priority(t *testing.T) {
	testCases := []struct {
		name           string
		configTimeout  string
		envTimeout     string
		expectedResult time.Duration
	}{
		{
			name:           "config_overrides_env",
			configTimeout:  "25s",
			envTimeout:     "40s",
			expectedResult: 25 * time.Second, // Config wins
		},
		{
			name:           "env_used_when_no_config",
			configTimeout:  "",
			envTimeout:     "35s",
			expectedResult: 35 * time.Second, // Env wins
		},
		{
			name:           "default_when_both_empty",
			configTimeout:  "",
			envTimeout:     "",
			expectedResult: 15 * time.Second, // Default wins
		},
		{
			name:           "config_invalid_fallback_to_env",
			configTimeout:  "invalid",
			envTimeout:     "50s",
			expectedResult: 50 * time.Second, // Fallback to env
		},
		{
			name:           "both_invalid_use_default",
			configTimeout:  "invalid",
			envTimeout:     "notvalid",
			expectedResult: 15 * time.Second, // Fallback to default
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set env
			if tc.envTimeout != "" {
				os.Setenv("SFDB_PROFILE_CONNECT_TIMEOUT", tc.envTimeout)
			} else {
				os.Unsetenv("SFDB_PROFILE_CONNECT_TIMEOUT")
			}
			defer os.Unsetenv("SFDB_PROFILE_CONNECT_TIMEOUT")

			// Set config
			var cfg interface{}
			if tc.configTimeout != "" {
				cfg = &MockConfig{timeout: tc.configTimeout}
			}

			timeout := ProfileConnectTimeout(cfg)

			if timeout != tc.expectedResult {
				t.Errorf("Expected %v, got %v", tc.expectedResult, timeout)
			}
		})
	}
}

// TestProfileConnectTimeout_VariousFormats tests various duration formats
func TestProfileConnectTimeout_VariousFormats(t *testing.T) {
	testCases := []struct {
		name     string
		duration string
		expected time.Duration
	}{
		{"milliseconds", "500ms", 500 * time.Millisecond},
		{"seconds", "30s", 30 * time.Second},
		{"minutes", "5m", 5 * time.Minute},
		{"hours", "1h", 1 * time.Hour},
		{"combined", "1m30s", 90 * time.Second},
		{"complex", "2h30m15s", 2*time.Hour + 30*time.Minute + 15*time.Second},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &MockConfig{timeout: tc.duration}
			timeout := ProfileConnectTimeout(cfg)

			if timeout != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, timeout)
			}
		})
	}
}

// TestProfileConnectTimeout_ZeroNotAllowed tests that zero timeout falls back to default
func TestProfileConnectTimeout_ZeroNotAllowed(t *testing.T) {
	cfg := &MockConfig{timeout: "0s"}
	timeout := ProfileConnectTimeout(cfg)

	// Zero or negative should fallback to default
	if timeout != 15*time.Second {
		t.Errorf("Expected fallback to default (15s) for zero timeout, got %v", timeout)
	}
}

// TestProfileConnectTimeout_NegativeNotAllowed tests that negative timeout falls back
func TestProfileConnectTimeout_NegativeNotAllowed(t *testing.T) {
	cfg := &MockConfig{timeout: "-10s"}
	timeout := ProfileConnectTimeout(cfg)

	// Negative should fallback to default
	if timeout != 15*time.Second {
		t.Errorf("Expected fallback to default (15s) for negative timeout, got %v", timeout)
	}
}

// TestProfileConnectTimeout_NilConfig tests nil config uses env or default
func TestProfileConnectTimeout_NilConfig(t *testing.T) {
	// Test with env set
	os.Setenv("SFDB_PROFILE_CONNECT_TIMEOUT", "25s")
	defer os.Unsetenv("SFDB_PROFILE_CONNECT_TIMEOUT")

	timeout := ProfileConnectTimeout(nil)

	if timeout != 25*time.Second {
		t.Errorf("Expected 25s from env when config is nil, got %v", timeout)
	}

	// Test with env unset
	os.Unsetenv("SFDB_PROFILE_CONNECT_TIMEOUT")
	timeout = ProfileConnectTimeout(nil)

	if timeout != 15*time.Second {
		t.Errorf("Expected default 15s when config nil and no env, got %v", timeout)
	}
}

// TestProfileConnectTimeout_NonConformingConfig tests config that doesn't implement interface
func TestProfileConnectTimeout_NonConformingConfig(t *testing.T) {
	type BadConfig struct {
		SomeOtherField string
	}

	badCfg := &BadConfig{SomeOtherField: "test"}
	
	// Set env to verify it falls through to env
	os.Setenv("SFDB_PROFILE_CONNECT_TIMEOUT", "35s")
	defer os.Unsetenv("SFDB_PROFILE_CONNECT_TIMEOUT")

	timeout := ProfileConnectTimeout(badCfg)

	// Should fallback to env since config doesn't implement interface
	if timeout != 35*time.Second {
		t.Errorf("Expected fallback to env (35s) for non-conforming config, got %v", timeout)
	}
}

// TestProfileConnectTimeout_EnvIsolation tests that env changes are respected
func TestProfileConnectTimeout_EnvIsolation(t *testing.T) {
	// First call with one env value
	os.Setenv("SFDB_PROFILE_CONNECT_TIMEOUT", "20s")
	timeout1 := ProfileConnectTimeout(nil)

	// Change env
	os.Setenv("SFDB_PROFILE_CONNECT_TIMEOUT", "40s")
	timeout2 := ProfileConnectTimeout(nil)

	// Cleanup
	os.Unsetenv("SFDB_PROFILE_CONNECT_TIMEOUT")

	if timeout1 != 20*time.Second {
		t.Errorf("First call: expected 20s, got %v", timeout1)
	}
	if timeout2 != 40*time.Second {
		t.Errorf("Second call: expected 40s, got %v", timeout2)
	}
}
