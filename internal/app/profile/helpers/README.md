# Profile Helpers

Utility packages untuk mendukung profile operations. Setiap subpackage bersifat independent, reusable, dan mengikuti single responsibility principle.

## Struktur

```
helpers/
├── common/           # String dan port utilities
├── importer/         # Import processing dan error handling
├── keys/             # Encryption key resolution
├── loader/           # Profile loading dengan fallback
├── parser/           # INI file parsing
├── password/         # Password encryption/decryption
├── paths/            # Path resolution
├── reader/           # External data source readers
├── resolver/         # Conflict resolution
├── selection/        # Interactive selection
└── snapshot/         # Profile snapshot management
```

## Subpackages

### common

String utilities dan port utilities untuk general-purpose operations.

**Files:**
- `string_utils.go`: String truncation, sanitization, formatting
- `port_utils.go`: Port validation, availability check, range validation

**Key Functions:**
```go
// String utilities
TruncateString(s string, maxLen int) string
SanitizeFileName(name string) string

// Port utilities
IsValidPort(port int) bool
IsPortAvailable(port int) bool
FindAvailablePort(start, end int) (int, error)
```

**Use Cases:**
- Display truncation untuk long strings
- File name sanitization (remove invalid characters)
- Port validation untuk SSH tunnel dan database
- Find available local port untuk auto-assignment

---

### importer

Processing dan error handling untuk bulk import operations.

**Files:**
- `row_processor.go`: Process CSV/XLSX rows menjadi ProfileInfo
- `error_handler.go`: Error aggregation dan reporting untuk batch operations

**Key Functions:**
```go
// Row processing
ProcessRow(row []string, headers []string) (*domain.ProfileInfo, error)
ValidateRow(row []string, headers []string) error

// Error handling
NewErrorCollector() *ErrorCollector
collector.Add(rowNum int, field string, err error)
collector.HasErrors() bool
collector.Report() string
```

**Use Cases:**
- Bulk import dari XLSX/Google Sheets
- Row validation dengan detailed error messages
- Error collection untuk batch reporting
- Skip invalid rows (optional mode)

**Error Format:**
```
Row 5: Invalid port "abc" (must be numeric 1-65535)
Row 12: Missing required field "host"
Row 23: SSH tunnel enabled but no auth method specified
```

---

### keys

Profile encryption key resolution dengan fallback chain.

**Files:**
- `profile_key.go`: Key resolution dan secure handling

**Key Functions:**
```go
// Key resolution dengan fallback chain
ResolveProfileEncryptionKey(key string, allowPrompt bool) (resolvedKey, source string, err error)
```

**Resolution Chain:**
1. Use provided key parameter (dari flag)
2. Fallback ke `SFDB_SOURCE_PROFILE_KEY` environment variable
3. Interactive prompt (jika allowPrompt=true dan terminal available)
4. Error: key not available

**Security:**
- Key tidak pernah di-log
- Prompt menggunakan hidden input (password mode)
- Source tracking untuk audit (flag/env/prompt)
- No key persistence (memory only)

**Use Cases:**
- Profile decryption
- Profile encryption (save operation)
- Key verification untuk edit operations

---

### loader

High-level profile loading dengan berbagai fallback mechanisms.

**Files:**
- `profile_loader.go`: Profile loading dengan fallback dan interactive selection

**Key Functions:**
```go
// Main loader dengan fallback chain
ResolveAndLoadProfile(opts ProfileLoadOptions) (*domain.ProfileInfo, error)

// Convenience wrappers
LoadSourceProfile(configDir, path, key string, allowInteractive bool) (*domain.ProfileInfo, error)
SelectExistingDBConfigWithSnapshot(configDir, prompt string) (info, originalName, snapshot, error)
```

**ProfileLoadOptions:**
```go
type ProfileLoadOptions struct {
    ConfigDir         string  // Profile directory
    ProfilePath       string  // Path/name ke profile
    ProfileKey        string  // Encryption key
    EnvProfilePath    string  // Env var name untuk path fallback
    EnvProfileKey     string  // Env var name untuk key fallback
    RequireProfile    bool    // Error jika not found
    ProfilePurpose    string  // Purpose untuk error message
    AllowInteractive  bool    // Allow interactive selection
    InteractivePrompt string  // Custom prompt text
}
```

**Use Cases:**
- Load profile untuk backup/restore operations
- Interactive selection dari multiple profiles
- Create snapshot untuk edit operations (baseline comparison)
- Environment variable fallback untuk automation

---

### parser

INI file parsing untuk plain text dan encrypted profiles.

**Files:**
- `profile_parse.go`: Parse INI file menjadi ProfileInfo
- `import_parser.go`: Parse import sources (XLSX/GSheet)

