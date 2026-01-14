package helpers

import (
	"sfdbtools/internal/app/profile/helpers/loader"
	"sfdbtools/internal/domain"
)

// ProfileLoadOptions berisi opsi untuk loading profile.
// Type alias agar tetap kompatibel dengan import path lama.
type ProfileLoadOptions = loader.ProfileLoadOptions

// ResolveAndLoadProfile me-resolve dan load profile dengan fallback ke environment variables.
func ResolveAndLoadProfile(opts ProfileLoadOptions) (*domain.ProfileInfo, error) {
	return loader.ResolveAndLoadProfile(opts)
}

// LoadSourceProfile loads source profile untuk backup/dbscan operations dengan interactive mode.
func LoadSourceProfile(configDir, profilePath, profileKey string, allowInteractive bool) (*domain.ProfileInfo, error) {
	return loader.LoadSourceProfile(configDir, profilePath, profileKey, allowInteractive)
}
