// Package executor mengimplementasikan operasi CRUD untuk profile management.
//
// Package ini adalah inti dari profile management yang menangani:
//   - Create: Membuat profile baru via wizard atau direct flags
//   - Edit: Mengedit profile existing via wizard
//   - Show: Menampilkan profile (single/all)
//   - Delete: Menghapus profile dengan confirmation
//   - Clone: Clone profile dengan modifikasi minimal
//   - Import: Bulk import dari XLSX/Google Sheets
//   - Save: Persist profile ke disk (encrypted)
//
// # Architecture Pattern
//
// Executor menggunakan Interface Segregation Principle (ISP) untuk dependencies:
//
//	type ProfileOps interface {
//		WizardProvider           // untuk interactive flows
//		ProfileSaver             // untuk persist ke file
//		NameUniquenessChecker    // untuk validasi uniqueness
//		SnapshotLoader           // untuk load existing profiles
//		ExistingConfigSelector   // untuk interactive selection
//		ConfigFormatter          // untuk format output
//	}
//
// Benefit: setiap method hanya depend pada interface yang dibutuhkan,
// memudahkan testing dan mengurangi coupling.
//
// # State Management
//
// Executor bekerja dengan ProfileState yang di-share across components:
//
//	type Executor struct {
//		State *profilemodel.ProfileState  // Shared state pointer
//		Ops   ProfileOps                   // Operations interface
//		// ... other fields
//	}
//
// State tidak perlu di-sync secara eksplisit karena semua components
// bekerja dengan pointer yang sama.
//
// # Execution Flow
//
// Create Flow:
//
//	Executor.Create()
//	  → [Interactive] Wizard.Run(create)
//	  → [Non-interactive] Direct from flags
//	  → Validation
//	  → SaveProfile()
//
// Edit Flow:
//
//	Executor.Edit()
//	  → Load existing + create snapshot
//	  → Wizard.Run(edit) with pre-filled values
//	  → Check HasMeaningfulChanges()
//	  → SaveProfile() if changed
//
// Import Flow:
//
//	Executor.Import()
//	  → Parse source (XLSX/GSheet)
//	  → Validate rows
//	  → Resolve conflicts
//	  → Batch save profiles
//
// # Error Handling
//
// Executor methods return errors yang bisa di-wrap untuk context:
//
//	if err := e.CreateProfile(); err != nil {
//		return fmt.Errorf("create profile failed: %w", err)
//	}
//
// Special errors:
//   - validation.ErrUserCancelled: User cancelled operation
//   - profileerrors.ErrProfileExists: Profile name conflict
//   - profileerrors.ErrConnectionFailed: Connection test failed
//
// # Testing
//
// Executor mudah di-test karena menggunakan interface:
//
//	mockOps := &MockProfileOps{}
//	executor := executor.New(logger, cfg, configDir, state, mockOps)
//	err := executor.CreateProfile()
package executor
