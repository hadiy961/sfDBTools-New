#!/usr/bin/env bash
set -euo pipefail

# release_github.sh - Sekali jalan untuk trigger GitHub Release via tag.
# Requirements:
# - Working tree bersih (tidak ada file modified/untracked yang ingin diikutkan)
# - Remote 'origin' sudah terset
# - GitHub Actions workflow .github/workflows/release.yml sudah ada
#
# Usage:
#   ./scripts/release_github.sh 1.0.0
#   ./scripts/release_github.sh v1.0.0

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd -- "${SCRIPT_DIR}/.." && pwd)"

if [[ $# -ne 1 ]]; then
  echo "Usage: $0 <version>  (contoh: 1.0.0 atau v1.0.0)" >&2
  exit 2
fi

RAW_VERSION="$1"
TAG="${RAW_VERSION}"
if [[ "${TAG}" != v* ]]; then
  TAG="v${TAG}"
fi

if [[ ! "${TAG}" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "Error: format versi harus semver: vMAJOR.MINOR.PATCH (contoh: v1.0.0)" >&2
  exit 2
fi

cd "${ROOT_DIR}"

if [[ -n "$(git status --porcelain=v1)" ]]; then
  echo "Error: working tree tidak bersih. Commit/stash dulu sebelum release." >&2
  git status --porcelain=v1 >&2
  exit 1
fi

if git rev-parse -q --verify "refs/tags/${TAG}" >/dev/null; then
  echo "Error: tag ${TAG} sudah ada." >&2
  exit 1
fi

if ! git remote get-url origin >/dev/null 2>&1; then
  echo "Error: remote 'origin' belum terset." >&2
  exit 1
fi

echo "→ Membuat tag ${TAG}"
git tag -a "${TAG}" -m "Release ${TAG}"

echo "→ Push tag ke origin (akan trigger GitHub Actions release)"
git push origin "${TAG}"

echo "✓ Selesai. Cek tab Releases di GitHub setelah workflow selesai."