# sfDBTools AI Coding Assistant Guide

## Project Overview
sfDBTools is a CLI tool for MariaDB/MySQL database management, focusing on backup, scanning, cleanup, and profile management. Built with Go 1.25 using Cobra for CLI.

## Architecture Patterns

### Service-Based Architecture
All features follow a consistent service pattern with dependency injection:
- **No global state** - services receive dependencies via constructors
- **BaseService embedding** - all services embed `pkg/servicehelper/BaseService` for graceful shutdown
- **Dependency injection** - `types.Dependencies` struct contains `Config` and `Logger`, injected via `main.go`
- Services live in `internal/{backup,cleanup,dbscan,profile}` with command wrappers in `cmd/cmd_{feature}/`

Example service initialization:
```go
// In command.go
svc := NewBackupService(deps.Logger, deps.Config, &parsedOpts)
// In service.go
type Service struct {
    servicehelper.BaseService
    Log    applog.Logger
    Config *appconfig.Config
}
```

### Strategy + Factory Pattern (Backup Modes)
Backup uses Strategy pattern with Factory (`internal/backup/modes/`):
- **ModeExecutor interface** - defines `Execute(ctx, databases) BackupResult`
- **Factory** - `modes.GetExecutor(mode, svc)` returns appropriate executor
- **Implementations** - `CombinedExecutor` (single file), `IterativeExecutor` (per-DB files)
- Modes: "all"/"combined" → single file, "single"/"primary"/"secondary"/"separated" → iterative

### Layer Structure
```
cmd/{feature}/          # Cobra commands, flag definitions
  ↓
internal/{feature}/command.go   # Entry point, dependency injection
  ↓
internal/{feature}/service.go   # Business logic, state management
  ↓
pkg/{helpers}/         # Reusable utilities
```

## Critical Conventions

### File Headers
Every Go file includes structured header:
```go
// File : path/to/file.go
// Deskripsi : Purpose description
// Author : Hadiyatna Muflihun
// Tanggal : YYYY-MM-DD
// Last Modified : YYYY-MM-DD
```

### Configuration Loading
- Config loaded from YAML via `internal/appconfig/LoadConfigFromEnv()`
- Environment variables prefixed with `SFDB_` (see `pkg/consts/consts_env.go`)
- Quiet mode (`SFDB_QUIET=1`) suppresses UI for piping
- Completion commands skip config/logging for clean output

### Database Connections
- Use `pkg/database.Client` for connections (wraps `sql.DB` with retry logic)
- Profile-based connections via encrypted `.cnf` files in `config/`
- Decrypt profiles using `pkg/encrypt` with keys from env vars (`SFDB_SOURCE_PROFILE_KEY`)
- Always defer `client.Close()` after setup

### Logging
- Custom logger via `internal/applog` (wraps logrus)
- Error logging to separate file via `pkg/errorlog.ErrorLogger`
- Quiet mode routes logs to stderr, stdout reserved for data

### Compression & Encryption
- Backup writer pipeline: `mysqldump → compress → encrypt → file`
- Compression: gzip/pgzip/xz via `pkg/compress` (pgzip default)
- Encryption: AES-256 via `pkg/encrypt` with passphrase from env/prompt
- Extensions: `.sql.gz.enc`, `.sql.xz.enc`, etc.

## Key Files Reference

### Entry Points
- `main.go` - initializes Config/Logger/Dependencies, calls `cmd.Execute(deps)`
- `cmd/cmd_root.go` - root Cobra command, `PersistentPreRunE` validates dependencies

### Core Services
- `internal/backup/service.go` - backup orchestration, implements `BackupService` interface
- `internal/backup/modes/{interface,factory,combined,iterative}.go` - backup strategy implementations
- `internal/cleanup/cleanup_main.go` - general-purpose file cleanup service (used by backup post-cleanup)
- `internal/dbscan/service.go` - database scanning with foreground/background modes

### Type Definitions
- `internal/types/types_global.go` - global `Dependencies` struct
- `internal/types/types_{feature}.go` - feature-specific types
- `internal/types/types_{feature}/` - complex nested types (e.g., `types_backup/`)

### Utilities
- `pkg/database/` - connection management, filtering, GTID capture
- `pkg/servicehelper/servicehelper_base.go` - BaseService for graceful shutdown
- `pkg/ui/` - terminal UI (headers, spinners, tables)
- `pkg/flags/` - shared flag parsing logic

## Development Workflow

### Build & Run
```bash
# Build and run (installs to /usr/bin/sfDBTools)
scripts/build_run.sh -- backup all --help

# Skip run (build only)
scripts/build_run.sh --skip-run

# Build with race detector
scripts/build_run.sh --race -- profile show
```

### Adding New Features
1. Create service in `internal/{feature}/{feature}_main.go` with `Service` struct embedding `BaseService`
2. Create command handler in `internal/{feature}/command.go` with `Execute{Feature}()` function
3. Create Cobra commands in `cmd/cmd_{feature}/` with flag definitions
4. Register commands in `cmd/cmd_root.go` via `rootCmd.AddCommand()`
5. Add types to `internal/types/types_{feature}.go`

### Adding New Backup Modes
1. Create executor in `internal/backup/modes/{mode}.go` implementing `ModeExecutor` interface
2. Register in `modes/factory.go` switch statement
3. Add mode config to `internal/backup/mode_config.go`

## Testing Notes
- No test files currently exist in codebase
- Services designed for testability via dependency injection
- Use `context.Context` for cancellation throughout

## Important Gotchas
- **Completion commands** - always check `cmd.Name() == "completion"` to skip heavy initialization
- **Background mode** - dbscan supports daemon mode via `ENV_DAEMON_MODE=1` flag (see `pkg/dbscanhelper/`)
- **Excluded databases** - system DBs filtered in `pkg/database/database_filter.go`
- **GTID capture** - replication info saved to metadata when enabled (`backup.replication.capture_gtid`)
- **User grants export** - mysqldump doesn't backup users, manually exported via `internal/backup/metadata/user.go`

## Documentation
Feature-specific docs in `internal/{feature}/`:
- `internal/backup/DEVELOPER_MANUAL.md` - comprehensive backup architecture guide
- `internal/cleanup/{README,USAGE_GUIDE,REFACTORING_NOTES}.md` - cleanup module docs

## Code Style
- Comments in Indonesian for business logic, English for technical
- Use `err != nil` checks immediately after calls
- Prefer explicit returns over naked returns
- Use `fmt.Errorf("msg: %w", err)` for error wrapping
