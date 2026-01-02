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
#
# Re-release (tag yang sama):
#   ./scripts/release_github.sh v1.0.0 --force
#   SFDBTOOLS_YES=1 ./scripts/release_github.sh v1.0.0 --force

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd -- "${SCRIPT_DIR}/.." && pwd)"

FORCE=false
if [[ $# -lt 1 || $# -gt 2 ]]; then
  echo "Usage: $0 <version> [--force]  (contoh: 1.0.0 atau v1.0.0)" >&2
  exit 2
fi

RAW_VERSION="$1"
if [[ ${#} -eq 2 ]]; then
  if [[ "$2" == "--force" ]]; then
    FORCE=true
  else
    echo "Error: argumen tidak dikenal: $2" >&2
    exit 2
  fi
fi
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
  if [[ "${FORCE}" != "true" ]]; then
    echo "Error: tag ${TAG} sudah ada. Pakai --force untuk re-release tag yang sama." >&2
    exit 1
  fi
fi

if ! git remote get-url origin >/dev/null 2>&1; then
  echo "Error: remote 'origin' belum terset." >&2
  exit 1
fi

remote_has_tag=false
if git ls-remote --tags origin "refs/tags/${TAG}" | grep -q .; then
  remote_has_tag=true
fi

if [[ "${FORCE}" == "true" ]]; then
  if [[ "${remote_has_tag}" == "true" ]]; then
    echo "⚠️  Tag ${TAG} sudah ada di remote. Ini akan menghapus tag remote dan membuat ulang." >&2
    echo "⚠️  Catatan: GitHub Release lama (jika ada) mungkin perlu dihapus manual di UI GitHub." >&2
  fi

  if [[ "${SFDBTOOLS_YES:-}" != "1" ]]; then
    read -r -p "Lanjut re-release ${TAG}? (y/N): " ans
    if [[ "${ans}" != "y" && "${ans}" != "Y" ]]; then
      echo "Dibatalkan." >&2
      exit 1
    fi
  fi

  # Delete local tag if present
  if git rev-parse -q --verify "refs/tags/${TAG}" >/dev/null; then
    echo "→ Menghapus tag lokal ${TAG}"
    git tag -d "${TAG}" >/dev/null
  fi

  # Delete remote tag if present
  if [[ "${remote_has_tag}" == "true" ]]; then
    echo "→ Menghapus tag remote ${TAG}"
    git push origin ":refs/tags/${TAG}" >/dev/null
  fi
fi

echo "→ Membuat tag ${TAG}"
git tag -a "${TAG}" -m "Release ${TAG}"

echo "→ Push tag ke origin (akan trigger GitHub Actions release)"
git push origin "${TAG}"

echo "✓ Selesai. Cek tab Releases di GitHub setelah workflow selesai."