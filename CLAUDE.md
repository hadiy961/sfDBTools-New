# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run

```bash
# Build dan install ke /usr/bin/sfdbtools (butuh root)
sudo bash bash/build_run.sh -- --help

# Build saja (tanpa run)
sudo bash bash/build_run.sh --skip-run

# Build + run dengan argumen tertentu
sudo bash bash/build_run.sh -- profile show

# Build dengan race detector
sudo bash bash/build_run.sh --race -- --help

# Build lokal tanpa install (untuk cek compile error)
go build ./...
```

Proyek ini tidak menggunakan unit test — tidak ada `go test ./...` di workflow normal. Verifikasi fitur dilakukan secara fungsional/integrasi.

## Arsitektur Tingkat Tinggi

### Alur Startup

`main.go` → bootstrap runtime flags → (opsional) auto-update → load config YAML → inject ke `cmd.Execute(deps)`.

`PersistentPreRunE` di `cmd/root.go` berjalan sebelum semua command: cek database SQLite sudah diinisialisasi (`sfdbtools init`), lalu override config YAML dengan nilai dari SQLite (`settings.SyncConfig`), dan (jika online) menjalankan auto-update + remote sync.

### Dependency Injection

`main.go` membuat `*appdeps.Dependencies` (berisi `Config`, `Logger`, `NotifyService`) dan menyimpannya ke global `appdeps.Deps`. Command dan service mengaksesnya lewat `internal/cli/deps`.

### Lapisan Kode

- **`cmd/`** — thin command definitions (Cobra), parsing flags saja. Sub-command dikelompokkan dalam sub-folder: `cmd/backup/`, `cmd/restore/`, `cmd/profile/`, dll.
- **`internal/app/`** — orchestration workflow per fitur (backup, restore, profile, cleanup, settings, sync).
- **`internal/services/`** — implementasi service (config loader, logger, notify).
- **`internal/shared/`** — utility reusable: compress, crypto, database, consts, execx, fsops, dll.
- **`internal/cli/`** — helpers CLI: deps injection, flag parsing, runner, resolver.
- **`internal/ui/`** — komponen UI terminal: menu interaktif (charmbracelet/huh), spinner, tabel, prompt.

### Inisialisasi (Wajib)

`sfdbtools init` **harus dijalankan sekali** sebelum command lain bisa dipakai. Init memigrasikan config YAML ke tabel `app_settings` di SQLite. Semua command (kecuali `version`, `update`, `init`, `completion`) akan `os.Exit(1)` jika SQLite belum diinisialisasi — cek dilakukan di `PersistentPreRunE` via `database.IsInitialized()`.

Jika sudah diinisialisasi, gunakan `sfdbtools settings` untuk mengubah konfigurasi operasional.

### Konfigurasi (Two-tier)

1. **`config.yaml`** — konfigurasi statis (path: `SFDB_APPS_CONFIG`, default `/etc/sfDBTools/config.yaml`, fallback `~/.config/sfdbtools/config.yaml`). Auto-generated saat pertama kali jalan.
2. **SQLite `app_settings`** (`storage.local_db`, default `/etc/sfDBTools/sfdbtools.db`) — konfigurasi operasional per-kategori. SQLite **mengoverride** nilai YAML saat startup via `settings.SyncConfig` di `cmd/root.go`.

### Settings System

`sfdbtools settings` (diimplementasi di `internal/app/settings/`) adalah UI utama untuk mengubah konfigurasi operasional yang tersimpan di SQLite. Menu utamanya:

- **View / Edit** — baca-tulis tabel `app_settings` secara interaktif per-kategori
- **Cloud Sync & Hub** — konfigurasi remote sync (host, port, user, mode: `push-only` / `pull-only` / `two-way`)
- **Health & Diagnostics** — cek konektivitas database dan internet
- **Systemd & Maintenance** — generate/update systemd timer (`sfdbtools-sync.service`) via `internal/shared/systemd/`

**Locked settings**: Setting yang di-lock dari remote hub (`is_locked=1` di `app_settings`) tidak bisa diubah dari UI lokal — enforced di `internal/app/settings/security.go`.

### Streaming Backup Pipeline

Core backup ada di `internal/app/backup/writer/engine.go`. Pipeline: `mariadb-dump` (fallback `mysqldump`) → optional `compress.Writer` → optional `encrypt.Writer` → file. **Jangan buffer seluruh dump ke memori** — invariant ini kritis.

Backup modes diorganisasi via factory + interface kecil: `internal/app/backup/modes/interface.go` dan `factory.go`. Setiap mode (`single`, `all`, `primary`, `filter`, dll) mengimplementasi `ModeExecutor`.

### Remote Sync

`internal/app/sync/` mengimplementasi sinkronisasi dua arah (settings, profiles, backup jobs) ke remote hub (MySQL atau Postgres). Interface `RemoteProvider` di `provider.go` — implementasinya adalah `SQLRemoteProvider`. Sync berjalan otomatis di startup jika `sync_auto=true` di settings SQLite.

### Crypto

`internal/crypto/` menangani AES-GCM encryption untuk file backup dan profile koneksi. `crypto.ResolveKey()` meresolve key dari: argumen CLI → env var → config → fallback. Profile koneksi disimpan terenkripsi sebagai `.cnf.enc`.

## Konvensi Wajib

- **Bahasa dokumentasi**: Gunakan **Bahasa Indonesia** untuk komentar di dalam kode dan dokumentasi teknis.
- **Header file**: Update `// Last Modified : <tanggal>` di setiap file yang diubah (format: `DD Januari 2026`).
- **`--quiet` / `SFDB_QUIET=1`**: Semua output banner/spinner harus ditekan di quiet mode. Command `completion`, `version`, `update` selalu quiet.

## Filosofi Desain Go (dari Copilot Instructions)

- **DRY vs dependensi**: Duplikasi beberapa baris lebih baik daripada coupling kompleks. Kemandirian paket diprioritaskan.
- **KISS**: Kode harus "membosankan" dan eksplisit. Hindari generics kompleks, reflection, atau one-liner "cerdas".
- **YAGNI**: Jangan buat interface jika hanya ada satu implementasi. Jangan buat folder dalam untuk "ekspansi masa depan".
- **ISP**: Interface sekecil mungkin — 1 metode (seperti `io.Reader`) jauh lebih baik dari interface dengan 10 metode. Lihat contohnya di `internal/app/backup/modes/interface.go`.

## Environment Variables Penting

Source of truth: `internal/shared/consts/consts_env.go`.

| Var | Keterangan |
|---|---|
| `SFDB_APPS_CONFIG` | Override path config YAML |
| `SFDB_QUIET=1` | Suppress banner/spinner (untuk pipeline/CI) |
| `SFDB_NO_AUTO_UPDATE=1` | Disable auto-update paksa |
| `SFDB_BACKUP_ENCRYPTION_KEY` | Default key enkripsi backup |
| `SFDB_SOURCE_PROFILE_KEY` | Key untuk dekripsi profile koneksi sumber |
| `SFDB_TARGET_PROFILE_KEY` | Key untuk dekripsi profile koneksi tujuan |
| `SFDB_ENCRYPTION_KEY` | Key generic untuk `crypto` commands |
| `SFDB_SCRIPT_KEY` | Key untuk bundle `script` |
| `SFDB_GITHUB_TOKEN` | Token GitHub (opsional, hindari rate limit auto-update) |
