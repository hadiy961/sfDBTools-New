package profile

import (
	"sfdbtools/pkg/helper/profileutil"
)

// TrimProfileSuffix menghapus suffix .cnf.enc dari nama jika ada.
func TrimProfileSuffix(name string) string {
	return profileutil.TrimProfileSuffix(name)
}

// ResolveConfigPath mengubah name/path menjadi absolute path di config dir dan nama normalized tanpa suffix.
func ResolveConfigPath(spec string) (string, string, error) {
	return profileutil.ResolveConfigPath(spec)
}
