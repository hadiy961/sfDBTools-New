package file

import (
	"os"
	"strings"

	"sfdbtools/internal/shared/consts"
)

// ListBackupFilesInDirectory membaca daftar file backup di direktori.
func ListBackupFilesInDirectory(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		// Hanya tampilkan dump backup, bukan file user grants.
		if strings.HasSuffix(name, consts.UsersSQLSuffix) {
			continue
		}
		if strings.Contains(name, consts.ExtSQL) {
			files = append(files, name)
		}
	}

	return files, nil
}
