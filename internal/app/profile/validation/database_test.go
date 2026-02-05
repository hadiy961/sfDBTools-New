package validation

import (
	"testing"

	"sfdbtools/internal/domain"
)

func TestValidateDBInfo(t *testing.T) {
	tests := []struct {
		name    string
		dbInfo  *domain.DBInfo
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil DBInfo",
			dbInfo:  nil,
			wantErr: true,
			errMsg:  "DB info is nil",
		},
		{
			name: "valid DBInfo",
			dbInfo: &domain.DBInfo{
				Host:     "10.0.0.5",
				Port:     3306,
				User:     "admin",
				Password: "secret",
			},
			wantErr: false,
		},
		{
			name: "valid with empty password",
			dbInfo: &domain.DBInfo{
				Host:     "10.0.0.5",
				Port:     3306,
				User:     "admin",
				Password: "", // Password dapat kosong
			},
			wantErr: false,
		},
		{
			name: "empty host",
			dbInfo: &domain.DBInfo{
				Host:     "",
				Port:     3306,
				User:     "admin",
				Password: "secret",
			},
			wantErr: true,
			errMsg:  "host",
		},
		{
			name: "host with only spaces",
			dbInfo: &domain.DBInfo{
				Host:     "   ",
				Port:     3306,
				User:     "admin",
				Password: "secret",
			},
			wantErr: true,
			errMsg:  "host",
		},
		{
			name: "port zero",
			dbInfo: &domain.DBInfo{
				Host:     "10.0.0.5",
				Port:     0,
				User:     "admin",
				Password: "secret",
			},
			wantErr: true,
			errMsg:  "port",
		},
		{
			name: "port negative",
			dbInfo: &domain.DBInfo{
				Host:     "10.0.0.5",
				Port:     -1,
				User:     "admin",
				Password: "secret",
			},
			wantErr: true,
			errMsg:  "port",
		},
		{
			name: "port above maximum",
			dbInfo: &domain.DBInfo{
				Host:     "10.0.0.5",
				Port:     70000,
				User:     "admin",
				Password: "secret",
			},
			wantErr: true,
			errMsg:  "port",
		},
		{
			name: "port at minimum (1)",
			dbInfo: &domain.DBInfo{
				Host:     "10.0.0.5",
				Port:     1,
				User:     "admin",
				Password: "secret",
			},
			wantErr: false,
		},
		{
			name: "port at maximum (65535)",
			dbInfo: &domain.DBInfo{
				Host:     "10.0.0.5",
				Port:     65535,
				User:     "admin",
				Password: "secret",
			},
			wantErr: false,
		},
		{
			name: "empty user",
			dbInfo: &domain.DBInfo{
				Host:     "10.0.0.5",
				Port:     3306,
				User:     "",
				Password: "secret",
			},
			wantErr: true,
			errMsg:  "user",
		},
		{
			name: "user with only spaces",
			dbInfo: &domain.DBInfo{
				Host:     "10.0.0.5",
				Port:     3306,
				User:     "   ",
				Password: "secret",
			},
			wantErr: true,
			errMsg:  "user",
		},
		{
			name: "all fields minimal valid",
			dbInfo: &domain.DBInfo{
				Host:     "h",
				Port:     1,
				User:     "u",
				Password: "",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDBInfo(tt.dbInfo)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDBInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" {
				// Check if error message contains expected substring
				if errStr := err.Error(); errStr == "" {
					t.Errorf("ValidateDBInfo() error message is empty")
				}
			}
		})
	}
}

func TestValidateDBInfo_PortEdgeCases(t *testing.T) {
	// Test specific port values yang sering dipakai
	commonPorts := []struct {
		port    int
		name    string
		wantErr bool
	}{
		{3306, "MySQL default", false},
		{3307, "MySQL alternative", false},
		{5432, "PostgreSQL", false},
		{1433, "SQL Server", false},
		{27017, "MongoDB", false},
		{6379, "Redis", false},
		{0, "zero (invalid)", true},
		{-1, "negative", true},
		{65536, "above max", true},
		{100000, "way above max", true},
	}

	for _, tc := range commonPorts {
		t.Run(tc.name, func(t *testing.T) {
			dbInfo := &domain.DBInfo{
				Host:     "localhost",
				Port:     tc.port,
				User:     "testuser",
				Password: "testpass",
			}
			err := ValidateDBInfo(dbInfo)
			if (err != nil) != tc.wantErr {
				t.Errorf("Port %d: error = %v, wantErr %v", tc.port, err, tc.wantErr)
			}
		})
	}
}
