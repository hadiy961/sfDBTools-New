package helper

import (
	"sfDBTools/pkg/compress"
	"sfDBTools/pkg/consts"
	"strings"
)

func IsEncryptedFile(path string) bool {
	return strings.HasSuffix(strings.ToLower(path), consts.ExtEnc)
}

// ValidBackupFileExtensionsForSelection returns the list of file extensions used by interactive file pickers.
// It includes plaintext sql, encrypted sql, compressed sql, and compressed+encrypted sql.
func ValidBackupFileExtensionsForSelection() []string {
	valid := []string{consts.ExtSQL, consts.ExtSQL + consts.ExtEnc, consts.ExtEnc}
	for _, cext := range compress.SupportedCompressionExtensions() {
		valid = append(valid, consts.ExtSQL+cext)
		valid = append(valid, consts.ExtSQL+cext+consts.ExtEnc)
	}
	return valid
}
