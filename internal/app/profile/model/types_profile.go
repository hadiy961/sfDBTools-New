package types

import "sfdbtools/internal/domain"

// ProfileCreateOptions - Options for creating a new profile
type ProfileCreateOptions struct {
	ProfileInfo domain.ProfileInfo
	OutputDir   string
	Interactive bool
}

// ProfileEditOptions - Flags for profile edit command
type ProfileEditOptions struct {
	ProfileInfo         domain.ProfileInfo
	Interactive         bool
	NewName             string
	NewProfileKey       string
	NewProfileKeySource string
}

// ProfileShowOptions - Flags for profile show and validate commands
type ProfileShowOptions struct {
	domain.ProfileInfo
	RevealPassword bool
	Interactive    bool
}

// ProfileDeleteOptions - Flags for profile delete command
type ProfileDeleteOptions struct {
	ProfileInfo domain.ProfileInfo
	Profiles    []string // List of profiles to delete
	Force       bool
	Interactive bool
}

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
	ProfileCreate       *ProfileCreateOptions
	ProfileEdit         *ProfileEditOptions
	ProfileShow         *ProfileShowOptions
	ProfileDelete       *ProfileDeleteOptions
	OriginalProfileName string
	OriginalProfileInfo *domain.ProfileInfo
}
