// File : internal/app/backup/helpers/lockpath/generator.go
// Deskripsi : Lock path generator untuk background backup mode
// Author : Hadiyatna Muflihun
// Tanggal : 20 Januari 2026
// Last Modified : 20 Januari 2026

package path

import (
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
)

const (
	// Lock directory preferences (in order of priority)
	varLockDir = "/var/lock"
	tmpDir     = "/tmp"

	// Lock file prefix
	lockPrefix = "sfdbtools-backup"
)

// GenerateForProfile generates unique lock file path for a given profile.
// Lock path is based on profile file path hash untuk prevent conflicts antara
// concurrent backup jobs dari profile berbeda.
//
// Example:
//   - /etc/sfDBTools/config/db_profile/prod.cnf.enc -> /var/lock/sfdbtools-backup-a1b2c3d4.lock
//   - /etc/sfDBTools/config/db_profile/staging.cnf.enc -> /var/lock/sfdbtools-backup-e5f6g7h8.lock
//
// Fallback to /tmp/ jika /var/lock/ tidak writable (non-root users).
func GenerateForProfile(profilePath string) string {
	// Generate hash dari profile path (8 chars hex)
	hash := md5.Sum([]byte(profilePath))
	lockName := fmt.Sprintf("%s-%x.lock", lockPrefix, hash[:4])

	// Try /var/lock first (standard Linux lock directory)
	if isWritable(varLockDir) {
		return filepath.Join(varLockDir, lockName)
	}

	// Fallback to /tmp for non-root users
	return filepath.Join(tmpDir, lockName)
}

// isWritable checks if directory is writable
func isWritable(dir string) bool {
	info, err := os.Stat(dir)
	if err != nil {
		return false
	}

	if !info.IsDir() {
		return false
	}

	// Test write permission dengan create temp file
	testFile := filepath.Join(dir, ".sfdbtools-write-test")
	file, err := os.Create(testFile)
	if err != nil {
		return false
	}
	file.Close()
	os.Remove(testFile)

	return true
}
