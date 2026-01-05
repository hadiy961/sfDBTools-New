# Contributing Guide: Extending the Backup Module (`internal/backup`)

This document is for contributors who want to add or modify features in the backup subsystem of **sfDBTools**.

Scope: **Go code under `internal/backup/**` and its integration points (CLI commands, flags, parsing, default values, types, and shared packages used by backup).**

---

## 1) High-Level Architecture (Mental Model)

The backup feature is designed as a pipeline with clear responsibility boundaries:

1. **CLI command layer** (`cmd/backup/*`)
   - Defines Cobra commands (`db-backup all|filter|single|primary|secondary`).
   - Attaches flags (via `internal/cli/flags/*`).
   - Calls the unified executor `internal/backup.ExecuteBackup(...)`.

2. **Options parsing** (`internal/cli/parsing/backup.go`)
   - Reads flags/env and produces a `types_backup.BackupDBOptions`.
   - Performs **fail-fast validation**, especially in non-interactive mode via `--quiet`.

3. **Service orchestration** (`internal/backup/*.go`)
   - `backup.Service` wires dependencies and provides “bridges” to subpackages.
   - `Service.PrepareBackupSession()` delegates to `internal/backup/setup`.
   - `Service.ExecuteBackup()` chooses a mode executor from `internal/backup/modes`.

4. **Mode strategies** (`internal/backup/modes/*`)
   - Strategy/factory pattern.
   - `CombinedExecutor`: all databases in one file.
   - `IterativeExecutor`: per-database loop (separated/single/primary/secondary).

5. **Execution engine** (`internal/backup/execution/*`)
   - Builds mysqldump args.
   - Runs mysqldump through writer.
   - Handles retries for common failures.
   - Builds metadata objects.

6. **Writer (streaming pipeline)** (`internal/backup/writer/*`)
   - Runs `mysqldump` and streams stdout to a file.
   - Optional compression and encryption using streaming writers.

7. **Metadata / user grants / GTID** (`internal/backup/metadata`, `internal/backup/grants`, `internal/backup/gtid`)
   - Writes `.meta.json` manifest (atomic write).
   - Exports `*_users.sql` (optional).
   - Captures GTID (optional).

---

## 2) Repository Conventions You Must Follow

### Keep the backup pipeline streaming
Backup should not buffer large dumps in memory. The expected flow is:

`mysqldump stdout → (optional) compression → (optional) encryption → file`

If you add a feature that transforms backup output, it should be done as an `io.Reader`/`io.Writer` streaming layer.

### Fail fast in non-interactive mode
Non-interactive mode is used for automation and must not prompt.

- `internal/cli/parsing/backup.go` enforces required inputs.
- `internal/backup/setup` enforces additional runtime validations.

If you add a required input (new flag), also add:

- non-interactive validation in parsing,
- and runtime validation in setup if appropriate.

### Prefer small, explicit code
Avoid over-abstracting. Keep code local to the module unless it’s clearly reusable and domain-agnostic.

---

## 3) How to Build and Run Backup Locally

Project build/run uses the helper script:

- Build only: `./scripts/build_run.sh --skip-run`
- Run a backup help: `./scripts/build_run.sh -- db-backup all --help`
- Run filter help: `./scripts/build_run.sh -- db-backup filter --help`

Quiet mode is useful for automation:

- `SFDB_QUIET=1 ./scripts/build_run.sh -- db-backup all --help`

---

## 4) Where to Change What (Common Feature Types)

### A) Add a new flag / option

Typical wiring steps:

1. **Define the option field**
   - Add/update structs in `internal/types/types_backup/options.go`.

2. **Expose the flag**
   - Add it in `internal/cli/flags/backup.go` (or `internal/cli/flags/shared.go` if it is truly shared).
   - Attach the flag in the command init (example: `cmd/backup/all.go` calls `flags.AddBackupAllFlags`).

3. **Parse the flag**
   - Implement parsing in `internal/cli/parsing/backup.go`.
   - If it can be set by env, use the same helper patterns as existing fields.

4. **Validate**
   - Add fail-fast validation (especially for non-interactive via `--quiet`).

5. **Use the option in the correct layer**
   - Setup-related? Put logic in `internal/backup/setup/*`.
   - Affects database selection/filtering? Put logic in `internal/backup/selection/*` or `pkg/database.FilterDatabases`.
   - Affects mysqldump args? Put logic in `internal/backup/execution/args.go`.
   - Affects output writing? Put logic in `internal/backup/writer/engine.go`.
   - Affects metadata only? Put logic in `internal/backup/execution/builder.go` and/or `internal/backup/metadata/*`.

### B) Add a brand-new backup mode

A “mode” is a high-level execution strategy (single file vs multi file vs custom business behavior).

Steps:

1. **Add/confirm a mode constant**
   - Mode strings are defined under `pkg/consts` (used across the app).

2. **Add CLI command (if needed)**
   - Create a new command under `cmd/backup/` and register it in `cmd/backup/main.go`.
   - Attach flags using `internal/cli/flags/backup.go`.

