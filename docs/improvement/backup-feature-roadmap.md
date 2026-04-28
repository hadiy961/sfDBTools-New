# sfDBTools Backup — Production-Grade Feature Roadmap (MariaDB)

> **Author**: AI-assisted  
> **Created**: 2026-04-24  
> **Status**: Draft / Open for Review

## Background

Berdasarkan analisis codebase yang sudah ada, `sfDBTools backup` saat ini sudah memiliki foundation yang solid:

### Fitur Yang Sudah Ada
| Capability | Status |
|---|---|
| Multi-mode backup (single, separated, combined, all, primary, secondary) | ✅ |
| Compression (gzip, zstd, xz, zlib, pgzip) | ✅ |
| Encryption (AES) | ✅ |
| GTID Capture (MariaDB `BINLOG_GTID_POS`) | ✅ |
| User & Grant Export (split CREATE USER + GRANT) | ✅ |
| Rich Metadata (`.meta.json` per backup) | ✅ |
| Retry Strategy (SSL mismatch, unsupported option) | ✅ |
| Graceful Shutdown + Partial File Cleanup | ✅ |
| Non-interactive / Background mode | ✅ |
| Systemd Timer Scheduler (in progress) | 🔨 |
| Cleanup by Retention Days | ✅ (basic) |
| Restore (interactive & companion-aware) | ✅ |

### Gap Untuk Production-Grade DBA Tooling

Yang **belum ada** dan sangat dibutuhkan seorang DBA di production:

---

## Proposed Features

### Feature 1: Backup Verification & Integrity Check
**Priority: 🔴 Critical** · **Complexity: Medium**

> *"Backup tanpa verification = tidak punya backup."*

Saat ini backup hanya dicek dari exit code `mariadb-dump`. Tidak ada validasi apakah file backup benar-benar restorable.

#### Key Capabilities
1. **Post-Backup Integrity Check** — Otomatis setelah backup selesai:
   - Checksum generation (SHA-256) disimpan di metadata
   - Header validation (cek SQL `-- MariaDB dump` header ada & valid)
   - Footer validation (cek `-- Dump completed` footer ada)
   - File size sanity check (> 0 bytes, dan > minimum threshold)
   
2. **Standalone Verify Command** — DBA bisa verify kapan saja:
   - `sfdbtools backup verify <file>` — Verify single file
   - `sfdbtools backup verify --dir <path>` — Verify semua backup di directory
   - `sfdbtools backup verify --latest` — Verify backup terbaru

3. **Restore Test (Dry-Run Restore)** — Level verifikasi tertinggi:
   - Parse SQL backup ke temporary database, tanpa persist
   - Cek semua `CREATE TABLE`, `INSERT` statement valid
   - Report: table count, row count estimate, schema consistency

#### Config Schema
```yaml
backup:
  verification:
    enabled: true
    checksum_algorithm: sha256  # sha256, md5
    post_backup_check: true     # auto-verify setelah backup
    header_footer_check: true   # validasi SQL dump structure
    min_file_size: 1KB          # minimum acceptable file size
```

#### CLI Commands
```bash
sfdbtools backup verify <backup-file>
sfdbtools backup verify --dir /media/ArchiveDB/20260424/
sfdbtools backup verify --latest --profile source.cnf.enc
sfdbtools backup verify --checksum-only <backup-file>
```

#### Proposed File Structure
```
internal/app/backup/
├── verify/
│   ├── checker.go          # Core verification engine
│   ├── checksum.go         # SHA-256/MD5 generation & comparison
│   ├── header_validator.go # SQL dump header/footer check
│   ├── size_validator.go   # File size sanity check
│   ├── restore_test.go     # Dry-run restore verification
│   └── report.go           # Verification result reporting
cmd/backup/
├── verify.go               # CLI command
```

---

### Feature 2: Backup Catalog & History
**Priority: ✅ Selesai** · **Complexity: Medium**

> *"Saya punya 500 file backup, mana yang terbaru untuk `dbsf_biznet_jtrust`?"*

