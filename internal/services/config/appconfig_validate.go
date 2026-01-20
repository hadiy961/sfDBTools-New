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
	"time"
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

		// Validate timeout format (Issue #55)
		if err := ValidateJobTimeout(job.Timeout, i, job.Name); err != nil {
			return err
		}
	}

	return nil
}

// ValidateJobTimeout melakukan validasi terhadap timeout format dan value
// Issue #55: Scheduler jobs harus memiliki reasonable timeout untuk mencegah hang indefinitely
func ValidateJobTimeout(timeoutStr string, jobIndex int, jobName string) error {
	// Empty timeout = will use default, OK
	if timeoutStr == "" {
		return nil
	}

	timeout, err := ParseTimeout(timeoutStr)
	if err != nil {
		return fmt.Errorf("backup.scheduler.jobs[%d] (%s): invalid timeout format '%s': %w", jobIndex, jobName, timeoutStr, err)
	}

	// Minimum timeout: 1 minute (prevent misconfiguration)
	if timeout < 1*time.Minute {
		return fmt.Errorf("backup.scheduler.jobs[%d] (%s): timeout too short (min: 1m): %v", jobIndex, jobName, timeout)
	}

	// Maximum timeout: 48 hours (warning only, tidak error)
	// Beberapa backup very large databases bisa butuh waktu lama
	if timeout > 48*time.Hour {
		// Log warning via fmt.Errorf tapi tidak return error (warning only)
		// Logger belum tersedia di validation phase, jadi kita skip warning di sini
		// Warning akan di-log saat RunJob() execution
	}

	return nil
}

// ParseTimeout mengkonversi timeout string ke time.Duration
// Format yang didukung:
//   - Duration string: "30m", "2h", "6h30m"
//   - Integer (backward compat): "6" = 6 hours
//   - Empty string: return 0 (caller akan use default)
func ParseTimeout(timeoutStr string) (time.Duration, error) {
	if timeoutStr == "" {
		return 0, nil
	}

	// Try parse as duration first (preferred format)
	timeout, err := time.ParseDuration(timeoutStr)
	if err == nil {
		return timeout, nil
	}

	// Backward compatibility: parse as integer hours
	// "6" → 6 hours
	var hours int
	if _, parseErr := fmt.Sscanf(timeoutStr, "%d", &hours); parseErr == nil {
		if hours < 0 {
			return 0, fmt.Errorf("negative timeout not allowed")
		}
		return time.Duration(hours) * time.Hour, nil
	}

	// Neither duration nor integer, return original error
	return 0, fmt.Errorf("invalid duration format (use '30m', '2h', etc): %w", err)
}
