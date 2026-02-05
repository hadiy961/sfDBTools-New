// Package helpers berisi utility functions yang digunakan oleh profile operations.
//
// Helpers di-organize dalam subpackages berdasarkan responsibility:
//
// # Subpackages
//
// common: String dan port utilities
//   - String truncation dan sanitization
//   - Port validation dan availability check
//
// importer: CSV/XLSX parsing dan error handling
//   - row_processor.go: Process setiap row menjadi ProfileInfo
//   - error_handler.go: Error aggregation untuk bulk operations
//
// keys: Profile encryption key resolution
//   - Fallback chain: CLI flag → Env var → Interactive prompt
//   - Secure key handling (tidak di-log)
//
// loader: Profile loading dengan fallback mechanisms
//   - File path (absolute/relative)
//   - ConfigDir search
//   - Interactive selection
//   - Environment variable fallback
//   - Snapshot creation untuk edit operations
//
// parser: INI file parsing
//   - Plain text (.cnf) dan encrypted (.cnf.enc)
//   - Validasi format dan required fields
//   - Marshal ProfileInfo ↔ INI format
//
// password: Password encryption/decryption
//   - AES-256-GCM encryption
//   - Base64 encoding untuk storage
//   - Secure key derivation
//
// paths: Path resolution utilities
//   - Relative → Absolute conversion
//   - ConfigDir-relative paths
//   - Name extraction dari path
//   - Extension handling (.cnf/.enc)
//
// reader: External data source readers
//   - XLSX local files (via excelize)
//   - Google Sheets (CSV export via public link)
//   - Multi-sheet/tab support
//   - Stream processing untuk large files
//
// resolver: Import conflict resolution strategies
//   - fail: Stop on conflict
//   - skip: Skip conflicting rows
//   - overwrite: Replace existing
//   - rename: Auto-rename dengan suffix
//
// selection: Interactive profile selection
//   - Fuzzy search
//   - Preview current values
//   - Multi-select support (untuk delete)
//
// snapshot: Profile snapshot untuk comparing changes
//   - Deep copy untuk baseline
//   - Diff detection (HasMeaningfulChanges)
//   - Used untuk edit operations
//
// # Design Principles
//
// Setiap subpackage:
//   - Bersifat independent dan reusable
//   - Single responsibility
//   - Minimal dependencies ke package lain
//   - No circular dependencies
//
// # Usage Patterns
//
// Loading profile dengan fallback:
//
//	profile, err := loader.ResolveAndLoadProfile(loader.ProfileLoadOptions{
//		ConfigDir:        cfg.ConfigDir.DatabaseProfile,
//		ProfilePath:      profilePath,
//		ProfileKey:       profileKey,
//		EnvProfilePath:   "SFDB_SOURCE_PROFILE",
//		EnvProfileKey:    "SFDB_SOURCE_PROFILE_KEY",
//		AllowInteractive: true,
//	})
//
// Path resolution:
//
//	absPath, name, err := paths.ResolveConfigPathInDir(configDir, profileSpec)
//
// Interactive selection:
//
//	profile, err := selection.SelectExistingDBConfig(configDir, "Pilih profile:")
//
// Import conflict resolution:
//
//	action, newName, err := resolver.ResolveConflict(existingName, strategy)
package helpers