DBA butuh satu tempat untuk melihat semua backup yang ada, history, dan status.
Fitur ini telah diimplementasikan mengikuti prinsip *Golang Enterprise Standard* (SOLID, Separation of Concerns). Terdapat lapisan *Repository* (untuk I/O `catalog.json`), *Service* (untuk filter, rebuild), dan *Delivery* (untuk CLI Command).

#### Key Capabilities
1. **Catalog Index** — Local JSON index yang otomatis di-*update* tiap backup:
   - Database name, backup time, file size, mode, status
   - Checksum, compression type, encrypted status
   - GTID position (untuk PITR chain tracking)

2. **List/Search Commands**:
   - `sfdbtools backup list` — Tampilkan semua backup terbaru
   - `sfdbtools backup list --db <name>` — Filter by database
   - `sfdbtools backup list --since 7d` — Filter by time range
   - `sfdbtools backup list --status failed` — Filter by status

3. **Backup Report** — Summary untuk reporting:
   - Total backups, success rate, dan Total Size.
   - Database coverage.

4. **Auto-catalog** — Setiap backup otomatis register ke catalog.

#### CLI Commands
```bash
# Listing
sfdbtools backup list
sfdbtools backup list --db dbsf_biznet_jtrust
sfdbtools backup list --since 7d --format table
sfdbtools backup list --since 30d --format json

# Report
sfdbtools backup report --period weekly
sfdbtools backup report --period monthly --format markdown

# Catalog management
sfdbtools backup catalog rebuild --dir /media/ArchiveDB
sfdbtools backup catalog prune   # hapus entry yang file-nya sudah tidak ada
```

#### Proposed File Structure
```
internal/app/backup/
├── catalog/
│   ├── store.go        # Catalog storage (JSON-based, upgrade path ke SQLite)
│   ├── indexer.go       # Auto-index setelah backup
│   ├── query.go         # Search & filter
│   ├── report.go        # Summary & reporting
│   └── model.go         # Catalog entry types
cmd/backup/
├── list.go              # CLI: backup list
├── report.go            # CLI: backup report
├── catalog.go           # CLI: catalog management
```

---

### Feature 3: Incremental / Differential Backup Support
**Priority: 🟡 High** · **Complexity: High**

> *"Full backup 200GB tiap hari terlalu lama dan boros storage."*

Untuk database besar, full backup setiap hari tidak feasible. DBA butuh strategi incremental.

#### Key Capabilities
1. **Mariabackup Integration** — Physical backup tool (bukan logical dump):
   - Full backup sebagai base
   - Incremental backup berdasarkan LSN (Log Sequence Number)
   - Differential backup (delta dari full terakhir)

2. **Backup Chain Management**:
   - Track full → incremental chain
   - Validate chain completeness sebelum restore
   - Auto-detect kapan perlu full backup baru (chain terlalu panjang)

3. **Strategy Configuration**:
   ```yaml
   backup:
     strategy:
       type: incremental     # full, incremental, differential
       full_backup_day: sunday  # kapan full backup
       max_chain_length: 6   # max incremental sebelum force full
       tool: mariabackup     # mariabackup atau mariadb-dump
   ```

4. **Hybrid Mode** — Logical dump (mariadb-dump) + Physical backup (mariabackup):
   - Logical: untuk portability & single-table restore
   - Physical: untuk speed & PITR

#### CLI Commands
```bash
# Incremental backup
sfdbtools backup incremental --profile source.cnf.enc --ticket INC-001

# Prepare (apply incremental ke full)
sfdbtools backup prepare --base <full-backup-dir> --incremental <inc-dir>

# Chain status
sfdbtools backup chain status --profile source.cnf.enc
```

> **⚠️ Note:** Feature ini membutuhkan `mariabackup` binary terinstall di server.
> Logical dump (mariadb-dump) **tidak mendukung** incremental secara native.
> Perlu evaluasi apakah ini align dengan workflow tim.

---

### Feature 4: Backup Health Monitoring & Alerting
**Priority: 🟡 High** · **Complexity: Medium**

