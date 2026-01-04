# sfDBTools - Backup Module Documentation

**Last Updated**: January 2, 2026

## Table of Contents
- [Overview](#overview)
- [Backup Modes](#backup-modes)
- [Common Features](#common-features)
- [Global Flags & Parameters](#global-flags--parameters)
- [Mode-Specific Documentation](#mode-specific-documentation)
  - [Backup All](#backup-all)
  - [Backup Filter](#backup-filter)
  - [Backup Single](#backup-single)
  - [Backup Primary](#backup-primary)
  - [Backup Secondary](#backup-secondary)
- [Output Files](#output-files)
- [Interactive vs Non-Interactive Mode](#interactive-vs-non-interactive-mode)
- [Metadata File Structure](#metadata-file-structure)
- [Advanced Features](#advanced-features)
- [Troubleshooting](#troubleshooting)
- [Best Practices](#best-practices)

---

## Overview

The backup module provides comprehensive MySQL/MariaDB database backup capabilities with multiple strategies to suit different use cases. All backup modes support:

- **AES-256-GCM Encryption** (OpenSSL-compatible)
- **Multi-format Compression** (zstd, gzip, pgzip, xz, zlib)
- **Streaming Pipeline** (no large memory buffers)
- **Metadata Tracking** (JSON manifest with backup details)
- **User Grants Export** (optional)
- **GTID Capture** (for replication consistency)
- **Interactive & Non-Interactive Modes** (for automation)

---

## Backup Modes

| Mode | Description | Output | Use Case |
|------|-------------|--------|----------|
| **all** | Backup all databases (with exclude filters) | Single file | Full instance backup |
| **filter** | Backup selected databases | Single or multiple files | Selective bulk backup |
| **single** | Backup one specific database | Single file | Quick single DB backup |
| **primary** | Backup primary databases (pattern-based) | Multiple files | Production client databases |
| **secondary** | Backup secondary databases (pattern-based) | Multiple files | Dev/staging databases |

---

## Common Features

### 1. Encryption
- **Algorithm**: AES-256-GCM with PBKDF2 key derivation (100,000 iterations)
- **Format**: OpenSSL-compatible (`Salted__` header + 8-byte salt + ciphertext)
- **Streaming**: No memory overhead for large files
- **Key Sources**: Flag (`--backup-key`), ENV (`SFDB_BACKUP_ENCRYPTION_KEY`), or interactive prompt

### 2. Compression
Supported types:
- `zstd` (Zstandard) - **Recommended** - balanced speed and ratio
- `gzip` - Standard, widely compatible
- `pgzip` - Parallel gzip (faster on multi-core systems)
- `xz` - Highest compression ratio (slower)
- `zlib` - Compatible with many tools
- `none` - No compression

Compression levels: `1-9` (1=fastest/lowest ratio, 9=slowest/highest ratio)

### 3. User Grants Export
- Automatically exports `CREATE USER` and `GRANT` statements
- Saves to `*_users.sql` file
- Can be filtered by database (for filter/primary/secondary modes)
- Useful for restoring permissions

### 4. GTID Capture
- Captures binlog position for replication consistency
- Stored in metadata for restore operations
- Only available for single-file modes (all/filter single-file)

### 5. Dry Run Mode
- Validates all settings without executing backup
- Tests connection, generates preview filenames
- Useful for testing automation scripts

---

## Global Flags & Parameters

### Required (for non-interactive mode)
- `-p, --profile` - Database profile path (encrypted .cnf file)
- `-k, --profile-key` - Profile encryption key (or ENV `SFDB_SOURCE_PROFILE_KEY`)
- `-t, --ticket` - Ticket number for audit trail
- `-K, --backup-key` - Backup encryption key (if encryption enabled, or ENV `SFDB_BACKUP_ENCRYPTION_KEY`)

### Common Optional
- `-o, --backup-dir` - Output directory (default: from config)
- `-f, --filename` - Custom filename base (without extension)
- `-q, --quiet` - Non-interactive mode (for automation/pipeline; menonaktifkan prompt interaktif)
- `-d, --dry-run` - Dry run mode (validation only)

### Filters
- `-x, --exclude-data` - Schema only (no data)
- `-S, --exclude-system` - Skip system databases
- `-E, --exclude-empty` - Skip empty databases

### Compression
- `-C, --skip-compress` - Disable compression
- `-c, --compress` - Compression type (zstd, gzip, xz, zlib, pgzip, none)
- `-l, --compress-level` - Compression level (1-9)

### Encryption
- `--skip-encrypt` - Disable encryption
- `-K, --backup-key` - Encryption key

### Mode-Specific
- `--database` - Database name (single mode only)
- `--client-code` - Client code filter (primary/secondary modes)
- `--instance` - Instance identifier (secondary mode only)
- `--include-dmart` - Include companion `_dmart` databases (primary/secondary modes)
- `--db` - Database list (filter mode)
- `--db-file` - File with database list (filter mode)
- `--mode` - Output mode: single-file or multi-file (filter mode)

---

## Mode-Specific Documentation

### Backup All

**Purpose**: Full instance backup with exclude filters

**Use Case**: 
- Scheduled full backups
- Disaster recovery preparation
- Instance migration

**Command**:
```bash
sfdbtools db-backup all [flags]
```

**Database Selection**:
- Backs up ALL databases on the server
- Applies exclude filters:
  - System databases (mysql, information_schema, performance_schema, sys)
  - Empty databases (optional)
  - Databases from exclude list/file (optional)

**Output**:
- Single SQL file containing all databases
- Format: `all_<hostname>_<timestamp>_<count>db.sql[.compression][.enc]`
- Example: `all_prod-server_20260102_143022_23db.sql.zst.enc`

**Special Features**:
- GTID capture (for replication consistency)
- User grants export (all users)
- Excluded databases list in metadata

**Interactive Flow**:
1. Select profile (if not provided)
2. Connect to database
3. Input ticket number (if not provided)
4. Apply exclude filters
5. Show preview and options
6. Edit options loop (Lanjutkan/Ubah opsi/Batalkan)
7. Execute backup
8. Display results

**Non-Interactive Example**:
```bash
sfdbtools db-backup all \
  --quiet \
  --profile /etc/sfDBTools/profiles/db_prod.cnf.enc \
  --profile-key "your-profile-key" \
  --ticket "BACKUP-2026-001" \
  --backup-key "your-backup-key" \
  --backup-dir /backup/daily
```

---

### Backup Filter

**Purpose**: Selective backup of multiple databases

**Use Case**:
- Backup specific application databases
- Bulk backup of related databases
- Custom database grouping

**Command**:
```bash
sfdbtools db-backup filter [flags]
```

**Database Selection**:
- **Interactive mode**: Multi-select from available databases
- **Non-interactive mode**: Specify via `--db` flag or `--db-file`

**Output Modes**:
1. **Single-file** (combined): All selected databases in one file
   - Format: `combined_<hostname>_<timestamp>_<count>db.sql[.compression][.enc]`
   - GTID capture available
   - User grants filtered by selected databases

2. **Multi-file** (separated): One file per database
   - Format: `<database>_<hostname>_<timestamp>.sql[.compression][.enc]`
   - Per-database metadata
   - Per-database user grants

**Interactive Flow**:
1. Select profile
2. Choose output mode (single-file or multi-file)
3. Multi-select databases
4. Show preview and options
5. Edit options loop
6. Execute backup
7. Display results

**Examples**:
```bash
# Interactive mode with multi-select
sfdbtools db-backup filter

# Single-file mode (non-interactive)
sfdbtools db-backup filter \
  --quiet \
  --mode single-file \
  --db "app_db,user_db,log_db" \
  --profile /path/to/profile.cnf.enc \
  --ticket "BACKUP-001"

# Multi-file mode with database list file
sfdbtools db-backup filter \
  --quiet \
  --mode multi-file \
  --db-file /path/to/db_list.txt \
  --profile /path/to/profile.cnf.enc \
  --ticket "BACKUP-001"
```

---

### Backup Single

**Purpose**: Backup one specific database

**Use Case**:
- Quick single database backup
- Testing/development backups
- Application-specific backups

**Command**:
```bash
sfdbtools db-backup single [flags]
```

**Database Selection**:
- Specify via `--database` flag (required for non-interactive)
- Interactive selection from available databases

**Output**:
- Single file: `<database>_<hostname>_<timestamp>.sql[.compression][.enc]`

**Special Features**:
- Optional `--include-dmart` flag to backup companion `<database>_dmart`
- User grants filtered by selected database

**Naming Exclusions**:
- Cannot backup databases ending with `_temp` or `_archive`
- These suffixes are reserved for temporary/archive databases

**Examples**:
```bash
# Interactive selection
sfdbtools db-backup single

# Specific database
sfdbtools db-backup single \
  --database "my_app_db" \
  --ticket "BACKUP-001"

# With companion database
sfdbtools db-backup single \
  --database "dbsf_nbc_client" \
  --include-dmart \
  --ticket "BACKUP-001"
```

---

### Backup Primary

**Purpose**: Backup primary client databases following naming convention

**Use Case**:
- Production client database backups
- Scheduled primary database backups
- Client-specific backup automation

**Command**:
```bash
sfdbtools db-backup primary [flags]
```

**Database Selection**:
- **Pattern**: `dbsf_nbc_<client_code>` (without `_secondary` suffix)
- **Exclusions**: System, `_dmart`, `_temp`, `_archive` suffixes
- **Optional filter**: `--client-code` to narrow selection

**Output**:
- One file per primary database
- Companion `_dmart` databases (if `--include-dmart` is enabled)
- Combined metadata for main + companions

**Companion Handling**:
- Automatically detects `<database>_dmart` companion
- Backs up both main and companion to separate files
- Creates combined metadata referencing all files

**Examples**:
```bash
# Backup all primary databases
sfdbtools db-backup primary \
  --ticket "BACKUP-001"

# Backup specific client with companion
sfdbtools db-backup primary \
  --client-code "adaro" \
  --include-dmart \
  --ticket "BACKUP-001"

# Non-interactive with auto-detection
sfdbtools db-backup primary \
  --quiet \
  --client-code "client123" \
  --include-dmart \
  --profile /path/to/profile.cnf.enc \
  --ticket "BACKUP-001"
```

**Output Files Example**:
```
/backup/
├── dbsf_nbc_adaro_prod-server_20260102_143022.sql.zst.enc
├── dbsf_nbc_adaro_dmart_prod-server_20260102_143022.sql.zst.enc
├── dbsf_nbc_adaro_prod-server_20260102_143022.sql.zst.enc.meta.json
└── dbsf_nbc_adaro_prod-server_20260102_143022_users.sql
```

---

### Backup Secondary

**Purpose**: Backup secondary (dev/staging) databases following naming convention

**Use Case**:
- Development environment backups
- Staging database backups
- Testing database backups

**Command**:
```bash
sfdbtools db-backup secondary [flags]
```

**Database Selection**:
- **Pattern**: `dbsf_nbc_<client_code>_secondary[_<instance>]`
- **Exclusions**: Databases without `_secondary` suffix, `_dmart`, `_temp`, `_archive`
- **Optional filters**: 
  - `--client-code`: Filter by client
  - `--instance`: Filter by instance number

**Output**:
- One file per secondary database
- Companion `_dmart` databases (if `--include-dmart` is enabled)

**Examples**:
```bash
# Backup all secondary databases
sfdbtools db-backup secondary \
  --ticket "BACKUP-001"

# Backup specific client's secondary
sfdbtools db-backup secondary \
  --client-code "client123" \
  --ticket "BACKUP-001"

# Backup specific instance
sfdbtools db-backup secondary \
  --client-code "client123" \
  --instance "1" \
  --include-dmart \
  --ticket "BACKUP-001"
```

**Pattern Examples**:
- `dbsf_nbc_adaro_secondary` - matches
- `dbsf_nbc_adaro_secondary_1` - matches with instance "1"
- `dbsf_nbc_adaro_secondary_2` - matches with instance "2"
- `dbsf_nbc_adaro` - does NOT match (no `_secondary`)

---

## Output Files

### File Naming Patterns

| Mode | Pattern |
|------|---------|
| all | `all_<hostname>_<timestamp>_<count>db.sql[.ext]` |
| filter (single) | `combined_<hostname>_<timestamp>_<count>db.sql[.ext]` |
| filter (multi) | `<database>_<hostname>_<timestamp>.sql[.ext]` |
| single | `<database>_<hostname>_<timestamp>.sql[.ext]` |
| primary | `<database>_<hostname>_<timestamp>.sql[.ext]` |
| secondary | `<database>_<hostname>_<timestamp>.sql[.ext]` |

**Extensions**:
- `.sql` - Base SQL dump
- `.zst` / `.gz` / `.xz` / `.zlib` - Compression
- `.enc` - Encryption

**Exclude-data suffix**:
- When `--exclude-data` is used, filename includes `_nodata`
- Example: `all_nodata_prod-server_20260102_143022_23db.sql.zst.enc`

### Generated Files

For each backup, up to 3 files are created:

1. **Backup file** (`.sql[.compression][.enc]`)
   - Main database dump
   - Compressed and/or encrypted

2. **Metadata file** (`.meta.json`)
   - Backup information and statistics
   - Database list, GTID info, warnings
   - Required for restore operations

3. **User grants file** (`_users.sql`)
   - User accounts and permissions
   - Optional (can be disabled)
   - Filtered by backed-up databases

### Directory Structure Example
```
/backup/daily/
├── all_prod-server_20260102_143022_23db.sql.zst.enc
├── all_prod-server_20260102_143022_23db.sql.zst.enc.meta.json
├── all_prod-server_20260102_143022_23db_users.sql
├── dbsf_nbc_client1_prod-server_20260102_143500.sql.zst.enc
├── dbsf_nbc_client1_prod-server_20260102_143500.sql.zst.enc.meta.json
└── dbsf_nbc_client1_prod-server_20260102_143500_users.sql
```

---

## Interactive vs Non-Interactive Mode

### Interactive Mode (Default)

**Characteristics**:
- User prompts for missing required inputs
- Profile selection menu
- Database multi-select (for filter mode)
- Options preview and edit loop
- Confirmation before execution

**Edit Options Menu**:
After showing the options preview, users can select:
- **Lanjutkan** - Proceed with backup
- **Ubah opsi** - Edit options
- **Batalkan** - Cancel

When selecting "Ubah opsi", a submenu appears with:
- Profile (reconnect with new profile)
- Ticket number
- Capture GTID (all/filter single-file modes)
- Export user grants
- Exclude system databases
- Exclude empty databases
- Exclude data (schema only)
- Backup directory
- Filename
- Encryption (toggle + key input)
- Compression (toggle + type + level)
- Kembali (back)

**Example Flow**:
```
1. Profile selection menu
2. Connect to database
3. Ticket input prompt (if not provided)
4. Database filtering/selection
5. Options preview display
6. "Lanjutkan/Ubah opsi/Batalkan" prompt
7. [If "Ubah opsi"] Edit menu loop
8. Backup execution
9. Results display
```

### Non-Interactive Mode

**Characteristics**:
- No user prompts
- Fail-fast validation
- All required inputs must be provided
- Suitable for automation/scheduling

**Required Inputs**:
- `--quiet` flag
- `--profile` - Profile path
- `--profile-key` - Profile key (or ENV)
- `--ticket` - Ticket number
- `--backup-key` - Encryption key (if encryption enabled, or ENV)
- Mode-specific requirements:
  - **single**: `--database`
  - **filter**: `--db` or `--db-file`

**Validation Rules**:
- Missing required inputs cause immediate failure
- No interactive prompts or menus
- Error messages indicate missing parameters

**Example**:
```bash
sfdbtools db-backup all \
  --quiet \
  --profile /etc/profiles/prod.cnf.enc \
  --profile-key "key123" \
  --ticket "BACKUP-2026-001" \
  --backup-key "backupkey123" \
  --backup-dir /backup/daily \
  --compress zstd \
  --compress-level 6
```

---

## Metadata File Structure

Metadata files use a **grouped/hierarchical JSON structure** for better readability:

```json
{
  "backup": {
    "file": "all_prod-server_20260102_143022_23db.sql.zst.enc",
    "type": "all",
    "status": "success",
    "databases": ["app_db", "users_db", "logs_db"],
    "excluded_databases": ["mysql", "information_schema", "sys", "performance_schema"],
    "exclude_data": false
  },
  "time": {
    "start_time": "2026-01-02 14:30:22",
    "end_time": "2026-01-02 14:35:48",
    "duration": "5m26s"
  },
  "file": {
    "size_bytes": 1234567890,
    "size_human": "1.15 GB"
  },
  "compression": {
    "enabled": true,
    "type": "zstd"
  },
  "encryption": {
    "enabled": true
  },
  "source": {
    "hostname": "prod-server.example.com",
    "host": "10.0.1.100",
    "port": 3306
  },
  "replication": {
    "user": "repl_user",
    "password": "repl_password",
    "gtid_info": "0-1-12345 (File: mysql-bin.000123, Pos: 12345678)"
  },
  "version": {
    "mysqldump": "mysqldump  Ver 10.19 Distrib 10.11.6-MariaDB",
    "mariadb": "10.11.6-MariaDB"
  },
  "additional_files": {
    "user_grants": "all_prod-server_20260102_143022_23db_users.sql"
  },
  "ticket": "BACKUP-2026-001",
  "generator": {
    "generated_by": "sfDBTools",
    "generated_at": "2026-01-02T14:35:48Z"
  },
  "warnings": []
}
```

**Status Values**:
- `"success"` - Backup completed without warnings
- `"success_with_warnings"` - Backup completed with mysqldump warnings
- `"dry-run"` - Dry run mode (no actual backup)

**Notes**:
- For multi-file backups (primary/secondary with companions), the main database metadata includes `database_details` array
- `gtid_info` is a combined string (not separate fields)
- `warnings` array contains mysqldump stderr output (if any)

---

## Advanced Features

### 1. GTID Capture

**What**: Global Transaction ID position capture for replication consistency

**When**: Available only for single-file backup modes (all, filter single-file)

**How**: 
- Executes `SHOW MASTER STATUS` before backup
- Captures binlog file name, position, and GTID
- Stores in metadata for restore operations

**Requirements**:
- Binary logging must be enabled on source database
- User must have `REPLICATION CLIENT` privilege

**Configuration**:
- Default: from config file (`backup.capture_gtid`)
- Can be toggled in interactive edit menu
- Stored in `replication.gtid_info` in metadata

### 2. User Grants Export

**What**: Exports user accounts and permissions to SQL file

**Modes**:
- **All mode**: Exports ALL user accounts
- **Filter/Primary/Secondary**: Exports only users with grants to backed-up databases
- **Single**: Exports users with grants to the single database

**Output**: `<backup_file_base>_users.sql`

**Content**:
- User creation statements (with password hashes)
- Grant statements for all privileges
- `FLUSH PRIVILEGES` at the end

**Use Case**: Restore user permissions after database restore

**Disable**: Set `ExcludeUser=true` in config or interactive menu

### 3. Companion Database Handling

**What**: Automatically backup related databases with `_dmart` suffix

**Applicable Modes**: primary, secondary, single (with `--include-dmart`)

**Pattern**: `<main_database>_dmart`

**Behavior**:
- Detects companion database existence
- Backs up to separate file
- Creates combined metadata
- Deletes individual companion metadata

**Example**:
```
Main: dbsf_nbc_client
Companion: dbsf_nbc_client_dmart

Output:
- dbsf_nbc_client_<timestamp>.sql.zst.enc
- dbsf_nbc_client_dmart_<timestamp>.sql.zst.enc
- dbsf_nbc_client_<timestamp>.sql.zst.enc.meta.json (combined)
- dbsf_nbc_client_<timestamp>_users.sql
```

### 4. Streaming Pipeline

**Architecture**: `mysqldump stdout → [compress] → [encrypt] → file`

**Benefits**:
- No large memory buffers
- Efficient for large databases
- Parallel compression (with pgzip)

**Order**: Encryption is applied AFTER compression for better compression ratios

### 5. Retry Strategies

**SSL Mismatch**:
- Detects "SSL is required, but server does not support it"
- Retries with `--skip-ssl` flag

**Unsupported Options**:
- Detects "unknown option" errors
- Removes problematic option
- Retries backup

**Cleanup**: Failed backup files are automatically removed before retry

---

## Troubleshooting

### Error: "ticket wajib diisi pada mode non-interaktif"
**Cause**: `--ticket` flag missing in non-interactive mode
**Solution**: Add `--ticket "TICKET-NUMBER"` to command

### Error: "profile-key wajib diisi pada mode non-interaktif"
**Cause**: Encrypted profile without key
**Solution**: 
- Add `--profile-key "KEY"`, or
- Set ENV `SFDB_SOURCE_PROFILE_KEY="KEY"`

### Error: "backup-key wajib diisi saat enkripsi aktif pada mode non-interaktif"
**Cause**: Encryption enabled without key
**Solution**:
- Add `--backup-key "KEY"`, or
- Set ENV `SFDB_BACKUP_ENCRYPTION_KEY="KEY"`, or
- Use `--skip-encrypt` if encryption not needed

### Error: "database wajib diisi pada mode backup single saat non-interaktif"
**Cause**: Single mode without database name
**Solution**: Add `--database "db_name"`

### Error: "mode backup filter non-interaktif membutuhkan include list"
**Cause**: Filter mode without database selection
**Solution**: Add `--db "db1,db2"` or `--db-file /path/to/list.txt`

### Warning: "SHOW MASTER STATUS tidak mengembalikan hasil"
**Cause**: Binary logging not enabled
**Impact**: GTID capture skipped
**Solution**: Enable binary logging in MySQL/MariaDB config if GTID needed

### Error: "backup tidak mendukung database dengan suffix '_temp' atau '_archive'"
**Cause**: Attempting to backup reserved suffix databases
**Solution**: These databases are intentionally excluded; backup manually if needed

### Backup Too Slow
**Solutions**:
- Use lower compression level (1-3 instead of 9)
- Use faster compression (zstd or pgzip instead of xz)
- Consider `--exclude-data` if only schema needed
- Use `--exclude-empty` to skip empty databases

### Backup File Too Large
**Solutions**:
- Use higher compression level (9)
- Use better compression (xz instead of gzip)
- Use `--exclude-data` for schema-only backup
- Split into multiple backups using filter mode

---

## Best Practices

### Security
1. ✅ **Always use encryption** in production environments
2. ✅ **Store keys securely** (use vault/secrets manager, not plaintext)
3. ✅ **Use ticket system** for audit trail and tracking
4. ✅ **Rotate encryption keys** periodically
5. ✅ **Restrict profile file permissions** (chmod 600)

### Reliability
6. ✅ **Test restore procedures** regularly
7. ✅ **Verify metadata files** after backup completion
8. ✅ **Monitor backup size and duration** for anomalies
9. ✅ **Use `--dry-run`** before scheduling new backup jobs
10. ✅ **Keep multiple backup generations** (daily, weekly, monthly)

### Automation
11. ✅ **Use non-interactive mode** for scheduled backups
12. ✅ **Set environment variables** for credentials (avoid command-line exposure)
13. ✅ **Implement backup monitoring** and alerting
14. ✅ **Log backup results** to centralized logging system
15. ✅ **Use SFDB_QUIET=1** for cleaner automation logs

### Performance
16. ✅ **Use zstd compression** for best balance (speed vs size)
17. ✅ **Use pgzip** for multi-core parallel compression
18. ✅ **Schedule backups** during low-traffic periods
19. ✅ **Monitor server load** during backup operations
20. ✅ **Consider incremental strategies** for very large databases

### Organization
21. ✅ **Follow naming conventions** for consistency
22. ✅ **Organize by date hierarchy** (YYYY/MM/DD structure)
23. ✅ **Document restore procedures** alongside backups
24. ✅ **Tag backups with tickets** for accountability
25. ✅ **Archive old backups** to cheaper storage

---

## Environment Variables

```bash
# Profile encryption key
export SFDB_SOURCE_PROFILE_KEY="your-profile-key"

# Backup file encryption key
export SFDB_BACKUP_ENCRYPTION_KEY="your-backup-key"

# Quiet mode for automation (logs to stderr)
export SFDB_QUIET=1
```

---

## Quick Command Reference

```bash
# Full instance backup
sfdbtools db-backup all --ticket "TICK-001"

# Selective multi-file backup
sfdbtools db-backup filter --mode multi-file --db "db1,db2,db3" --ticket "TICK-001"

# Selective single-file backup
sfdbtools db-backup filter --mode single-file --db-file /path/to/list.txt --ticket "TICK-001"

# Single database backup
sfdbtools db-backup single --database "my_app_db" --ticket "TICK-001"

# Primary databases with companions
sfdbtools db-backup primary --client-code "client123" --include-dmart --ticket "TICK-001"

# Secondary databases
sfdbtools db-backup secondary --client-code "client123" --instance "1" --ticket "TICK-001"

# Non-interactive automated backup
sfdbtools db-backup all \
  --quiet \
  --profile /etc/profiles/prod.cnf.enc \
  --ticket "DAILY-BACKUP-$(date +%Y%m%d)" \
  --compress zstd \
  --compress-level 6

# Schema-only backup
sfdbtools db-backup all --exclude-data --ticket "SCHEMA-BACKUP-001"

# Dry run for testing
sfdbtools db-backup all --dry-run --ticket "TEST-001"
```

---

**For more information**: See [Contributing Guide](contributing_backup_module.md) for developers wanting to extend the backup module.
