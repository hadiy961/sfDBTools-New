package types

import "time"

// SSHTunnelConfig menyimpan konfigurasi SSH tunnel untuk akses database melalui bastion.
// Catatan: ResolvedLocalPort hanya runtime (tidak perlu disimpan ke file).
type SSHTunnelConfig struct {
	Enabled           bool
	Host              string
	Port              int
	User              string
	Password          string
	IdentityFile      string
	LocalPort         int
	ResolvedLocalPort int
}

// ProfileCreateOptions - Options for creating a new profile
type ProfileCreateOptions struct {
	ProfileInfo ProfileInfo
	OutputDir   string
	Interactive bool
}

// ProfileInfo - Struct to hold profile information
type ProfileInfo struct {
	Name             string
	DBInfo           DBInfo
	SSHTunnel        SSHTunnelConfig
	EncryptionKey    string
	EncryptionSource string
	Size             string
	LastModified     time.Time
	Path             string
}

// ProfileEditOptions - Flags for profile edit command
type ProfileEditOptions struct {
	ProfileInfo ProfileInfo
	Interactive bool
	NewName     string
}

// ProfileShowOptions - Flags for profile show and validate commands
type ProfileShowOptions struct {
	ProfileInfo
	RevealPassword bool
	Interactive    bool
}

// ProfileDeleteOptions - Flags for profile delete command
type ProfileDeleteOptions struct {
	ProfileInfo ProfileInfo
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