> *"Backup gagal 3 hari berturut-turut dan tidak ada yang tahu."*

#### Key Capabilities
1. **Health Check Command**:
   - Cek backup terbaru per database (freshness check)
   - Alert jika backup lebih dari N jam/hari
   - Cek disk space availability sebelum backup
   - Cek backup chain integrity

2. **Notification Hooks** (Webhook-based):
   - Kirim alert ke Slack/Telegram/Discord on failure
   - Kirim daily summary report
   - Configurable notification level (failure only / all)

3. **Exit Code Convention** (untuk monitoring tools seperti Nagios/Zabbix):
   - `0` = OK, semua backup fresh
   - `1` = WARNING, beberapa backup mendekati threshold
   - `2` = CRITICAL, backup expired/missing

#### Config Schema
```yaml
backup:
  monitoring:
    freshness_threshold: 24h   # alert jika backup > 24 jam
    disk_space_warning: 10GB   # warning jika sisa disk < 10GB
    disk_space_critical: 2GB   # critical jika sisa disk < 2GB
    notifications:
      enabled: false
      webhook_url: ""          # Slack/Telegram/Discord webhook
      on_failure: true
      on_success: false
      daily_summary: true
      summary_time: "08:00"
```

#### CLI Commands
```bash
# Health check
sfdbtools backup health
sfdbtools backup health --db dbsf_biznet_jtrust
sfdbtools backup health --nagios   # output Nagios-compatible

# Disk check
sfdbtools backup disk-check --dir /media/ArchiveDB

# Send test notification
sfdbtools backup notify test
```

#### Proposed File Structure
```
internal/app/backup/
├── health/
│   ├── checker.go       # Health check engine
│   ├── freshness.go     # Backup freshness validation
│   ├── disk.go          # Disk space checker
│   ├── nagios.go        # Nagios/monitoring exit codes
│   └── notify/
│       ├── webhook.go   # Generic webhook sender
│       ├── slack.go     # Slack formatting
│       └── telegram.go  # Telegram formatting
cmd/backup/
├── health.go            # CLI command
```

---

### Feature 5: Smart Backup Rotation & Lifecycle (GFS)
**Priority: 🟡 High** · **Complexity: Medium**

> *"Saya ingin simpan 7 daily, 4 weekly, 3 monthly — seperti GFS."*

Saat ini cleanup hanya berdasarkan `retention_days`. DBA butuh Grandfather-Father-Son rotation.

#### Key Capabilities
1. **GFS Rotation Strategy**:
   - Daily: keep last N days
   - Weekly: keep last N weeks (backup hari tertentu)
   - Monthly: keep last N months (backup tanggal tertentu)

2. **Smart Cleanup**:
   - Jangan hapus backup yang menjadi base dari incremental chain
   - Jangan hapus backup terakhir (failsafe)
   - Dry-run mode untuk preview sebelum delete
   - Generate cleanup report (apa yang akan dihapus dan mengapa)

3. **Per-Database Retention** — Override retention per database:
   - Database critical (production) = longer retention
   - Database dev/staging = shorter retention

#### Config Schema
```yaml
backup:
  rotation:
    strategy: gfs          # simple (days only), gfs
    daily:
      keep: 7
    weekly:
      keep: 4
      day: sunday          # hari referensi
    monthly:
      keep: 3
      day: 1               # tanggal referensi
    protect_last_backup: true  # never delete the very last backup
    per_database:
      dbsf_production:
        daily_keep: 14
        monthly_keep: 6
```

#### CLI Commands
```bash
# Preview rotation
sfdbtools backup rotate --dry-run --dir /media/ArchiveDB

# Execute rotation
sfdbtools backup rotate --dir /media/ArchiveDB

# Override per database
sfdbtools backup rotate --db dbsf_production --keep-daily 14
```

---

### Feature 6: Point-in-Time Recovery (PITR) Support
**Priority: 🟢 Medium** · **Complexity: High**

