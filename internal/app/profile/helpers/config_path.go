package helpers

import (
	"sfdbtools/internal/app/profile/helpers/paths"
)

func ResolveConfigPathInDir(configDir string, spec string) (string, string, error) {
	return paths.ResolveConfigPathInDir(configDir, spec)
}

// ResolveConfigPath mengubah name/path menjadi absolute path di config dir dan nama normalized tanpa suffix.
func ResolveConfigPath(spec string) (string, string, error) {
	return paths.ResolveConfigPath(spec)
}
