package types

import "sfdbtools/internal/domain"

// ProfileOptions adalah union-style interface untuk menyimpan opsi aktif
// (create/edit/show/delete) dalam ProfileState.
// Tujuan: menghindari banyak pointer options yang hanya dipakai salah satu per eksekusi.
type ProfileOptions interface {
	Mode() string
	IsInteractive() bool
}

// ProfileCreateOptions - Options for creating a new profile
type ProfileCreateOptions struct {
	ProfileInfo domain.ProfileInfo
	OutputDir   string
	Interactive bool
}

func (o *ProfileCreateOptions) Mode() string { return "create" }

func (o *ProfileCreateOptions) IsInteractive() bool { return o != nil && o.Interactive }

// ProfileEditOptions - Flags for profile edit command
type ProfileEditOptions struct {
	ProfileInfo         domain.ProfileInfo
	Interactive         bool
	NewName             string
	NewProfileKey       string
	NewProfileKeySource string
}

func (o *ProfileEditOptions) Mode() string { return "edit" }

func (o *ProfileEditOptions) IsInteractive() bool { return o != nil && o.Interactive }

// ProfileShowOptions - Flags for profile show and validate commands
type ProfileShowOptions struct {
	domain.ProfileInfo
	RevealPassword bool
	Interactive    bool
}

func (o *ProfileShowOptions) Mode() string { return "show" }

func (o *ProfileShowOptions) IsInteractive() bool { return o != nil && o.Interactive }

// ProfileDeleteOptions - Flags for profile delete command
type ProfileDeleteOptions struct {
	ProfileInfo domain.ProfileInfo
	Profiles    []string // List of profiles to delete
	Force       bool
	Interactive bool
}

func (o *ProfileDeleteOptions) Mode() string { return "delete" }

func (o *ProfileDeleteOptions) IsInteractive() bool { return o != nil && o.Interactive }

// ProfileEntryConfig menyimpan konfigurasi untuk entry point profile operations
type ProfileEntryConfig struct {
	HeaderTitle string // UI header title
	Mode        string // "create", "show", "edit", "delete"
	SuccessMsg  string // Success message
	LogPrefix   string // Log prefix for tracking
}

// ProfileState adalah single source of truth untuk semua state yang perlu di-share
// antara Service, Wizard, Executor, dan Display components.
// Menggunakan pointer shared ini mengeliminasi kebutuhan sync functions yang repetitif.
type ProfileState struct {
	ProfileInfo         *domain.ProfileInfo
	Options             ProfileOptions
	OriginalProfileName string
	OriginalProfileInfo *domain.ProfileInfo
}

func (s *ProfileState) CreateOptions() (*ProfileCreateOptions, bool) {
	o, ok := s.Options.(*ProfileCreateOptions)
	return o, ok
}

func (s *ProfileState) EditOptions() (*ProfileEditOptions, bool) {
	o, ok := s.Options.(*ProfileEditOptions)
	return o, ok
}

func (s *ProfileState) ShowOptions() (*ProfileShowOptions, bool) {
	o, ok := s.Options.(*ProfileShowOptions)
	return o, ok
}

func (s *ProfileState) DeleteOptions() (*ProfileDeleteOptions, bool) {
	o, ok := s.Options.(*ProfileDeleteOptions)
	return o, ok
}

func (s *ProfileState) IsInteractive() bool {
	if s == nil || s.Options == nil {
		return false
	}
	return s.Options.IsInteractive()
}

func (s *ProfileState) Mode() string {
	if s == nil || s.Options == nil {
		return ""
	}
	return s.Options.Mode()
}
