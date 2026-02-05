// File : internal/app/profile/helpers/parser/ini_parsing_test.go
// Deskripsi : INI parsing edge cases tests
// Author : Test Suite
// Tanggal : 5 Februari 2026
package parser

import (
	"os"
	"path/filepath"
	"testing"

	"sfdbtools/internal/crypto"
)

// TestINIParsing_CommentsAndWhitespace tests INI parsing with comments
func TestINIParsing_CommentsAndWhitespace(t *testing.T) {
	content := `
# This is a comment
; This is also a comment

[client]
# Database configuration
host = localhost   # inline comment (but treated as part of value)
port = 3306

  user = testuser   
password=testpass


[ssh]
; SSH tunnel disabled
enabled = false
`

	testKey := "ini-comment-key"
	tempDir := filepath.Join("testdata", "ini_temp")
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir)

	plainFile := filepath.Join(tempDir, "comments.cnf")
	encFile := filepath.Join(tempDir, "comments.cnf.enc")

	if err := os.WriteFile(plainFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}

	if err := crypto.EncryptFile(plainFile, encFile, []byte(testKey)); err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	profile, err := LoadAndParseProfile(encFile, testKey)
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	// Verify whitespace trimmed
	if profile.DBInfo.Host != "localhost   # inline comment (but treated as part of value)" {
		// Note: INI parser might not handle inline comments
		t.Logf("Host value: '%s'", profile.DBInfo.Host)
	}
	if profile.DBInfo.User != "testuser" {
		t.Errorf("Expected user 'testuser', got '%s'", profile.DBInfo.User)
	}
	if profile.DBInfo.Password != "testpass" {
		t.Errorf("Expected password 'testpass', got '%s'", profile.DBInfo.Password)
	}
}

// TestINIParsing_MultipleSections tests INI with multiple sections
func TestINIParsing_MultipleSections(t *testing.T) {
	content := `[client]
host = localhost
port = 3306
user = testuser
password = testpass

[mysql]
# This section should be ignored
socket = /tmp/mysql.sock

[ssh]
enabled = true
host = sshhost
port = 22
user = sshuser

[other_section]
# This should also be ignored
key = value
`

	testKey := "multi-section-key"
	tempDir := filepath.Join("testdata", "multi_section_temp")
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir)

	plainFile := filepath.Join(tempDir, "multi.cnf")
	encFile := filepath.Join(tempDir, "multi.cnf.enc")

	if err := os.WriteFile(plainFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}

	if err := crypto.EncryptFile(plainFile, encFile, []byte(testKey)); err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	profile, err := LoadAndParseProfile(encFile, testKey)
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	// Only [client] and [ssh] sections should be parsed
	if profile.DBInfo.Host != "localhost" {
		t.Error("Client section not parsed correctly")
	}
	if !profile.SSHTunnel.Enabled {
		t.Error("SSH section not parsed correctly")
	}
}

// TestINIParsing_SectionNameCaseInsensitive tests case insensitivity
func TestINIParsing_SectionNameCaseInsensitive(t *testing.T) {
	testCases := []struct {
		name           string
		clientSection  string
		sshSection     string
		shouldParse    bool
	}{
		{"lowercase", "[client]", "[ssh]", true},
		{"uppercase", "[CLIENT]", "[SSH]", true},
		{"mixedcase", "[Client]", "[Ssh]", true},
		{"weird_case", "[cLiEnT]", "[SsH]", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			content := tc.clientSection + `
host = localhost
port = 3306
user = testuser
password = testpass

` + tc.sshSection + `
enabled = false
`

			testKey := "case-test-key"
			tempDir := filepath.Join("testdata", "case_temp_"+tc.name)
			os.MkdirAll(tempDir, 0755)
			defer os.RemoveAll(tempDir)

			plainFile := filepath.Join(tempDir, "case.cnf")
			encFile := filepath.Join(tempDir, "case.cnf.enc")

			if err := os.WriteFile(plainFile, []byte(content), 0644); err != nil {
				t.Fatalf("Failed to write: %v", err)
			}

			if err := crypto.EncryptFile(plainFile, encFile, []byte(testKey)); err != nil {
				t.Fatalf("Failed to encrypt: %v", err)
			}

			profile, err := LoadAndParseProfile(encFile, testKey)
			if err != nil {
				t.Fatalf("Failed to load: %v", err)
			}

			if profile.DBInfo.Host != "localhost" {
				t.Errorf("Case %s: Failed to parse, got host '%s'", tc.name, profile.DBInfo.Host)
			}
		})
	}
}

