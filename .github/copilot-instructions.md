# sfdbtools Copilot Instructions

## Project Overview
**sfdbtools** is a production-grade CLI utility built to simplify and standardize day-to-day work for **Database Administrators (DBAs)** managing **MariaDB/MySQL and Microsoft SQL Server** environments. It supports operational workflows across **SaaS**, **hosting**, and **on‑premise** deployments—focused on repeatability, automation, and reducing human error.

Built in **Go (Golang)** and aligned with **Clean Architecture**, sfdbtools prioritizes **data safety** and **security**, including **AES‑256** encryption for sensitive assets (such as backups and connection profiles), while keeping execution reliable and suitable for production operations.

**Core Features**: Backup (multi-mode with encryption/compression), Restore (with companion database handling), Profile management, DB scanning, Cleanup, and Crypto utilities.

## Go Design Philosophy (CRITICAL)
Follow these specific principles when writing or refactoring code for this project:

- **DRY vs. Dependency (The "Go Way")**:
  - **Principle**: "A little copying is better than a little dependency."
  - **Guideline**: Do not create a giant shared library just to satisfy DRY. It is better to duplicate a few lines of simple logic in two places than to couple them to a shared function that creates complex dependencies.
  - **Goal**: Code independence is prioritized over strict deduplication.

- **KISS (Keep It Simple, Stupid)**:
  - **Principle**: Code should be "boring" and explicit.
  - **Guideline**: Avoid complex Generics (unless absolutely necessary), Reflection, or "clever" one-liners. If you need to open 5 files to understand one function, it is too complex.
  - **Constraint**: Go does not have ternary operators; do not try to emulate them with complex logic.

