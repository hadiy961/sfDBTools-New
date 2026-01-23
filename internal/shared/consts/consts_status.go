package consts

// Backup status values persisted in results/metadata.
const (
	BackupStatusDryRun              = "dry-run"
	BackupStatusSuccess             = "success"
	BackupStatusSuccessWithWarnings = "success_with_warnings"
)

// Exit codes untuk semantic error handling
const (
	ExitCodeSuccess         = 0 // Success
	ExitCodePermanentError  = 1 // Permanent error
	ExitCodeTransientError  = 2 // Transient error
	ExitCodeConfigError     = 3 // Configuration error
	ExitCodeValidationError = 4 // Validation error
	ExitCodeCancelled       = 5 // Operation cancelled by user/signal
)