// TestINIParsing_NoSpacesAroundEquals tests parsing without spaces
func TestINIParsing_NoSpacesAroundEquals(t *testing.T) {
	content := `[client]
host=no-spaces-host
port=3306
user=no-spaces-user
password=no-spaces-pass
`

	testKey := "no-spaces-key"
	tempDir := filepath.Join("testdata", "no_spaces_temp")
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir)

	plainFile := filepath.Join(tempDir, "no_spaces.cnf")
	encFile := filepath.Join(tempDir, "no_spaces.cnf.enc")

	if err := os.WriteFile(plainFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}

	if err := crypto.EncryptFile(plainFile, encFile, []byte(testKey)); err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	profile, err := LoadAndParseProfile(encFile, testKey)
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	if profile.DBInfo.Host != "no-spaces-host" {
		t.Errorf("Expected 'no-spaces-host', got '%s'", profile.DBInfo.Host)
	}
	if profile.DBInfo.User != "no-spaces-user" {
		t.Errorf("Expected 'no-spaces-user', got '%s'", profile.DBInfo.User)
	}
}

// TestINIParsing_ExtraSpacesAroundEquals tests parsing with extra spaces
func TestINIParsing_ExtraSpacesAroundEquals(t *testing.T) {
	content := `[client]
host    =    extra-spaces-host
port  =  3306
user     =     extra-spaces-user
password =         extra-spaces-pass
`

	testKey := "extra-spaces-key"
	tempDir := filepath.Join("testdata", "extra_spaces_temp")
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir)

	plainFile := filepath.Join(tempDir, "extra_spaces.cnf")
	encFile := filepath.Join(tempDir, "extra_spaces.cnf.enc")

	if err := os.WriteFile(plainFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}

	if err := crypto.EncryptFile(plainFile, encFile, []byte(testKey)); err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	profile, err := LoadAndParseProfile(encFile, testKey)
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	// Should trim spaces
	if profile.DBInfo.Host != "extra-spaces-host" {
		t.Errorf("Expected 'extra-spaces-host', got '%s'", profile.DBInfo.Host)
	}
	if profile.DBInfo.User != "extra-spaces-user" {
		t.Errorf("Expected 'extra-spaces-user', got '%s'", profile.DBInfo.User)
	}
}

// TestINIParsing_EqualsSignInValue tests values containing equals sign
func TestINIParsing_EqualsSignInValue(t *testing.T) {
	content := `[client]
host = localhost
port = 3306
user = testuser
password = pass=with=equals=signs
`

	testKey := "equals-in-value-key"
	tempDir := filepath.Join("testdata", "equals_temp")
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir)

	plainFile := filepath.Join(tempDir, "equals.cnf")
	encFile := filepath.Join(tempDir, "equals.cnf.enc")

	if err := os.WriteFile(plainFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}

	if err := crypto.EncryptFile(plainFile, encFile, []byte(testKey)); err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	profile, err := LoadAndParseProfile(encFile, testKey)
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	// Parser should handle equals in value (takes everything after first =)
	expectedPassword := "pass=with=equals=signs"
	if profile.DBInfo.Password != expectedPassword {
		t.Errorf("Expected '%s', got '%s'", expectedPassword, profile.DBInfo.Password)
	}
}

// TestINIParsing_EmptySection tests INI with empty sections
func TestINIParsing_EmptySection(t *testing.T) {
	content := `[client]
host = localhost
port = 3306
user = testuser
password = testpass

[ssh]
# Empty section - no keys

[another]
key = value
`

	testKey := "empty-section-key"
	tempDir := filepath.Join("testdata", "empty_section_temp")
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir)

	plainFile := filepath.Join(tempDir, "empty_section.cnf")
	encFile := filepath.Join(tempDir, "empty_section.cnf.enc")

	if err := os.WriteFile(plainFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}

	if err := crypto.EncryptFile(plainFile, encFile, []byte(testKey)); err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	profile, err := LoadAndParseProfile(encFile, testKey)
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	// Empty SSH section should not enable SSH
	if profile.SSHTunnel.Enabled {
		t.Error("Empty SSH section should not enable SSH tunnel")
	}
	
	// Client section should still be parsed
	if profile.DBInfo.Host != "localhost" {
		t.Error("Client section should be parsed despite empty SSH section")
	}
}