- **YAGNI (You Ain't Gonna Need It)**:
  - **Principle**: Do not design for a hypothetical future.
  - **Guideline**:
    - **Do NOT** create an Interface if there is currently only one implementation.
    - **Do NOT** create deep folder structures for "future expansion."
    - Refactoring in Go is easy; build for *now*.

- **SOLID Adaptation**:
  - **SRP (Single Responsibility)**: A package must have one clear purpose (e.g., `net/http`).
  - **ISP (Interface Segregation)**: **Crucial**. Keep interfaces tiny. An interface with 1 method (like `io.Reader`) is far better than one with 10.
  - **DIP (Dependency Inversion)**: Functions should accept interfaces but return concrete structs (generally).

## Architecture & Structural Patterns
- **Clean Architecture Layers**:
  - `cmd/`: Entry points (Cobra commands). Keep thin—only flag parsing and command setup.
  - `internal/`: Core business logic (Backup, Restore, Profile, DBScan, Cleanup, Crypto).
  - `pkg/`: Reusable, domain-agnostic libraries (encryption, compression, database, validation).

- **Dependency Injection**:
  - Global dependencies (`Config`, `Logger`) injected via `types.Dependencies`.
  - Flow: `main.go` → `cmd.Execute(deps)` → `types.Deps` (global) → `PersistentPreRunE` validation.
  - Each module creates its own Service with injected dependencies (e.g., `backup.NewBackupService(logs, cfg, opts)`).

- **Strategy Pattern (Backup/Restore Modes)**:
  - **Location**: `internal/backup/modes/` and `internal/restore/modes/`
  - **Pattern**: Factory (`GetExecutor(mode)`) returns `ModeExecutor` interface implementations.
  - **Implementations**:
    - **Backup**: `CombinedExecutor` (all/combined), `IterativeExecutor` (single/primary/secondary/separated)
    - **Restore**: `SingleMode`, `PrimaryMode`, `AllMode`, `SelectionMode`
  - **Key Interface** (`internal/backup/modes/interface.go`): `ModeExecutor.Execute(ctx, databases) Result`
  - Add new modes by implementing interface + updating factory.

- **Service Layer Pattern**:
  - Each feature has a `Service` struct (e.g., `backup.Service`, `restore.Service`).
  - Services embed `servicehelper.BaseService` for common operations (locking, context management).
  - Services own feature-specific state and orchestrate business logic.

## Build & Development Workflows
- **Build Script**: ALWAYS use the helper script at `./scripts/build_run.sh`.
  - **Build & Run**: `./scripts/build_run.sh -- [args]`
  - **Build Only**: `./scripts/build_run.sh --skip-run`
  - **With Race Detector**: `./scripts/build_run.sh --race -- [args]`
  - **Examples**:
    - `./scripts/build_run.sh -- backup single --help`
    - `./scripts/build_run.sh -- profile show --file config/my.cnf.enc`
  - **Output**: Binary compiled to `/usr/bin/sfdbtools`

- **Environment Variables**:
  - Defined in `pkg/consts/consts_env.go`
  - **Key Variables**:
    - `SFDB_QUIET=1`: Suppresses banners/spinners for pipeline usage (logs to stderr)
    - `SFDB_SOURCE_PROFILE_KEY`: Encryption key for source profile
    - `SFDB_TARGET_PROFILE_KEY`: Encryption key for target profile
    - `SFDB_BACKUP_ENCRYPTION_KEY`: Key for backup file encryption

## Data Flow & Streaming Architecture
- **Streaming Pipeline Philosophy**: NO large memory buffers—use `io.Reader`/`io.Writer` chains.
- **Backup Pipeline** (see `internal/backup/execution_helpers.go`):
  ```
  mysqldump → compress.Writer → encrypt.Writer → file.Writer
  ```
- **Restore Pipeline**:
  ```
  file.Reader → decrypt.Reader → decompress.Reader → mysql client stdin
  ```
- **Key Packages**:
  - `pkg/compress/`: Compression writers (zstd, gzip, pgzip, xz, zlib)
  - `pkg/encrypt/`: Streaming AES-256-GCM encryption (`EncryptingWriter`, `DecryptingReader`)
  - `pkg/backuphelper/mysqldump.go`: Executes `mysqldump` with streaming stdout

## Domain-Specific Patterns

### Companion Database Handling
- **Context**: Production DBs have "companion" databases (e.g., `dbsf_nbc_client` + `dbsf_nbc_client_dmart`)
- **Suffixes**: `_dmart`, `_temp`, `_archive` are companions
- **Restore Logic** (`internal/restore/companion_helpers.go`):
  - Automatically detects and restores companions when restoring primary
  - Controlled by flags: `--include-dmart`, `--auto-detect-dmart`, `--companion-file`
- **Safety**: Primary databases (pattern: `dbsf_nbc_*` or `dbsf_biznet_*` WITHOUT suffix) cannot be restored if they already exist.

### Validation & Safety (Fail-Fast)
- **Restore Safety** (`internal/restore/validation_helpers.go`, `internal/restore/helpers/validation.go`):
  - **Rule**: Cannot restore to existing primary database—prevents accidental data loss.
  - **Validation**: `ValidateNotPrimaryDatabase()` checks DB existence and naming pattern.
  - **Application Password**: `restore primary` requires app password (`consts.ENV_PASSWORD_APP`)
- **Path Validation** (`pkg/validation/validation_backup_dir.go`):
  - Validates directory patterns against path traversal (`..`), absolute paths
  - Token validation for dynamic path generation (`{database}`, `{year}`, `{timestamp}`, etc.)

### Path Pattern System
- **Dynamic Paths** (`pkg/helper/helper_path.go`):
  - **Tokens**: `{database}`, `{hostname}`, `{year}`, `{month}`, `{day}`, `{hour}`, `{minute}`, `{second}`, `{timestamp}`
  - **Example Pattern**: `{year}/{month}/{database}_{timestamp}_{hostname}.sql`
  - **Replacer**: `PathPatternReplacer.ReplacePattern(pattern, excludeHostname)`
  - **Usage**: Backup file naming and directory structure generation

## Coding Conventions
- **File Naming**:
  - Explicit naming: `pkg/helper/encrypt.go` NOT `pkg/helper/helper_encrypt.go`
  - Types split by domain: `internal/types/types_backup.go`, `types_restore.go`, etc.

- **Logging**:
  - Use `sfdbtools/internal/applog.Logger` interface
  - Respect `consts.ENV_QUIET` for pipeline usage (routes logs to stderr)
  - **Error Logging**: Use `pkg/errorlog.ErrorLogger` for feature-specific error logs

- **Error Handling**:
  - **Fail-Fast**: Validate early (connections, file existence, safety rules) before heavy operations
  - **Context Propagation**: Pass `context.Context` for cancellation support
  - **Wrapped Errors**: Use `fmt.Errorf("descriptive msg: %w", err)` for error chains

- **Concurrency & Safety**:
  - Services use `servicehelper.BaseService.WithLock()` for state mutations
  - Signal handling for graceful shutdown (CTRL+C cleanup)
  - Context cancellation for long-running operations

## Key Dependencies (go.mod)
- **CLI**: `github.com/spf13/cobra` (commands), `github.com/AlecAivazis/survey/v2` (interactive prompts)
- **Database**: `github.com/go-sql-driver/mysql`
- **Compression**: `github.com/klauspost/compress` (zstd), `github.com/klauspost/pgzip`, `github.com/ulikunitz/xz`
- **Logging**: `github.com/sirupsen/logrus`
- **Display**: `github.com/olekukonko/tablewriter`, `github.com/dustin/go-humanize`

## Testing & Debugging
- **Manual Testing**: Use `./scripts/build_run.sh -- [command]`
- **Race Detection**: `./scripts/build_run.sh --race -- [command]`
- **Quiet Mode Testing**: `SFDB_QUIET=1 ./scripts/build_run.sh -- [command]`

## Consistency & Maintenance
- Regularly review code for adherence to Go design principles.
- Refactor services and modes to maintain clarity and simplicity.
- Update documentation and comments to reflect architectural decisions and patterns.
- Ensure all new features follow established patterns for ease of maintenance.
- Keep dependencies minimal and relevant to avoid bloat.
- Encourage code reviews focusing on design philosophy adherence and code quality.
- Use indonesian language for comments and documentation where applicable.
- Update last modified date on each modification on header comments of each file.