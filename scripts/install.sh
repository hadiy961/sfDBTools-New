#!/usr/bin/env bash
set -euo pipefail

# install.sh - One-click installer for sfdbtools from GitHub Releases.
# Default: installs latest release for linux/amd64.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/hadiy961/sfdbtools-New/main/scripts/install.sh | sudo bash
#   curl -fsSL https://raw.githubusercontent.com/hadiy961/sfdbtools-New/main/scripts/install.sh | bash
#
# Options (env vars):
#   SFDBTOOLS_REPO   (default: hadiy961/sfdbtools-New)
#   SFDBTOOLS_METHOD (deb|rpm|tar|auto) default: auto
#   SFDBTOOLS_PREFIX (default: /usr/local for root, ~/.local for non-root) used for tar installs

REPO="${SFDBTOOLS_REPO:-hadiy961/sfdbtools-New}"
METHOD="${SFDBTOOLS_METHOD:-auto}"
DEFAULT_PREFIX="/usr/local"
if [[ "$(id -u)" -ne 0 ]]; then
  DEFAULT_PREFIX="${HOME:-/usr/local}/.local"
fi
PREFIX="${SFDBTOOLS_PREFIX:-${DEFAULT_PREFIX}}"

if [[ "$(uname -s)" != "Linux" ]]; then
  echo "Error: installer ini hanya untuk Linux." >&2
  exit 1
fi

ARCH="$(uname -m)"
if [[ "${ARCH}" != "x86_64" && "${ARCH}" != "amd64" ]]; then
  echo "Error: arsitektur tidak didukung: ${ARCH}. Saat ini hanya linux/amd64." >&2
  exit 1
fi

is_root=false
if [[ "$(id -u)" -eq 0 ]]; then
  is_root=true
fi

need_cmd() {
  command -v "$1" >/dev/null 2>&1
}

if ! need_cmd curl; then
  echo "Error: butuh 'curl'" >&2
  exit 1
fi

detect_method() {
  if [[ "${METHOD}" != "auto" ]]; then
    echo "${METHOD}"; return
  fi

  if [[ -r /etc/os-release ]]; then
    . /etc/os-release
    case "${ID:-}" in
      ubuntu|debian|linuxmint|pop|elementary)
        echo "deb"; return
        ;;
      centos|rhel|rocky|almalinux|fedora)
        echo "rpm"; return
        ;;
    esac
  fi

  # Fallback
  echo "tar"
}

METHOD_REAL="$(detect_method)"

# Use GitHub "releases/latest" redirect so we don't need API/jq.
BASE="https://github.com/${REPO}/releases/latest/download"

tmpdir="$(mktemp -d)"
cleanup() { rm -rf "${tmpdir}"; }
trap cleanup EXIT

install_deb() {
  local url="${BASE}/sfdbtools_latest_amd64.deb"
  local out="${tmpdir}/sfdbtools_latest_amd64.deb"

  # Backward-compatible: if asset naming is versioned, try to discover by hitting HEAD on common pattern.
  # Prefer the stable 'latest' name if you keep it in releases; otherwise fallback to versioned name.
  if ! curl -fsSLI "${url}" >/dev/null 2>&1; then
    # Fallback to wildcard-like approach: not possible without API.
    echo "Error: asset .deb tidak ditemukan di latest release." >&2
    echo "Hint: pastikan release mengupload 'sfdbtools_latest_amd64.deb' atau jalankan install via tar." >&2
    exit 1
  fi

  echo "→ Downloading ${url}"
  curl -fsSL "${url}" -o "${out}"

  if ! need_cmd apt-get; then
    echo "Error: apt-get tidak ditemukan. Gunakan METHOD=rpm atau METHOD=tar." >&2
    exit 1
  fi

  if [[ "${is_root}" != "true" ]]; then
    echo "Error: install .deb butuh root. Jalankan dengan sudo." >&2
    exit 1
  fi

  echo "→ Installing .deb"
  apt-get update -y >/dev/null
  apt-get install -y "${out}"

  echo "✓ Installed: sfdbtools"
}

install_rpm() {
  local url="${BASE}/sfdbtools-latest-1.x86_64.rpm"
  local out="${tmpdir}/sfdbtools-latest-1.x86_64.rpm"

  if ! curl -fsSLI "${url}" >/dev/null 2>&1; then
    echo "Error: asset .rpm tidak ditemukan di latest release." >&2
    echo "Hint: pastikan release mengupload 'sfdbtools-latest-1.x86_64.rpm' atau jalankan install via tar." >&2
    exit 1
  fi

  echo "→ Downloading ${url}"
  curl -fsSL "${url}" -o "${out}"

  if [[ "${is_root}" != "true" ]]; then
    echo "Error: install .rpm butuh root. Jalankan dengan sudo." >&2
    exit 1
  fi

  if need_cmd dnf; then
    echo "→ Installing .rpm (dnf)"
    dnf -y install "${out}"
  elif need_cmd yum; then
    echo "→ Installing .rpm (yum)"
    yum -y localinstall "${out}"
  elif need_cmd rpm; then
    echo "→ Installing .rpm (rpm)"
    rpm -Uvh "${out}"
  else
    echo "Error: tidak ada dnf/yum/rpm untuk install .rpm. Gunakan METHOD=tar." >&2
    exit 1
  fi

  echo "✓ Installed: sfdbtools"
}

install_tar() {
  local url="${BASE}/sfdbtools_linux_amd64.tar.gz"
  local out="${tmpdir}/sfdbtools_linux_amd64.tar.gz"

  if ! curl -fsSLI "${url}" >/dev/null 2>&1; then
    echo "Error: asset tar.gz tidak ditemukan di latest release." >&2
    echo "Hint: pastikan release mengupload 'sfdbtools_linux_amd64.tar.gz'." >&2
    exit 1
  fi

  echo "→ Downloading ${url}"
  curl -fsSL "${url}" -o "${out}"

  local bindir
  if [[ "${is_root}" == "true" ]]; then
    # Root install: /usr/bin is almost always in PATH
    bindir="/usr/bin"
  else
    # Non-root install: default to ~/.local/bin
    bindir="${PREFIX}/bin"
  fi
  mkdir -p "${bindir}"

  echo "→ Extracting to ${bindir}"
  tar -xzf "${out}" -C "${tmpdir}"
  install -m 0755 "${tmpdir}/sfdbtools" "${bindir}/sfdbtools"
  ln -sf "sfdbtools" "${bindir}/sfdbtools" || true

  echo "✓ Installed: ${bindir}/sfdbtools"
  if [[ "${is_root}" != "true" ]]; then
    case ":${PATH}:" in
      *":${bindir}:"*) : ;;
      *)
        echo "ℹ️  PATH kamu belum berisi ${bindir}." >&2
        echo "    Tambahkan ini (contoh bash):" >&2
        echo "    echo 'export PATH=\"${bindir}:\$PATH\"' >> ~/.bashrc" >&2
        echo "    source ~/.bashrc" >&2
        ;;
    esac
  fi
}

case "${METHOD_REAL}" in
  deb) install_deb ;;
  rpm) install_rpm ;;
  tar) install_tar ;;
  *)
    echo "Error: method tidak dikenal: ${METHOD_REAL}" >&2
    exit 1
    ;;
esac

# Run a quick version check (non-fatal)
if command -v sfdbtools >/dev/null 2>&1; then
  sfdbtools version || true
fi
