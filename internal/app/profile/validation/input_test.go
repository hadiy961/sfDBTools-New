package validation

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateNoLeadingTrailingSpace(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		fieldName string
		wantErr   bool
	}{
		{
			name:      "valid input",
			input:     "hostname",
			fieldName: "Host",
			wantErr:   false,
		},
		{
			name:      "leading space",
			input:     " hostname",
			fieldName: "Host",
			wantErr:   true,
		},
		{
			name:      "trailing space",
			input:     "hostname ",
			fieldName: "Host",
			wantErr:   true,
		},
		{
			name:      "both leading and trailing",
			input:     " hostname ",
			fieldName: "Host",
			wantErr:   true,
		},
		{
			name:      "spaces in middle (ok)",
			input:     "host name",
			fieldName: "Description",
			wantErr:   false,
		},
		{
			name:      "empty field name",
			input:     " hostname",
			fieldName: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNoLeadingTrailingSpace(tt.input, tt.fieldName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateNoLeadingTrailingSpace() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateNoControlChars(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		fieldName string
		wantErr   bool
	}{
		{
			name:      "valid input",
			input:     "hostname-123",
			fieldName: "Host",
			wantErr:   false,
		},
		{
			name:      "with null byte",
			input:     "host\x00name",
			fieldName: "Host",
			wantErr:   true,
		},
		{
			name:      "with bell character",
			input:     "host\x07name",
			fieldName: "Host",
			wantErr:   true,
		},
		{
			name:      "newline allowed",
			input:     "line1\nline2",
			fieldName: "Description",
			wantErr:   false,
		},
		{
			name:      "tab allowed",
			input:     "column1\tcolumn2",
			fieldName: "Data",
			wantErr:   false,
		},
		{
			name:      "carriage return allowed",
			input:     "line1\r\nline2",
			fieldName: "Text",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNoControlChars(tt.input, tt.fieldName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateNoControlChars() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateNoSpaces(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		fieldName string
		wantErr   bool
	}{
		{
			name:      "valid no spaces",
			input:     "hostname",
			fieldName: "Host",
			wantErr:   false,
		},
		{
			name:      "with dash",
			input:     "host-name",
			fieldName: "Host",
			wantErr:   false,
		},
		{
			name:      "with underscore",
			input:     "host_name",
			fieldName: "Host",
			wantErr:   false,
		},
		{
			name:      "with space",
			input:     "host name",
			fieldName: "Host",
			wantErr:   true,
		},
		{
			name:      "multiple spaces",
			input:     "host  name  server",
			fieldName: "Host",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNoSpaces(tt.input, tt.fieldName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateNoSpaces() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateNotEmpty(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		fieldName string
		wantErr   bool
	}{
		{
			name:      "valid non-empty",
			input:     "value",
			fieldName: "Field",
			wantErr:   false,
		},
		{
			name:      "empty string",
			input:     "",
			fieldName: "Field",
			wantErr:   true,
		},
		{
			name:      "only spaces",
			input:     "   ",
			fieldName: "Field",
			wantErr:   true,
		},
		{
			name:      "single character",
			input:     "a",
			fieldName: "Field",
			wantErr:   false,
		},
		{
			name:      "whitespace trimmed",
			input:     "  value  ",
			fieldName: "Field",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNotEmpty(tt.input, tt.fieldName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateNotEmpty() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateIntInRange(t *testing.T) {
	tests := []struct {
		name      string
		value     int
		min       int
		max       int
		allowZero bool
		fieldName string
		wantErr   bool
	}{
		{
			name:      "valid port",
			value:     3306,
			min:       1,
			max:       65535,
			allowZero: false,
			fieldName: "Port",
			wantErr:   false,
		},
		{
			name:      "port at minimum",
			value:     1,
			min:       1,
			max:       65535,
			allowZero: false,
			fieldName: "Port",
			wantErr:   false,
		},
		{
			name:      "port at maximum",
			value:     65535,
			min:       1,
			max:       65535,
			allowZero: false,
			fieldName: "Port",
			wantErr:   false,
		},
		{
			name:      "port below minimum",
			value:     0,
			min:       1,
			max:       65535,
			allowZero: false,
			fieldName: "Port",
			wantErr:   true,
		},
		{
			name:      "port above maximum",
			value:     70000,
			min:       1,
			max:       65535,
			allowZero: false,
			fieldName: "Port",
			wantErr:   true,
		},
		{
			name:      "zero allowed and provided",
			value:     0,
			min:       1,
			max:       65535,
			allowZero: true,
			fieldName: "LocalPort",
			wantErr:   false,
		},
		{
			name:      "negative value",
			value:     -1,
			min:       1,
			max:       100,
			allowZero: false,
			fieldName: "Count",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateIntInRange(tt.value, tt.min, tt.max, tt.allowZero, tt.fieldName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateIntInRange() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateFileAccessible(t *testing.T) {
	// Create temp file for testing
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test-file.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	tests := []struct {
		name      string
		path      string
		fieldName string
		wantErr   bool
	}{
		{
			name:      "empty path (optional)",
			path:      "",
			fieldName: "IdentityFile",
			wantErr:   false,
		},
		{
			name:      "whitespace only (optional)",
			path:      "   ",
			fieldName: "IdentityFile",
			wantErr:   false,
		},
		{
			name:      "valid file",
			path:      tmpFile,
			fieldName: "IdentityFile",
			wantErr:   false,
		},
		{
			name:      "directory not file",
			path:      tmpDir,
			fieldName: "IdentityFile",
			wantErr:   true,
		},
		{
			name:      "non-existent file",
			path:      filepath.Join(tmpDir, "not-exists.txt"),
			fieldName: "IdentityFile",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFileAccessible(tt.path, tt.fieldName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFileAccessible() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateConfigName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid simple name",
			input:   "prod-db",
			wantErr: false,
		},
		{
			name:    "valid with underscore",
			input:   "prod_db_master",
			wantErr: false,
		},
		{
			name:    "valid with numbers",
			input:   "db-server-01",
			wantErr: false,
		},
		{
			name:    "empty name",
			input:   "",
			wantErr: true,
		},
		{
			name:    "only spaces",
			input:   "   ",
			wantErr: true,
		},
		{
			name:    "leading space",
			input:   " prod-db",
			wantErr: true,
		},
		{
			name:    "trailing space",
			input:   "prod-db ",
			wantErr: true,
		},
		{
			name:    "with control char",
			input:   "prod\x00db",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfigName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfigName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateHost(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid hostname",
			input:   "database.example.com",
			wantErr: false,
		},
		{
			name:    "valid IP",
			input:   "10.0.0.5",
			wantErr: false,
		},
		{
			name:    "localhost",
			input:   "localhost",
			wantErr: false,
		},
		{
			name:    "empty host",
			input:   "",
			wantErr: true,
		},
		{
			name:    "with space",
			input:   "host name",
			wantErr: true,
		},
		{
			name:    "leading space",
			input:   " 10.0.0.5",
			wantErr: true,
		},
		{
			name:    "trailing space",
			input:   "10.0.0.5 ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateHost(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateHost() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid simple username",
			input:   "admin",
			wantErr: false,
		},
		{
			name:    "valid with underscore",
			input:   "db_user",
			wantErr: false,
		},
		{
			name:    "valid with numbers",
			input:   "user123",
			wantErr: false,
		},
		{
			name:    "empty username",
			input:   "",
			wantErr: true,
		},
		{
			name:    "only spaces",
			input:   "   ",
			wantErr: true,
		},
		{
			name:    "leading space",
			input:   " admin",
			wantErr: true,
		},
		{
			name:    "trailing space",
			input:   "admin ",
			wantErr: true,
		},
		{
			name:    "spaces in middle (allowed for usernames)",
			input:   "admin user",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUsername(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUsername() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
