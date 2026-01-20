// File : internal/app/backup/scheduler/validate.go
// Deskripsi : Validation utilities untuk scheduler job configuration
// Author : Hadiyatna Muflihun
// Tanggal : 2026-01-20
// Last Modified : 2026-01-20
package scheduler

import (
	appconfig "sfdbtools/internal/services/config"
)

// ValidateJobName adalah alias untuk appconfig.ValidateJobName
// Wrapper ini untuk backwards compatibility dan convenience
func ValidateJobName(jobName string) error {
	return appconfig.ValidateJobName(jobName)
}
