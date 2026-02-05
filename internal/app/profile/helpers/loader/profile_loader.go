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

// ProfileLoadOptions berisi opsi untuk loading profile dengan berbagai fallback mechanisms.
//
// Fields:
//   - ConfigDir: Directory tempat profile disimpan (opsional, default current dir)
//   - ProfilePath: Path ke profile file (absolute/relative/name saja)
//   - ProfileKey: Encryption key untuk decrypt profile
//   - EnvProfilePath: Nama env var untuk fallback profile path
//   - EnvProfileKey: Nama env var untuk fallback encryption key
//   - RequireProfile: Jika true, error jika profile tidak ditemukan
//   - ProfilePurpose: Purpose string untuk error message (e.g., "source", "target")
//   - AllowInteractive: Allow interactive selection jika path kosong
//   - InteractivePrompt: Custom prompt text untuk interactive selection
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

// ResolveAndLoadProfile me-resolve path dan load profile dengan fallback ke environment variables.
//
// Resolution chain untuk profile path:
//  1. Gunakan opts.ProfilePath jika tersedia
//  2. Fallback ke environment variable (opts.EnvProfilePath)
//  3. Interactive selection (jika opts.AllowInteractive=true)
//  4. Error jika RequireProfile=true
//
// Resolution chain untuk encryption key:
//  1. Gunakan opts.ProfileKey jika tersedia
//  2. Fallback ke environment variable (opts.EnvProfileKey)
//  3. Interactive prompt (ditangani di parser layer)
//
// Returns:
//   - *domain.ProfileInfo: Loaded profile dengan Path dan Name sudah di-set
//   - error: Error jika profile tidak ditemukan atau gagal parse/decrypt
//
// Example:
//
//	profile, err := loader.ResolveAndLoadProfile(loader.ProfileLoadOptions{
//		ConfigDir:        "/etc/sfdbtools/profiles",
//		ProfilePath:      "prod-db",
//		ProfileKey:       "secret-key",
//		EnvProfilePath:   "SFDB_PROFILE",
//		EnvProfileKey:    "SFDB_PROFILE_KEY",
//		RequireProfile:   true,
//		AllowInteractive: true,
//	})
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
//
// Fungsi ini adalah convenience wrapper untuk ResolveAndLoadProfile dengan:
//   - RequireProfile=true (wajib ada profile)
//   - ProfilePurpose="source" (untuk error message)
//   - Interactive prompt yang sesuai context
//
// Parameters:
//   - configDir: Directory tempat profile disimpan
//   - profilePath: Path/name profile (bisa kosong jika allowInteractive=true)
//   - profileKey: Encryption key (bisa kosong, akan di-prompt)
//   - allowInteractive: Allow interactive selection jika profilePath kosong
//
// Returns:
//   - *domain.ProfileInfo: Loaded source profile
//   - error: Error jika profile tidak ditemukan atau gagal load
//
// Use case: db-backup, db-scan, db-restore dengan source profile.
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

// SelectExistingDBConfigWithSnapshot memilih profile secara interaktif dan membuat snapshot untuk baseline comparison.
//
// Fungsi ini digunakan untuk operasi edit/delete yang membutuhkan:
//  1. Profile selection via interactive picker
//  2. Original snapshot untuk detecting changes
//  3. Original name untuk rename detection
//
// Parameters:
//   - configDir: Directory tempat profile disimpan
//   - promptText: Custom prompt text untuk selection UI
//
// Returns:
//   - info: Profile yang loaded (working copy untuk edit)
//   - originalName: Nama profile original (untuk detect rename)
//   - snapshot: Deep copy dari profile (baseline untuk diff)
//   - err: Error jika selection cancelled atau load gagal
//
// Example:
//
//	profile, origName, snapshot, err := loader.SelectExistingDBConfigWithSnapshot(
//		configDir,
//		"Pilih profile yang ingin diedit:",
//	)
//	// profile dapat di-edit, snapshot tetap immutable untuk comparison
//	if hasChanges(profile, snapshot) {
//		// ... save changes
//	}
//
// Use case: profile edit, profile delete dengan confirmation.
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
