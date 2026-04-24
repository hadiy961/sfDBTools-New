package fsops

import (
	"os"
	"strconv"

	applog "sfdbtools/internal/services/log"
)

// ParseFilePermissions mengkonversi string permissions (e.g., "0600") ke os.FileMode.
// Jika parsing gagal atau permissions kosong, return defaultPerm.
func ParseFilePermissions(permStr string, defaultPerm os.FileMode, logger applog.Logger) os.FileMode {
	if permStr == "" {
		return defaultPerm
	}

	perm, err := strconv.ParseUint(permStr, 8, 32)
	if err != nil {
		if logger != nil {
			logger.Warnf("Invalid file_permissions '%s', using default: %v", permStr, err)
		}
		return defaultPerm
	}

	return os.FileMode(perm)
}
