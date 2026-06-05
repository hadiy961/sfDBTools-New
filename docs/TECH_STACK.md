# Tech Stack — sfDBTools

> Dokumen ini merangkum seluruh teknologi dan library yang digunakan dalam project sfDBTools.
> Di-generate otomatis berdasarkan `go.mod` dan arsitektur kode per 07 Mei 2026.

---

## Bahasa & Runtime

| Komponen | Detail |
|---|---|
| **Language** | Go 1.25.6 |
| **Module Name** | `sfdbtools` |
| **Build** | Binary CLI tunggal, di-install ke `/usr/bin/sfdbtools` |

---

## CLI Framework

| Library | Kegunaan |
|---|---|
| `spf13/cobra` | Framework utama CLI & sub-command routing |
| `spf13/pflag` | Flag parsing (dependency cobra) |

---

## TUI / Terminal UI

| Library | Kegunaan |
|---|---|
| `charmbracelet/huh` | Form interaktif, wizard settings |
| `charmbracelet/bubbletea` | TUI framework (BubbleTea model) |
| `charmbracelet/bubbles` | Komponen TUI (input, list, spinner) |
| `charmbracelet/lipgloss` | Styling terminal (warna, layout) |
| `olekukonko/tablewriter` | Render tabel di terminal |
| `briandowns/spinner` | Spinner/loading animation |
| `fatih/color` | Warna output terminal |
| `AlecAivazis/survey/v2` | Prompt interaktif (confirm, select, input) |

---

## Database

| Library | Kegunaan |
|---|---|
| `modernc.org/sqlite` | SQLite (pure Go, tanpa CGO) — database lokal `sfdbtools.db` |
| `go-sql-driver/mysql` | Driver MySQL/MariaDB (koneksi ke database target) |
| `lib/pq` | Driver PostgreSQL (remote hub sync) |

---

## Compression & Archiving

| Library | Kegunaan |
|---|---|
| `klauspost/compress` | Kompresi gzip, zstd, dll |
| `klauspost/pgzip` | Parallel gzip (streaming backup) |
| `ulikunitz/xz` | Kompresi XZ/LZMA |

---

## Kriptografi & Security

| Library | Kegunaan |
|---|---|
| `golang.org/x/crypto` | AES-GCM encryption untuk backup & profile koneksi |
| `cespare/xxhash/v2` | Fast non-crypto hashing (integritas file) |
| `google/uuid` | UUID generation |

---

## Konfigurasi & Logging

| Library | Kegunaan |
|---|---|
| `joho/godotenv` | Load `.env` file |
| `gopkg.in/yaml.v3` | Parse `config.yaml` |
| `sirupsen/logrus` | Structured logging |
| `gopkg.in/natefinch/lumberjack.v2` | Log rotation |

---

## Export & Reporting

| Library | Kegunaan |
|---|---|
| `xuri/excelize/v2` | Generate file Excel (`.xlsx`) untuk export |
| `dustin/go-humanize` | Format angka/ukuran human-readable (e.g. `1.2 GB`) |

---

## File & Path Utilities

| Library | Kegunaan |
|---|---|
| `bmatcuk/doublestar/v4` | Glob pattern matching untuk file path |
| `mitchellh/hashstructure/v2` | Hashing struct Go |

---

## Testing

| Library | Kegunaan |
|---|---|
| `stretchr/testify` | Assertions untuk unit test (dipakai minimal) |

> **Catatan**: Project ini tidak menggunakan unit test secara aktif. Verifikasi fitur dilakukan secara fungsional/integrasi.

---

## External Tools (Runtime Dependency)

| Tool | Kegunaan |
|---|---|
| `mariadbd-dump` / `mysqldump` | Streaming dump MariaDB (backup pipeline) |
| `systemd` | Timer service untuk auto-sync terjadwal |

---

## Arsitektur Internal

```
cmd/          → Cobra command definitions (thin layer)
internal/
  app/        → Business logic per fitur (backup, restore, sync, settings)
  services/   → Logger, NotifyService, config loader
  shared/     → Utility reusable (compress, crypto, database, execx, fsops)
  cli/        → Dependency injection, flag helpers
  ui/         → TUI components (menu, spinner, tabel)
  crypto/     → AES-GCM key resolution & encryption
  domain/     → Domain types/interfaces
  autoupdate/ → GitHub release auto-update
```

---

## Alur Startup

```
main.go
  → bootstrap runtime flags
  → (opsional) auto-update
  → load config YAML
  → inject ke cmd.Execute(deps)

PersistentPreRunE (cmd/root.go)
  → cek SQLite sudah diinisialisasi
  → override config YAML dengan nilai dari SQLite (settings.SyncConfig)
  → (jika online) auto-update + remote sync
```

---

## Sistem Konfigurasi (Two-Tier)

| Tier | Path | Keterangan |
|---|---|---|
| **YAML** | `/etc/sfDBTools/config.yaml` | Konfigurasi statis, auto-generated saat pertama jalan |
| **SQLite** | `/etc/sfDBTools/sfdbtools.db` | Konfigurasi operasional, **mengoverride** YAML saat startup |

---

## Environment Variables

| Variable | Keterangan |
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
