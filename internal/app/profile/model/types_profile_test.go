package model

import (
	"testing"

	"sfdbtools/internal/domain"
	"sfdbtools/internal/shared/consts"
)

// TestProfileCreateOptions_Mode tests ProfileCreateOptions implements ProfileOptions
func TestProfileCreateOptions_Mode(t *testing.T) {
	opts := &ProfileCreateOptions{}
	if got := opts.Mode(); got != consts.ProfileModeCreate {
		t.Errorf("Mode() = %v, want %v", got, consts.ProfileModeCreate)
	}
}

func TestProfileCreateOptions_IsInteractive(t *testing.T) {
	tests := []struct {
		name string
		opts *ProfileCreateOptions
		want bool
	}{
		{
			name: "nil options",
			opts: nil,
			want: false,
		},
		{
			name: "interactive true",
			opts: &ProfileCreateOptions{Interactive: true},
			want: true,
		},
		{
			name: "interactive false",
			opts: &ProfileCreateOptions{Interactive: false},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.opts.IsInteractive(); got != tt.want {
				t.Errorf("IsInteractive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProfileEditOptions_Mode(t *testing.T) {
	opts := &ProfileEditOptions{}
	if got := opts.Mode(); got != consts.ProfileModeEdit {
		t.Errorf("Mode() = %v, want %v", got, consts.ProfileModeEdit)
	}
}

func TestProfileShowOptions_Mode(t *testing.T) {
	opts := &ProfileShowOptions{}
	if got := opts.Mode(); got != consts.ProfileModeShow {
		t.Errorf("Mode() = %v, want %v", got, consts.ProfileModeShow)
	}
}

func TestProfileDeleteOptions_Mode(t *testing.T) {
	opts := &ProfileDeleteOptions{}
	if got := opts.Mode(); got != consts.ProfileModeDelete {
		t.Errorf("Mode() = %v, want %v", got, consts.ProfileModeDelete)
	}
}

func TestProfileCloneOptions_Mode(t *testing.T) {
	opts := &ProfileCloneOptions{}
	if got := opts.Mode(); got != consts.ProfileModeClone {
		t.Errorf("Mode() = %v, want %v", got, consts.ProfileModeClone)
	}
}

func TestProfileImportOptions_Mode(t *testing.T) {
	opts := &ProfileImportOptions{}
	if got := opts.Mode(); got != consts.ProfileModeImport {
		t.Errorf("Mode() = %v, want %v", got, consts.ProfileModeImport)
	}
}

func TestProfileState_HasMeaningfulChanges(t *testing.T) {
	baseProfile := &domain.ProfileInfo{
		Name: "test-db",
		DBInfo: domain.DBInfo{
			Host:     "10.0.0.5",
			Port:     3306,
			User:     "admin",
			Password: "secret",
		},
		SSHTunnel: domain.SSHTunnelConfig{
			Enabled:  false,
			Host:     "",
			Port:     22,
			User:     "",
			Password: "",
		},
	}

	tests := []struct {
		name     string
		state    *ProfileState
		want     bool
		mutate   func(*domain.ProfileInfo) // Function to modify current profile
		describe string
	}{
		{
			name:  "nil state",
			state: nil,
			want:  false,
		},
		{
			name: "nil original profile",
			state: &ProfileState{
				ProfileInfo:         baseProfile,
				OriginalProfileInfo: nil,
			},
			want: false,
		},
		{
			name: "nil current profile",
			state: &ProfileState{
				ProfileInfo:         nil,
				OriginalProfileInfo: baseProfile,
			},
			want: false,
		},
		{
			name: "no changes",
			state: &ProfileState{
				ProfileInfo:         cloneProfile(baseProfile),
				OriginalProfileInfo: cloneProfile(baseProfile),
			},
			want:     false,
			describe: "Identical profiles should have no changes",
		},
		{
			name: "name changed",
			state: &ProfileState{
				ProfileInfo:         cloneProfile(baseProfile),
				OriginalProfileInfo: cloneProfile(baseProfile),
			},
			mutate: func(p *domain.ProfileInfo) {
				p.Name = "test-db-updated"
			},
			want:     true,
			describe: "Name change is meaningful",
		},
		{
			name: "host changed",
			state: &ProfileState{
				ProfileInfo:         cloneProfile(baseProfile),
				OriginalProfileInfo: cloneProfile(baseProfile),
			},
			mutate: func(p *domain.ProfileInfo) {
				p.DBInfo.Host = "10.0.0.10"
			},
			want:     true,
			describe: "Host change is meaningful",
		},
		{
			name: "port changed",
			state: &ProfileState{
				ProfileInfo:         cloneProfile(baseProfile),
				OriginalProfileInfo: cloneProfile(baseProfile),
			},
			mutate: func(p *domain.ProfileInfo) {
				p.DBInfo.Port = 3307
			},
			want:     true,
			describe: "Port change is meaningful",
		},
		{
			name: "user changed",
			state: &ProfileState{
				ProfileInfo:         cloneProfile(baseProfile),
				OriginalProfileInfo: cloneProfile(baseProfile),
			},
			mutate: func(p *domain.ProfileInfo) {
				p.DBInfo.User = "newuser"
			},
			want:     true,
			describe: "User change is meaningful",
		},
		{
			name: "password changed",
			state: &ProfileState{
				ProfileInfo:         cloneProfile(baseProfile),
				OriginalProfileInfo: cloneProfile(baseProfile),
			},
			mutate: func(p *domain.ProfileInfo) {
				p.DBInfo.Password = "newsecret"
			},
			want:     true,
			describe: "Password change is meaningful",
		},
		{
			name: "ssh tunnel enabled",
			state: &ProfileState{
				ProfileInfo:         cloneProfile(baseProfile),
				OriginalProfileInfo: cloneProfile(baseProfile),
			},
			mutate: func(p *domain.ProfileInfo) {
				p.SSHTunnel.Enabled = true
			},
			want:     true,
			describe: "SSH tunnel enable is meaningful",
		},
		{
			name: "ssh host changed",
			state: &ProfileState{
				ProfileInfo:         cloneProfile(baseProfile),
				OriginalProfileInfo: cloneProfile(baseProfile),
			},
			mutate: func(p *domain.ProfileInfo) {
				p.SSHTunnel.Host = "bastion.example.com"
			},
			want:     true,
			describe: "SSH host change is meaningful",
		},
		{
			name: "metadata path changed (non-meaningful)",
			state: &ProfileState{
				ProfileInfo:         cloneProfile(baseProfile),
				OriginalProfileInfo: cloneProfile(baseProfile),
			},
			mutate: func(p *domain.ProfileInfo) {
				p.Path = "/new/path/profile.cnf.enc"
			},
			want:     false,
			describe: "Path change is NOT meaningful (runtime metadata)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mutate != nil && tt.state != nil && tt.state.ProfileInfo != nil {
				tt.mutate(tt.state.ProfileInfo)
			}

			got := tt.state.HasMeaningfulChanges()
			if got != tt.want {
				t.Errorf("HasMeaningfulChanges() = %v, want %v", got, tt.want)
				if tt.describe != "" {
					t.Logf("Description: %s", tt.describe)
				}
			}
		})
	}
}

func TestProfileState_CreateOptions(t *testing.T) {
	createOpts := &ProfileCreateOptions{Interactive: true}
	state := &ProfileState{Options: createOpts}

	got, ok := state.CreateOptions()
	if !ok {
		t.Error("CreateOptions() returned ok=false, want ok=true")
	}
	if got != createOpts {
		t.Error("CreateOptions() returned different instance")
	}

	// Test with wrong type
	state.Options = &ProfileEditOptions{}
	_, ok = state.CreateOptions()
	if ok {
		t.Error("CreateOptions() returned ok=true for EditOptions, want ok=false")
	}
}

func TestProfileState_EditOptions(t *testing.T) {
	editOpts := &ProfileEditOptions{Interactive: true}
	state := &ProfileState{Options: editOpts}

	got, ok := state.EditOptions()
	if !ok {
		t.Error("EditOptions() returned ok=false, want ok=true")
	}
	if got != editOpts {
		t.Error("EditOptions() returned different instance")
	}
}

func TestProfileState_IsInteractive(t *testing.T) {
	tests := []struct {
		name  string
		state *ProfileState
		want  bool
	}{
		{
			name:  "nil state",
			state: nil,
			want:  false,
		},
		{
			name:  "nil options",
			state: &ProfileState{Options: nil},
			want:  false,
		},
		{
			name:  "interactive create",
			state: &ProfileState{Options: &ProfileCreateOptions{Interactive: true}},
			want:  true,
		},
		{
			name:  "non-interactive create",
			state: &ProfileState{Options: &ProfileCreateOptions{Interactive: false}},
			want:  false,
		},
		{
			name:  "interactive edit",
			state: &ProfileState{Options: &ProfileEditOptions{Interactive: true}},
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.state.IsInteractive(); got != tt.want {
				t.Errorf("IsInteractive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProfileState_Mode(t *testing.T) {
	tests := []struct {
		name  string
		state *ProfileState
		want  string
	}{
		{
			name:  "nil state",
			state: nil,
			want:  "",
		},
		{
			name:  "nil options",
			state: &ProfileState{Options: nil},
			want:  "",
		},
		{
			name:  "create mode",
			state: &ProfileState{Options: &ProfileCreateOptions{}},
			want:  consts.ProfileModeCreate,
		},
		{
			name:  "edit mode",
			state: &ProfileState{Options: &ProfileEditOptions{}},
			want:  consts.ProfileModeEdit,
		},
		{
			name:  "show mode",
			state: &ProfileState{Options: &ProfileShowOptions{}},
			want:  consts.ProfileModeShow,
		},
		{
			name:  "delete mode",
			state: &ProfileState{Options: &ProfileDeleteOptions{}},
			want:  consts.ProfileModeDelete,
		},
		{
			name:  "clone mode",
			state: &ProfileState{Options: &ProfileCloneOptions{}},
			want:  consts.ProfileModeClone,
		},
		{
			name:  "import mode",
			state: &ProfileState{Options: &ProfileImportOptions{}},
			want:  consts.ProfileModeImport,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.state.Mode(); got != tt.want {
				t.Errorf("Mode() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper function to clone profile for testing
func cloneProfile(p *domain.ProfileInfo) *domain.ProfileInfo {
	if p == nil {
		return nil
	}
	clone := *p
	return &clone
}