**Key Functions:**
```go
// Profile parsing
LoadAndParseProfile(path, key string) (*domain.ProfileInfo, error)
ParseINIContent(content []byte) (*domain.ProfileInfo, error)

// Marshal/Unmarshal
MarshalProfileToINI(profile *domain.ProfileInfo) ([]byte, error)
UnmarshalINIToProfile(data []byte) (*domain.ProfileInfo, error)
```

**Supported Formats:**
- Plain text INI (`.cnf`)
- Encrypted INI (`.cnf.enc`) - AES-256-GCM
- Google Sheets CSV export
- XLSX (via excelize)

**INI Structure:**
```ini
[database]
host = 10.0.0.5
port = 3306
user = admin
password = <encrypted_base64>

[ssh_tunnel]
enabled = true
host = bastion.example.com
port = 22
user = sshuser
identity_file = ~/.ssh/id_rsa
local_port = 0
```

**Use Cases:**
- Load existing profiles
- Save new/edited profiles
- Parse import sources
- Format conversion (ProfileInfo ↔ INI)

---

### password

Password encryption dan decryption untuk database credentials.

**Files:**
- `profile_password.go`: Password crypto operations

**Key Functions:**
```go
// Encryption
EncryptPassword(plainPassword, encryptionKey string) (string, error)

// Decryption
DecryptPassword(encryptedPassword, encryptionKey string) (string, error)
```

**Algorithm:**
- AES-256-GCM (authenticated encryption)
- Random nonce per encryption (stored with ciphertext)
- PBKDF2 key derivation (dari encryption key)
- Base64 encoding untuk storage

**Security Properties:**
- Authenticated encryption (detect tampering)
- Unique nonce per encryption (no pattern leakage)
- Strong key derivation (protect weak keys)
- Constant-time comparison (prevent timing attacks)

**Use Cases:**
- Encrypt DB password sebelum save
- Decrypt password saat connection
- Double encryption: password encrypted, lalu entire file encrypted

---

### paths

Path resolution utilities untuk profile files.

**Files:**
- `resolve.go`: Path resolution helpers
- `resolver.go`: PathResolver class

**Key Functions:**
```go
// Path resolution
ResolveConfigPath(spec string) (absPath, name string, err error)
ResolveConfigPathInDir(dir, spec string) (absPath, name string, err error)

// PathResolver dengan ConfigDir context
type PathResolver struct {
    ConfigDir string
}
resolver.Resolve(spec string) (absPath, name, error)
```

**Resolution Rules:**

1. **Absolute path** (`/path/to/profile.cnf.enc`)
   - Use as-is
   - Extract name dari basename

2. **Relative path** (`./configs/profile.cnf.enc`, `configs/profile.cnf.enc`)
   - Resolve relative to current directory
   - Extract name dari basename

3. **Name only** (`prod-db`, `prod-db.cnf.enc`)
   - Search in ConfigDir
   - Auto-add `.cnf.enc` extension jika tidak ada

4. **ConfigDir-relative** (via ResolveConfigPathInDir)
   - Resolve relative to ConfigDir
   - Prefer ini untuk profile operations

**Name Extraction:**
- Strip directory path
- Remove `.cnf.enc` atau `.cnf` extension
- Sanitize (trim spaces)

**Use Cases:**
- Resolve user input (--file flag)
- Build profile file paths
- Extract display names
- Validate file existence

---

### reader

External data source readers untuk import operations.

**Files:**
- `source_reader.go`: Main reader interface dan dispatcher
- `xlsx_sheets.go`: XLSX local file reader
- `gsheet_tabs.go`: Google Sheets reader (via CSV export)

**Key Functions:**
```go
// Main reader dispatcher
ReadSource(source ImportSource) (rows [][]string, err error)

// XLSX reader
ReadXLSXSheet(path, sheetName string) ([][]string, error)
ListXLSXSheets(path string) ([]string, error)

// Google Sheets reader
ReadGoogleSheet(url string, gid int) ([][]string, error)
ExtractSpreadsheetID(url string) (string, error)
```

**Import Sources:**

1. **XLSX Local File**
   - Via excelize library
   - Support multiple sheets
   - Stream processing untuk large files
   - Auto-detect sheet jika tidak specified

2. **Google Sheets**
   - Public URL (edit atau share link)
   - CSV export via `https://docs.google.com/spreadsheets/d/{id}/export?format=csv&gid={gid}`
   - No authentication needed (public link only)
   - Support multiple tabs (via gid parameter)

**Error Handling:**
- File not found
- Invalid format (not XLSX)
- Sheet not found
- Network errors (untuk GSheet)
- Parse errors

