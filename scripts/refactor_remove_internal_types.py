#!/usr/bin/env python3
"""Refactor helper: remove sfDBTools/internal/types usages.

Tujuan:
- Ganti import "sfDBTools/internal/types" dengan paket baru sesuai mapping.
- Ganti selector `types.X` ke `domain.X`, `dbscanmodel.X`, `profilemodel.X`.
- Rapikan import block (heuristik sederhana, bukan go/ast rewriter).

Catatan:
- Script ini sengaja konservatif: hanya handle simbol yang sudah dipindah di refactor ini.
- Jalankan dari root repo: `python3 scripts/refactor_remove_internal_types.py --apply`.

Last Modified : 2026-01-05
"""

from __future__ import annotations

import argparse
import os
import re
from dataclasses import dataclass
from pathlib import Path
from typing import Iterable, List, Set, Tuple

REPO_ROOT = Path(__file__).resolve().parents[1]

SKIP_DIRS = {
    ".git",
    "backup",
    "logs",
    "bin",
    "dist",
    "vendor",
    "node_modules",
}

# Mapping: types.<Symbol> -> <qualifier>.<Symbol>
SYMBOL_MAP = {
    # shared domain
    "DBInfo": "domain",
    "ProfileInfo": "domain",
    "SSHTunnelConfig": "domain",
    "FilterOptions": "domain",
    "FilterStats": "domain",
    "CompressionOptions": "domain",
    "EncryptionOptions": "domain",

    # dbscan feature
    "ScanEntryConfig": "dbscanmodel",
    "ScanOptions": "dbscanmodel",
    "ScanResult": "dbscanmodel",
    "DatabaseDetailInfo": "dbscanmodel",
    "ScanEntryConfig": "dbscanmodel",

    # profile feature
    "ProfileShowOptions": "profilemodel",
    "ProfileCreateOptions": "profilemodel",
    "ProfileEditOptions": "profilemodel",
    "ProfileDeleteOptions": "profilemodel",
    "ProfileEntryConfig": "profilemodel",

    # restore feature
    "RestoreSelectionEntry": "restoremodel",
    "RestoreSelectionOptions": "restoremodel",
    "RestoreSingleOptions": "restoremodel",
    "RestorePrimaryOptions": "restoremodel",
    "RestoreSecondaryOptions": "restoremodel",
    "RestoreAllOptions": "restoremodel",
    "RestoreCustomOptions": "restoremodel",
    "RestoreBackupOptions": "restoremodel",
    "RestoreResult": "restoremodel",
}

IMPORTS_BY_QUALIFIER = {
    "domain": "sfDBTools/internal/domain",
    "dbscanmodel": "sfDBTools/internal/app/dbscan/model",
    "profilemodel": "sfDBTools/internal/app/profile/model",
    "restoremodel": "sfDBTools/internal/app/restore/model",
}


@dataclass
class FileEdit:
    path: Path
    original: str
    updated: str


def iter_go_files(root: Path) -> Iterable[Path]:
    for dirpath, dirnames, filenames in os.walk(root):
        rel = Path(dirpath).resolve().relative_to(root)
        parts = set(rel.parts)
        if parts & SKIP_DIRS:
            dirnames[:] = []
            continue
        # prune hidden dirs (except .github maybe)
        dirnames[:] = [d for d in dirnames if not d.startswith(".") or d == ".github"]
        for name in filenames:
            if name.endswith(".go"):
                yield Path(dirpath) / name


def replace_types_selectors(src: str) -> Tuple[str, Set[str]]:
    """Replace `types.Symbol` occurrences and return needed qualifiers."""
    needed: Set[str] = set()

    def repl(m: re.Match[str]) -> str:
        sym = m.group(1)
        qualifier = SYMBOL_MAP.get(sym)
        if not qualifier:
            return m.group(0)
        needed.add(qualifier)
        return f"{qualifier}.{sym}"

    # only replace selector form: types.X
    out = re.sub(r"\btypes\.([A-Za-z_][A-Za-z0-9_]*)\b", repl, src)
    return out, needed


