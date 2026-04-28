# Phase 1 Point 2: Backup Catalog & History

> **Context:** Saat ini, DBA harus navigasi filesystem secara manual (`ls`, `find`) untuk menemukan backup, dan membaca file `.meta.json` satu per satu. Tidak ada satu tempat sentral untuk melihat semua backup, mencari berdasarkan database name, atau melihat tren storage.

## Keputusan Desain

> [!IMPORTANT]
> **1. Storage Backend: JSON File-based**
> Menggunakan **JSON file** (`catalog.json`) sebagai storage backend. Alasan:
>
> - Zero dependency (tidak perlu CGo atau SQLite binary)
> - Portable (copy 1 file untuk backup/share)
> - Mudah di-inspect oleh DBA (`cat`, `jq`)
> - Sudah cukup untuk ratusan entry (typical DBA use case)
>
> Jika nanti catalog tumbuh >10.000 entries, bisa di-upgrade ke SQLite tanpa mengubah interface.

> [!IMPORTANT]
> **2. Catalog Location: config-driven**
> Default: `/etc/sfDBTools/catalog.json` (berdampingan dengan config lainnya).
> Bisa di-override lewat `config.yaml`:
> ```yaml
> backup:
>   catalog:
>     file_path: /etc/sfDBTools/catalog.json
> ```
> Jika DBA mau catalog "travel" bersama backup di NFS/NAS, tinggal ubah path.

> [!IMPORTANT]
> **3. Dual-Mode: Interaktif & Non-Interaktif**
> Semua command CLI (`list`, `report`, `catalog`) mendukung dua mode:
>
> - **Interaktif (default):** Wizard dengan `survey` prompt — DBA dipandu melalui filter, pilihan format, dll.
> - **Non-interaktif (`--quiet`):** Semua parameter lewat flags/args — cocok untuk scripting, cron, pipeline, dan scheduler systemd.
>
> Pattern ini konsisten dengan command lain (`backup filter`, `profile create`, dll).

> [!WARNING]
> **4. Auto-catalog: Enabled by Default**
> Setiap backup yang berhasil akan otomatis terdaftar di catalog melalui integrasi di `builder.go`. DBA bisa disable via config `catalog.enabled: false` jika tidak mau overhead I/O tambahan.

## Proposed Changes

---

### Component 1: Catalog Config

#### [MODIFY] [appconfig_types.go](file:///home/dbado/sfDBTools/internal/services/config/appconfig_types.go)

Tambah `CatalogConfig` di `BackupConfig`:

```diff
 type BackupConfig struct {
     ...
     Scheduler     SchedulerConfig    `yaml:"scheduler"`
+    Catalog       CatalogConfig      `yaml:"catalog"`
 }

+type CatalogConfig struct {
+    Enabled  bool   `yaml:"enabled"`   // default: true
+    FilePath string `yaml:"file_path"` // default: /etc/sfDBTools/catalog.json
+}
```

Config YAML:
```yaml
backup:
  catalog:
    enabled: true
    file_path: /etc/sfDBTools/catalog.json
```

---

### Component 2: Catalog Data Model

#### [NEW] `internal/app/backup/catalog/model.go`

Defines the core data structures for catalog entries and the catalog file itself.

```go
// CatalogEntry merepresentasikan satu record backup di catalog
type CatalogEntry struct {
    ID              string    `json:"id"`               // UUID unik per entry
    BackupFile      string    `json:"backup_file"`      // Absolute path ke file backup
    MetadataFile    string    `json:"metadata_file"`    // Path ke .meta.json
    DatabaseNames   []string  `json:"database_names"`   // List DB yang di-backup
    Hostname        string    `json:"hostname"`          // Source DB hostname
    BackupType      string    `json:"backup_type"`       // combined, separated, single, etc.
    BackupMode      string    `json:"backup_mode"`       // all, filter, single, primary, secondary
    BackupStatus    string    `json:"backup_status"`     // success, partial, failed
    BackupTime      time.Time `json:"backup_time"`       // Waktu mulai backup
    FileSizeBytes   int64     `json:"file_size_bytes"`
    FileSizeHuman   string    `json:"file_size_human"`
    Compressed      bool      `json:"compressed"`
    CompressionType string    `json:"compression_type,omitempty"`
    Encrypted       bool      `json:"encrypted"`
    Ticket          string    `json:"ticket,omitempty"`
    GTIDInfo        string    `json:"gtid_info,omitempty"`
    ChecksumHash    string    `json:"checksum_hash,omitempty"`  // dari Phase 1 Point 1
    ProfileUsed     string    `json:"profile_used,omitempty"`   // profile .cnf.enc yang dipakai
    RegisteredAt    time.Time `json:"registered_at"`            // Kapan entry ini masuk catalog
}

// Catalog adalah root structure dari catalog file
type Catalog struct {
    Version   string         `json:"version"`    // Schema version (e.g. "1.0")
    UpdatedAt time.Time      `json:"updated_at"`
    Entries   []CatalogEntry `json:"entries"`
}
```

