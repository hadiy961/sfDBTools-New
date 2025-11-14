#!/usr/bin/env bash
set -euo pipefail

# build_run.sh - Build and run sfDBTools
# Usage:
#   scripts/build_run.sh [--skip-run] [--race] [--clean] [--tags "tag1,tag2"] [--] [args...]
# Examples:
#   scripts/build_run.sh --skip-run                        # Only build
#   scripts/build_run.sh                                   # Build then run with no args
#   scripts/build_run.sh -- --help                         # Build then run 'sfdbtools --help'
#   scripts/build_run.sh -- profile show                   # Build then run 'sfdbtools profile show'
#   scripts/build_run.sh --race -- --help                  # Build with -race, then run

# Resolve project root (repo root = parent of this script's directory)
SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd -- "${SCRIPT_DIR}/.." && pwd)"
BIN_DIR="/usr/bin"
BIN_PATH="${BIN_DIR}/sfDBTools"

# Defaults
SKIP_RUN=false
WITH_RACE=false
CLEAN=false
BUILD_TAGS=""

# Parse arguments
PASSTHRU_ARGS=()
while [[ $# -gt 0 ]]; do
  case "${1}" in
    --skip-run)
      SKIP_RUN=true; shift ;;
    --race)
      WITH_RACE=true; shift ;;
    --clean)
      CLEAN=true; shift ;;
    --tags)
      BUILD_TAGS="${2:-}"; shift 2 ;;
    --)
      shift; PASSTHRU_ARGS=("${@}"); break ;;
    -h|--help)
      sed -n '2,25p' "${BASH_SOURCE[0]}"; exit 0 ;;
    *)
      # Treat unknown tokens before "--" as binary args too (quality-of-life)
      PASSTHRU_ARGS+=("${1}"); shift ;;
  esac
done

# Env checks
if ! command -v go >/dev/null 2>&1; then
  echo "Error: 'go' is not installed or not in PATH" >&2
  exit 1
fi

mkdir -p "${BIN_DIR}"

# Optional clean
if [[ "${CLEAN}" == "true" ]]; then
  rm -f "${BIN_PATH}"
fi

# Build flags
GO_BUILD_FLAGS=("-trimpath" "-ldflags" "-s -w")
if [[ "${WITH_RACE}" == "true" ]]; then
  GO_BUILD_FLAGS+=("-race")
fi
if [[ -n "${BUILD_TAGS}" ]]; then
  GO_BUILD_FLAGS+=("-tags" "${BUILD_TAGS}")
fi

# Build the main module (root contains main.go)
(
  cd "${ROOT_DIR}"
  echo "→ Building sfdbtools …"
  GO111MODULE=on go build "${GO_BUILD_FLAGS[@]}" -o "${BIN_PATH}" .
  echo "✓ Built: ${BIN_PATH}"
)

# Run (unless skipped)
if [[ "${SKIP_RUN}" == "false" ]]; then
  echo "→ Running: ${BIN_PATH} ${PASSTHRU_ARGS[*]:-}"
  "${BIN_PATH}" "${PASSTHRU_ARGS[@]:-}"
else
  echo "ℹ️  Build completed. Skipping run (use without --skip-run to execute)."
fi
