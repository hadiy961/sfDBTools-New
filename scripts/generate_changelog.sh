#!/usr/bin/env bash
set -euo pipefail

# generate_changelog.sh - Auto-generate changelog sederhana berbasis git log.
#
# Output: Markdown dengan format section per versi.
#
# Usage:
#   ./scripts/generate_changelog.sh --version 1.0.1 --from v1.0.0 --to v1.0.1 --out dist/CHANGELOG_1.0.1.md
#   ./scripts/generate_changelog.sh --version 1.0.0 --to v1.0.0 --out dist/CHANGELOG_1.0.0.md

VERSION=""
FROM_TAG=""
TO_TAG=""
OUT_PATH=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --version)
      VERSION="$2"; shift 2 ;;
    --from)
      FROM_TAG="$2"; shift 2 ;;
    --to)
      TO_TAG="$2"; shift 2 ;;
    --out)
      OUT_PATH="$2"; shift 2 ;;
    -h|--help)
      echo "Usage: $0 --version <x.y.z> [--from <tag>] --to <tag> --out <path>" >&2
      exit 0
      ;;
    *)
      echo "Error: argumen tidak dikenal: $1" >&2
      exit 2
      ;;
  esac
done

if [[ -z "$VERSION" || -z "$TO_TAG" || -z "$OUT_PATH" ]]; then
  echo "Error: --version, --to, dan --out wajib diisi" >&2
  exit 2
fi

if ! git rev-parse --git-dir >/dev/null 2>&1; then
  echo "Error: harus dijalankan di dalam git repo" >&2
  exit 1
fi

RANGE="$TO_TAG"
if [[ -n "$FROM_TAG" ]]; then
  RANGE="$FROM_TAG..$TO_TAG"
fi

DATE_UTC="$(date -u +%Y-%m-%d)"

mkdir -p "$(dirname "$OUT_PATH")"

{
  echo "# Changelog"
  echo
  echo "## v${VERSION} (${DATE_UTC})"
  echo
  echo "Perubahan sejak ${FROM_TAG:-initial} → ${TO_TAG}:"
  echo

  # Gunakan subject line. Skip merge commit supaya tidak noisy.
  # Note: --reverse agar urutan kronologis (lama → baru).
  if ! git log "$RANGE" --no-merges --pretty=format:'- %s (%h)' --reverse | sed '/^$/d'; then
    true
  fi

  echo
} > "$OUT_PATH"

echo "✓ Changelog generated: $OUT_PATH" >&2