> *"Ada data terhapus jam 14:30, saya perlu restore ke jam 14:29."*

#### Key Capabilities
1. **Binlog Backup** — Otomatis backup binlog files:
   - `mariadb-binlog --read-from-remote-server`
   - Continuous binlog streaming
   - Binlog rotation tracking

2. **PITR Restore**:
   - Restore full backup + replay binlog sampai timestamp tertentu
   - `--stop-datetime` support
   - `--stop-position` support (GTID-based)

3. **PITR Chain Validation**:
   - Cek apakah ada gap di binlog chain
   - Verify GTID continuity dari full backup ke binlog terbaru

#### CLI Commands
```bash
# Backup binlog
sfdbtools backup binlog --profile source.cnf.enc --output /backup/binlog/

# PITR restore
sfdbtools restore pitr --backup <full-backup> --until "2026-04-24 14:29:00"
sfdbtools restore pitr --backup <full-backup> --gtid "0-1-12345"

# Check PITR capability
sfdbtools backup pitr-check --profile source.cnf.enc
```

> **ℹ️ Note:** PITR membutuhkan binlog aktif di MariaDB (`log_bin = ON`).
> Fitur ini merupakan extension dari GTID capture yang sudah ada (`internal/app/backup/gtid/`).

---

## Implementation Roadmap

### Phase 1 — Foundation (1-2 sesi AI)
Fitur yang **langsung berguna** dan **independent** dari fitur lain:

| # | Feature | Priority | Effort (AI-assisted) |
|---|---------|----------|--------|
| 1 | **Backup Verification & Integrity** | 🔴 Critical | ~1 sesi |
| 2 | **Backup Catalog & History** | 🔴 Critical | ~1 sesi |
| 3 | **Scheduler** (systemd timer) | ✅ Selesai | ~1 sesi |

### Phase 2 — Operations (1-2 sesi AI)
Fitur untuk **daily operations** DBA:

| # | Feature | Priority | Effort (AI-assisted) |
|---|---------|----------|--------|
| 4 | **Health Monitoring & Alerting** | 🟡 High | ~1 sesi |
| 5 | **Smart Rotation (GFS)** | 🟡 High | ~1 sesi |

### Phase 3 — Advanced (2-3 sesi AI)
Fitur advanced yang butuh **architectural changes**:

| # | Feature | Priority | Effort (AI-assisted) |
|---|---------|----------|--------|
| 6 | **Incremental/Differential (mariabackup)** | 🟡 High | ~1-2 sesi |
| 7 | **PITR (Binlog Backup + Replay)** | 🟢 Medium | ~1-2 sesi |

---

## Open Questions

1. **Catalog Storage** — Prefer JSON file-based (simple, portable) atau SQLite (queryable, scalable)? JSON bisa upgrade ke SQLite nanti.
2. **Incremental Backup** — Apakah `mariabackup` sudah terinstall di production servers? Atau mau tetap fokus di `mariadb-dump` saja?
3. **Notification** — Webhook generic sudah cukup, atau butuh native integration (Slack App, Telegram Bot)?
4. **PITR** — Apakah binlog sudah aktif di semua production MariaDB? Apakah ada kebutuhan PITR yang mendesak?
5. **Feature Priority** — Dari 6 fitur di atas, mana yang paling urgent untuk dikerjakan duluan?

---

## Summary: Minimum Viable untuk Production-Grade

```
✅ Sudah ada:  Backup + Compress + Encrypt + GTID + Grants + Retry + Cleanup + Scheduler (systemd timer)


📋 HARUS ADA (Phase 1):
   1. Backup Verification (checksum + header/footer) - ✅ SELESAI (including Performance/UX enhancements)
   2. Backup Catalog (list + search + report)

📋 SANGAT DIBUTUHKAN (Phase 2):
   3. Health Monitoring (freshness check + alerting)
   4. GFS Rotation (smart cleanup)

📋 NICE TO HAVE (Phase 3):
   5. Incremental Backup (mariabackup)
   6. PITR (binlog backup + replay)
```
