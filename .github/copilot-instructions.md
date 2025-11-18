# sfDBTools Development Guide

## Project Overview
sfDBTools is a production-grade MariaDB/MySQL database backup and management utility written in Go. It provides secure database operations including backup, scanning, cleanup, restore, and profile management with encryption support.

## Architecture

### Dependency Injection Pattern
The application uses a global dependency injection model via `types.Dependencies`:
- `types.Deps` is a global variable injected in `main.go` and passed to `cmd.Execute(deps)`
- All commands access config and logger through `types.Deps.Config` and `types.Deps.Logger`
- Check `types.Deps != nil` before accessing in `PersistentPreRunE` hooks

### Command Structure (Cobra-based)
- Root command in `cmd/cmd_root.go` registers all subcommands in `init()`
- Each feature has a main command file (e.g., `cmd_backup_main.go`) that acts as a parent
- Subcommands are added to feature commands (e.g., `backup combined`, `backup separated`)
- Flag definitions are centralized in `pkg/flags/flags_*.go` files

### Service Layer Pattern with Helper Composition
Each feature uses a service struct pattern with embedded helper services:
```go
type Service struct {
    servicehelper.BaseService // Embed for mutex & cancel func
    Config *appconfig.Config
    Log    applog.Logger
    // Feature-specific fields
}
```
- Services are instantiated with `New*Service()` constructors
- Type switching in constructors handles different option types
- Services encapsulate business logic separate from CLI layer
- **BaseService** provides: mutex locking (`WithLock()`), graceful shutdown (`SetCancelFunc()`)
- Restore service uses `servicehelper.TrackProgress()` for progress tracking with defer pattern

### Configuration System
- Primary config: `config/sfDBTools_config.yaml` loaded via `appconfig.LoadConfigFromEnv()`
- Structure defined in `internal/appconfig/appconfig_structs.go` with nested YAML tags
- Database profiles stored as encrypted `.cnf.enc` files in `config/database_profile/`
- Environment variables prefixed with `SFDB_` (see `pkg/consts/consts_env.go`)

### Helper Package Architecture (Modular Patterns)
The codebase follows a highly modular pattern with reusable helpers in `pkg/`:

**Profile & Connection Helpers** (`pkg/profilehelper/`):
- `LoadSourceProfile()` - Unified profile loading with interactive selector
- `ConnectWithProfile()` / `ConnectWithTargetProfile()` - Database connection setup
- Eliminates 94 lines of duplication across backup/restore/dbscan packages

**Service Helpers** (`pkg/servicehelper/`):
- `BaseService` - Embed for mutex (`WithLock()`) and cancel func management
- `TrackProgress()` - Progress tracking with defer pattern for restore operations
- Eliminates 21 lines of state management boilerplate

**Time Tracking** (`pkg/helper/`):
- `NewTimer()` / `Elapsed()` - Unified timer pattern across all features
- Replaces `time.Now()` + `time.Since()` pattern in 13 locations

**File Operations** (`pkg/fsops/`):
- `FileExists()` / `DirExists()` / `PathExists()` - Simplified existence checks
- `FileExistsWithInfo()` - Returns `os.FileInfo` when file exists
- Replaces verbose `os.Stat()` pattern in 7 locations

**Database Filtering** (`pkg/database/`):
- `FilterFromBackupOptions()` / `FilterFromScanOptions()` - Generate filter from options
- Eliminates 45 lines of filter creation duplication

**UI Helpers** (`pkg/ui/`):
- `SpinnerWithElapsed()` - Spinner with elapsed time display
- Eliminates 52 lines of spinner boilerplate across backup/restore

## Critical Workflows

### Building and Running
Use the build script, not direct `go build`:
```bash
scripts/build_run.sh                    # Build and run
scripts/build_run.sh -- backup combined # Build and run with args
scripts/build_run.sh --skip-run         # Build only
scripts/build_run.sh --race -- --help   # Build with race detector
```
Binary output: `bin/sfdbtools`

### Database Profile Management
Profiles are encrypted MariaDB config files stored in `config/database_profile/`:
- Create: `sfdbtools profile create` (interactive wizard)
- Files have `.cnf.enc` extension and contain encrypted credentials
- Encryption uses AES-256-GCM compatible with OpenSSL (see `pkg/encrypt/encrypt_aes.go`)
- Profile selection via `profileselect` package provides interactive menu
- Decryption key from `SFDB_ENCRYPTION_KEY` env var or interactive prompt

### Backup Flow
1. Profile selection (`profilehelper.LoadSourceProfile()`)
2. Database filtering (`database.FilterFromBackupOptions()`)
3. Connection via `profilehelper.ConnectWithProfile()`
4. Backup execution with `mysqldump` process management (`internal/backup/backup_mysqldump.go`)
5. Compression (gzip/zstd/xz/pgzip) via streaming writers (`pkg/compress/`)
6. Optional encryption using `encrypt.NewEncryptWriter()` (`pkg/encrypt/encrypt_writer.go`)
7. Graceful shutdown handling via `BaseService.SetCancelFunc()` - cleans up partial files on interrupt

