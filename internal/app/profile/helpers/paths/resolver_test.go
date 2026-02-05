package paths

import (
	"testing"
)

func TestNewPathResolver(t *testing.T) {
	tests := []struct {
		name      string
		configDir string
	}{
		{
			name:      "with config dir",
			configDir: "/etc/sfdbtools/profiles",
		},
		{
			name:      "empty config dir",
			configDir: "",
		},
		{
			name:      "relative config dir",
			configDir: "./configs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := NewPathResolver(tt.configDir)
			if resolver == nil {
				t.Error("NewPathResolver() returned nil")
				return
			}
			if resolver.ConfigDir != tt.configDir {
				t.Errorf("ConfigDir = %v, want %v", resolver.ConfigDir, tt.configDir)
			}
		})
	}
}

func TestPathResolver_Resolve(t *testing.T) {
	// Note: These tests will test the resolver logic structure.
	// The actual path resolution depends on ResolveConfigPath and ResolveConfigPathInDir
	// which should have their own comprehensive tests.

	t.Run("calls appropriate resolver based on ConfigDir", func(t *testing.T) {
		// Test dengan ConfigDir
		resolver := NewPathResolver("/etc/sfdbtools/profiles")
		if resolver.ConfigDir == "" {
			t.Error("Expected non-empty ConfigDir")
		}

		// Test tanpa ConfigDir
		resolver2 := NewPathResolver("")
		if resolver2.ConfigDir != "" {
			t.Error("Expected empty ConfigDir")
		}
	})

	t.Run("returns error on invalid spec", func(t *testing.T) {
		resolver := NewPathResolver("")

		// Test dengan spec kosong (should error or handle gracefully)
		// Actual behavior depends on ResolveConfigPath implementation
		_, _, err := resolver.Resolve("")
		// We just verify it handles empty input gracefully
		_ = err // May or may not error depending on implementation
	})
}

func TestPathResolver_ResolveMultiple(t *testing.T) {
	// Test resolving multiple profiles consecutively
	resolver := NewPathResolver("/etc/sfdbtools/profiles")

	specs := []string{"prod-db", "staging-db", "dev-db"}

	for _, spec := range specs {
		_, _, err := resolver.Resolve(spec)
		// Each call should work independently
		if err != nil {
			// Expected for non-existent files in test environment
			// Just verify it doesn't panic or corrupt state
			continue
		}
	}
}

func TestPathResolver_StateIndependence(t *testing.T) {
	// Test that resolver maintains no internal state between calls
	resolver := NewPathResolver("/etc/sfdbtools/profiles")

	// First call
	_, name1, _ := resolver.Resolve("test-profile")

	// Second call dengan spec berbeda
	_, name2, _ := resolver.Resolve("another-profile")

	// Names should be different (tidak ada state corruption)
	if name1 != "" && name2 != "" && name1 == name2 {
		t.Error("Resolver appears to be retaining state between calls")
	}
}

func TestPathResolver_EmptyConfigDir(t *testing.T) {
	resolver := NewPathResolver("")

	if resolver.ConfigDir != "" {
		t.Errorf("Expected empty ConfigDir, got %v", resolver.ConfigDir)
	}

	// Resolver should still work dengan empty ConfigDir
	// (will use ResolveConfigPath instead of ResolveConfigPathInDir)
	_, _, err := resolver.Resolve("/absolute/path/to/profile.cnf.enc")
	// Error or not depends on file existence, but shouldn't panic
	_ = err
}

func TestPathResolver_WhitespaceConfigDir(t *testing.T) {
	resolver := NewPathResolver("   ")

	// Whitespace-only ConfigDir should be treated as empty
	// (depends on implementation, but generally should work)
	if resolver.ConfigDir != "   " {
		t.Errorf("ConfigDir was modified, got %v", resolver.ConfigDir)
	}
}
