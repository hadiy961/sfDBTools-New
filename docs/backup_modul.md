# sfDBTools - Modul Backup Database

## Overview

Modul Backup menyediakan berbagai mode pencadangan database MariaDB/MySQL dengan fitur:
- 🔐 **Encryption**: AES-256-GCM dengan PBKDF2
- 🗜️ **Compression**: zstd, gzip, xz, pgzip, zlib
- 📊 **Metadata**: Comprehensive backup tracking
- 🎯 **Flexibility**: 5 mode backup untuk berbagai kebutuhan
- 🤖 **Automation**: Non-interactive mode untuk scheduling

---

## Mode Backup

| Mode | Use Case | Status |
|------|----------|--------|
| **All** | Backup seluruh instance | ✅ [Dokumentasi Lengkap](backup_all.md) |
| **Primary** | Database utama client | 📝 Coming soon |
| **Secondary** | Database sekunder client | 📝 Coming soon |
| **Single** | Backup 1 database spesifik | 📝 Coming soon |
| **Filter** | Backup custom selection | 📝 Coming soon |

---

## Quick Start

```bash
# Backup semua database (interactive)
sfdbtools db-backup all

# Backup dengan automation
sfdbtools db-backup all \
  --non-interactive \
  --profile /path/to/profile.cnf.enc \
  --ticket "BACKUP-001" \
  --backup-dir /backup/daily
```

---

## Mode Backup - Detail

### 1. Backup All ✅
**Full Instance Backup - Semua Database**

📖 **[Dokumentasi Lengkap →](backup_all.md)**

Backup seluruh database dalam satu file dengan metadata lengkap, user grants, dan GTID capture.

**Fitur:**
- Full instance snapshot
- GTID capture untuk replication
- User grants export
- Metadata JSON
- Exclude filters (system/empty/data)

**Command:**
```bash
sfdbtools db-backup all [flags]
```

---

### 2. Backup Primary 📝
Backup database utama client (naming: `dbsf_nbc_<client>` atau `dbsf_biznet_<client>`).

*Dokumentasi coming soon*

---

### 3. Backup Secondary 📝
Backup database sekunder client dengan instance identifier.

*Dokumentasi coming soon*

---

### 4. Backup Single 📝
Backup satu database spesifik dengan kontrol penuh.

*Dokumentasi coming soon*

---

### 5. Backup Filter 📝
Backup beberapa database yang dipilih secara custom.

*Dokumentasi coming soon*

---

## Common Features

### Encryption
- **Algorithm**: AES-256-GCM + PBKDF2 (100k iterations)
- **Format**: OpenSSL-compatible
- **Flag**: `--backup-key` atau ENV `SFDB_BACKUP_ENCRYPTION_KEY`

### Compression
- **Types**: zstd (recommended), gzip, pgzip, xz, zlib
- **Levels**: 1-9 (default: dari config)
- **Flag**: `--compress <type>` dan `--compress-level <1-9>`

### Output Files
```
<backup-dir>/
├── <mode>_<hostname>_<timestamp>_<info>.sql[.zst][.enc]
├── <backup-file>.meta.json
└── <backup-file>_users.sql
```

---

## Global Flags

**Required (non-interactive):**
- `-p, --profile` - Database profile path
- `-k, --profile-key` - Profile encryption key
- `-t, --ticket` - Ticket number
- `-K, --backup-key` - Backup encryption key (jika encrypt aktif)

**Common:**
- `-o, --backup-dir` - Output directory
- `-n, --non-interactive` - Non-interactive mode
- `-d, --dry-run` - Simulation mode

**Filters:**
- `-x, --exclude-data` - Schema only
- `-S, --exclude-system` - Skip system databases
- `-E, --exclude-empty` - Skip empty databases

---

## Environment Variables

```bash
export SFDB_SOURCE_PROFILE_KEY="profile-key"
export SFDB_BACKUP_ENCRYPTION_KEY="backup-key"
```

---

## Best Practices

- ✅ Gunakan encryption di production
- ✅ Set ticket number untuk audit trail
- ✅ Gunakan `--non-interactive` untuk scheduled backups
- ✅ Test restore secara berkala
- ✅ Monitor backup size dan duration
- ✅ Keep multiple backup generations

---

**Last Updated**: December 30, 2025