### Restore Flow
1. Profile selection and connection via `profilehelper.ConnectWithTargetProfile()`
2. Progress tracking using `defer servicehelper.TrackProgress(service)()`
3. File validation and metadata loading (`restore_verify.go`)
4. Reader pipeline setup: decrypt → decompress (`restore_reader.go`)
5. Pre-backup execution (safety backup before restore) via backup service integration
6. Database preparation: check exists → drop (if flag) → create (`restore_database.go`)
7. MySQL restore execution with streaming from reader pipeline
8. Result building using builder pattern (`restore_result.go`)

## Project-Specific Conventions

### File Headers
All Go files include standard headers in Indonesian:
```go
// File : <path>
// Deskripsi : <description>
// Author : Hadiyatna Muflihun
// Tanggal : <date>
// Last Modified : <date>
```

### Error Handling
- Sentinel errors in `internal/types/types_error.go` (e.g., `ErrUserCancelled`)
- `validation.HandleInputError()` for user input cancellation handling
- Services log errors via injected logger before returning
- Main commands exit with `os.Exit(1)` on fatal errors

### UI Patterns
- `ui.Headers()` clears screen and shows app header (suppress in quiet mode)
- Check `SFDB_QUIET` env var to skip UI banners for pipeline-friendly output
- Quiet mode routes logs to stderr, leaving stdout clean for data
- Progress indicators use `github.com/briandowns/spinner`
- Tables rendered with `github.com/olekukonko/tablewriter`

### Package Organization
- `internal/`: Application-specific logic (backup, dbscan, profile, restore, cleanup)
- `internal/types/`: Shared type definitions (options, results, errors)
- `pkg/`: Reusable utility packages organized by concern:
  - `profilehelper/` - Profile loading and connection helpers
  - `servicehelper/` - Base service functionality (mutex, progress tracking)
  - `helper/` - General utilities (timer, env vars, path resolution)
  - `fsops/` - File system operations (existence checks, read/write)
  - `database/` - Database operations (client wrapper, filtering, queries)
  - `encrypt/` - AES-256-GCM encryption/decryption
  - `compress/` - Multi-format compression (gzip, zstd, xz, pgzip)
  - `ui/` - Terminal UI components (headers, spinners, tables)
  - `validation/` - Input validation and error handling
  - `flags/` - Cobra flag definitions per feature
- `cmd/`: Cobra command definitions organized by feature in subdirectories
- `config/`: Runtime configuration files (YAML config, encrypted profiles, include/exclude lists)

### Refactoring Philosophy
When identifying duplicate patterns:
1. Extract to helper package if used in 3+ locations
2. Use embedding for common service behaviors (BaseService pattern)
3. Prefer defer cleanup patterns for resource management
4. Timer pattern: `timer := helper.NewTimer()` ... `duration := timer.Elapsed()`
5. Progress tracking: `defer servicehelper.TrackProgress(service)()`
6. File checks: Use `fsops.FileExists()` instead of `os.Stat()` patterns

### Database Operations
- Use `database.Client` wrapper around `sql.DB` from `pkg/database/`
- Client has connection pooling, retry logic, and context-based timeout handling
- Database credentials come from profiles or environment variables
- System databases filtered via `pkg/database/database_filter.go`
- Query helpers in `pkg/database/database_dbscan_query.go` for metadata collection

### Flag Management
- Flags defined in `pkg/flags/flags_*.go` organized by feature
- Common flag patterns: `AddProfileFlags`, `AddBackupFlags`, `AddCompressionFlags`
- Bind flags to option structs passed by reference
- Default values set in option struct initialization before adding flags

### Encryption Specifics
- OpenSSL-compatible AES-256-GCM with PBKDF2 key derivation (100,000 iterations)
- Encrypted files start with "Salted__" header for OpenSSL compatibility
- Streaming encryption via `io.WriteCloser` for memory-efficient large file handling
- Key sources: env var, flag, interactive prompt via `encrypt.PromptEncryptionKey()`

### Background Process Handling
- Graceful shutdown via context cancellation (`Service.SetCancelFunc()`)
- Track in-progress backups to clean up partial files on interrupt
- Mutex-protected state in service structs for concurrent access safety

## Testing Considerations
- Test database connectivity with `db-scan all-local` before running backups
- Dry-run mode available for backup commands (`--dry-run` flag)
- Validate profiles with `profile show` command
- Use `--show-options` flag to preview backup configuration before execution

## Common Patterns
- Interactive prompts use `github.com/AlecAivazis/survey/v2`
- File path validation centralized in `pkg/validation/`
- Helper functions for env vars in `pkg/helper/` (e.g., `GetEnvOrDefault`)
- Date/time formatting respects config locale settings
- Compression type detection via file extension helpers in `pkg/helper/`

## Additional note
- All log messages and user-facing text are in Indonesian to maintain consistency across the application.
- Follow existing code style and conventions for new features or modifications.
- Ensure proper error handling and logging in all new code paths.
- No need to create documentation if not requested by the user.