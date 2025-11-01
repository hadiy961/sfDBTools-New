package helper

import "strings"

var (
	// backupExtensions mendefinisikan ekstensi file yang dianggap sebagai file backup.
	backupExtensions = []string{".sql", ".gz", ".zst", ".lz4", ".enc"}
)

// helper.TrimProfileSuffix menghapus suffix .cnf.enc dari nama jika ada.
func TrimProfileSuffix(name string) string {
	return strings.TrimSuffix(strings.TrimSuffix(name, ".enc"), ".cnf")
}

// IsBackupFile memeriksa apakah sebuah file dianggap sebagai file backup berdasarkan ekstensinya.
func IsBackupFile(filename string) bool {
	lowerFilename := strings.ToLower(filename)
	for _, ext := range backupExtensions {
		if strings.HasSuffix(lowerFilename, ext) {
			return true
		}
	}
	return false
}
