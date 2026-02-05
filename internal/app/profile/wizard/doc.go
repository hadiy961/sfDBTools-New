// Package wizard menyediakan interactive wizard flows untuk profile operations.
//
// Wizard menangani user interaction untuk:
//   - Create: Wizard lengkap untuk profile baru (DB info + SSH tunnel)
//   - Edit: Wizard dengan pre-filled values untuk edit profile
//   - Clone: Wizard dengan pre-fill dari source profile
//
// # Architecture
//
// Wizard menggunakan dependency injection untuk flexibility:
//
//	type Runner struct {
//		State     *profilemodel.ProfileState  // Shared state
//		Validator NameValidator                // Name uniqueness check
//		Loader    SnapshotLoader              // Load existing profiles
//	}
//
// # Flow Pattern
//
// Setiap wizard flow mengikuti pattern:
//
//  1. Display header/info
//  2. Prompt user input (dengan validation)
//  3. Update state
//  4. Display summary
//  5. Confirm save
//  6. Return control ke Executor
//
// # Create Flow
//
//	runCreateFlow()
//	  → promptProfileName()
//	  → promptDatabaseInfo()
//	  → promptSSHTunnel()
//	  → Display summary
//	  → Confirm save
//
// # Edit Flow
//
//	runEditFlow()
//	  → Display current values
//	  → Show action menu (Edit Name/DB/SSH/Done)
//	  → Apply changes to State
//	  → Display diff (original vs new)
//	  → Confirm save
//
// # Clone Flow
//
//	runCloneFlow()
//	  → State already pre-filled from source
//	  → Show action menu (Edit fields/Save)
//	  → Apply modifications
//	  → Save clone
//
// # State Updates
//
// Wizard bekerja directly dengan shared ProfileState:
//
//	// Update state
//	r.State.ProfileInfo.Name = newName
//	r.State.ProfileInfo.DBInfo.Host = newHost
//
// Tidak perlu sync karena state di-share via pointer.
//
// # Validation
//
// Wizard melakukan validation di beberapa level:
//
//  1. Input validation (per-field): format, range, dll
//  2. Business validation: name uniqueness, port availability
//  3. Comprehensive validation: full profile structure before save
//
// # Error Handling
//
// Wizard menangani errors dengan user-friendly approach:
//
//   - Validation error → Retry prompt
//   - User cancelled (Ctrl+C) → Return ErrUserCancelled
//   - System error → Bubble up ke caller
//
// # Prompts
//
// Wizard menggunakan survey package untuk prompts dengan features:
//   - Default values (untuk edit mode)
//   - Validation functions
//   - Help text
//   - Conditional prompts (skip jika tidak relevan)
//
// # User Experience
//
// Wizard dirancang untuk UX yang smooth:
//   - Clear instructions dan labels
//   - Sane defaults (port 3306, SSH port 22)
//   - Pre-filled values untuk edit/clone
//   - Immediate feedback untuk validation errors
//   - Summary sebelum save untuk review
//   - Retry option jika validation gagal
//
// # Example Usage
//
//	runner := wizard.New(logger, configDir, state, validator, loader)
//	err := runner.Run(consts.ProfileModeCreate)
//	if err == validation.ErrUserCancelled {
//		// User cancelled, handle gracefully
//	}
package wizard
