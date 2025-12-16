package profilehelper

// File : pkg/profilehelper/profilehelper_loader.go
// Deskripsi : Helper functions untuk loading dan resolving database profiles
// Author : Hadiyatna Muflihun
// Tanggal : 2025-11-11
// Last Modified : 2025-11-11

import (
	"fmt"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/helper"
	"sfDBTools/pkg/profileselect"
)

// ProfileLoadOptions berisi opsi untuk loading profile
type ProfileLoadOptions struct {
	ConfigDir         string // Config directory untuk profile files
	ProfilePath       string // Path ke profile file (bisa dari flag)
	ProfileKey        string // Encryption key (bisa dari flag)
	EnvProfilePath    string // Environment variable name untuk profile path (e.g., ENV_TARGET_PROFILE)
	EnvProfileKey     string // Environment variable name untuk profile key (e.g., ENV_TARGET_PROFILE_KEY)
	RequireProfile    bool   // Jika true, error jika profile tidak ditemukan
	ProfilePurpose    string // Deskripsi purpose untuk error message (e.g., "target", "source")
	AllowInteractive  bool   // Jika true, allow interactive profile selection
	InteractivePrompt string // Prompt message untuk interactive selection
}

// ResolveAndLoadProfile me-resolve dan load profile dengan fallback ke environment variables
// Function ini menggabungkan pattern yang sama dari backup dan dbscan packages
func ResolveAndLoadProfile(opts ProfileLoadOptions) (*types.ProfileInfo, error) {
	profilePath := opts.ProfilePath
	profileKey := opts.ProfileKey

	// Step 1: Resolve profile path dari environment variable jika tidak ada
	if profilePath == "" {
		if opts.EnvProfilePath != "" {
			profilePath = helper.GetEnvOrDefault(opts.EnvProfilePath, "")
		}

		// Jika masih kosong dan interactive mode enabled, gunakan selector
		if profilePath == "" {
			if opts.AllowInteractive {
				prompt := opts.InteractivePrompt
				if prompt == "" {
					prompt = "Pilih file konfigurasi database:"
				}
				info, err := profileselect.SelectExistingDBConfig(opts.ConfigDir, prompt)
				if err != nil {
					return nil, fmt.Errorf("gagal memilih konfigurasi database: %w", err)
				}
				return &info, nil
			}

			// Tidak ada profile path dan tidak bisa interactive
			if opts.RequireProfile {
				purpose := opts.ProfilePurpose
				if purpose == "" {
					purpose = "database"
				}
				envVar := opts.EnvProfilePath
				if envVar == "" {
					envVar = "(environment variable not specified)"
				}
				return nil, fmt.Errorf("%s profile tidak tersedia, gunakan flag --profile atau env %s", purpose, envVar)
			}
		}
	}

	// Step 2: Resolve profile key dari environment variable jika tidak ada
	if profileKey == "" && opts.EnvProfileKey != "" {
		profileKey = helper.GetEnvOrDefault(opts.EnvProfileKey, "")
	}

	// Step 3: Resolve dan normalize profile path
	absPath, name, err := helper.ResolveConfigPath(profilePath)
	if err != nil {
		return nil, fmt.Errorf("gagal memproses path konfigurasi: %w", err)
	}

	// Step 4: Load and parse profile
	profile, err := profileselect.LoadAndParseProfile(absPath, profileKey)
	if err != nil {
		return nil, fmt.Errorf("gagal load profile: %w", err)
	}

	// Update path dan name dengan yang sudah di-resolve
	profile.Path = absPath
	profile.Name = name

	return profile, nil
}

// LoadSourceProfile loads source profile untuk backup/dbscan operations dengan interactive mode
// Wrapper function dengan default values untuk backup/dbscan
func LoadSourceProfile(configDir, profilePath, profileKey string, allowInteractive bool) (*types.ProfileInfo, error) {
	return ResolveAndLoadProfile(ProfileLoadOptions{
		ConfigDir:         configDir,
		ProfilePath:       profilePath,
		ProfileKey:        profileKey,
		EnvProfilePath:    "", // Backup/dbscan tidak menggunakan env var untuk path
		EnvProfileKey:     "", // Key di-resolve dalam LoadAndParseProfile
		RequireProfile:    true,
		ProfilePurpose:    "source",
		AllowInteractive:  allowInteractive,
		InteractivePrompt: "Pilih file konfigurasi database sumber:",
	})
}
