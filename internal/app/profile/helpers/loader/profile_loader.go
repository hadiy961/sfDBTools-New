package loader

import (
	"fmt"

	"sfdbtools/internal/app/profile/helpers/parser"
	"sfdbtools/internal/app/profile/helpers/paths"
	"sfdbtools/internal/app/profile/helpers/selection"
	"sfdbtools/internal/app/profile/merger"
	"sfdbtools/internal/domain"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/envx"
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
			profilePath = envx.GetEnvOrDefault(opts.EnvProfilePath, "")
		}

		if profilePath == "" {
			if opts.AllowInteractive {
				prompt := opts.InteractivePrompt
				if prompt == "" {
					prompt = "Pilih file konfigurasi database:"
				}
				info, err := selection.SelectExistingDBConfig(opts.ConfigDir, prompt)
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
		profileKey = envx.GetEnvOrDefault(opts.EnvProfileKey, "")
	}

	var (
		absPath string
		name    string
		err     error
	)
	if opts.ConfigDir != "" {
		absPath, name, err = paths.ResolveConfigPathInDir(opts.ConfigDir, profilePath)
	} else {
		absPath, name, err = paths.ResolveConfigPath(profilePath)
	}
	if err != nil {
		return nil, fmt.Errorf("gagal memproses path konfigurasi: %w", err)
	}

	profile, err := parser.LoadAndParseProfile(absPath, profileKey)
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

// SelectExistingDBConfigWithSnapshot memilih profile secara interaktif lalu mengembalikan:
// - info (hasil load+parse profile)
// - originalName (nama profile yang dipilih)
// - snapshot (salinan info untuk baseline/original display)
func SelectExistingDBConfigWithSnapshot(configDir string, promptText string) (info *domain.ProfileInfo, originalName string, snapshot *domain.ProfileInfo, err error) {
	loaded, err := ResolveAndLoadProfile(ProfileLoadOptions{
		ConfigDir:         configDir,
		ProfilePath:       "",
		AllowInteractive:  true,
		InteractivePrompt: promptText,
		RequireProfile:    true,
	})
	if err != nil {
		return nil, "", nil, err
	}
	if loaded == nil {
		return nil, "", nil, fmt.Errorf("%sprofile tidak tersedia (hasil load nil)", consts.ProfileMsgNonInteractivePrefix)
	}

	originalName = loaded.Name
	snapshot = merger.CloneAsOriginalProfileInfo(loaded)
	return loaded, originalName, snapshot, nil
}
