// Package model berisi domain models dan state management untuk profile operations.
//
// Package ini menyediakan:
//   - ProfileState: Single source of truth untuk shared state
//   - ProfileOptions: Union interface untuk operation-specific options
//   - ProfileEntryConfig: Configuration untuk entry point operations
//   - Concrete option types untuk setiap operation mode
//
// # ProfileState
//
// Central state management untuk semua profile operations:
//
//	type ProfileState struct {
//		ProfileInfo         *domain.ProfileInfo  // Current/working state
//		Options             ProfileOptions       // Operation-specific options
//		OriginalProfileName string              // For edit: original name
//		OriginalProfileInfo *domain.ProfileInfo // For edit: baseline snapshot
//	}
//
// State lifecycle:
//  1. Created: Di Service layer saat initialization
//  2. Initialized: Options dan ProfileInfo di-set berdasarkan operation mode
//  3. Shared: Pointer di-inject ke Executor, Wizard, Display
//  4. Mutated: Components update state directly (shared pointer)
//  5. Read: Semua components baca state yang sama (always up-to-date)
//
// Benefits:
//   - Eliminasi sync functions (no manual state sync)
//   - Real-time state visibility across components
//   - No state drift issues
//   - Simpler code flow (no return values untuk state updates)
//
// # ProfileOptions Interface
//
// Union-style interface untuk operation-specific options:
//
//	type ProfileOptions interface {
//		Mode() string
//		IsInteractive() bool
//	}
//
// Concrete implementations:
//   - ProfileCreateOptions
//   - ProfileEditOptions
//   - ProfileShowOptions
//   - ProfileDeleteOptions
//   - ProfileCloneOptions
//   - ProfileImportOptions
//
// Design rationale:
//   - Avoid multiple option pointers (only one active per operation)
//   - Type-safe access via helper methods (State.CreateOptions(), etc)
//   - Common interface untuk mode detection dan interactive check
//
// # Operation Options
//
// ProfileCreateOptions:
//   - Interactive: Enable wizard flow
//   - OutputDir: Custom output directory (override ConfigDir)
//   - ProfileInfo: Working profile data
//
// ProfileEditOptions:
//   - Interactive: Enable wizard dengan pre-fill
//   - NewName: Target name untuk rename (optional)
//   - NewProfileKey: New encryption key (optional)
//   - ProfileInfo: Current profile data
//
// ProfileShowOptions:
//   - Interactive: Enable connection test dan selection
//   - RevealPassword: Show plain password (dengan confirmation)
//   - ProfileInfo: Profile to display
//
// ProfileDeleteOptions:
//   - Interactive: Enable confirmation prompts
//   - Force: Skip confirmation (dangerous!)
//   - Profiles: List of profiles to delete
//
// ProfileCloneOptions:
//   - Interactive: Enable wizard dengan pre-fill dari source
//   - SourceProfile: Source profile path/name
//   - TargetName: Target profile name
//   - TargetHost: Override host (optional)
//   - TargetPort: Override port (optional)
//   - ProfileKey: Source profile key
//   - NewProfileKey: Target profile key (optional)
//
// ProfileImportOptions:
//   - Input: XLSX file path (mutually exclusive dengan GSheetURL)
//   - GSheetURL: Google Spreadsheet URL
//   - Sheet/GID: Sheet identifier
//   - OnConflict: Conflict resolution strategy (fail/skip/overwrite/rename)
//   - SkipConfirm: Skip user confirmation (automation mode)
//   - SkipInvalidRows: Skip invalid rows instead of failing
//   - ContinueOnError: Continue saat connection test/save failed
//   - SkipConnTest: Skip connection test (fast mode)
//   - ConnTestDone: Flag untuk indicate test sudah dilakukan
//
// # State Helpers
//
// Type-safe option access:
//
//	if createOpts, ok := state.CreateOptions(); ok {
//		// Access create-specific fields
//		outputDir := createOpts.OutputDir
//	}
//
// Mode detection:
//
//	mode := state.Mode() // "create", "edit", "show", dll
//
// Interactive check:
//
//	if state.IsInteractive() {
//		// Run wizard atau prompts
//	}
//
// Change detection:
//
//	if state.HasMeaningfulChanges() {
//		// Save needed
//	}
//
// # ProfileEntryConfig
//
// Configuration untuk ExecuteProfileCommand router:
//
//	type ProfileEntryConfig struct {
//		HeaderTitle string // UI header title
//		Mode        string // "create", "show", "edit", "delete"
//		SuccessMsg  string // Success message
//		LogPrefix   string // Log prefix for tracking
//	}
//
// Usage:
//
//	err := service.ExecuteProfileCommand(ProfileEntryConfig{
//		Mode:        consts.ProfileModeCreate,
//		HeaderTitle: "Buat Profile Baru",
//		SuccessMsg:  "Profile berhasil dibuat",
//		LogPrefix:   "create",
//	})
//
// # Change Detection
//
// HasMeaningfulChanges() logic:
//   - Compare ProfileInfo dengan OriginalProfileInfo
//   - Ignore metadata fields (Path, Size, LastModified, ResolvedLocalPort)
//   - Focus pada fields yang affect profile file content
//   - Used untuk edit mode: detect if save needed
//
// Meaningful changes:
//   - Name change
//   - Database info (host, port, user, password)
//   - SSH tunnel config (all fields)
//
// Non-meaningful changes (ignored):
//   - Path (runtime metadata)
//   - Size (runtime metadata)
//   - LastModified (runtime metadata)
//   - ResolvedLocalPort (runtime, ephemeral)
//
// # Example Usage
//
// Initialize state untuk create:
//
//	opts := &ProfileCreateOptions{
//		Interactive: true,
//	}
//	state := &ProfileState{
//		Options:     opts,
//		ProfileInfo: &opts.ProfileInfo,
//	}
//
// Initialize state untuk edit:
//
//	opts := &ProfileEditOptions{
//		Interactive: true,
//	}
//	state := &ProfileState{
//		Options:             opts,
//		ProfileInfo:         &opts.ProfileInfo, // Working copy
//		OriginalProfileName: loaded.Name,
//		OriginalProfileInfo: snapshot,          // Immutable baseline
//	}
//
// Check mode dan act:
//
//	switch state.Mode() {
//	case consts.ProfileModeCreate:
//		// Handle create
//	case consts.ProfileModeEdit:
//		if !state.HasMeaningfulChanges() {
//			return nil // No-op
//		}
//		// Handle edit
//	}
package model
