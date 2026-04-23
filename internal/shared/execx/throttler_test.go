package execx

import (
	"testing"
)

func TestParseSpeed(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int64
		wantErr bool
	}{
		{"empty", "", 0, false},
		{"bytes only", "1024", 1024, false},
		{"KB", "10KB", 10 * 1024, false},
		{"MB", "5MB", 5 * 1024 * 1024, false},
		{"GB", "2 GB", 2 * 1024 * 1024 * 1024, false},
		{"MB/s format", "10MB/s", 10 * 1024 * 1024, false},
		{"lowercase", "500kb", 500 * 1024, false},
		{"invalid unit", "10XY", 0, true},
		{"invalid format", "abc", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSpeed(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSpeed() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseSpeed() = %v, want %v", got, tt.want)
			}
		})
	}
}
