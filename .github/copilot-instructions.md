# sfDBTools Development Guide

## Project Overview
sfDBTools is a production-grade MariaDB/MySQL database backup and management utility written in Go. It provides secure database operations including backup, scanning, cleanup, and profile management with encryption support.

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

### Service Layer Pattern
Each feature uses a service struct pattern (see `internal/backup/backup_main.go`, `internal/dbscan/dbscan_main.go`):
```go
type Service struct {
    Config *appconfig.Config
    Log    applog.Logger
    // Feature-specific fields
}
```
- Services are instantiated with `New*Service()` constructors
- Type switching in constructors handles different option types
- Services encapsulate business logic separate from CLI layer

### Configuration System
- Primary config: `config/sfDBTools_config.yaml` loaded via `appconfig.LoadConfigFromEnv()`
- Structure defined in `internal/appconfig/appconfig_structs.go` with nested YAML tags
- Database profiles stored as encrypted `.cnf.enc` files in `config/database_profile/`
- Environment variables prefixed with `SFDB_` (see `pkg/consts/consts_env.go`)

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
1. Profile selection (`internal/profileselect/profileselect_selector.go`)
2. Database filtering (include/exclude lists from config or flags)
3. Connection via `pkg/database/database_connection.go` (`ConnectToSourceDatabase`)
4. Backup execution with `mysqldump` process management (`internal/backup/backup_mysqldump.go`)
5. Compression (gzip/zstd/xz/pgzip) via streaming writers (`pkg/compress/`)
6. Optional encryption using `encrypt.NewEncryptWriter()` (`pkg/encrypt/encrypt_writer.go`)
7. Verification and cleanup if configured

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
- `internal/`: Application-specific logic (backup, dbscan, profile, etc.)
- `internal/types/`: Shared type definitions (options, results, errors)
- `pkg/`: Reusable utility packages (database, encrypt, ui, validation, etc.)
- `cmd/`: Cobra command definitions organized by feature in subdirectories
- `config/`: Runtime configuration files (YAML config, encrypted profiles, include/exclude lists)

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
