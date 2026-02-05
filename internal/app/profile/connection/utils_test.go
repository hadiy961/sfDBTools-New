// File : internal/app/profile/connection/utils_test.go
// Deskripsi : Tests for utility functions (TrimProfileSuffix, etc)
// Author : Test Suite
// Tanggal : 5 Februari 2026
package connection

import (
	"testing"
)

// TestTrimProfileSuffix_Basic tests basic suffix trimming
func TestTrimProfileSuffix_Basic(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"double_extension", "prod-db.cnf.enc", "prod-db"},
		{"cnf_only", "prod-db.cnf", "prod-db"},
		{"enc_only", "prod-db.enc", "prod-db"},
		{"no_extension", "prod-db", "prod-db"},
		{"empty_string", "", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := TrimProfileSuffix(tc.input)
			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

// TestTrimProfileSuffix_Order tests that .enc is trimmed before .cnf
func TestTrimProfileSuffix_Order(t *testing.T) {
	// Should trim .enc first, then .cnf
	result := TrimProfileSuffix("profile.cnf.enc")
	expected := "profile"

	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

// TestTrimProfileSuffix_Whitespace tests whitespace handling
func TestTrimProfileSuffix_Whitespace(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"leading_space", "  profile.cnf.enc", "profile"},
		{"trailing_space", "profile.cnf.enc  ", "profile"},
		{"both_spaces", "  profile.cnf.enc  ", "profile"},
		{"only_spaces", "   ", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := TrimProfileSuffix(tc.input)
			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

// TestTrimProfileSuffix_MultipleExtensions tests profiles with multiple extensions
func TestTrimProfileSuffix_MultipleExtensions(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"double_enc", "profile.enc.enc", "profile.enc"}, // Only one .enc trimmed
		{"double_cnf", "profile.cnf.cnf", "profile.cnf"}, // Only one .cnf trimmed
		{"mixed", "profile.enc.cnf", "profile.enc"},      // Only .cnf trimmed (order matters)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := TrimProfileSuffix(tc.input)
			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

// TestTrimProfileSuffix_SpecialNames tests special profile names
func TestTrimProfileSuffix_SpecialNames(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"with_dash", "prod-db-01.cnf.enc", "prod-db-01"},
		{"with_underscore", "prod_db_01.cnf.enc", "prod_db_01"},
		{"with_dot_in_name", "prod.db.cnf.enc", "prod.db"},
		{"numbers_only", "12345.cnf.enc", "12345"},
		{"mixed_case", "ProdDB.CNF.ENC", "ProdDB.CNF.ENC"}, // Case-sensitive
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := TrimProfileSuffix(tc.input)
			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

// TestTrimProfileSuffix_EdgeCases tests edge cases
func TestTrimProfileSuffix_EdgeCases(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"just_cnf_enc", ".cnf.enc", ""},
		{"just_cnf", ".cnf", ""},
		{"just_enc", ".enc", ""},
		{"double_dot", "profile..cnf.enc", "profile."},
		{"unicode", "プロファイル.cnf.enc", "プロファイル"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := TrimProfileSuffix(tc.input)
			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

// TestTrimProfileSuffix_PathsNotHandled tests that paths are not handled
func TestTrimProfileSuffix_PathsNotHandled(t *testing.T) {
	// Function only handles basename, not full paths
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"with_slash", "/path/to/profile.cnf.enc", "/path/to/profile"},
		{"windows_path", "C:\\profiles\\prod.cnf.enc", "C:\\profiles\\prod"},
		{"relative_path", "./profiles/test.cnf.enc", "./profiles/test"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := TrimProfileSuffix(tc.input)
			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

// TestTrimProfileSuffix_CasesSensitivity tests case sensitivity
func TestTrimProfileSuffix_CaseSensitivity(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"uppercase_ext", "profile.CNF.ENC", "profile.CNF.ENC"}, // Not trimmed (case-sensitive)
		{"mixedcase_ext", "profile.Cnf.Enc", "profile.Cnf.Enc"}, // Not trimmed
		{"lowercase", "profile.cnf.enc", "profile"},             // Trimmed
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := TrimProfileSuffix(tc.input)
			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

// TestTrimProfileSuffix_LongNames tests long profile names
func TestTrimProfileSuffix_LongNames(t *testing.T) {
	longName := "this-is-a-very-long-profile-name-with-many-characters-and-dashes-and-underscores_123.cnf.enc"
	expected := "this-is-a-very-long-profile-name-with-many-characters-and-dashes-and-underscores_123"

	result := TrimProfileSuffix(longName)
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

// TestTrimProfileSuffix_Idempotent tests that running twice gives same result
func TestTrimProfileSuffix_Idempotent(t *testing.T) {
	input := "profile.cnf.enc"
	
	// First trim
	result1 := TrimProfileSuffix(input)
	
	// Second trim on result
	result2 := TrimProfileSuffix(result1)

	// Should be the same
	if result1 != result2 {
		t.Errorf("Function should be idempotent: '%s' != '%s'", result1, result2)
	}

	if result2 != "profile" {
		t.Errorf("Expected 'profile', got '%s'", result2)
	}
}

// TestTrimProfileSuffix_MultipleProfiles tests batch processing
func TestTrimProfileSuffix_MultipleProfiles(t *testing.T) {
	profiles := []struct {
		input    string
		expected string
	}{
		{"prod-db1.cnf.enc", "prod-db1"},
		{"prod-db2.cnf.enc", "prod-db2"},
		{"staging-db.cnf.enc", "staging-db"},
		{"dev-db.cnf", "dev-db"},
		{"test-db", "test-db"},
	}

	for _, p := range profiles {
		result := TrimProfileSuffix(p.input)
		if result != p.expected {
			t.Errorf("For '%s': expected '%s', got '%s'", p.input, p.expected, result)
		}
	}
}
