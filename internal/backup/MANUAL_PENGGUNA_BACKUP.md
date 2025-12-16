# Manual Pengguna - Fitur Backup Database sfDBTools

## Daftar Isi

1. [Pengantar](#pengantar)
2. [Konsep Dasar](#konsep-dasar)
3. [Mode Backup](#mode-backup)
4. [Arsitektur dan Alur Proses](#arsitektur-dan-alur-proses)
5. [Panduan Penggunaan](#panduan-penggunaan)
6. [Opsi dan Flag](#opsi-dan-flag)
7. [Fitur Tambahan](#fitur-tambahan)
8. [Contoh Penggunaan](#contoh-penggunaan)
9. [Troubleshooting](#troubleshooting)

---

## Pengantar

sfDBTools adalah aplikasi berbasis Go untuk melakukan backup dan manajemen database MariaDB/MySQL dengan fitur enkripsi dan kompresi. Fitur backup menyediakan berbagai mode untuk memenuhi kebutuhan backup yang beragam, mulai dari backup single database hingga backup semua database dengan filter yang fleksibel.

### Fitur Utama Backup

- ✅ **Multi-mode backup**: Single, Filter, All, Primary, Secondary
- ✅ **Kompresi otomatis**: Gzip, Zstd, XZ, Pgzip, Zlib
- ✅ **Enkripsi AES-256-GCM**: Kompatibel dengan OpenSSL
- ✅ **GTID Capture**: Untuk point-in-time recovery
- ✅ **User Grants Export**: Backup otomatis untuk user permissions
- ✅ **Graceful Shutdown**: Cleanup otomatis file parsial saat interupsi
- ✅ **Auto Cleanup**: Hapus backup lama secara otomatis
- ✅ **Dry-run Mode**: Preview tanpa eksekusi

---

## Konsep Dasar

### 1. Database Profile

Profile adalah file konfigurasi terenkripsi (`.cnf.enc`) yang menyimpan kredensial koneksi database. Profile harus dibuat terlebih dahulu sebelum melakukan backup.

**Lokasi Profile**: `config/database_profile/`

**Membuat Profile**:
```bash
sfdbtools profile create
```

### 2. Backup Mode

sfDBTools menyediakan berbagai mode backup untuk kebutuhan yang berbeda:

| Mode | Deskripsi | Output |
|------|-----------|--------|
| **single** | Backup satu database spesifik | 1 file per database |
| **filter** | Backup database pilihan dengan multi-select | 1 file (single-file) atau multi-file |
| **all** | Backup semua database dengan exclude filter | 1 file gabungan |
| **primary** | Backup database primary (tanpa suffix `_secondary`) | 1 file per database |
| **secondary** | Backup database secondary (dengan suffix `_secondary`) | 1 file per database |

### 3. Compression Types

Kompresi otomatis untuk menghemat ruang penyimpanan:

| Tipe | Ekstensi | Kecepatan | Rasio Kompresi |
|------|----------|-----------|----------------|
| gzip | `.gz` | Sedang | Baik |
| zstd | `.zst` | Sangat Cepat | Sangat Baik |
| xz | `.xz` | Lambat | Terbaik |
| pgzip | `.gz` | Cepat (parallel) | Baik |
| zlib | `.zlib` | Sedang | Baik |
| none | - | - | Tidak ada |

### 4. Enkripsi

Enkripsi AES-256-GCM kompatibel dengan OpenSSL untuk keamanan data:

- **Algoritma**: AES-256-GCM
- **Key Derivation**: PBKDF2 (100,000 iterasi)
- **Kompatibilitas**: OpenSSL decryption
- **Format Header**: "Salted__" untuk kompatibilitas

---

## Mode Backup

### Mode 1: Single Database Backup

Backup satu database spesifik dengan opsi untuk backup database companion (dmart, temp, archive).

**Karakteristik**:
- Fokus pada satu database utama
- Opsi backup database companion otomatis
- GTID capture untuk single snapshot
- User grants export untuk database terpilih

**Command**:
```bash
sfdbtools backup single [flags]
```

---

### Mode 2: Filter Backup

Backup database pilihan dengan multi-select interaktif atau dari daftar.

**Karakteristik**:
- Multi-select interaktif atau dari file/flag
- Dua sub-mode:
  - **single-file**: Gabungkan semua dalam satu file
  - **multi-file**: Pisahkan per database
- Fleksibel untuk kebutuhan khusus

**Command**:
```bash
sfdbtools backup filter [flags]
```

---

### Mode 3: All Database Backup

Backup semua database dalam satu file dengan kemampuan exclude.

**Karakteristik**:
- Backup semua database sekaligus
- Exclude system databases (information_schema, mysql, dll)
- Exclude database spesifik dengan flag atau file
- Exclude empty databases (opsional)
- Exclude data (hanya struktur)

**Command**:
```bash
sfdbtools backup all [flags]
```

---

### Mode 4: Primary Database Backup

Backup database primary (tanpa suffix `_secondary`).

**Karakteristik**:
- Filter otomatis: hanya database tanpa `_secondary`
- Satu file per database
- Cocok untuk backup database produksi utama

**Command**:
```bash
sfdbtools backup primary [flags]
```

---

### Mode 5: Secondary Database Backup

Backup database secondary (dengan suffix `_secondary`).

**Karakteristik**:
- Filter otomatis: hanya database dengan `_secondary`
- Satu file per database
- Cocok untuk backup database replika

**Command**:
```bash
sfdbtools backup secondary [flags]
```

---

## Arsitektur dan Alur Proses

### Diagram Arsitektur Sistem

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           sfDBTools Backup System                         │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                           │
│  ┌──────────────┐      ┌──────────────┐      ┌──────────────┐          │
│  │   CLI Layer  │─────▶│ Command Layer│─────▶│ Service Layer│          │
│  │   (Cobra)    │      │  (cmd/)      │      │ (internal/)  │          │
│  └──────────────┘      └──────────────┘      └──────────────┘          │
│         │                      │                      │                  │
│         │                      │                      │                  │
│         ▼                      ▼                      ▼                  │
│  ┌──────────────────────────────────────────────────────────┐          │
│  │              Backup Mode Executor Factory                 │          │
│  ├──────────────────────────────────────────────────────────┤          │
│  │  • CombinedExecutor    (all, filter-single-file)         │          │
│  │  • IterativeExecutor   (single, separated, primary,      │          │
│  │                         secondary, filter-multi-file)     │          │
│  └──────────────────────────────────────────────────────────┘          │
│                              │                                           │
│                              ▼                                           │
│  ┌─────────────────────────────────────────────────────────┐           │
│  │            Core Backup Engine Components                 │           │
│  ├─────────────────────────────────────────────────────────┤           │
│  │  • Profile Loader      • Database Filter                │           │
│  │  • Connection Manager  • GTID Capture                   │           │
│  │  • Mysqldump Executor  • User Grants Export             │           │
│  │  • Writer Pipeline     • Metadata Generator             │           │
│  └─────────────────────────────────────────────────────────┘           │
│                              │                                           │
│                              ▼                                           │
│  ┌─────────────────────────────────────────────────────────┐           │
│  │              Stream Processing Pipeline                  │           │
│  ├─────────────────────────────────────────────────────────┤           │
│  │  mysqldump → Compression → Encryption → File Write      │           │
│  │  (streaming, memory-efficient)                           │           │
│  └─────────────────────────────────────────────────────────┘           │
│                                                                           │
└─────────────────────────────────────────────────────────────────────────┘
```

### Diagram Alur Proses Backup

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    Flow Diagram: Backup Execution                         │
└─────────────────────────────────────────────────────────────────────────┘

START
  │
  ▼
┌─────────────────────────┐
│ 1. Parse Command & Flags│
│    - Mode selection     │
│    - Options parsing    │
└───────────┬─────────────┘
            │
            ▼
┌─────────────────────────┐
│ 2. Load Profile         │
│    - Decrypt .cnf.enc   │
│    - Validate credentials│
└───────────┬─────────────┘
            │
            ▼
┌─────────────────────────┐
│ 3. Connect to Database  │
│    - Test connection    │
│    - Set session vars   │
└───────────┬─────────────┘
            │
            ▼
┌─────────────────────────┐
│ 4. Database Filtering   │
│    - Get database list  │
│    - Apply filters      │
│    - Validate selection │
└───────────┬─────────────┘
            │
            ▼
┌─────────────────────────┐
│ 5. Setup Execution      │
│    - Prompt for ticket  │
│    - Create output dir  │
│    - Setup encryption   │
└───────────┬─────────────┘
            │
            ▼
┌─────────────────────────┐
│ 6. Mode Routing         │
│    ┌─────────────────┐  │
│    │ Combined Mode?  │──┼──Yes──▶ Combined Executor
│    └────────┬────────┘  │
│             │ No         │
│             ▼            │
│    ┌─────────────────┐  │
│    │ Iterative Mode? │──┼──Yes──▶ Iterative Executor
│    └─────────────────┘  │
└───────────┬─────────────┘
            │
            ▼
┌────────────────────────────────────────────────────────────┐
│ 7. Execute Backup (Mode-specific)                          │
│                                                             │
│  Combined Mode:                    Iterative Mode:         │
│  ┌──────────────────┐             ┌──────────────────┐    │
│  │ • Capture GTID   │             │ • Loop databases │    │
│  │ • Dump all DBs   │             │ • Dump per DB    │    │
│  │ • One file       │             │ • Multiple files │    │
│  └──────────────────┘             └──────────────────┘    │
│           │                                 │              │
│           └─────────────┬───────────────────┘              │
│                         ▼                                  │
│            ┌────────────────────────┐                      │
│            │ Stream Processing:     │                      │
│            │  mysqldump stdout      │                      │
│            │       ↓                │                      │
│            │  Compression (opt)     │                      │
│            │       ↓                │                      │
│            │  Encryption (opt)      │                      │
│            │       ↓                │                      │
│            │  File Write            │                      │
│            └────────────────────────┘                      │
└────────────────────────────┬───────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────┐
│ 8. Post-Backup Operations                   │
│    - Export user grants (if needed)         │
│    - Generate metadata JSON                 │
│    - Update GTID info in metadata           │
│    - Calculate checksums                    │
└───────────────────────┬─────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────┐
│ 9. Cleanup Old Backups (if enabled)         │
│    - Apply retention policy                 │
│    - Remove old files                       │
└───────────────────────┬─────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────┐
│ 10. Display Results                         │
│    - Success/failed count                   │
│    - File sizes                             │
│    - Duration                               │
│    - GTID info                              │
└───────────────────────┬─────────────────────┘
                        │
                        ▼
                       END

┌─────────────────────────────────────────────┐
│ Error Handling (at any step):               │
│  • Log error                                │
│  • Cleanup partial files                    │
│  • Return error result                      │
│                                             │
│ Interrupt Handling (CTRL+C):               │
│  • Graceful shutdown                        │
│  • Remove incomplete files                  │
│  • Close connections                        │
└─────────────────────────────────────────────┘
```

### Diagram Alur Filter Mode

```
┌────────────────────────────────────────────────────────────┐
│              Filter Mode Selection Flow                     │
└────────────────────────────────────────────────────────────┘

START (backup filter)
  │
  ▼
┌─────────────────────────┐
│ Flag --mode provided?   │
└───┬─────────────────┬───┘
    │ No              │ Yes
    │                 │
    ▼                 ▼
┌──────────────┐   ┌──────────────────┐
│ Interactive  │   │ Use provided mode│
│ Selection:   │   └────────┬─────────┘
│ • single-file│            │
│ • multi-file │            │
└───────┬──────┘            │
        │                   │
        └─────────┬─────────┘
                  │
                  ▼
        ┌────────────────────┐
        │ Mode = single-file?│
        └───┬────────────┬───┘
            │ Yes        │ No (multi-file)
            │            │
            ▼            ▼
    ┌──────────────┐  ┌──────────────┐
    │ Internal:    │  │ Internal:    │
    │ mode=combined│  │ mode=separated│
    └──────┬───────┘  └──────┬───────┘
           │                 │
           ▼                 ▼
    ┌──────────────┐  ┌──────────────┐
    │ Get DB list  │  │ Get DB list  │
    │ (multi-select│  │ (multi-select│
    │  or from     │  │  or from     │
    │  flags/file) │  │  flags/file) │
    └──────┬───────┘  └──────┬───────┘
           │                 │
           ▼                 ▼
    ┌──────────────┐  ┌──────────────┐
    │ Backup all   │  │ Loop: backup │
    │ selected DBs │  │ each DB to   │
    │ to ONE file  │  │ separate file│
    └──────────────┘  └──────────────┘
```

### Diagram Data Flow - Writer Pipeline

```
┌────────────────────────────────────────────────────────────┐
│              Writer Pipeline (Streaming)                    │
└────────────────────────────────────────────────────────────┘

┌─────────────────┐
│ mysqldump Process│
│   (subprocess)   │
└────────┬─────────┘
         │ stdout pipe
         ▼
┌─────────────────┐
│  Raw SQL Data   │
│  (streaming)    │
└────────┬─────────┘
         │
         ▼
┌─────────────────────────────────────┐
│ Compression Writer (optional)        │
│  • gzip.Writer                       │
│  • zstd.Encoder                      │
│  • xz.Writer                         │
│  • pgzip.Writer                      │
└────────┬────────────────────────────┘
         │ compressed stream
         ▼
┌─────────────────────────────────────┐
│ Encryption Writer (optional)         │
│  • AES-256-GCM                       │
│  • PBKDF2 key derivation             │
│  • Salted header                     │
└────────┬────────────────────────────┘
         │ encrypted stream
         ▼
┌─────────────────────────────────────┐
│ File Writer                          │
│  • os.Create                         │
│  • Buffered I/O                      │
└────────┬────────────────────────────┘
         │
         ▼
┌─────────────────┐
│ Backup File     │
│ (.sql.gz.enc)   │
└─────────────────┘

Note: Semua operasi streaming untuk efisiensi memori
```

---

## Panduan Penggunaan

### Prasyarat

1. **Buat Database Profile**:
   ```bash
   sfdbtools profile create
   ```

2. **Set Environment Variable** (opsional):
   ```bash
   export SFDB_ENCRYPTION_KEY="your-encryption-key"
   export SFDB_CONFIG_PATH="/path/to/config"
   ```

3. **Verifikasi Koneksi**:
   ```bash
   sfdbtools db-scan all-local
   ```

---

### Workflow Umum

#### Step 1: Pilih Mode Backup

Tentukan mode backup sesuai kebutuhan:
- **Single**: Untuk satu database spesifik
- **Filter**: Untuk beberapa database pilihan
- **All**: Untuk semua database
- **Primary/Secondary**: Untuk kategori database tertentu

#### Step 2: Siapkan Flag dan Opsi

Tentukan opsi yang diperlukan:
- **Profile**: `--profile /path/to/profile.cnf.enc`
- **Database**: `--db database_name` atau `--db-file list.txt`
- **Compression**: `--compress-type zstd --compress-level 3`
- **Encryption**: `--encryption-key "your-key"`
- **Output**: `--output-dir /backup/dir`
- **Ticket**: `--ticket TICKET-12345` (wajib)

#### Step 3: Jalankan Backup

```bash
sfdbtools backup <mode> [flags]
```

#### Step 4: Verifikasi Hasil

Periksa:
- File backup di output directory
- Metadata JSON (`.meta.json`)
- User grants file (`.users.sql`, jika ada)
- GTID info (`.gtid`, jika capture-gtid enabled)

---

## Opsi dan Flag

### Flag Global (Semua Mode)

| Flag | Shorthand | Tipe | Default | Deskripsi |
|------|-----------|------|---------|-----------|
| `--profile` | `-p` | string | - | Path ke database profile (.cnf.enc) |
| `--encryption-key` | `-K` | string | - | Kunci enkripsi untuk profile |
| `--ticket` | - | string | - | Ticket number untuk request (WAJIB) |
| `--output-dir` | - | string | config | Direktori output untuk backup |
| `--compress-type` | `-C` | string | zstd | Tipe kompresi (gzip/zstd/xz/pgzip/zlib/none) |
| `--compress-level` | - | int | 3 | Level kompresi (1-9) |
| `--capture-gtid` | - | bool | false | Capture GTID info untuk PITR |
| `--exclude-user` | - | bool | false | Exclude user grants dari export |
| `--dry-run` | - | bool | false | Preview tanpa eksekusi |
| `--force` | - | bool | false | Tampilkan opsi sebelum eksekusi |

### Flag Mode-Specific

#### Mode: Single, Primary, Secondary

| Flag | Tipe | Deskripsi |
|------|------|-----------|
| `--db` | string | Nama database untuk backup |
| `--filename` | string | Custom filename untuk output |

#### Mode: Filter

| Flag | Tipe | Deskripsi |
|------|------|-----------|
| `--mode` | string | Sub-mode: single-file atau multi-file |
| `--db` | string[] | List database untuk backup |
| `--db-file` | string | File berisi list database (satu per baris) |

#### Mode: All

| Flag | Tipe | Deskripsi |
|------|------|-----------|
| `--exclude-system` | bool | Exclude system databases |
| `--exclude-db` | string[] | List database untuk dikecualikan |
| `--exclude-db-file` | string | File berisi list database exclude |
| `--exclude-data` | bool | Hanya backup struktur (no data) |
| `--exclude-empty` | bool | Skip database kosong (no tables) |

---

## Fitur Tambahan

### 1. GTID Capture

GTID (Global Transaction Identifier) digunakan untuk point-in-time recovery.

**Enable GTID Capture**:
```bash
sfdbtools backup all --capture-gtid --ticket TICKET-001
```

**Output**:
- File `.gtid` berisi informasi GTID snapshot
- Metadata JSON mencatat GTID info

**Use Case**:
- Replication setup
- Point-in-time recovery
- Disaster recovery planning

---

### 2. User Grants Export

Export otomatis untuk user permissions, berguna saat restore.

**Behavior**:
- **Mode All**: Export semua user grants
- **Mode Filter/Single**: Export user dengan grants ke database terpilih
- **Mode Primary/Secondary**: Export user untuk database dalam kategori

**Output**:
- File `.users.sql` berisi CREATE USER dan GRANT statements

**Disable Export**:
```bash
sfdbtools backup all --exclude-user --ticket TICKET-002
```

---

### 3. Auto Cleanup Old Backups

Hapus backup lama secara otomatis berdasarkan retention policy.

**Konfigurasi** (`sfDBTools_config.yaml`):
```yaml
backup:
  cleanup:
    enabled: true
    retention_days: 7
    min_backups: 3
```

**Behavior**:
- Berjalan setelah backup selesai
- Hapus file lebih lama dari `retention_days`
- Pertahankan minimal `min_backups` file

---

### 4. Dry-Run Mode

Preview backup tanpa eksekusi nyata, berguna untuk testing.

**Usage**:
```bash
sfdbtools backup all --dry-run --ticket TICKET-003
```

**Output**:
- List database yang akan di-backup
- File paths yang akan dibuat
- Konfigurasi yang akan digunakan
- Tidak membuat file backup

---

### 5. Graceful Shutdown

Penanganan interrupt (CTRL+C) untuk cleanup otomatis.

**Behavior**:
- Deteksi SIGINT/SIGTERM
- Hentikan proses mysqldump
- Hapus file backup parsial
- Close database connections
- Log status interrupt

---

### 6. Metadata JSON

Setiap backup menghasilkan metadata file untuk tracking.

**Content**:
```json
{
  "database_name": "mydb",
  "backup_date": "2025-12-16T10:30:00Z",
  "backup_type": "all",
  "compression": {
    "enabled": true,
    "type": "zstd",
    "level": 3
  },
  "encryption": {
    "enabled": true,
    "algorithm": "AES-256-GCM"
  },
  "gtid_info": {
    "executed": "uuid:1-100",
    "purged": "uuid:1-50"
  },
  "file_size": 1048576,
  "user_grants_file": "mydb.users.sql",
  "ticket": "TICKET-001"
}
```

---

## Contoh Penggunaan

### Contoh 1: Backup Single Database

```bash
# Basic backup
sfdbtools backup single --db myapp_db --ticket TICKET-001

# Dengan custom filename dan enkripsi
sfdbtools backup single \
  --db myapp_db \
  --filename myapp_backup.sql \
  --encryption-key "my-secret-key" \
  --compress-type zstd \
  --compress-level 5 \
  --ticket TICKET-001
```

### Contoh 2: Backup dengan Filter (Multi-Select)

```bash
# Interactive mode
sfdbtools backup filter --ticket TICKET-002

# Dengan flag
sfdbtools backup filter \
  --mode single-file \
  --db app_db \
  --db user_db \
  --db log_db \
  --ticket TICKET-002

# Dari file list
sfdbtools backup filter \
  --mode multi-file \
  --db-file /path/to/db_list.txt \
  --ticket TICKET-002
```

### Contoh 3: Backup All dengan Exclude

```bash
# Exclude system databases
sfdbtools backup all \
  --exclude-system \
  --ticket TICKET-003

# Exclude specific databases
sfdbtools backup all \
  --exclude-db test_db \
  --exclude-db temp_db \
  --exclude-db-file /path/to/exclude_list.txt \
  --ticket TICKET-003

# Exclude empty databases
sfdbtools backup all \
  --exclude-system \
  --exclude-empty \
  --ticket TICKET-003
```

### Contoh 4: Backup Primary dengan GTID

```bash
sfdbtools backup primary \
  --capture-gtid \
  --compress-type xz \
  --compress-level 9 \
  --ticket TICKET-004
```

### Contoh 5: Backup Secondary

```bash
sfdbtools backup secondary \
  --output-dir /backup/secondary \
  --ticket TICKET-005
```

### Contoh 6: Dry-Run untuk Testing

```bash
# Test backup all configuration
sfdbtools backup all \
  --exclude-system \
  --dry-run \
  --ticket TICKET-006

# Preview dengan verbose
sfdbtools backup filter \
  --mode single-file \
  --db-file databases.txt \
  --dry-run \
  --force \
  --ticket TICKET-006
```

### Contoh 7: Backup dengan Kompresi Maksimal

```bash
sfdbtools backup all \
  --compress-type xz \
  --compress-level 9 \
  --ticket TICKET-007
```

### Contoh 8: Backup Hanya Struktur (No Data)

```bash
sfdbtools backup all \
  --exclude-data \
  --exclude-system \
  --ticket TICKET-008
```

### Contoh 9: Backup dengan Custom Output Directory

```bash
sfdbtools backup single \
  --db myapp_db \
  --output-dir /mnt/backup/$(date +%Y-%m-%d) \
  --ticket TICKET-009
```

### Contoh 10: Backup dengan Environment Variables

```bash
export SFDB_ENCRYPTION_KEY="my-secure-key"
export SFDB_CONFIG_PATH="/opt/sfdbtools/config"

sfdbtools backup all --ticket TICKET-010
```

---

## Troubleshooting

### Error: Dependencies tidak tersedia

**Penyebab**: Aplikasi tidak diinisialisasi dengan benar.

**Solusi**:
```bash
# Build dan run dengan script
./scripts/build_run.sh -- backup all --ticket TEST-001
```

---

### Error: Profile tidak ditemukan

**Penyebab**: File profile tidak ada atau path salah.

**Solusi**:
```bash
# List available profiles
ls config/database_profile/

# Create new profile
sfdbtools profile create

# Specify profile path
sfdbtools backup single --profile /path/to/profile.cnf.enc --ticket TEST-002
```

---

### Error: Koneksi database gagal

**Penyebab**: Kredensial salah atau database tidak reachable.

**Solusi**:
```bash
# Test connection
sfdbtools db-scan all-local

# Verify profile
sfdbtools profile show --profile /path/to/profile.cnf.enc

# Check network
ping database-host
telnet database-host 3306
```

---

### Error: Permission denied pada output directory

**Penyebab**: User tidak memiliki write permission.

**Solusi**:
```bash
# Create directory dengan permission
sudo mkdir -p /backup/mysql
sudo chown $USER:$USER /backup/mysql
chmod 755 /backup/mysql

# Atau gunakan directory user
sfdbtools backup all --output-dir ~/backups --ticket TEST-003
```

---

### Error: Out of disk space

**Penyebab**: Tidak cukup ruang untuk backup.

**Solusi**:
```bash
# Check disk space
df -h

# Enable cleanup old backups
# Edit: config/sfDBTools_config.yaml
backup:
  cleanup:
    enabled: true
    retention_days: 7

# Atau gunakan kompresi lebih tinggi
sfdbtools backup all \
  --compress-type xz \
  --compress-level 9 \
  --ticket TEST-004
```

---

### Warning: GTID capture gagal

**Penyebab**: GTID tidak enabled di database.

**Solusi**:
```bash
# Check GTID status
mysql -e "SELECT @@gtid_mode;"

# Enable GTID di my.cnf
[mysqld]
gtid_mode = ON
enforce_gtid_consistency = ON

# Restart MySQL
sudo systemctl restart mysql

# Atau skip GTID capture
sfdbtools backup all --ticket TEST-005
# (GTID capture default false, tidak perlu flag)
```

---

### Error: Backup interrupted oleh user

**Penyebab**: CTRL+C ditekan saat backup.

**Behavior**: Graceful shutdown akan:
- Hentikan proses mysqldump
- Hapus file parsial
- Log interrupt status

**Solusi**: Jalankan ulang backup.

---

### Error: Compression gagal

**Penyebab**: Compressor binary tidak tersedia.

**Solusi**:
```bash
# Install required compressor
# For zstd
sudo apt-get install zstd

# For xz
sudo apt-get install xz-utils

# Atau gunakan gzip (built-in)
sfdbtools backup all \
  --compress-type gzip \
  --ticket TEST-006
```

---

### Error: Encryption gagal - key tidak valid

**Penyebab**: Encryption key format salah atau tidak cukup panjang.

**Solusi**:
```bash
# Generate strong key
openssl rand -base64 32

# Set environment variable
export SFDB_ENCRYPTION_KEY="generated-key"

# Atau pass via flag
sfdbtools backup all \
  --encryption-key "your-key" \
  --ticket TEST-007
```

---

### Performance: Backup sangat lambat

**Penyebab**: Kompresi tinggi atau enkripsi.

**Solusi**:
```bash
# Gunakan kompresi lebih cepat
sfdbtools backup all \
  --compress-type zstd \
  --compress-level 1 \
  --ticket TEST-008

# Atau gunakan pgzip (parallel gzip)
sfdbtools backup all \
  --compress-type pgzip \
  --compress-level 3 \
  --ticket TEST-009

# Atau disable kompresi
sfdbtools backup all \
  --compress-type none \
  --ticket TEST-010
```

---

### Error: Database list kosong

**Penyebab**: Filter terlalu ketat atau tidak ada database yang match.

**Solusi**:
```bash
# List all databases
sfdbtools db-scan all

# Check filters
sfdbtools backup all \
  --exclude-system \
  --dry-run \
  --force \
  --ticket TEST-011

# Adjust filters
sfdbtools backup filter \
  --mode single-file \
  --db your_database \
  --ticket TEST-012
```

---

## Appendix

### A. Naming Convention

**Backup File Format**:
```
{database_name}_{timestamp}_{ticket}.sql[.compression_ext][.enc]
```

**Contoh**:
- `myapp_20251216_103000_TICKET001.sql.zst.enc`
- `userdb_20251216_103500_TICKET002.sql.gz`
- `all_databases_20251216_110000_TICKET003.sql.xz.enc`

**Metadata File**:
```
{backup_filename}.meta.json
```

**User Grants File**:
```
{backup_filename}.users.sql
```

**GTID File**:
```
{backup_filename}.gtid
```

---

### B. File Extensions

| Extension | Deskripsi |
|-----------|-----------|
| `.sql` | Raw SQL dump |
| `.gz` | Gzip compressed |
| `.zst` | Zstandard compressed |
| `.xz` | XZ compressed |
| `.zlib` | Zlib compressed |
| `.enc` | AES-256-GCM encrypted |
| `.meta.json` | Metadata file |
| `.users.sql` | User grants export |
| `.gtid` | GTID information |

---

### C. Environment Variables

| Variable | Deskripsi | Default |
|----------|-----------|---------|
| `SFDB_CONFIG_PATH` | Path ke config directory | `./config` |
| `SFDB_ENCRYPTION_KEY` | Default encryption key | - |
| `SFDB_PROFILE_PATH` | Default profile path | - |
| `SFDB_OUTPUT_DIR` | Default output directory | `{config}/backup` |
| `SFDB_QUIET` | Quiet mode (no UI banners) | `false` |
| `SFDB_LOG_LEVEL` | Log level (debug/info/warn/error) | `info` |

---

### D. Best Practices

1. **Selalu gunakan ticket number** untuk tracking dan audit.

2. **Enable enkripsi** untuk backup sensitive data:
   ```bash
   --encryption-key "secure-key"
   ```

3. **Capture GTID** untuk backup yang mungkin butuh PITR:
   ```bash
   --capture-gtid
   ```

4. **Gunakan kompresi optimal** berdasarkan kebutuhan:
   - **zstd level 3**: Balance speed/ratio (recommended)
   - **xz level 9**: Maximum compression (slow)
   - **pgzip level 3**: Fast parallel compression

5. **Enable auto-cleanup** untuk manajemen ruang disk:
   ```yaml
   backup:
     cleanup:
       enabled: true
       retention_days: 7
       min_backups: 3
   ```

6. **Test dengan dry-run** sebelum production backup:
   ```bash
   --dry-run --force
   ```

7. **Monitor disk space** sebelum backup besar:
   ```bash
   df -h
   ```

8. **Backup metadata dan user grants** untuk restore lengkap.

9. **Verifikasi backup** setelah selesai:
   ```bash
   ls -lh /backup/dir
   cat /backup/dir/backup.meta.json
   ```

10. **Dokumentasikan backup schedule** dan retention policy.

---

### E. Support

Untuk bantuan lebih lanjut:

1. **Check documentation**: `docs/` directory
2. **View logs**: `logs/` directory
3. **Report issues**: Contact system administrator
4. **Check error logs**: `logs/backup/` directory

---

**Version**: 1.0  
**Last Updated**: 2025-12-16  
**Author**: Hadiyatna Muflihun