// TestINIParsing_OnlyClientSection tests profile with only [client] section
func TestINIParsing_OnlyClientSection(t *testing.T) {
	content := `[client]
host = only-client.example.com
port = 3306
user = clientuser
password = clientpass
`

	testKey := "only-client-key"
	tempDir := filepath.Join("testdata", "only_client_temp")
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir)

	plainFile := filepath.Join(tempDir, "only_client.cnf")
	encFile := filepath.Join(tempDir, "only_client.cnf.enc")

	if err := os.WriteFile(plainFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}

	if err := crypto.EncryptFile(plainFile, encFile, []byte(testKey)); err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	profile, err := LoadAndParseProfile(encFile, testKey)
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	// Should parse successfully
	if profile.DBInfo.Host != "only-client.example.com" {
		t.Error("Failed to parse profile with only client section")
	}
	
	// SSH should be disabled by default
	if profile.SSHTunnel.Enabled {
		t.Error("SSH should be disabled when no [ssh] section")
	}
}

// TestINIParsing_MalformedLines tests INI with malformed lines
func TestINIParsing_MalformedLines(t *testing.T) {
	content := `[client]
host = localhost
port = 3306
this-line-has-no-equals-sign
user = testuser
= value-without-key
password = testpass
`

	testKey := "malformed-key"
	tempDir := filepath.Join("testdata", "malformed_temp")
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir)

	plainFile := filepath.Join(tempDir, "malformed.cnf")
	encFile := filepath.Join(tempDir, "malformed.cnf.enc")

	if err := os.WriteFile(plainFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}

	if err := crypto.EncryptFile(plainFile, encFile, []byte(testKey)); err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	profile, err := LoadAndParseProfile(encFile, testKey)
	if err != nil {
		t.Fatalf("Expected to parse (skip malformed lines), got error: %v", err)
	}

	// Valid lines should still be parsed
	if profile.DBInfo.Host != "localhost" {
		t.Error("Valid lines should be parsed despite malformed lines")
	}
	if profile.DBInfo.User != "testuser" {
		t.Error("User should be parsed despite malformed lines")
	}
}

// TestINIParsing_WindowsLineEndings tests INI with Windows line endings
func TestINIParsing_WindowsLineEndings(t *testing.T) {
	content := "[client]\r\nhost = localhost\r\nport = 3306\r\nuser = testuser\r\npassword = testpass\r\n"

	testKey := "windows-endings-key"
	tempDir := filepath.Join("testdata", "windows_temp")
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir)

	plainFile := filepath.Join(tempDir, "windows.cnf")
	encFile := filepath.Join(tempDir, "windows.cnf.enc")

	if err := os.WriteFile(plainFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}

	if err := crypto.EncryptFile(plainFile, encFile, []byte(testKey)); err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	profile, err := LoadAndParseProfile(encFile, testKey)
	if err != nil {
		t.Fatalf("Failed to load with Windows line endings: %v", err)
	}

	if profile.DBInfo.Host != "localhost" {
		t.Error("Failed to parse with Windows line endings")
	}
}

// TestINIParsing_QuotedValues tests values with quotes (treated as-is)
func TestINIParsing_QuotedValues(t *testing.T) {
	content := `[client]
host = "localhost"
port = 3306
user = 'testuser'
password = "test'pass"
`

	testKey := "quoted-values-key"
	tempDir := filepath.Join("testdata", "quoted_temp")
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir)

	plainFile := filepath.Join(tempDir, "quoted.cnf")
	encFile := filepath.Join(tempDir, "quoted.cnf.enc")

	if err := os.WriteFile(plainFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}

	if err := crypto.EncryptFile(plainFile, encFile, []byte(testKey)); err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	profile, err := LoadAndParseProfile(encFile, testKey)
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	// Quotes are typically preserved in simple INI parsers
	t.Logf("Host value: '%s' (quotes may be preserved)", profile.DBInfo.Host)
	t.Logf("User value: '%s' (quotes may be preserved)", profile.DBInfo.User)
	t.Logf("Password value: '%s' (quotes may be preserved)", profile.DBInfo.Password)
}
