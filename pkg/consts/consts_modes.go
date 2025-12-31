package consts

// Shared mode strings used across backup and restore.
// Keep these centralized so cmd/, internal/, and pkg/ can reuse them consistently.
const (
	// Core modes (backup and restore)
	ModeSingle    = "single"
	ModePrimary   = "primary"
	ModeSecondary = "secondary"
	ModeAll       = "all"

	// Backup : filter mode
	ModeFilter = "filter"

	// Backup: filter output variants
	// - single-file: combined
	// - multi-file: separated
	ModeCombined  = "combined"
	ModeSeparated = "separated"

	// Restore modes
	ModeSelection = "selection"
	ModeCustom    = "custom"
)
