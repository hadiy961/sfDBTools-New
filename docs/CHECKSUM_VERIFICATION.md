# Backup Checksum Verification - Feature Documentation

## Overview
Fitur checksum verification adalah implementasi best practice untuk memastikan integritas backup files. Fitur ini menghitung dan memverifikasi checksum secara otomatis saat proses backup berlangsung.

## Best Practices yang Diimplementasikan

### 1. **Multi-Hash Algorithm**
- **SHA256**: Cryptographically secure hash untuk security dan integrity
- **MD5**: Faster hash untuk backward compatibility dan quick verification
- Kedua hash dihitung secara bersamaan untuk flexibility

### 2. **Streaming Hash Calculation**
- Hash dihitung saat data ditulis (on-the-fly)
- Memory-efficient bahkan untuk file backup yang sangat besar
- Tidak perlu re-reading file untuk kalkulasi checksum
- Menggunakan `MultiHashWriter` wrapper di backup pipeline

### 3. **Separate Checksum Files**
- Format: `backup_file.sql.gz.sha256` dan `backup_file.sql.gz.md5`
- Compatible dengan standard tools: `sha256sum -c` dan `md5sum -c`
- Format: `<hash>  <filename>` (space ganda sesuai standard)
- Portable dan mudah di-share/verify secara manual

### 4. **Post-Backup Verification** (Optional)
- Automatic re-reading dan verification setelah backup selesai
- Mendeteksi corruption saat write process
- Configurable via `backup.verification.verify_after_write`

### 5. **Comprehensive Metadata Manifest**
- JSON manifest dengan complete backup information
- Format: `backup_file.sql.gz.meta.json`
- Contains: checksums, timestamps, database list, compression info, dll

## Architecture

### File Structure
```
internal/backup/
├── backup_checksum_writer.go      # MultiHashWriter untuk streaming hash
├── backup_checksum_verifier.go    # Verification logic
├── backup_checksum_files.go       # Generator untuk checksum files & metadata
├── backup_writer.go               # Integration dengan backup pipeline
├── backup_combined.go             # Combined backup dengan checksum
└── backup_separated.go            # Separated backup dengan checksum
```

### Data Flow
```
mysqldump → MultiHashWriter → Compression → Encryption → File
                    ↓
         (Calculate SHA256 & MD5 dari SQL mentah)
                    ↓
    Write .sha256 & .md5 files
                    ↓
     Write .meta.json manifest
                    ↓
       (Optional) Verify checksums (hanya jika tidak terenkripsi)
```

**PENTING:** 
- Checksum dihitung dari **data SQL mentah** (sebelum compression/encryption)
- Encryption menggunakan random salt/nonce, sehingga setiap kali hasil berbeda
- Post-backup verification **TIDAK DAPAT dilakukan** untuk file terenkripsi
- Untuk verify encrypted backup, harus decrypt dulu kemudian bandingkan checksum

### Checksum Scope

**Checksum mewakili:**
- ✅ Integritas data SQL asli dari mysqldump
- ✅ Dapat digunakan untuk compare antar backup yang sama
- ✅ Detect corruption dalam data source

**Checksum TIDAK mewakili:**
- ❌ Integritas file encrypted/compressed di disk (karena random salt)
- ❌ Tidak bisa verify dengan membaca file encrypted langsung

## Configuration

### Enable/Disable Checksum
File: `config/sfDBTools_config.yaml`
```yaml
backup:
    verification:
        compare_checksums: true      # Enable checksum calculation
        verify_after_write: true     # Enable post-backup verification
        disk_space_check: true       # Check disk space before backup
```

### Settings Impact
- `compare_checksums: true` → Generate checksums saat backup
- `compare_checksums: false` → Skip checksum calculation (faster)
- `verify_after_write: true` → Re-read dan verify setelah backup
- `verify_after_write: false` → Skip verification (faster)

## Generated Files

### 1. Backup File
```
dbsf_biznet_20251105_143022_localhost.sql.gz.enc
```

### 2. SHA256 Checksum File
```
dbsf_biznet_20251105_143022_localhost.sql.gz.enc.sha256
```
Content:
```
a1b2c3d4e5f6...  dbsf_biznet_20251105_143022_localhost.sql.gz.enc
```

### 3. MD5 Checksum File
```
dbsf_biznet_20251105_143022_localhost.sql.gz.enc.md5
```
Content:
```
9876543210abc...  dbsf_biznet_20251105_143022_localhost.sql.gz.enc
```

