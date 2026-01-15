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

// HasMeaningfulChanges mengembalikan true jika ada perubahan yang akan berdampak
// pada file profil yang disimpan.
//
// Catatan: field metadata runtime seperti Path/Size/LastModified dan
// ResolvedLocalPort tidak dianggap perubahan.
func (s *ProfileState) HasMeaningfulChanges() bool {
	if s == nil || s.OriginalProfileInfo == nil || s.ProfileInfo == nil {
		return false
	}
	orig := s.OriginalProfileInfo
	cur := s.ProfileInfo

	if orig.Name != cur.Name {
		return true
	}
	if orig.DBInfo.Host != cur.DBInfo.Host {
		return true
	}
	if orig.DBInfo.Port != cur.DBInfo.Port {
		return true
	}
	if orig.DBInfo.User != cur.DBInfo.User {
		return true
	}
	if orig.DBInfo.Password != cur.DBInfo.Password {
		return true
	}

	if orig.SSHTunnel.Enabled != cur.SSHTunnel.Enabled {
		return true
	}
	if orig.SSHTunnel.Host != cur.SSHTunnel.Host {
		return true
	}
	if orig.SSHTunnel.Port != cur.SSHTunnel.Port {
		return true
	}
	if orig.SSHTunnel.User != cur.SSHTunnel.User {
		return true
	}
	if orig.SSHTunnel.Password != cur.SSHTunnel.Password {
		return true
	}
	if orig.SSHTunnel.IdentityFile != cur.SSHTunnel.IdentityFile {
		return true
	}
	if orig.SSHTunnel.LocalPort != cur.SSHTunnel.LocalPort {
		return true
	}

	return false
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
