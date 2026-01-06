package profile

import (
	"fmt"

	"sfdbtools/internal/domain"
	"sfdbtools/pkg/helper/env"
	"sfdbtools/pkg/helper/profileutil"
)

// ProfileLoadOptions berisi opsi untuk loading profile.
type ProfileLoadOptions struct {
	ConfigDir         string
	ProfilePath       string
	ProfileKey        string
	EnvProfilePath    string
	EnvProfileKey     string
	RequireProfile    bool
	ProfilePurpose    string
	AllowInteractive  bool
	InteractivePrompt string
}

// ResolveAndLoadProfile me-resolve dan load profile dengan fallback ke environment variables.
func ResolveAndLoadProfile(opts ProfileLoadOptions) (*domain.ProfileInfo, error) {
	profilePath := opts.ProfilePath
	profileKey := opts.ProfileKey

	if profilePath == "" {
		if opts.EnvProfilePath != "" {
			profilePath = env.GetEnvOrDefault(opts.EnvProfilePath, "")
		}

		if profilePath == "" {
			if opts.AllowInteractive {
				prompt := opts.InteractivePrompt
				if prompt == "" {
					prompt = "Pilih file konfigurasi database:"
				}
				info, err := SelectExistingDBConfig(opts.ConfigDir, prompt)
				if err != nil {
					return nil, fmt.Errorf("gagal memilih konfigurasi database: %w", err)
				}
				return &info, nil
			}

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

	if profileKey == "" && opts.EnvProfileKey != "" {
		profileKey = env.GetEnvOrDefault(opts.EnvProfileKey, "")
	}

	absPath, name, err := profileutil.ResolveConfigPath(profilePath)
	if err != nil {
		return nil, fmt.Errorf("gagal memproses path konfigurasi: %w", err)
	}

	profile, err := LoadAndParseProfile(absPath, profileKey)
	if err != nil {
		return nil, fmt.Errorf("gagal load profile: %w", err)
	}

	profile.Path = absPath
	profile.Name = name

	return profile, nil
}

// LoadSourceProfile loads source profile untuk backup/dbscan operations dengan interactive mode.
func LoadSourceProfile(configDir, profilePath, profileKey string, allowInteractive bool) (*domain.ProfileInfo, error) {
	return ResolveAndLoadProfile(ProfileLoadOptions{
		ConfigDir:         configDir,
		ProfilePath:       profilePath,
		ProfileKey:        profileKey,
		EnvProfilePath:    "",
		EnvProfileKey:     "",
		RequireProfile:    true,
		ProfilePurpose:    "source",
		AllowInteractive:  allowInteractive,
		InteractivePrompt: "Pilih file konfigurasi database sumber:",
	})
}
