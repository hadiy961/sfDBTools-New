// Package validation menyediakan validasi untuk profile input dan business rules.
//
// Package ini menangani berbagai level validasi:
//
// # Input Validation
//
// Validasi format dan constraint dasar:
//   - Profile name: format, length, allowed characters
//   - Host: IP address atau hostname format
//   - Port: range validation (1-65535)
//   - User: non-empty, length
//   - Password: non-empty (optional untuk SSH key auth)
//
// # Business Rule Validation
//
// Validasi business logic:
//   - Name uniqueness: check existing profiles di ConfigDir
//   - SSH tunnel consistency: jika enabled, wajib ada host/user
//   - Auth method: password XOR identity file (tidak boleh keduanya kosong)
//   - Port availability: check local port tidak bentrok (untuk SSH tunnel)
//
// # Profile Structure Validation
//
// Validasi comprehensive untuk full profile:
//   - Database info completeness
//   - SSH tunnel config consistency
//   - Required fields presence
//   - Cross-field validation (dependent fields)
//
// # Import Validation
//
// Validasi khusus untuk bulk import:
//   - Row-by-row validation
//   - Required column presence
//   - Data type validation
//   - Conflict detection dengan existing profiles
//   - Error aggregation untuk batch reporting
//
// # Validation Functions
//
// ValidateProfileInfo(profile):
//   - Comprehensive validation untuk full profile
//   - Used sebelum save operation
//   - Return first error found (fail-fast)
//
// ValidateProfileName(name):
//   - Name format validation
//   - Allowed characters: alphanumeric, dash, underscore, dot
//   - Length: 1-64 characters
//
// ValidateDatabaseInfo(dbInfo):
//   - Host, port, user, password validation
//   - Port range check
//   - Non-empty required fields
//
// ValidateSSHTunnel(sshInfo):
//   - Conditional validation (jika enabled)
//   - Host, port, user required
//   - Auth method validation (password OR identity file)
//   - Identity file existence check
//
// CheckNameUniqueness(name, configDir):
//   - Scan existing profiles
//   - Case-insensitive comparison
//   - Return error jika conflict
//
// # Error Handling
//
// Validation errors adalah user-facing errors dengan clear messages:
//
//	err := validation.ValidateProfileName("invalid name!")
//	// Error: "profile name contains invalid characters: !"
//
// Error types:
//   - ErrInvalidFormat: Format tidak sesuai
//   - ErrRequired: Required field kosong
//   - ErrOutOfRange: Value di luar range yang diizinkan
//   - ErrConflict: Uniqueness constraint violation
//
// # Usage Examples
//
// Basic validation:
//
//	if err := validation.ValidateProfileInfo(profile); err != nil {
//		return fmt.Errorf("validation failed: %w", err)
//	}
//
// Name uniqueness check:
//
//	if err := validation.CheckNameUniqueness(name, configDir); err != nil {
//		// Handle conflict (prompt rename, etc)
//	}
//
// SSH tunnel validation:
//
//	if profile.SSHTunnel.Enabled {
//		if err := validation.ValidateSSHTunnel(&profile.SSHTunnel); err != nil {
//			return err
//		}
//	}
//
// # Integration
//
// Validation di-trigger di berbagai points:
//   - Wizard: Per-field validation saat prompt (immediate feedback)
//   - Executor: Full profile validation sebelum save
//   - Import: Batch validation dengan error collection
//   - CLI: Flag validation untuk non-interactive mode
package validation
