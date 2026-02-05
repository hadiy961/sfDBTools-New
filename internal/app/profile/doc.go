// Package profile menyediakan manajemen profil koneksi database untuk sfdbtools.
//
// # Arsitektur
//
// Package ini mengikuti pola layered architecture dengan separation of concerns:
//
//   - Service Layer (service.go): Entry point dan orchestration
//   - Executor Layer (executor/): Eksekusi operasi CRUD profile
//   - Wizard Layer (wizard/): Interactive UI flow untuk create/edit/clone
//   - Validation Layer (validation/): Validasi input dan business rules
//   - Display Layer (display/): Formatting dan rendering output
//   - Model Layer (model/): Domain models dan state management
//   - Connection Layer (connection/): Database connection handling
//   - Process Layer (process/): SSH tunnel dan background process management
//   - Helpers: Utility functions yang dipakai lintas layer
//
// # Flow Eksekusi Umum
//
//  1. Command Layer (cmd) → Service.ExecuteProfileCommand()
//  2. Service → Executor (create/edit/show/delete/clone/import)
//  3. Executor → Wizard (untuk interactive mode)
//  4. Wizard → Validation → Display
//  5. Executor → Save/Show hasil akhir
//
// # State Management
//
// Menggunakan ProfileState sebagai single source of truth yang di-share
// antar semua komponen untuk menghindari state synchronization issues.
// State di-inject sebagai pointer ke Service, Executor, dan Wizard.
//
// # Operasi yang Didukung
//
//   - create: Membuat profile baru (wizard/non-interactive)
//   - show: Menampilkan profile existing
//   - edit: Mengedit profile existing (wizard)
//   - delete: Menghapus profile
//   - clone: Clone profile dengan modifikasi minimal
//   - import: Bulk import dari XLSX/Google Sheets
//
// # Subpackages
//
//   - connection: Database connection handling (direct/SSH tunnel)
//   - display: Output formatting dan rendering (tables, summaries)
//   - errors: Custom error types untuk profile operations
//   - executor: Implementasi operasi profile (CRUD + import)
//   - merger: Profile merging dan cloning utilities
//   - model: Domain models (ProfileInfo, ProfileState, Options)
//   - process: SSH tunnel dan background process management
//   - validation: Input dan business rule validation
//   - wizard: Interactive wizard flows (create/edit/clone)
//   - helpers/*: Utility functions dan shared logic
//
// # Contoh Penggunaan
//
// Create profile via service:
//
//	// Prepare options
//	opts := &profilemodel.ProfileCreateOptions{
//		Interactive: true,
//	}
//
//	// Create service
//	svc, err := profile.NewProfileService(cfg, logger, opts)
//	if err != nil {
//		return err
//	}
//
//	// Execute operation
//	err = svc.ExecuteProfileCommand(profilemodel.ProfileEntryConfig{
//		Mode:        consts.ProfileModeCreate,
//		HeaderTitle: "Buat Profile Baru",
//		LogPrefix:   "create",
//	})
//
// Edit profile interactively:
//
//	opts := &profilemodel.ProfileEditOptions{
//		Interactive: true,
//	}
//	svc, _ := profile.NewProfileService(cfg, logger, opts)
//	err := svc.ExecuteProfileCommand(profilemodel.ProfileEntryConfig{
//		Mode: consts.ProfileModeEdit,
//	})
//
// # Security
//
// Profile files disimpan dalam format encrypted (.cnf.enc) menggunakan AES-256-GCM.
// Encryption key harus disediakan via:
//   - Flag --profile-key
//   - Environment variable SFDB_SOURCE_PROFILE_KEY
//   - Interactive prompt (untuk terminal mode)
//
// # File Format
//
// Profile disimpan dalam format INI dengan struktur:
//
//	[database]
//	host = 10.0.0.5
//	port = 3306
//	user = admin
//	password = encrypted_password
//
//	[ssh_tunnel]
//	enabled = true
//	host = bastion.example.com
//	port = 22
//	user = sshuser
//	# ... SSH tunnel config
package profile
