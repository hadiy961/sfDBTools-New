# sfDBTools Copilot Instructions

## Project Overview
**sfDBTools** is a production-grade CLI utility for MySQL/MariaDB database management, built in **Go (Golang)**. It follows **Clean Architecture** principles and prioritizes data safety, security (AES-256), and automation.

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
  - `cmd/`: Entry points (Cobra commands). Keep thin.
  - `internal/`: Core business logic (Backup, Restore, Profile).
  - `pkg/`: Reusable, domain-agnostic libraries.
- **Dependency Injection**:
  - Dependencies (`Config`, `Logger`) are injected via `types.Dependencies`.
  - Passed from `main.go` -> `cmd.Execute` -> `PersistentPreRunE`.

## Build & Development Workflows
- **Build Script**: ALWAYS use the helper script.
  - **Build & Run**: `./scripts/build_run.sh -- [args]`
  - **Build Only**: `./scripts/build_run.sh --skip-run`
  - **Example**: `./scripts/build_run.sh -- backup single --help`

## Coding Conventions
- **File Naming**:
  - Explicit naming is preferred: `pkg/helper/encrypt.go` instead of `pkg/helper/helper_encrypt.go`.
- **Logging**:
  - Use `sfDBTools/internal/applog`.
  - Respect `consts.ENV_QUIET` for pipeline usage.
- **Safety**:
  - **Fail-Fast**: Validate connections and paths immediately.
  - **Streaming**: Prefer `io.Reader`/`io.Writer` pipelines over memory buffers.