**Use Cases:**
- Bulk profile import
- Multi-profile setup
- Migration dari existing config management
- Team collaboration via shared spreadsheet

---

### resolver

Conflict resolution untuk import operations.

**Files:**
- `conflict_resolver.go`: Conflict detection dan resolution

**Key Functions:**
```go
// Conflict resolution
ResolveConflict(existingName, newName string, strategy ConflictStrategy) (action ConflictAction, finalName string, err error)

// Conflict detection
DetectConflicts(profiles []ProfileInfo, configDir string) ([]Conflict, error)
```

**Conflict Strategies:**

1. **fail** (default)
   - Stop import saat first conflict
   - Return error dengan conflict details
   - Safe option (no data overwrite)

2. **skip**
   - Skip conflicting profiles
   - Continue dengan profiles yang tidak conflict
   - Log skipped profiles

3. **overwrite**
   - Replace existing profiles
   - Backup old profiles (optional)
   - Dangerous: data loss possible

4. **rename**
   - Auto-rename conflicting profiles
   - Append suffix: `-2`, `-3`, dll
   - Find first available name
   - No data loss, all profiles imported

**Conflict Detection:**
- Case-insensitive name comparison
- Extension-agnostic (ignore `.cnf.enc`)
- ConfigDir scan untuk existing profiles

**Use Cases:**
- Bulk import dengan duplicate handling
- Team setup (multiple users import same profiles)
- Re-import (update existing profiles)
- Migration (preserve existing configs)

---

### selection

Interactive profile selection dengan fuzzy search.

**Files:**
- `profile_select.go`: Interactive profile picker

**Key Functions:**
```go
// Single selection
SelectExistingDBConfig(configDir, prompt string) (domain.ProfileInfo, error)

// Multi-selection
SelectMultipleProfiles(configDir, prompt string) ([]domain.ProfileInfo, error)
```

**Features:**
- Fuzzy search (filter profiles by name)
- Preview profile details (hover/select)
- Keyboard navigation (arrow keys, enter, esc)
- Multi-select support (space bar untuk toggle)
- Empty directory handling (clear error message)

**Display Format:**
```
? Pilih profile yang ingin diedit:
  ❯ prod-db-master (10.0.0.5:3306)
    prod-db-replica (10.0.0.6:3306)
    staging-db (192.168.1.10:3306)
    dev-local (localhost:3306)
```

**Use Cases:**
- Edit profile (select from list)
- Show profile (select from list)
- Delete profiles (multi-select)
- Clone source selection

---

### snapshot

Profile snapshot untuk baseline comparison.

**Files:**
- `profile_snapshot.go`: Deep copy dan diff utilities

**Key Functions:**
```go
// Create immutable snapshot
CreateSnapshot(profile *domain.ProfileInfo) *domain.ProfileInfo

// Compare snapshots
CompareSnapshots(original, current *domain.ProfileInfo) []Change

// Check if has changes
HasMeaningfulChanges(original, current *domain.ProfileInfo) bool
```

**Snapshot Use Cases:**

1. **Edit Operations**
   - Load profile → create snapshot
   - User edit fields
   - Compare current vs snapshot
   - Display diff sebelum save
   - Detect no-op edits

2. **Change Tracking**
   - Audit trail (what changed?)
   - Confirmation prompts (show changes)
   - Rollback support (restore from snapshot)

**Meaningful Changes:**
- Fields yang affect saved profile file
- Ignore runtime metadata (Path, Size, LastModified)
- Ignore ephemeral values (ResolvedLocalPort)

**Diff Display:**
```
Changes detected:
  Name:     prod-db → prod-db-updated
  Host:     10.0.0.5 → 10.0.0.10
  Password: (changed)
  SSH Tunnel: Disabled → Enabled
    SSH Host: - → bastion.example.com
    SSH Port: - → 22
    SSH User: - → sshuser
```

---

## Common Patterns

### Error Handling

Semua helper functions return error dengan context:

```go
if err != nil {
    return fmt.Errorf("load profile failed: %w", err)
}
```

Error wrapping dengan `%w` untuk error chain yang bisa di-unwrap.

### Validation

Input validation di layer helper untuk early error detection:

```go
func ValidatePort(port int) error {
    if port < 1 || port > 65535 {
        return fmt.Errorf("invalid port %d (must be 1-65535)", port)
    }
    return nil
}
```

### Logging

Helpers tidak melakukan logging (tanggung jawab caller). Return errors untuk caller log jika needed.

### Null Safety

Check nil pointers sebelum access:

```go
if profile == nil {
    return nil, errors.New("profile is nil")
}
```

---

## Testing

