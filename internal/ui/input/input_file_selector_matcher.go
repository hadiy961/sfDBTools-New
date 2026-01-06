package input

import "strings"

func matchesExtension(fileName string, extensions []string) bool {
	lowerName := strings.ToLower(fileName)
	for _, ext := range extensions {
		if strings.HasSuffix(lowerName, ext) {
			return true
		}
	}
	return false
}
