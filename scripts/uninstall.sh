#!/usr/bin/env bash
set -euo pipefail

# uninstall.sh - Uninstaller for sfDBTools.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/hadiy961/sfDBTools-New/main/scripts/uninstall.sh | sudo bash
#   curl -fsSL https://raw.githubusercontent.com/hadiy961/sfDBTools-New/main/scripts/uninstall.sh | sudo bash -s -- --purge
#
# Flags:
#   --purge    Hapus juga config dan data user (HATI-HATI)
# Env:
#   SFDBTOOLS_PREFIX (default: /usr/local) lokasi install tar non-root
#   SFDBTOOLS_YES=1  skip prompt (untuk --purge)

PREFIX="${SFDBTOOLS_PREFIX:-/usr/local}"
PURGE=false

if [[ ${#} -gt 1 ]]; then
  echo "Usage: $0 [--purge]" >&2
  exit 2
fi
if [[ ${#} -eq 1 ]]; then
  if [[ "$1" == "--purge" ]]; then
    PURGE=true
  else
    echo "Error: argumen tidak dikenal: $1" >&2
    exit 2
  fi
fi

if [[ "$(uname -s)" != "Linux" ]]; then
  echo "Error: uninstaller ini hanya untuk Linux." >&2
  exit 1
fi

is_root=false
if [[ "$(id -u)" -eq 0 ]]; then
  is_root=true
fi

need_cmd() { command -v "$1" >/dev/null 2>&1; }

pkg_removed=false

remove_deb() {
  if [[ "${is_root}" != "true" ]]; then
    echo "Error: uninstall paket .deb butuh root. Jalankan dengan sudo." >&2
    exit 1
  fi
  if need_cmd apt-get; then
    if [[ "${PURGE}" == "true" ]]; then
      apt-get purge -y sfdbtools
    else
      apt-get remove -y sfdbtools
    fi
  elif need_cmd dpkg; then
    if [[ "${PURGE}" == "true" ]]; then
      dpkg --purge sfdbtools
    else
      dpkg -r sfdbtools
    fi
  else
    echo "Error: dpkg/apt-get tidak ditemukan." >&2
    exit 1
  fi
}

remove_rpm() {
  if [[ "${is_root}" != "true" ]]; then
    echo "Error: uninstall paket .rpm butuh root. Jalankan dengan sudo." >&2
    exit 1
  fi
  if need_cmd dnf; then
    dnf -y remove sfdbtools
  elif need_cmd yum; then
    yum -y remove sfdbtools
  elif need_cmd rpm; then
    rpm -e sfdbtools
  else
    echo "Error: dnf/yum/rpm tidak ditemukan." >&2
    exit 1
  fi
}

remove_tar_files() {
  # Hapus binary dari lokasi umum.
  local candidates=(
    "/usr/bin/sfdbtools" "/usr/bin/sfDBTools"
    "/usr/local/bin/sfdbtools" "/usr/local/bin/sfDBTools"
    "${PREFIX}/bin/sfdbtools" "${PREFIX}/bin/sfDBTools"
  )

  for p in "${candidates[@]}"; do
    if [[ -e "${p}" || -L "${p}" ]]; then
      if [[ "${is_root}" != "true" ]]; then
        # Non-root: hanya hapus kalau user punya permission.
        rm -f "${p}" 2>/dev/null || true
      else
        rm -f "${p}" || true
      fi
    fi
  done
}

# 1) Coba uninstall via package manager kalau terinstall sebagai paket.
if need_cmd dpkg-query && dpkg-query -W -f='${Status}' sfdbtools 2>/dev/null | grep -q "install ok installed"; then
  echo "→ Uninstall paket (deb): sfdbtools"
  remove_deb
  pkg_removed=true
elif need_cmd rpm && rpm -q sfdbtools >/dev/null 2>&1; then
  echo "→ Uninstall paket (rpm): sfdbtools"
  remove_rpm
  pkg_removed=true
fi

# 2) Hapus binary hasil install tar (atau sisa-sisa).
if [[ "${pkg_removed}" != "true" ]]; then
  echo "→ Menghapus binary sfdbtools/sfDBTools (tar/manual install)"
  remove_tar_files
fi

# 3) Purge config/data (opsional)
if [[ "${PURGE}" == "true" ]]; then
  if [[ "${SFDBTOOLS_YES:-}" != "1" ]]; then
    echo "⚠️  --purge akan menghapus konfigurasi di /etc dan home user." >&2
    read -r -p "Lanjut purge? (y/N): " ans
    if [[ "${ans}" != "y" && "${ans}" != "Y" ]]; then
      echo "Dibatalkan." >&2
      exit 1
    fi
  fi

  if [[ "${is_root}" == "true" ]]; then
    rm -rf /etc/sfDBTools || true
  else
    echo "ℹ️  Skip hapus /etc/sfDBTools (butuh sudo)" >&2
  fi

  # User config
  if [[ -n "${XDG_CONFIG_HOME:-}" ]]; then
    rm -rf "${XDG_CONFIG_HOME}/sfDBTools" 2>/dev/null || true
  fi
  if [[ -n "${HOME:-}" ]]; then
    rm -rf "${HOME}/.config/sfDBTools" 2>/dev/null || true
  fi
fi

echo "✓ Uninstall selesai"