Setiap helper subpackage memiliki unit tests:

```bash
# Test single subpackage
go test ./internal/app/profile/helpers/loader/

# Test all helpers
go test ./internal/app/profile/helpers/...

# With coverage
go test -cover ./internal/app/profile/helpers/...
```

Test patterns:
- Table-driven tests untuk input variations
- Mock filesystem untuk path operations
- Mock network untuk reader operations
- Error case coverage

---

## Dependencies

Helpers minimize external dependencies:

**Internal dependencies:**
- `internal/domain`: Domain models (ProfileInfo, DBInfo)
- `internal/shared`: Shared utilities (validation, fsops)
- `internal/ui`: UI components (prompt, print untuk selection)

**External dependencies:**
- `github.com/xuri/excelize/v2`: XLSX parsing
- Standard library: net, os, path, io, dll

**No dependencies:**
- Database drivers (not needed di helper layer)
- SSH clients (handled by process layer)
- Crypto libraries (wrapped by internal/crypto)

---

## Usage Examples

### Complete Profile Load Flow

```go
// With fallback chain
profile, err := loader.ResolveAndLoadProfile(loader.ProfileLoadOptions{
    ConfigDir:         "/etc/sfdbtools/profiles",
    ProfilePath:       "",  // Empty: will fallback
    ProfileKey:        "",  // Empty: will fallback
    EnvProfilePath:    "SFDB_PROFILE",
    EnvProfileKey:     "SFDB_PROFILE_KEY",
    RequireProfile:    true,
    AllowInteractive:  true,
    InteractivePrompt: "Select database profile:",
})
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Loaded profile: %s (%s:%d)\n", 
    profile.Name, profile.DBInfo.Host, profile.DBInfo.Port)
```

### Import with Conflict Resolution

```go
// Read XLSX
rows, err := reader.ReadXLSXSheet("profiles.xlsx", "Profiles")
if err != nil {
    return err
}

// Process rows
var profiles []*domain.ProfileInfo
for i, row := range rows[1:] { // Skip header
    profile, err := importer.ProcessRow(row, rows[0])
    if err != nil {
        log.Printf("Row %d invalid: %v", i+2, err)
        continue
    }
    profiles = append(profiles, profile)
}

// Detect conflicts
conflicts, err := resolver.DetectConflicts(profiles, configDir)
if err != nil {
    return err
}

// Resolve conflicts
for _, conflict := range conflicts {
    action, newName, err := resolver.ResolveConflict(
        conflict.ExistingName,
        conflict.NewName,
        resolver.StrategyRename,
    )
    if action == resolver.ActionRename {
        conflict.Profile.Name = newName
    }
}

// Save profiles
for _, profile := range profiles {
    if err := saveProfile(profile); err != nil {
        log.Printf("Failed to save %s: %v", profile.Name, err)
    }
}
```

### Interactive Selection with Snapshot

```go
// Select and create snapshot
profile, origName, snapshot, err := loader.SelectExistingDBConfigWithSnapshot(
    configDir,
    "Select profile to edit:",
)
if err != nil {
    return err
}

// User edits (via wizard or direct field updates)
profile.DBInfo.Host = "10.0.0.10"
profile.DBInfo.Port = 3307

// Check if changes are meaningful
if snapshot.HasMeaningfulChanges(profile) {
    // Display diff
    changes := snapshot.CompareSnapshots(snapshot, profile)
    for _, change := range changes {
        fmt.Printf("%s: %s → %s\n", change.Field, change.Old, change.New)
    }
    
    // Confirm and save
    if confirm("Save changes?") {
        saveProfile(profile)
    }
} else {
    fmt.Println("No meaningful changes detected")
}
```

---

## Best Practices

### Do's
✅ Keep helpers focused (single responsibility)  
✅ Return errors dengan context (wrap errors)  
✅ Validate input early (fail fast)  
✅ Document public functions (godoc format)  
✅ Write unit tests (coverage > 80%)  
✅ Use type-safe APIs (avoid `interface{}`)  

### Don'ts
❌ Don't log errors (return to caller)  
❌ Don't panic (return errors instead)  
❌ Don't depend on other app packages (only domain/shared)  
❌ Don't mutate input parameters (immutable or explicit)  
❌ Don't ignore errors (handle or propagate)  

---

## Contributing

When adding new helpers:

1. Create new subpackage under `helpers/`
2. Add `README.md` section di file ini
3. Write unit tests
4. Document exported functions
5. Update dependencies jika needed
6. Follow existing patterns

## Questions?

Refer to:
- Package documentation (`doc.go` files)
- Architecture documentation (`../ARCHITECTURE.md`)
- Code examples di unit tests
