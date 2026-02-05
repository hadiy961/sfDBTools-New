# Profile Package Architecture

## Overview

Package `profile` mengelola database connection profiles untuk sfdbtools. Profile disimpan dalam format encrypted INI files dan mendukung koneksi direct maupun via SSH tunnel.

## Component Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                         CLI Layer                            │
│                    (cmd/profile/*.go)                        │
│  - create.go, edit.go, show.go, delete.go                   │
│  - clone.go, import.go                                       │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                      Service Layer                           │
│                  (profile/service.go)                        │
│                                                              │
│  - NewProfileService(cfg, logger, options)                  │
│  - ExecuteProfileCommand(config) [Router]                   │
│                                                              │
│  Responsibilities:                                           │
│  - Initialize ProfileState                                   │
│  - Route operations ke executor methods                      │
│  - Manage shared dependencies                                │
└───────────────────────────┬─────────────────────────────────┘
                            │
        ┌───────────────────┴───────────────────┐
        │                                       │
        ▼                                       ▼
┌──────────────────────────┐         ┌─────────────────────────┐
│    Executor Layer        │         │     Wizard Layer        │
│    (executor/)           │◄───────►│     (wizard/)           │
│                          │         │                         │
│  - executor.go (main)    │         │  - runner.go (main)     │
│  - create.go             │         │  - create_flow.go       │
│  - edit.go               │         │  - edit_flow.go         │
│  - show.go               │         │  - clone_flow.go        │
│  - delete.go             │         │  - prompts.go           │
│  - clone.go              │         │  - validators.go        │
│  - import.go             │         │                         │
│  - save.go               │         │  Responsibilities:      │
│                          │         │  - User interaction     │
│  Responsibilities:       │         │  - Input prompts        │
│  - Orchestrate operations│         │  - Field validation     │
│  - Call wizard for UI    │         │  - Display summaries    │
│  - Validate & save       │         │  - Confirmation flows   │
│  - Handle errors         │         │                         │
└────────┬─────────────────┘         └─────────┬───────────────┘
         │                                     │
         │                                     │
         └────────────┬────────────────────────┘
                      │
      ┌───────────────┼───────────────┬────────────────┐
      │               │               │                │
      ▼               ▼               ▼                ▼
┌────────────┐  ┌────────────┐  ┌───────────┐  ┌─────────────┐
│ Validation │  │  Display   │  │Connection │  │   Helpers   │
│            │  │            │  │           │  │             │
│ - Input    │  │ - Formatter│  │ - Direct  │  │ - Loader    │
│ - Rules    │  │ - Tables   │  │ - Tunnel  │  │ - Parser    │
│ - SSH      │  │ - Summary  │  │ - Test    │  │ - Keys      │
│ - Profile  │  │ - Password │  │ - Preflight│ │ - Paths     │
│            │  │            │  │           │  │ - Reader    │
└────────────┘  └────────────┘  └───────────┘  └─────────────┘
```

## Layered Architecture

```
┌──────────────────────────────────────────────────────────┐
│                    Presentation Layer                     │
│                   (CLI Commands + Wizard)                 │
└────────────────────────┬─────────────────────────────────┘
                         │
┌────────────────────────▼─────────────────────────────────┐
│                   Application Layer                       │
│              (Service + Executor + Display)               │
└────────────────────────┬─────────────────────────────────┘
                         │
┌────────────────────────▼─────────────────────────────────┐
│                     Domain Layer                          │
│              (Models + Validation + Errors)               │
└────────────────────────┬─────────────────────────────────┘
                         │
┌────────────────────────▼─────────────────────────────────┐
│                  Infrastructure Layer                     │
│       (Helpers: Parser, Loader, Reader, Paths, etc)      │
└──────────────────────────────────────────────────────────┘
```

## Data Flow Diagrams

### 1. Create Profile Flow

```
┌──────────┐
│   User   │
└────┬─────┘
     │ sfdbtools profile create [--flags]
     ▼
┌─────────────────────┐
│  CLI Command        │
│  (cmd/profile)      │
└─────────┬───────────┘
          │ Parse flags → ProfileCreateOptions
          ▼
┌─────────────────────────────┐
│  Service.CreateProfile()    │
│                             │
│  Initialize State with opts │
└─────────┬───────────────────┘
          │
          ▼
┌─────────────────────────────┐
│  Executor.Create()          │
└─────────┬───────────────────┘
          │
    ┌─────┴──────┐
    │            │
    │ Interactive? 
    │            │
    ▼            ▼
  [YES]        [NO]
    │            │
    │            └──> Use flags directly
    │
    ▼
┌──────────────────────────┐
│  Wizard.Run(create)      │
│                          │
│  1. Prompt profile name  │
│  2. Prompt DB info       │
│  3. Prompt SSH tunnel    │
│  4. Display summary      │
│  5. Confirm save         │
└──────┬───────────────────┘
       │
       ▼
┌──────────────────────────────┐
│  Validation                  │
│  - ValidateProfileInfo()     │
│  - CheckNameUnique()         │
└──────┬───────────────────────┘
       │
       ▼
┌──────────────────────────────┐
│  Executor.SaveProfile()      │
│                              │
│  1. Test connection (if !skip)│
│  2. Encrypt password         │
│  3. Format to INI            │
│  4. Encrypt file (AES-256)   │
│  5. Write to disk            │
└──────┬───────────────────────┘
       │
       ▼
┌──────────────┐
│   Success    │
│  Display msg │
└──────────────┘
```

### 2. Edit Profile Flow

```
┌──────────┐
│   User   │
└────┬─────┘
     │ sfdbtools profile edit [--file profile.enc]
     ▼
┌─────────────────────┐
│  CLI Command        │
└─────────┬───────────┘
          │
          ▼
┌─────────────────────────────┐
│  Service.EditProfile()      │
└─────────┬───────────────────┘
          │
          ▼
┌─────────────────────────────────────┐
│  Executor.Edit()                    │
│                                     │
│  1. Resolve profile path            │
│  2. Load existing profile           │
│  3. Create snapshot (baseline)      │
│  4. Store in State.OriginalProfile  │
└─────────┬───────────────────────────┘
          │
          ▼
┌──────────────────────────────────────┐
│  Wizard.Run(edit)                    │
│                                      │
│  1. Display current values           │
│  2. Show action menu:                │
│     - Edit Name                      │
│     - Edit Database Info             │
│     - Edit SSH Tunnel                │
│     - Cancel                         │
│     - Done                           │
│  3. Apply changes to State           │
│  4. Display diff (original vs new)   │
│  5. Confirm save                     │
└─────────┬────────────────────────────┘
          │
          ▼
┌─────────────────────────────────────┐
│  State.HasMeaningfulChanges() ?     │
└─────────┬───────────┬───────────────┘
          │           │
      [NO changes] [Has changes]
          │           │
          │           ▼
          │    ┌──────────────────┐
          │    │  SaveProfile()   │
          │    │  - Validate      │
          │    │  - Test conn     │
          │    │  - Overwrite file│
          │    └──────────────────┘
          │
          ▼
┌─────────────────────┐
│  Cancel (no-op)     │
│  Return cancelled   │
└─────────────────────┘
```

### 3. Clone Profile Flow

```
┌──────────┐
│   User   │
└────┬─────┘
     │ sfdbtools profile clone --source "prod"
     ▼
┌─────────────────────┐
│  CLI Command        │
└─────────┬───────────┘
          │
          ▼
┌──────────────────────────────┐
│  Service.CloneProfile()      │
└─────────┬────────────────────┘
          │
          ▼
┌─────────────────────────────────────┐
│  Executor.Clone()                   │
│                                     │
│  1. Load source profile             │
│  2. Create copy for target          │
│  3. Pre-fill State with source data │
│  4. Apply overrides (host/port)     │
└─────────┬───────────────────────────┘
          │
          ▼
┌──────────────────────────────────────┐
│  Wizard.Run(clone)                   │
│                                      │
│  State already pre-filled!           │
│                                      │
│  1. Display source profile info      │
│  2. Show action menu:                │
│     - Edit Target Name               │
│     - Edit Database Info             │
│     - Edit SSH Tunnel                │
│     - Save Clone                     │
│     - Cancel                         │
│  3. Apply modifications              │
│  4. Confirm & save                   │
└─────────┬────────────────────────────┘
          │
          ▼
┌─────────────────────────────────────┐
│  SaveProfile()                      │
│  - New name validation              │
│  - New encryption key (optional)    │
│  - Save as new file                 │
└─────────────────────────────────────┘
```

### 4. Import Profiles Flow (Bulk)

```
┌──────────┐
│   User   │
└────┬─────┘
     │ sfdbtools profile import --input profiles.xlsx
     ▼
┌─────────────────────┐
│  CLI Command        │
└─────────┬───────────┘
          │
          ▼
┌──────────────────────────────┐
│  Service.ImportProfiles()    │
└─────────┬────────────────────┘
          │
          ▼
┌─────────────────────────────────────────┐
│  Executor.Import()                      │
│                                         │
│  1. Detect source (XLSX/GSheet)         │
│  2. Parse rows → []ProfileInfo          │
│  3. Validate each row                   │
│  4. Detect conflicts with existing      │
│  5. Resolve conflicts (fail/skip/       │
│     overwrite/rename)                   │
└─────────┬───────────────────────────────┘
          │
          ▼
┌──────────────────────────────────────────┐
│  Display import plan                     │
│  - Valid: X profiles                     │
│  - Invalid: Y profiles                   │
│  - Conflicts: Z profiles                 │
│                                          │
│  Show conflict resolution strategy       │
└─────────┬────────────────────────────────┘
          │
          ▼
┌──────────────────────────────────────────┐
│  User confirmation (if !skip-confirm)    │
└─────────┬────────────────────────────────┘
          │
          ▼
┌──────────────────────────────────────────┐
│  Batch processing                        │
│                                          │
│  For each valid profile:                 │
│    1. Test connection (if !skip-test)    │
│    2. Encrypt & save                     │
│    3. Aggregate results                  │
│                                          │
│  Continue-on-error handling              │
└─────────┬────────────────────────────────┘
          │
          ▼
┌──────────────────────────────────────────┐
│  Display summary                         │
│  - Successfully imported: X              │
│  - Failed: Y                             │
│  - Skipped: Z                            │
│  - Error details (if any)                │
└──────────────────────────────────────────┘
```

## State Management

### ProfileState Structure

```go
type ProfileState struct {
    ProfileInfo         *domain.ProfileInfo  // Current/working state
    Options             ProfileOptions       // Operation-specific options
    OriginalProfileName string              // For edit: original name
    OriginalProfileInfo *domain.ProfileInfo // For edit: baseline snapshot
}
```

### State Sharing Pattern

```
                    ProfileState (heap)
                          │
        ┌─────────────────┼─────────────────┐
        │                 │                 │
        ▼                 ▼                 ▼
    Service           Executor           Wizard
    (pointer)         (pointer)         (pointer)
        │                 │                 │
        └─────────────────┴─────────────────┘
                All see same state!
```

**Benefits:**
- Eliminasi sync functions
- Real-time state visibility
- Simpler code flow
- No state drift issues

### State Lifecycle

```
1. Create (Service)
   state := &ProfileState{}

2. Initialize (Service) 
   state.Options = createOptions
   state.ProfileInfo = &createOptions.ProfileInfo

3. Share (Injection)
   executor := executor.New(..., state, ...)
   wizard := wizard.New(..., state, ...)

4. Mutate (Wizard/Executor)
   state.ProfileInfo.Name = "new-name"
   state.ProfileInfo.DBInfo.Host = "10.0.0.5"

5. Read (Anywhere)
   if state.HasMeaningfulChanges() { ... }
   mode := state.Mode()
```

## Dependency Injection & Interfaces

### Interface Segregation (ISP)

Executor tidak depend pada concrete Service, tapi pada interface yang dibutuhkan:

```go
// Executor hanya butuh methods ini
type ProfileOps interface {
    WizardProvider           // NewWizard()
    ProfileSaver             // SaveProfile()
    NameUniquenessChecker    // CheckNameUnique()
    SnapshotLoader           // LoadSnapshot()
    ExistingConfigSelector   // PromptSelect()
    ConfigFormatter          // FormatToINI()
}

// Service implements semua interfaces
type Service struct { ... }

func (s *Service) NewWizard() *wizard.Runner { ... }
func (s *Service) SaveProfile() error { ... }
// ... dst
```

**Benefits:**
- Easy testing (mock interfaces)
- Reduced coupling
- Clear contracts
- Flexibility untuk swap implementations

## Security Model

### Encryption Flow

```
┌──────────────┐
│  Plain Text  │
│  Password    │
└──────┬───────┘
       │
       ▼
┌────────────────────────────┐
│  AES-256-GCM Encryption    │
│  (via password helper)     │
└──────┬─────────────────────┘
       │
       ▼
┌──────────────┐
│  Base64      │
│  Encoded     │
└──────┬───────┘
       │
       ▼
┌──────────────────────┐
│  Store in INI file   │
│  password = <base64> │
└──────┬───────────────┘
       │
       ▼
┌────────────────────────────┐
│  Encrypt entire file       │
│  with profile key          │
│  → .cnf.enc                │
└────────────────────────────┘
```

### Key Resolution Chain

```
Profile Encryption Key:
1. CLI flag: --profile-key
2. Env var: SFDB_SOURCE_PROFILE_KEY
3. Interactive prompt (if terminal)
4. Error: key not available

Backup Encryption Key:
1. CLI flag: --encryption-key
2. Env var: SFDB_BACKUP_ENCRYPTION_KEY
3. Interactive prompt
4. Error: key required
```

## Error Handling Strategy

### Error Categories

```go
// User errors (recoverable)
validation.ErrUserCancelled      // User pressed Ctrl+C
profileerrors.ErrProfileNotFound  // Profile doesn't exist
profileerrors.ErrInvalidInput     // Bad input format

// Business errors
profileerrors.ErrProfileExists    // Name conflict
profileerrors.ErrConnectionFailed // DB connection failed

// System errors
profileerrors.ErrFileNotWritable  // Permission denied
profileerrors.ErrEncryptionFailed // Crypto error
```

### Error Propagation

```
Wizard/Validator (detect error)
         │
         ▼ return error
    Executor (add context)
         │
         ▼ fmt.Errorf("...: %w", err)
    Service (handle/log)
         │
         ▼ 
    CLI (display to user)
```

## Testing Strategy

### Unit Tests

```
✓ Model: State management logic
✓ Validation: Input validators
✓ Helpers: Utility functions
✓ Parser: INI parsing
✓ Connection: Preflight checks
```

### Integration Tests

```
✓ Executor: Full CRUD operations
✓ Wizard: Flow completion
✓ Import: Bulk processing
```

### Test Doubles

```go
// Mock interfaces untuk testing
type MockProfileOps struct {
    SaveProfileFunc      func() error
    CheckNameUniqueFunc  func() error
    // ... dst
}

// Usage
executor := executor.New(logger, cfg, configDir, state, &MockProfileOps{
    SaveProfileFunc: func() error {
        return nil // success
    },
})
```

## Performance Considerations

### Memory Efficiency

- Stream processing untuk large imports (XLSX)
- Connection pooling untuk bulk operations
- Profile caching untuk repeated access (di loader)

### Startup Performance

- Lazy initialization (wizard hanya di-create saat interactive)
- Config validation di-defer sampai execution
- SSH tunnel hanya start saat benar-benar dipakai

## Extension Points

### Adding New Operations

1. Add method ke Service:
   ```go
   func (s *Service) NewOperation() error
   ```

2. Add executor method:
   ```go
   func (e *Executor) ExecuteNewOp() error
   ```

3. Update router di ExecuteProfileCommand:
   ```go
   case consts.ProfileModeNewOp:
       return s.NewOperation()
   ```

### Adding New Import Sources

Implement reader interface di `helpers/reader`:

```go
func ReadFromNewSource(url string) ([][]string, error) {
    // Parse new source
    // Return CSV-like data
}
```

## File Organization

```
profile/
├── doc.go                    # Package documentation
├── ARCHITECTURE.md           # This file
├── service.go               # Service layer
├── service_methods.go       # Service methods
├── executor_ops.go          # ProfileOps implementation
├── command.go               # Command setup
├── setup.go                 # Initialization helpers
│
├── connection/              # Database connection
│   ├── doc.go
│   ├── connector.go         # Main connection logic
│   ├── preflight.go         # Pre-connection checks
│   ├── test.go              # Connection testing
│   └── errors.go            # Connection errors
│
├── display/                 # Output formatting
│   ├── displayer.go
│   ├── show_formatter.go
│   ├── summary.go
│   └── import_*.go
│
├── errors/                  # Custom errors
│   └── errors.go
│
├── executor/                # Operation execution
│   ├── doc.go
│   ├── executor.go          # Base executor
│   ├── create.go
│   ├── edit.go
│   ├── show.go
│   ├── delete.go
│   ├── clone.go
│   ├── import.go
│   ├── save.go
│   └── retry_helper.go
│
├── helpers/                 # Utilities
│   ├── doc.go
│   ├── common/              # Common utils
│   ├── importer/            # Import helpers
│   ├── keys/                # Key resolution
│   ├── loader/              # Profile loading
│   ├── parser/              # INI parsing
│   ├── password/            # Encryption
│   ├── paths/               # Path resolution
│   ├── reader/              # Data readers
│   ├── resolver/            # Conflict resolution
│   ├── selection/           # Interactive selection
│   └── snapshot/            # Snapshot management
│
├── merger/                  # Profile merging
│   └── merger.go
│
├── model/                   # Domain models
│   ├── types_profile.go     # Profile types
│   └── types_import.go      # Import types
│
├── process/                 # Background processes
│   ├── auth.go
│   ├── tunnel.go
│   ├── forwarding.go
│   └── known_hosts.go
│
├── validation/              # Validators
│   ├── database.go
│   ├── input.go
│   ├── profile.go
│   ├── ssh.go
│   ├── name_uniqueness.go
│   └── import_validation.go
│
└── wizard/                  # Interactive flows
    ├── doc.go
    ├── runner.go            # Main runner
    ├── create_flow.go
    ├── edit_flow.go
    ├── clone_flow.go
    ├── import_*.go
    ├── prompts.go
    ├── validators.go
    └── helpers.go
```

## Future Improvements

### Potential Enhancements

- [ ] Profile templates (pre-defined configs)
- [ ] Profile groups/tags
- [ ] Profile validation in background (health checks)
- [ ] Profile auto-discovery (scan network)
- [ ] Profile sync across machines
- [ ] Web UI untuk profile management
- [ ] Profile versioning (git-like)
- [ ] Multi-profile operations (batch edit)

### Technical Debt

- [ ] Add more integration tests
- [ ] Refactor wizard prompts (too long)
- [ ] Improve error messages (more actionable)
- [ ] Add telemetry/metrics
- [ ] Connection pooling optimization