def has_import(src: str, import_path: str) -> bool:
    return re.search(rf'\n\s*"{re.escape(import_path)}"\s*\n', src) is not None


def remove_import_types(src: str) -> str:
    # remove single-line import
    src = re.sub(r'\n\s*"sfDBTools/internal/types"\s*\n', "\n", src)
    # remove inline form: import "sfDBTools/internal/types"
    src = re.sub(r'\n\s*import\s+"sfDBTools/internal/types"\s*\n', "\n", src)
    return src


def ensure_imports(src: str, needed_qualifiers: Set[str]) -> str:
    needed_paths: List[Tuple[str, str]] = []  # (qualifier, path)
    for q in sorted(needed_qualifiers):
        p = IMPORTS_BY_QUALIFIER.get(q)
        if not p:
            continue
        needed_paths.append((q, p))

    if not needed_paths:
        return src

    # Determine if file uses an import block.
    m = re.search(r"\nimport\s*\(\n", src)
    if not m:
        # no import block: if there are existing single imports, convert is complex; skip.
        return src

    # find end of import block
    start = m.end()
    end_m = re.search(r"\n\)\s*\n", src[start:])
    if not end_m:
        return src
    end = start + end_m.start()
    block = src[start:end]

    # build a set of already imported paths (rough)
    imported_paths = set(re.findall(r'\n\s*(?:[A-Za-z_][A-Za-z0-9_]*\s+)?"([^"]+)"', "\n" + block))

    additions: List[str] = []
    for qualifier, path in needed_paths:
        if path in imported_paths:
            continue
        if qualifier == "domain":
            additions.append(f'\t"{path}"')
        else:
            additions.append(f'\t{qualifier} "{path}"')

    if not additions:
        return src

    # insert additions near top of block, but after any standard library imports if present.
    # Simple strategy: append at end of block.
    if not block.endswith("\n"):
        block += "\n"
    block += "\n".join(additions) + "\n"

    return src[:start] + block + src[end:]


def maybe_update_last_modified(src: str) -> str:
    # If file has a Last Modified header line in comment, update to 2026-01-05
    return re.sub(
        r"^(\s*//\s*Last Modified\s*:\s*).*$",
        r"\1 2026-01-05",
        src,
        flags=re.MULTILINE,
    )


def rewrite_file(path: Path, apply: bool, update_last_modified: bool) -> FileEdit | None:
    original = path.read_text(encoding="utf-8")
    if "sfDBTools/internal/types" not in original and "types." not in original:
        return None

    updated = original
    updated = remove_import_types(updated)
    updated, needed = replace_types_selectors(updated)

    # if we replaced anything, ensure imports
    updated = ensure_imports(updated, needed)

    if update_last_modified and updated != original:
        updated = maybe_update_last_modified(updated)

    if updated == original:
        return None

    if apply:
        path.write_text(updated, encoding="utf-8")

    return FileEdit(path=path, original=original, updated=updated)


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--apply", action="store_true", help="Write changes to files")
    ap.add_argument("--update-last-modified", action="store_true", default=True, help="Update // Last Modified lines")
    ap.add_argument("--no-update-last-modified", action="store_false", dest="update_last_modified")
    args = ap.parse_args()

    edits: List[FileEdit] = []
    for go_file in iter_go_files(REPO_ROOT):
        e = rewrite_file(go_file, apply=args.apply, update_last_modified=args.update_last_modified)
        if e:
            edits.append(e)

    print(f"Scanned: {REPO_ROOT}")
    print(f"Files changed: {len(edits)}")

    # print a small summary
    for e in edits[:40]:
        rel = e.path.relative_to(REPO_ROOT)
        print(f"- {rel}")
    if len(edits) > 40:
        print(f"... and {len(edits) - 40} more")

    if not args.apply and edits:
        print("\nDry-run only. Re-run with --apply to write changes.")

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
