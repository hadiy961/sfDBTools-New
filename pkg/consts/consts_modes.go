package consts

// Shared mode strings used across backup and restore.
// Keep these centralized so cmd/, internal/, and pkg/ can reuse them consistently.
const (
	ModeSingle    = "single"
	ModePrimary   = "primary"
	ModeSecondary = "secondary"
	ModeSeparated = "separated"
	ModeSeparate  = "separate"
	ModeCombined  = "combined"
	ModeAll       = "all"
	ModeSelection = "selection"
)
