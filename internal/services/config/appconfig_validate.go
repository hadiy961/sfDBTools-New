// File : internal/services/config/appconfig_validate.go
// Deskripsi : Validation functions untuk Config
// Author : Hadiyatna Muflihun
// Tanggal : 2026-01-20
// Last Modified : 2026-01-20
package appconfig

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	// JobNameMaxLength adalah panjang maksimal untuk job name (prevent path too long error)
	JobNameMaxLength = 64
)

var (
	// jobNamePattern adalah whitelist untuk karakter yang diperbolehkan di job name
	// Hanya alphanumeric, dash, dan underscore untuk mencegah path traversal dan injection
	jobNamePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

	// windowsReservedNames adalah nama-nama yang reserved di Windows filesystem
	// Kita validate ini untuk cross-platform compatibility
	windowsReservedNames = []string{
		"CON", "PRN", "AUX", "NUL",
		"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
		"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9",
	}
)

// ValidateJobName melakukan validasi comprehensive terhadap job name untuk mencegah:
// - Path traversal (../, ..\, dll)
// - Log injection (newline characters)
// - Filesystem injection (special characters)
// - Path too long error (length limit)
// - Windows reserved names
func ValidateJobName(jobName string) error {
	// 1. Empty check
	if len(jobName) == 0 {
		return fmt.Errorf("job name cannot be empty")
	}

	// 2. Length check (prevent path too long error)
	if len(jobName) > JobNameMaxLength {
		return fmt.Errorf("job name too long (max %d chars): '%s'", JobNameMaxLength, jobName)
	}

	// 3. Character whitelist (alphanumeric + dash/underscore only)
	// Ini mencegah:
	// - Path traversal: '../', '..\' tidak bisa dibuat
	// - Log injection: newline '\n' tidak diperbolehkan
	// - Filesystem injection: special chars seperti ';', '|', '&' ditolak
	if !jobNamePattern.MatchString(jobName) {
		return fmt.Errorf("invalid job name '%s': only alphanumeric, dash, and underscore allowed", jobName)
	}

	// 4. Reserved names check (cross-platform compatibility)
	// Windows reserved names tidak boleh digunakan, bahkan di Linux
	for _, reserved := range windowsReservedNames {
		if strings.EqualFold(jobName, reserved) {
			return fmt.Errorf("job name '%s' is reserved (Windows compatibility)", jobName)
		}
	}

	// 5. Dot-only names check (prevent '.' dan '..')
	if jobName == "." || jobName == ".." {
		return fmt.Errorf("job name '%s' is reserved", jobName)
	}

	return nil
}

// validateSchedulerJobs melakukan validasi terhadap semua scheduler jobs di config
// Dipanggil saat config load untuk fail-fast sebelum scheduler berjalan
func validateSchedulerJobs(jobs []BackupSchedulerJob) error {
	if len(jobs) == 0 {
		return nil
	}

	seen := make(map[string]bool)
	for i, job := range jobs {
		// Validate job name
		if err := ValidateJobName(job.Name); err != nil {
			return fmt.Errorf("backup.scheduler.jobs[%d]: %w", i, err)
		}

		// Check duplicate names
		if seen[job.Name] {
			return fmt.Errorf("backup.scheduler.jobs[%d]: duplicate job name '%s'", i, job.Name)
		}
		seen[job.Name] = true
	}

	return nil
}