### 4. Metadata Manifest
```
dbsf_biznet_20251105_143022_localhost.sql.gz.enc.meta.json
```
Content:
```json
{
  "backup_file": "/path/to/backup.sql.gz.enc",
  "backup_type": "separated",
  "database_names": ["dbsf_biznet"],
  "hostname": "localhost",
  "backup_start_time": "2025-11-05T14:30:22Z",
  "backup_end_time": "2025-11-05T14:35:18Z",
  "backup_duration": "4m56s",
  "file_size_bytes": 1073741824,
  "file_size_human": "1.00 GB",
  "compressed": true,
  "compression_type": "gzip",
  "encrypted": true,
  "checksums": [
    {
      "algorithm": "sha256",
      "hash": "a1b2c3d4...",
      "calculated_at": "2025-11-05T14:35:18Z",
      "file_size": 1073741824
    },
    {
      "algorithm": "md5",
      "hash": "9876543210...",
      "calculated_at": "2025-11-05T14:35:18Z",
      "file_size": 1073741824
    }
  ],
  "mariadb_version": "10.6.23",
  "backup_status": "success",
  "generated_by": "sfDBTools v1.0.0",
  "generated_at": "2025-11-05T14:35:18Z"
}
```

## Manual Verification

### Using Standard Tools
```bash
# Verify SHA256
sha256sum -c backup_file.sql.gz.enc.sha256

# Verify MD5
md5sum -c backup_file.sql.gz.enc.md5

# Expected output on success:
# backup_file.sql.gz.enc: OK
```

### Using sfDBTools (Future)
```bash
# Verify single file
sfdbtools verify --file backup_file.sql.gz.enc

# Verify all backups in directory
sfdbtools verify --dir /backups/2025/11/05/
```

## Performance Impact

### Checksum Calculation Overhead
- **SHA256**: ~2-3% additional CPU usage
- **MD5**: ~1-2% additional CPU usage
- **Combined**: ~3-5% total overhead
- **Memory**: Minimal (streaming calculation)

### Verification Overhead
- Requires re-reading entire backup file
- Time: Typically 20-30% of backup time
- Only runs if `verify_after_write: true`

## Use Cases

### 1. Production Backups
```yaml
verification:
    compare_checksums: true
    verify_after_write: true   # Extra safety
```

### 2. Fast Development Backups
```yaml
verification:
    compare_checksums: true
    verify_after_write: false  # Skip verification for speed
```

### 3. Minimal Overhead
```yaml
verification:
    compare_checksums: false   # Fastest, no checksums
    verify_after_write: false
```

## Security Benefits

1. **Integrity Verification**: Detect file corruption or tampering
2. **Transfer Validation**: Ensure backups transferred correctly
3. **Long-term Storage**: Verify backups haven't degraded over time
4. **Audit Trail**: Metadata provides complete backup history

## Implementation Details

### MultiHashWriter
```go
type MultiHashWriter struct {
    writer     io.Writer
    sha256Hash hash.Hash
    md5Hash    hash.Hash
    bytesWritten int64
}
```
- Implements `io.Writer` interface
- Updates both hashes on every Write() call
- No memory buffering of data

### Checksum Storage
```go
type ChecksumInfo struct {
    Algorithm     string
    Hash          string
    CalculatedAt  time.Time
    FileSize      int64
    VerifiedAt    time.Time
    VerifyStatus  string
}
```

## Error Handling

### Checksum Generation Failure
- Log warning message
- Backup continues successfully
- No checksum files generated

### Verification Failure
- Log error message
- Mark backup status as "success_with_warnings"
- Backup file retained for manual inspection
- Checksum files retained for debugging

## Future Enhancements

1. **Incremental Checksums**: Track checksums over time
2. **Checksum Database**: Store checksums in central DB
3. **Automated Verification**: Periodic background verification
4. **Remote Verification**: Verify backups on remote storage
5. **Checksum Comparison**: Compare backups across environments

## Troubleshooting

### Checksum Not Generated
- Check `compare_checksums: true` in config
- Verify write permissions on backup directory
- Check logs for error messages

### Verification Failed
- File may be corrupted during write
- Storage device issues
- Retry backup with same settings
- Check disk health

### Performance Issues
- Disable `verify_after_write` for large files
- Consider disabling checksums for very large backups
- Use faster compression algorithms