3. **Add parsing & default values**
   - Ensure `internal/cli/defaults.DefaultBackupOptions(mode)` supports your mode.
   - Update `internal/cli/parsing/backup.go` to handle any mode-specific flags.

4. **Add an executor**
   - Implement `modes.ModeExecutor` under `internal/backup/modes/`.
   - Keep the executor focused on orchestration; push heavy work to `execution.Engine`.

5. **Register it in the factory**
   - Update `internal/backup/modes/factory.go`.

6. **Add an execution config**
   - Update `internal/backup/mode_config.go` so the mode has consistent header/log/success message.

7. **Update docs**
   - Add/extend user documentation in `docs/`.

### C) Change which databases are included/excluded

Database selection logic is split by intent:

- General filtering from flags/files/system rules: `internal/backup/setup/filter.go` → `pkg/database.FilterDatabases`.
- Interactive multi-select for the `filter` command: `internal/backup/selection/selector.go`.
- Primary/secondary naming rules and companion handling: `internal/backup/selection/*`.

If you change exclusion rules (e.g., new suffix rules), update selection code and ensure the behavior is consistent across:

- interactive selection
- non-interactive include lists
- metadata fields (e.g., excluded databases for `all` mode)

### D) Add or change mysqldump arguments

- Build logic is in `internal/backup/execution/args.go`.
- Retry heuristics are in `internal/backup/execution/retry.go`.

Guidelines:

- Avoid logging raw args containing passwords.
- If you add args that may not be supported by older mysqldump versions, consider adding a retry strategy similar to `RemoveUnsupportedMysqldumpOption`.

### E) Extend the output pipeline (compression/encryption/format)

Writer logic lives in `internal/backup/writer/engine.go`.

Important constraints:

- Keep it streaming (`io.Writer` chaining).
- Ensure all closers are closed in reverse order.
- If you add a new pipeline layer, consider how it affects:
  - file extensions (naming)
  - metadata fields
  - restore compatibility (restore pipeline expects the inverse order)

### F) Add metadata fields

Metadata is generated in:

- `internal/backup/execution/builder.go` (creates `types_backup.BackupMetadata`)
- `internal/backup/metadata/generator.go` (constructs the object)
- `internal/backup/metadata/writer.go` (atomic file write)

If you add new fields:

- Add them to `internal/types/types_backup/results.go` (`BackupMetadata`).
- Populate them in generator/builder.
- Confirm JSON output remains backward compatible if other tooling reads it.

---

## 5) Safety & UX Requirements

### Graceful shutdown must cleanup partial files
Backup is expected to remove partial output on cancellation.

- State tracking is managed by `backup.Service.SetCurrentBackupFile()` / `ClearCurrentBackupFile()`.
- `Service.HandleShutdown()` removes the current output file best-effort.

If you create additional output files (temporary, sidecar, etc.), decide whether they should also be cleaned up and where.

### Non-interactive must never prompt
If you add interactive prompts, they must be behind `!opts.NonInteractive` checks.

### Encryption key handling
- If encryption is enabled, key must be available (flag or env `SFDB_BACKUP_ENCRYPTION_KEY`).
- Setup layer enforces minimal validation before the actual backup runs.

---

## 6) Logging & Error Reporting

- Use `applog.Logger` for normal logs.
- Use `errorlog.ErrorLogger` to capture execution failures with stderr output.
- When returning errors, wrap them with context (`fmt.Errorf("...: %w", err)`) so callers can add higher-level meaning.

Avoid leaking secrets:

- Do not print plaintext keys or passwords.
- When logging mysqldump args, use a masked version.

---

## 7) Suggested Checklist for Backup PRs

Before opening a PR for a backup feature:

- CLI help text updated (command + flags)
- Non-interactive (`--quiet`) validations updated
- Interactive flow still works (no unexpected prompts)
- Output naming/extensions consistent with compression/encryption choices
- Metadata remains correct and is written atomically
- Graceful shutdown still removes partial output
- Documentation updated under `docs/`

---

## 8) Quick Pointers (Files You Will Touch Often)

- Commands: `cmd/backup/*.go`
- Flags: `internal/cli/flags/backup.go`
- Parsing: `internal/cli/parsing/backup.go`
- Orchestration: `internal/backup/command.go`, `internal/backup/service.go`, `internal/backup/executor.go`
- Modes: `internal/backup/modes/*`
- Setup: `internal/backup/setup/*`
- Selection: `internal/backup/selection/*`
- Execution: `internal/backup/execution/*`
- Writer: `internal/backup/writer/*`
- Metadata: `internal/backup/metadata/*`
- Grants/GTID: `internal/backup/grants/*`, `internal/backup/gtid/*`

---

## 9) Example: Adding “Retry Strategy X” (Mini Walkthrough)

If you want to add a new retry strategy for a mysqldump failure pattern:

1. Add a detector function in `internal/backup/execution/retry.go`.
2. Add an args modifier function that returns `(newArgs, modified)`.
3. In `internal/backup/execution/engine.go`, extend `attemptRetries(...)` with your new rule.
4. Ensure failed partial outputs are cleaned up before retry (`cleanupFailedBackup`).

Keep retry strategies small and very targeted to known, common failures.

---

Last updated: 2026-01-02
