// Package display menyediakan formatting dan rendering untuk profile output.
//
// Package ini menangani semua aspek visual dari profile operations:
//
// # Displayer
//
// Main component yang orchestrate display operations:
//
//	type Displayer struct {
//		ConfigDir string
//		State     *profilemodel.ProfileState
//	}
//
// Methods:
//   - DisplayProfileDetails(): Show single profile dengan connection test
//   - DisplayProfileListTable(): List all profiles dalam table format
//   - DisplayImportSummary(): Summary hasil bulk import
//   - DisplayImportConflicts(): Conflicts yang perlu di-resolve
//
// # Display Modes
//
// Create/Edit/Clone Summary:
//   - Profile name dan metadata
//   - Database info (host, port, user)
//   - SSH tunnel config (jika enabled)
//   - Diff display (untuk edit: original vs new)
//   - Change indicators (untuk field yang berubah)
//
// Show Profile:
//   - Comprehensive profile details
//   - Connection test results (interactive mode)
//   - Health status (HEALTHY/UNHEALTHY/SKIPPED)
//   - Latency information
//   - Error hints jika connection failed
//
// List Profiles:
//   - Table dengan Name, Path, Last Modified
//   - Sortable columns
//   - Pagination support (untuk large lists)
//   - Quick actions (show/edit/delete)
//
// Import Summary:
//   - Total profiles processed
//   - Success/Failed/Skipped counts
//   - Conflict resolution summary
//   - Invalid rows dengan error messages
//   - Execution time
//
// # Formatting
//
// Password Formatting:
//   - Masked: "********" (default)
//   - Revealed: plain text (dengan confirmation)
//   - State indicator: "(set)" atau "(not set)"
//
// SSH Tunnel Formatting:
//   - Enabled/Disabled indicator
//   - Auth method display (password atau key file)
//   - Local port display (auto-assigned atau specified)
//
// Diff Formatting:
//   - Original value → New value
//   - Color coding: red (removed), green (added), yellow (changed)
//   - Field-by-field comparison
//   - Highlight meaningful changes only (ignore metadata)
//
// # Table Rendering
//
// Using internal/ui/table for consistent formatting:
//   - Auto column width adjustment
//   - Header styling
//   - Border customization
//   - Alignment support (left/right/center)
//
// # Connection Test Display
//
// Visual progress untuk connection test steps:
//
//	DNS Resolution    ✓ 10.0.0.5
//	TCP Connection    ✓ 3306
//	SSH Tunnel        ✓ bastion:22 → localhost:33060
//	Authentication    ✓ user@host
//	DB Version        ✓ MariaDB 10.11.8
//	Health           HEALTHY
//	Latency          125ms
//
// Error display:
//   - Step-by-step status (success/failed/skipped)
//   - Error message dengan context
//   - Actionable hints untuk troubleshooting
//
// # Import Displays
//
// Import Plan:
//   - Valid profiles to import
//   - Invalid profiles dengan errors
//   - Conflicts dengan resolution strategy
//   - Estimated time (untuk large batches)
//
// Import Progress:
//   - Real-time progress bar
//   - Current profile being processed
//   - Success/Failed counters
//   - Elapsed time
//
// Import Summary:
//   - Final statistics
//   - Failed profiles dengan error details
//   - Skipped profiles dengan reasons
//   - Saved profiles locations
//
// # Color Scheme
//
// Consistent color usage:
//   - Green: Success, healthy, added
//   - Red: Error, unhealthy, removed
//   - Yellow: Warning, changed, pending
//   - Blue: Info, neutral, metadata
//   - Gray: Disabled, skipped, not applicable
//
// # Responsive Display
//
// Adapt display based on:
//   - Terminal width (wrap/truncate)
//   - Interactive vs non-interactive mode
//   - Quiet mode (minimal output)
//   - TTY detection (no colors for pipes)
//
// # Usage Examples
//
// Display profile details:
//
//	displayer := display.NewDisplayer(configDir, state)
//	displayer.DisplayProfileDetails()
//
// Show diff untuk edit:
//
//	// State contains both ProfileInfo (current) and OriginalProfileInfo
//	displayer.DisplayProfileDetails() // Auto-detect edit mode dan show diff
//
// Display import summary:
//
//	summary := &display.ImportSummary{
//		Total:     100,
//		Success:   95,
//		Failed:    3,
//		Skipped:   2,
//		Errors:    errorDetails,
//	}
//	displayer.DisplayImportSummary(summary)
package display
