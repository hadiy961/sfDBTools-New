package input

import "strings"

// SanitizeFileName membersihkan nama file dari karakter ilegal.
func SanitizeFileName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "database"
	}

	illegal := " /\\:*?\"<>|"
	name = strings.Map(func(r rune) rune {
		if strings.ContainsRune(illegal, r) {
			return '_'
		}
		return r
	}, name)

	return name
}