---

### Component 3: Catalog Store (Read/Write)

#### [NEW] `internal/app/backup/catalog/store.go`

Handles all I/O operations for the catalog JSON file — load, save, atomic write. Uses the same atomic write pattern (`tmp` + `rename`) as the existing [writer.go](file:///home/dbado/sfDBTools/internal/app/backup/metadata/writer.go).

Key functions:

- `LoadCatalog(path string) (*Catalog, error)` — Load catalog dari disk (auto-create jika belum ada)
- `SaveCatalog(catalog *Catalog, path string) error` — Atomic write catalog ke disk
- `DefaultCatalogPath() string` — Return default path dari config atau fallback

#### [NEW] `internal/app/backup/catalog/store_test.go`

Unit test for store operations (load empty, load existing, save, concurrent safety).

---

### Component 4: Catalog Indexer (Auto-register)

#### [NEW] `internal/app/backup/catalog/indexer.go`

Handles automatic registration of new backups into the catalog. This is called from the backup execution engine after a backup completes.

Key functions:

- `RegisterBackup(meta *types_backup.BackupMetadata, mode string, profile string) error` — Tambah entry baru dari metadata
- `TryRegisterBackup(meta, mode, profile, logger)` — Safe wrapper yang log error tapi tidak propagate (untuk integrasi di builder.go)
- `RebuildFromDirectory(dir string, recursive bool) (int, error)` — Scan directory, baca semua `.meta.json`, rebuild catalog

Rebuild flow:

1. Walk directory recursively
2. Find all `*.meta.json` files
3. Parse each `BackupMetadata`
4. Deduplicate by `BackupFile` path
5. Create `CatalogEntry` for each
6. Save catalog

---

### Component 5: Catalog Query & Filter

#### [NEW] `internal/app/backup/catalog/query.go`

Provides search and filter capabilities on the catalog.

Key functions:

- `ListAll() []CatalogEntry` — Semua entries, sorted by backup_time DESC
- `FilterByDatabase(dbName string) []CatalogEntry` — Filter by database name (substring match)
- `FilterBySince(duration string) []CatalogEntry` — Filter by time range ("7d", "30d", "24h")
- `FilterByStatus(status string) []CatalogEntry` — Filter by status (success/failed/partial)
- `FilterByHostname(hostname string) []CatalogEntry` — Filter by source hostname
- `GetLatest(dbName string) *CatalogEntry` — Get most recent backup for a DB
- `Prune() (removed int, err error)` — Remove entries whose files no longer exist on disk

**Multi-filter composition** — Filters bisa di-chain:
```go
type QueryOptions struct {
    Database string
    Since    string
    Status   string
    Hostname string
    Limit    int
}

func Query(catalog *Catalog, opts QueryOptions) []CatalogEntry
```

---

### Component 6: Catalog Report

#### [NEW] `internal/app/backup/catalog/report.go`

Generates summary reports from catalog data.

Key functions:

- `GenerateReport(catalog *Catalog, period string) *CatalogReport`
- `DisplayReport(report *CatalogReport, format string)` — Render ke terminal (table/json/markdown)

```go
type CatalogReport struct {
    Period           string
    TotalBackups     int
    TotalSizeBytes   int64
    TotalSizeHuman   string
    SuccessCount     int
    FailedCount      int
    SuccessRate      float64
    DatabaseCoverage []DatabaseCoverage
    StorageTrend     []StorageTrendPoint
}

type DatabaseCoverage struct {
    DatabaseName string
    LastBackup   time.Time
    BackupCount  int
    TotalSize    int64
}

type StorageTrendPoint struct {
    Date        string
    SizeBytes   int64
    SizeHuman   string
    BackupCount int
}
```

Output contoh `backup report`:
```
📊 Backup Report — Weekly (17 Apr 2026 - 24 Apr 2026)
────────────────────────────────────────────────────────────────────────────
Summary:
  Total Backups   : 42
  Total Size      : 15.2 GB
  Success Rate    : 97.6% (41/42)
  Failed          : 1

Database Coverage:
┌──────────────────────┬─────────────────────┬───────┬──────────┐
│ DATABASE             │ LAST BACKUP         │ COUNT │ SIZE     │
├──────────────────────┼─────────────────────┼───────┼──────────┤
│ dbsaas_host          │ 2026-04-24 15:30    │ 7     │ 3.4 GB   │
│ dbsf_biznet_jtrust   │ 2026-04-24 15:31    │ 7     │ 2.1 GB   │
│ mysql                │ 2026-04-24 15:32    │ 7     │ 156 KB   │
│ ⚠ information_schema │ NEVER               │ 0     │ -        │
└──────────────────────┴─────────────────────┴───────┴──────────┘
```

---

### Component 7: CLI Commands (Dual-Mode: Interaktif & Non-Interaktif)

Semua command mengikuti pattern yang sudah ada di codebase:
- **Tanpa flags + TTY** → mode interaktif (survey prompts)
- **`--quiet` atau flags lengkap** → mode non-interaktif (langsung execute, fail-fast)

#### [NEW] `cmd/backup/list.go`

Registers `sfdbtools backup list` — **List & search backup catalog.**

##### Non-Interaktif (flags)

```bash
# List semua backup terbaru
sfdbtools backup list --quiet

# Filter by database
sfdbtools backup list --quiet --db dbsaas_host

# Filter by time range
sfdbtools backup list --quiet --since 7d

# Filter by status
sfdbtools backup list --quiet --status failed

# Filter by hostname
sfdbtools backup list --quiet --host dbserver1

# Combine filters + limit
sfdbtools backup list --quiet --db dbsaas_host --since 30d --status success --limit 10

# Output JSON (untuk scripting)
sfdbtools backup list --quiet --format json

# Pipeline: list failed backups dalam 24 jam terakhir
sfdbtools backup list --quiet --since 24h --status failed --format json | jq '.[].backup_file'
```

##### Interaktif (wizard)

Jika dijalankan tanpa `--quiet` dan tanpa filter flags:

```
=== Backup Catalog - List ===

? Pilih filter yang ingin digunakan:  [Use arrows to move, space to select, enter to submit]
  [x] Database Name
  [ ] Time Range
  [ ] Status
  [ ] Hostname

? Masukkan nama database (partial match): dbsaas
? Pilih format output: (Use arrow keys)
  > Table (default)
    JSON
    Compact

? Jumlah maksimum entry: (50)
```

##### Flags

| Flag | Default | Deskripsi |
|------|---------|-----------|
| `--db <name>` | - | Filter by database name (substring) |
| `--since <duration>` | - | Filter by time ("24h", "7d", "30d") |
| `--status <status>` | - | Filter: success, failed, partial |
| `--host <hostname>` | - | Filter by source hostname |
| `--format <fmt>` | table | Output: table, json, compact |
| `--limit <n>` | 50 | Max entries to show |

##### Output table columns

`#` · `DATABASE` · `BACKUP TIME` · `SIZE` · `STATUS` · `TYPE` · `COMPRESSED` · `TICKET`

---

#### [NEW] `cmd/backup/report.go`

Registers `sfdbtools backup report` — **Summary reporting.**

##### Non-Interaktif (flags)

```bash
# Weekly report (default)
sfdbtools backup report --quiet

# Monthly report
sfdbtools backup report --quiet --period monthly

# JSON output untuk monitoring tools
sfdbtools backup report --quiet --period daily --format json

# Markdown output untuk Slack/Teams webhook
sfdbtools backup report --quiet --period weekly --format markdown
```

##### Interaktif (wizard)

```
=== Backup Report ===

? Pilih periode report: (Use arrow keys)
  > Harian (Daily)
    Mingguan (Weekly)
    Bulanan (Monthly)

? Pilih format output: (Use arrow keys)
  > Table
    JSON
    Markdown

📊 Generating report...
```

##### Flags

| Flag | Default | Deskripsi |
|------|---------|-----------|
| `--period <period>` | weekly | Report period: daily, weekly, monthly |
| `--format <fmt>` | table | Output: table, json, markdown |

---

#### [NEW] `cmd/backup/catalog.go`

Registers `sfdbtools backup catalog` subcommand group — **Catalog management.**

##### Sub-commands

**1. `catalog rebuild`** — Rebuild catalog dari directory scan

```bash
# Non-interaktif: rebuild dari directory tertentu
sfdbtools backup catalog rebuild --quiet --dir /media/ArchiveDB

# Non-interaktif: rebuild dari default backup directory (config)
sfdbtools backup catalog rebuild --quiet

# Interaktif: wizard pilih directory
sfdbtools backup catalog rebuild
```

Interaktif flow:
```
=== Catalog Rebuild ===

? Masukkan directory untuk scan: /media/ArchiveDB
? Scan subdirectory secara rekursif? (Y/n)
? Ini akan menimpa catalog yang ada. Lanjutkan? (y/N)

🔍 Scanning...
   Found 125 .meta.json files
   Registered 120 new entries (5 duplicates skipped)
✓ Catalog rebuild complete.
```

**2. `catalog prune`** — Hapus entries yang file-nya sudah tidak ada di disk

```bash
# Non-interaktif: langsung prune
sfdbtools backup catalog prune --quiet

# Interaktif: konfirmasi sebelum hapus
sfdbtools backup catalog prune
```

Interaktif flow:
```
=== Catalog Prune ===

🔍 Checking catalog entries...
   Total entries: 120
   Files missing: 8

   Entries to remove:
   1. dbsaas_host_20260410_153025_dbserver1.sql.zst.enc (deleted from disk)
   2. dbsf_biznet_20260411_020015_dbserver1.sql.zst.enc (deleted from disk)
   ...

? Hapus 8 entry dari catalog? (y/N)

✓ Pruned 8 entries. Remaining: 112 entries.
```

**3. `catalog stats`** — Quick summary

```bash
# Selalu non-interaktif (no prompts needed)
sfdbtools backup catalog stats
sfdbtools backup catalog stats --format json
```

Output:
```
📋 Catalog Statistics
────────────────────────────────
  Total Entries    : 120
  Total Size       : 45.2 GB
  Oldest Backup    : 2026-03-01 02:00
  Newest Backup    : 2026-04-24 15:30
  Unique Databases : 15
  Success Rate     : 98.3% (118/120)
  Catalog File     : /etc/sfDBTools/catalog.json
  Catalog Size     : 24 KB
```

##### Flags (catalog rebuild)

| Flag | Default | Deskripsi |
|------|---------|-----------|
| `--dir <path>` | config base_directory | Directory to scan |
| `--recursive` | true | Scan subdirectories |
| `--force` | false | Skip confirmation prompt |

---

### Component 8: Integration ke Backup Engine

#### [MODIFY] [builder.go](file:///home/dbado/sfDBTools/internal/app/backup/execution/builder.go)

Di `buildRealBackupInfo()`, setelah `TrySaveBackupMetadata()` berhasil, panggil `catalog.TryRegisterBackup()` untuk auto-register backup baru ke catalog.

```diff
 func (e *Engine) buildRealBackupInfo(...) types_backup.DatabaseBackupInfo {
     ...
     manifestPath := ""
     if e.Config.Backup.Output.SaveBackupInfo {
         manifestPath = metadata.TrySaveBackupMetadata(meta, ...)
     }
+    // Auto-register ke catalog (jika enabled)
+    if e.Config.Backup.Catalog.Enabled {
+        catalog.TryRegisterBackup(meta, e.Options.BackupMode, e.Options.Profile.Path, e.Log)
+    }

     return (&metadata.DatabaseBackupInfoBuilder{...}).Build()
 }
```

#### [MODIFY] [main.go](file:///home/dbado/sfDBTools/cmd/backup/main.go)

Register new subcommands:

```diff
 func init() {
     ...
     CmdBackupMain.AddCommand(CmdBackupSchedule)
     CmdBackupMain.AddCommand(CmdBackupVerify)
+    CmdBackupMain.AddCommand(CmdBackupList)
+    CmdBackupMain.AddCommand(CmdBackupReport)
+    CmdBackupMain.AddCommand(CmdBackupCatalog)
 }
```

---

## File Structure Summary

```
internal/app/backup/catalog/
├── model.go          # CatalogEntry, Catalog structs
├── store.go          # Load/Save (JSON, atomic write)
├── store_test.go     # Unit tests
├── indexer.go        # RegisterBackup, RebuildFromDirectory
├── query.go          # ListAll, Filter*, Query, Prune
└── report.go         # GenerateReport, DisplayReport, CatalogReport struct

cmd/backup/
├── list.go           # CLI: sfdbtools backup list (interaktif + --quiet)
├── report.go         # CLI: sfdbtools backup report (interaktif + --quiet)
└── catalog.go        # CLI: sfdbtools backup catalog (rebuild/prune/stats)
```

## Interaktif vs Non-Interaktif — Summary

| Command | Interaktif (default) | Non-Interaktif (`--quiet` / flags) |
|---------|------|----------|
| `backup list` | Wizard: pilih filter → format → limit | Flags: `--db`, `--since`, `--status`, `--host`, `--format`, `--limit` |
| `backup report` | Wizard: pilih period → format | Flags: `--period`, `--format` |
| `catalog rebuild` | Wizard: input dir → confirm overwrite | Flags: `--dir`, `--force` |
| `catalog prune` | Tampilkan preview → confirm delete | Flag: `--quiet` (langsung hapus) |
| `catalog stats` | Langsung tampilkan (tidak perlu prompt) | Flag: `--format json` untuk scripting |

> [!TIP]
> **Pattern Detection:** Semua command mengecek `cmd.Root().PersistentFlags().GetBool("quiet")` untuk menentukan mode, konsisten dengan `filter.go`, `profile create`, dll.

## Verification Plan

### Automated Tests

1. **Unit test `store`** — Load/save cycle, verify data integrity, test atomic write
2. **Unit test `query`** — Filter by DB, since, status, combine filters, prune
3. **Unit test `report`** — Generate report, verify summary calculations
4. **Integration test** — Run `sfdbtools backup single`, then `sfdbtools backup list --quiet` dan pastikan entry baru muncul
5. **Rebuild test** — `sfdbtools backup catalog rebuild --quiet --dir /backup/test` pada directory yang sudah ada `.meta.json`

### Manual Verification

1. **Interaktif — backup list:**
   ```bash
   sfdbtools backup list
   # Pastikan wizard muncul dan bisa di-navigate
   ```

2. **Non-interaktif — backup list (scripting):**
   ```bash
   sfdbtools backup list --quiet --db dbsaas_host --since 7d --format json
   # Pastikan output JSON valid dan filter bekerja
   ```

3. **Auto-catalog setelah backup:**
   ```bash
   sfdbtools backup single --db dbsaas_host --profile localhost_3306
   cat /etc/sfDBTools/catalog.json | jq '.entries | length'
   # Pastikan entry bertambah
   ```

4. **Catalog rebuild:**
   ```bash
   sfdbtools backup catalog rebuild --quiet --dir /media/ArchiveDB
   sfdbtools backup catalog stats
   # Pastikan semua backup ter-index
   ```

5. **Non-interaktif — report (cron):**
   ```bash
   sfdbtools backup report --quiet --period weekly --format markdown > /tmp/weekly-report.md
   # Pastikan file markdown valid
   ```

6. **Catalog prune:**
   ```bash
   # Hapus satu file backup secara manual
   rm /backup/test/some_old_backup.sql.zst.enc
   sfdbtools backup catalog prune --quiet
   sfdbtools backup catalog stats
   # Pastikan entry terhapus dari catalog
   ```
