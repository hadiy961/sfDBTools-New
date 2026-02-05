package keys

import (
	"os"
	"testing"
)

func TestResolveProfileEncryptionKey_FromFlag(t *testing.T) {
	// Test dengan key yang sudah ada (dari flag/state)
	existing := "my-secret-key"

	key, source, err := ResolveProfileEncryptionKey(existing, false)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if key != existing {
		t.Errorf("key = %v, want %v", key, existing)
	}
	if source != "flag/state" {
		t.Errorf("source = %v, want 'flag/state'", source)
	}
}

func TestResolveProfileEncryptionKey_FromFlagWithWhitespace(t *testing.T) {
	// Test key dengan whitespace (should be trimmed)
	existing := "  my-secret-key  "

	key, source, err := ResolveProfileEncryptionKey(existing, false)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if key != "my-secret-key" {
		t.Errorf("key = %v, want trimmed 'my-secret-key'", key)
	}
	if source != "flag/state" {
		t.Errorf("source = %v, want 'flag/state'", source)
	}
}

func TestResolveProfileEncryptionKey_FromEnv(t *testing.T) {
	// Setup: Set env var
	testKey := "env-secret-key"
	originalTarget := os.Getenv("SFDB_TARGET_PROFILE_KEY")
	originalSource := os.Getenv("SFDB_SOURCE_PROFILE_KEY")

	os.Setenv("SFDB_TARGET_PROFILE_KEY", testKey)

	// Cleanup
	defer func() {
		if originalTarget != "" {
			os.Setenv("SFDB_TARGET_PROFILE_KEY", originalTarget)
		} else {
			os.Unsetenv("SFDB_TARGET_PROFILE_KEY")
		}
		if originalSource != "" {
			os.Setenv("SFDB_SOURCE_PROFILE_KEY", originalSource)
		} else {
			os.Unsetenv("SFDB_SOURCE_PROFILE_KEY")
		}
	}()

	// Test: Key from env
	key, source, err := ResolveProfileEncryptionKey("", false)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if key != testKey {
		t.Errorf("key = %v, want %v", key, testKey)
	}
	if source != "env" {
		t.Errorf("source = %v, want 'env'", source)
	}
}

func TestResolveProfileEncryptionKey_FallbackToSource(t *testing.T) {
	// Test fallback dari TARGET ke SOURCE env var
	testKey := "source-secret-key"
	originalTarget := os.Getenv("SFDB_TARGET_PROFILE_KEY")
	originalSource := os.Getenv("SFDB_SOURCE_PROFILE_KEY")

	// Unset TARGET, set SOURCE
	os.Unsetenv("SFDB_TARGET_PROFILE_KEY")
	os.Setenv("SFDB_SOURCE_PROFILE_KEY", testKey)

	// Cleanup
	defer func() {
		if originalTarget != "" {
			os.Setenv("SFDB_TARGET_PROFILE_KEY", originalTarget)
		} else {
			os.Unsetenv("SFDB_TARGET_PROFILE_KEY")
		}
		if originalSource != "" {
			os.Setenv("SFDB_SOURCE_PROFILE_KEY", originalSource)
		} else {
			os.Unsetenv("SFDB_SOURCE_PROFILE_KEY")
		}
	}()

	// Test: Should fallback to SOURCE
	key, source, err := ResolveProfileEncryptionKey("", false)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if key != testKey {
		t.Errorf("key = %v, want %v", key, testKey)
	}
	if source != "env" {
		t.Errorf("source = %v, want 'env'", source)
	}
}

func TestResolveProfileEncryptionKey_NonInteractiveError(t *testing.T) {
	// Setup: Clear env vars
	originalTarget := os.Getenv("SFDB_TARGET_PROFILE_KEY")
	originalSource := os.Getenv("SFDB_SOURCE_PROFILE_KEY")

	os.Unsetenv("SFDB_TARGET_PROFILE_KEY")
	os.Unsetenv("SFDB_SOURCE_PROFILE_KEY")

	// Cleanup
	defer func() {
		if originalTarget != "" {
			os.Setenv("SFDB_TARGET_PROFILE_KEY", originalTarget)
		}
		if originalSource != "" {
			os.Setenv("SFDB_SOURCE_PROFILE_KEY", originalSource)
		}
	}()

	// Test: Non-interactive mode should error when no key available
	_, _, err := ResolveProfileEncryptionKey("", false)
	if err == nil {
		t.Error("expected error for non-interactive without key, got nil")
	}
}

func TestResolveProfileEncryptionKey_PriorityOrder(t *testing.T) {
	// Test priority: flag > env
	flagKey := "flag-key"
	envKey := "env-key"

	originalTarget := os.Getenv("SFDB_TARGET_PROFILE_KEY")
	os.Setenv("SFDB_TARGET_PROFILE_KEY", envKey)

	defer func() {
		if originalTarget != "" {
			os.Setenv("SFDB_TARGET_PROFILE_KEY", originalTarget)
		} else {
			os.Unsetenv("SFDB_TARGET_PROFILE_KEY")
		}
	}()

	// Flag should take priority over env
	key, source, err := ResolveProfileEncryptionKey(flagKey, false)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if key != flagKey {
		t.Errorf("key = %v, want flag key %v (not env key %v)", key, flagKey, envKey)
	}
	if source != "flag/state" {
		t.Errorf("source = %v, want 'flag/state'", source)
	}
}

func TestResolveProfileEncryptionKey_EmptyString(t *testing.T) {
	// Test dengan empty string (should fallback to env/prompt)
	originalTarget := os.Getenv("SFDB_TARGET_PROFILE_KEY")
	testKey := "env-fallback-key"
	os.Setenv("SFDB_TARGET_PROFILE_KEY", testKey)

	defer func() {
		if originalTarget != "" {
			os.Setenv("SFDB_TARGET_PROFILE_KEY", originalTarget)
		} else {
			os.Unsetenv("SFDB_TARGET_PROFILE_KEY")
		}
	}()

	// Empty string should trigger fallback
	key, source, err := ResolveProfileEncryptionKey("", false)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if key != testKey {
		t.Errorf("key = %v, want %v", key, testKey)
	}
	if source != "env" {
		t.Errorf("source = %v, want 'env'", source)
	}
}

func TestResolveProfileEncryptionKey_WhitespaceOnly(t *testing.T) {
	// Test dengan whitespace only (should be treated as empty)
	originalTarget := os.Getenv("SFDB_TARGET_PROFILE_KEY")
	testKey := "env-fallback-key"
	os.Setenv("SFDB_TARGET_PROFILE_KEY", testKey)

	defer func() {
		if originalTarget != "" {
			os.Setenv("SFDB_TARGET_PROFILE_KEY", originalTarget)
		} else {
			os.Unsetenv("SFDB_TARGET_PROFILE_KEY")
		}
	}()

	// Whitespace-only should trigger fallback
	key, source, err := ResolveProfileEncryptionKey("   ", false)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if key != testKey {
		t.Errorf("key = %v, want %v", key, testKey)
	}
	if source != "env" {
		t.Errorf("source = %v, want 'env'", source)
	}
}
