package types

import (
	"sfdbtools/internal/domain"
	"sfdbtools/internal/shared/consts"
)

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

func (o *ProfileCreateOptions) Mode() string { return consts.ProfileModeCreate }

func (o *ProfileCreateOptions) IsInteractive() bool { return o != nil && o.Interactive }

// ProfileEditOptions - Flags for profile edit command
type ProfileEditOptions struct {
	ProfileInfo         domain.ProfileInfo
	Interactive         bool
	NewName             string
	NewProfileKey       string
	NewProfileKeySource string
}

func (o *ProfileEditOptions) Mode() string { return consts.ProfileModeEdit }

func (o *ProfileEditOptions) IsInteractive() bool { return o != nil && o.Interactive }

// ProfileShowOptions - Flags for profile show and validate commands
type ProfileShowOptions struct {
	domain.ProfileInfo
	RevealPassword bool
	Interactive    bool
}

func (o *ProfileShowOptions) Mode() string { return consts.ProfileModeShow }

func (o *ProfileShowOptions) IsInteractive() bool { return o != nil && o.Interactive }

// ProfileDeleteOptions - Flags for profile delete command
type ProfileDeleteOptions struct {
	ProfileInfo domain.ProfileInfo
	Profiles    []string // List of profiles to delete
	Force       bool
	Interactive bool
}

func (o *ProfileDeleteOptions) Mode() string { return consts.ProfileModeDelete }

func (o *ProfileDeleteOptions) IsInteractive() bool { return o != nil && o.Interactive }

// ProfileCloneOptions - Options for cloning an existing profile
type ProfileCloneOptions struct {
	SourceProfile string              // Source profile path/name to clone from
	TargetName    string              // Target profile name
	TargetHost    string              // Override host (optional)
	TargetPort    int                 // Override port (optional)
	ProfileKey    string              // Encryption key for source profile
	NewProfileKey string              // Encryption key for target profile (optional, defaults to same as source)
	OutputDir     string              // Output directory for cloned profile
	Interactive   bool                // Interactive mode with pre-fill wizard
	SourceInfo    *domain.ProfileInfo // Loaded source profile info (filled during execution)
}

func (o *ProfileCloneOptions) Mode() string { return consts.ProfileModeClone }

func (o *ProfileCloneOptions) IsInteractive() bool { return o != nil && o.Interactive }

// ProfileImportOptions - Options for importing profiles (bulk) dari XLSX atau Google Spreadsheet.
// Catatan: 1 row = 1 profile = 1 encryption key.
type ProfileImportOptions struct {
	Input           string // Path file XLSX lokal (mutually exclusive dengan GSheetURL)
	Sheet           string // Nama sheet di XLSX (opsional; default sheet pertama)
	GSheetURL       string // Google Spreadsheet URL (edit/share)
	GID             int    // Sheet/tab gid untuk export CSV (Google)
	OnConflict      string // fail|skip|overwrite|rename
	SkipConfirm     bool   // Jika true: tidak ada prompt (wajib untuk automation)
	SkipInvalidRows bool   // Jika true: row invalid di-skip (default false)
	ContinueOnError bool   // Jika true: error saat conn-test/save tidak menghentikan proses
	SkipConnTest    bool   // Jika true: skip connection test (default false)

	// ConnTestDone ditetapkan saat runtime setelah tahap conn-test selesai,
	// agar SaveProfile tidak melakukan conn-test dua kali.
	ConnTestDone bool

	Interactive bool
}

func (o *ProfileImportOptions) Mode() string { return consts.ProfileModeImport }

func (o *ProfileImportOptions) IsInteractive() bool { return o != nil && o.Interactive }

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

func (s *ProfileState) CloneOptions() (*ProfileCloneOptions, bool) {
	o, ok := s.Options.(*ProfileCloneOptions)
	return o, ok
}

func (s *ProfileState) ImportOptions() (*ProfileImportOptions, bool) {
	o, ok := s.Options.(*ProfileImportOptions)
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
