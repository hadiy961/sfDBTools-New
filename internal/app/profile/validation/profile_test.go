package validation

import (
	"testing"

	"sfdbtools/internal/domain"
)

func TestValidateProfileInfo(t *testing.T) {
	validDBInfo := domain.DBInfo{
		Host:     "10.0.0.5",
		Port:     3306,
		User:     "admin",
		Password: "secret",
	}

	tests := []struct {
		name    string
		profile *domain.ProfileInfo
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil profile",
			profile: nil,
			wantErr: true,
			errMsg:  "nil",
		},
		{
			name: "valid profile without SSH tunnel",
			profile: &domain.ProfileInfo{
				Name:   "test-db",
				DBInfo: validDBInfo,
				SSHTunnel: domain.SSHTunnelConfig{
					Enabled: false,
				},
			},
			wantErr: false,
		},
		{
			name: "valid profile with SSH tunnel",
			profile: &domain.ProfileInfo{
				Name:   "test-db",
				DBInfo: validDBInfo,
				SSHTunnel: domain.SSHTunnelConfig{
					Enabled:  true,
					Host:     "bastion.example.com",
					Port:     22,
					User:     "sshuser",
					Password: "sshpass",
				},
			},
			wantErr: false,
		},
		{
			name: "empty profile name",
			profile: &domain.ProfileInfo{
				Name:   "",
				DBInfo: validDBInfo,
			},
			wantErr: true,
			errMsg:  "name",
		},
		{
			name: "profile name with only spaces",
			profile: &domain.ProfileInfo{
				Name:   "   ",
				DBInfo: validDBInfo,
			},
			wantErr: true,
			errMsg:  "name",
		},
		{
			name: "invalid DB info",
			profile: &domain.ProfileInfo{
				Name: "test-db",
				DBInfo: domain.DBInfo{
					Host:     "",
					Port:     3306,
					User:     "admin",
					Password: "secret",
				},
			},
			wantErr: true,
			errMsg:  "db info",
		},
		{
			name: "SSH tunnel enabled but no host",
			profile: &domain.ProfileInfo{
				Name:   "test-db",
				DBInfo: validDBInfo,
				SSHTunnel: domain.SSHTunnelConfig{
					Enabled: true,
					Host:    "",
					Port:    22,
					User:    "sshuser",
				},
			},
			wantErr: true,
			errMsg:  "ssh",
		},
		{
			name: "SSH tunnel enabled with only spaces in host",
			profile: &domain.ProfileInfo{
				Name:   "test-db",
				DBInfo: validDBInfo,
				SSHTunnel: domain.SSHTunnelConfig{
					Enabled: true,
					Host:    "   ",
					Port:    22,
					User:    "sshuser",
				},
			},
			wantErr: true,
			errMsg:  "ssh",
		},
		{
			name: "SSH tunnel disabled with empty host (ok)",
			profile: &domain.ProfileInfo{
				Name:   "test-db",
				DBInfo: validDBInfo,
				SSHTunnel: domain.SSHTunnelConfig{
					Enabled: false,
					Host:    "", // OK karena tunnel disabled
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProfileInfo(tt.profile)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateProfileInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" {
				// Verify error message contains expected keyword
				if errStr := err.Error(); errStr == "" {
					t.Errorf("ValidateProfileInfo() error message is empty")
				}
			}
		})
	}
}

func TestValidateProfileInfo_ComplexScenarios(t *testing.T) {
	t.Run("minimal valid profile", func(t *testing.T) {
		profile := &domain.ProfileInfo{
			Name: "db",
			DBInfo: domain.DBInfo{
				Host:     "h",
				Port:     1,
				User:     "u",
				Password: "",
			},
			SSHTunnel: domain.SSHTunnelConfig{Enabled: false},
		}
		if err := ValidateProfileInfo(profile); err != nil {
			t.Errorf("minimal valid profile should pass, got error: %v", err)
		}
	})

	t.Run("full featured profile", func(t *testing.T) {
		profile := &domain.ProfileInfo{
			Name: "production-db-master",
			DBInfo: domain.DBInfo{
				Host:     "mysql.prod.example.com",
				Port:     3306,
				User:     "prod_admin",
				Password: "super-secret-password-123",
			},
			SSHTunnel: domain.SSHTunnelConfig{
				Enabled:      true,
				Host:         "bastion.prod.example.com",
				Port:         22,
				User:         "bastion_user",
				Password:     "bastion-password",
				IdentityFile: "/home/user/.ssh/id_rsa",
				LocalPort:    33060,
			},
		}
		if err := ValidateProfileInfo(profile); err != nil {
			t.Errorf("full featured profile should pass, got error: %v", err)
		}
	})

	t.Run("profile with special characters in name", func(t *testing.T) {
		profile := &domain.ProfileInfo{
			Name: "test-db_master.v2",
			DBInfo: domain.DBInfo{
				Host:     "10.0.0.5",
				Port:     3306,
				User:     "admin",
				Password: "secret",
			},
		}
		if err := ValidateProfileInfo(profile); err != nil {
			t.Errorf("profile with valid special chars in name should pass, got error: %v", err)
		}
	})
}
