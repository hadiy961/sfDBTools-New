# Phase 1 Point 1: Backup Verification & Integrity Check

> **Context:** Saat ini backup hanya dicek dari exit code `mariadb-dump`. Tidak ada validasi apakah file backup benar-benar restorable. DBA tidak tahu apakah file `.sql.zst.enc` yang 50GB itu valid atau korup sampai saatnya restore darurat.

## Keputusan Desain

> [!IMPORTANT]
> **1. Checksum pada file mentah (on-disk), bukan pada SQL content**
> Checksum dihitung terhadap file backup _sebagaimana tersimpan di disk_ (termasuk kompresi & enkripsi). Alasan:
> - Lebih cepat (baca file sekali, tanpa decrypt/decompress)
> - Mendeteksi korupsi di level storage/transfer (bit-rot, NFS glitch)
> - Cocok untuk validasi setelah `rsync`/`scp` ke remote server
>
> Jika file berubah 1 byte pun (korupsi disk, transfer gagal), checksum akan mismatch.

> [!IMPORTANT]
> **2. Header/Footer check = opsional & default OFF untuk post-backup**
> Validasi SQL header (`-- MariaDB dump`) dan footer (`-- Dump completed`) memerlukan streaming decrypt → decompress → baca dari awal sampai akhir. Untuk file besar (>10GB) ini memakan waktu signifikan.
>
> - **Post-backup:** default `header_footer_check: false` (hanya checksum + size)
> - **Manual verify:** selalu aktif (`sfdbtools backup verify <file>`)
>
> DBA bisa enable post-backup header check di config jika bersedia mengorbankan waktu.

> [!WARNING]
> **3. Dry-Run Restore ditunda ke iterasi berikutnya**
> Fitur restore test (parse SQL ke temporary database) memiliki risiko side-effect pada production DB server. Untuk iterasi pertama, fokus pada:
> - ✅ Checksum generation & comparison
> - ✅ File size sanity check
> - ✅ SQL header/footer validation (streaming)
> - ⏳ Dry-run restore → Phase 2 atau sub-task terpisah

## Proposed Changes

---

### Component 1: Verification Config

Extend `VerificationConfig` yang sudah ada di [appconfig_types.go](file:///home/dbado/sfDBTools/internal/services/config/appconfig_types.go) (saat ini hanya `disk_space_check`).

#### [MODIFY] `internal/services/config/appconfig_types.go`

```diff
 type VerificationConfig struct {
     DiskSpaceCheck bool `yaml:"disk_space_check"`
+    // Post-backup verification settings
+    ChecksumAlgorithm string `yaml:"checksum_algorithm"` // "sha256" (default) atau "md5"
+    PostBackupCheck   bool   `yaml:"post_backup_check"`  // auto-verify setelah backup selesai
+    HeaderFooterCheck bool   `yaml:"header_footer_check"` // validasi SQL dump structure (butuh decrypt+decompress)
+    MinFileSize       string `yaml:"min_file_size"`        // minimum acceptable file size (e.g. "1KB", "100B")
 }
```

Config YAML update:
```yaml
backup:
  verification:
    disk_space_check: false
    checksum_algorithm: sha256  # sha256 atau md5
    post_backup_check: true     # auto-verify setelah backup
    header_footer_check: false  # default OFF (butuh streaming full file)
    min_file_size: 1KB          # minimum file size dianggap valid
```

---

### Component 2: Verification Result Model

Tambah field `Verification` di `BackupMetadata` agar hasil verifikasi tersimpan di `.meta.json`.

#### [MODIFY] `internal/app/backup/model/types_backup/results.go`

Tambah struct baru dan field di `BackupMetadata`:

```go
// VerificationResult menyimpan hasil verifikasi integritas backup
type VerificationResult struct {
    ChecksumAlgo   string     `json:"checksum_algo,omitempty"`   // "sha256" atau "md5"
    ChecksumHash   string     `json:"checksum_hash,omitempty"`   // hex-encoded hash
    HeaderValid    *bool      `json:"header_valid,omitempty"`    // nil = not checked
    FooterValid    *bool      `json:"footer_valid,omitempty"`    // nil = not checked
    SizeValid      *bool      `json:"size_valid,omitempty"`      // nil = not checked
    FileSizeBytes  int64      `json:"file_size_bytes,omitempty"` // actual file size saat verify
    VerifiedAt     *time.Time `json:"verified_at,omitempty"`     // timestamp verifikasi
    VerifyStatus   string     `json:"verify_status,omitempty"`   // "passed", "failed", "partial"
    FailureReason  string     `json:"failure_reason,omitempty"`  // alasan jika gagal
}
```

Tambahkan field di `BackupMetadata`:
```diff
 type BackupMetadata struct {
     ...
     SourcePort          int    `json:"source_port,omitempty"`
+    Verification        *VerificationResult `json:"verification,omitempty"`
 }
```

Update `MarshalJSON` dan `UnmarshalJSON` untuk menyertakan group `"verification"` di output JSON.

---

### Component 3: Core Verification Package

#### [NEW] `internal/app/backup/verify/checker.go`

Main orchestrator. Menjalankan semua check dan menghasilkan `VerificationResult`.

```go
// CheckOptions mengontrol check mana yang dijalankan
type CheckOptions struct {
    Checksum        bool   // generate/compare checksum
    ChecksumAlgo    string // "sha256" atau "md5"
    HeaderFooter    bool   // validasi SQL header/footer (butuh streaming)
    SizeCheck       bool   // validasi minimum file size
    MinFileSize     int64  // minimum size dalam bytes
    ExpectedHash    string // jika non-empty, compare dengan hash ini
    EncryptionKey   string // key untuk decrypt (jika header/footer check pada .enc file)
}

// Check menjalankan verifikasi pada satu backup file
func Check(filePath string, opts CheckOptions, logger applog.Logger) (*types_backup.VerificationResult, error)

// CheckFromMetadata menjalankan verifikasi menggunakan info dari .meta.json
// Otomatis mendeteksi compression type dan encryption dari metadata
func CheckFromMetadata(metaPath string, opts CheckOptions, logger applog.Logger) (*types_backup.VerificationResult, error)
```

Flow:
1. Validasi file exists & readable
2. **Size check** → `os.Stat()`, bandingkan dengan `MinFileSize`
3. **Checksum** → streaming hash (baca file mentah on-disk)
4. **Header/Footer** (jika enabled) → buka reader pipeline (decrypt → decompress → scan)
5. Compile hasil ke `VerificationResult`

#### [NEW] `internal/app/backup/verify/checksum.go`

Streaming checksum generation. Baca file sekali, hasilkan hex hash.

```go
// GenerateChecksum menghitung hash dari file on-disk (tanpa decrypt/decompress)
// Menggunakan io.Copy dari file ke hash.Hash untuk memory-efficient streaming
func GenerateChecksum(filePath string, algo string) (string, error)

// CompareChecksum menghitung hash dan membandingkan dengan expected
func CompareChecksum(filePath string, algo string, expected string) (match bool, actual string, err error)
```

Implementasi:
- `sha256`: `crypto/sha256` → `io.Copy(hasher, file)` → `hex.EncodeToString`
- `md5`: `crypto/md5` → sama
- Buffer: 256KB read buffer untuk throughput optimal

#### [NEW] `internal/app/backup/verify/header_validator.go`

Validasi SQL dump structure. **Memerlukan reader pipeline** (decrypt → decompress) karena harus baca SQL content.

```go
// ValidateHeader membaca N bytes pertama dari stream dan cek apakah mengandung
// signature MariaDB/MySQL dump (e.g. "-- MariaDB dump", "-- MySQL dump")
func ValidateHeader(reader io.Reader) (bool, error)

// ValidateFooter membaca stream sampai habis dan cek apakah N bytes terakhir
// mengandung "-- Dump completed" marker
func ValidateFooter(reader io.Reader) (bool, error)

// ValidateHeaderFooter menjalankan keduanya dalam satu pass
// Baca header dari awal, lalu stream sampai akhir sambil track tail buffer
func ValidateHeaderFooter(reader io.Reader) (headerOK bool, footerOK bool, err error)
```

Strategi untuk footer (tanpa seek pada compressed stream):
- Gunakan **rolling tail buffer** (simpan 4KB terakhir) saat streaming
- Setelah EOF, cek apakah tail buffer mengandung `-- Dump completed`
- Ini berarti kita baca seluruh stream sekali, tapi tidak perlu simpan semuanya di memory

#### [NEW] `internal/app/backup/verify/size_validator.go`

Validasi file size.

```go
// ParseMinFileSize mengkonversi string size (e.g. "1KB", "100B", "5MB") ke bytes
func ParseMinFileSize(sizeStr string) (int64, error)

// ValidateSize mengecek apakah file size >= minimum threshold
func ValidateSize(filePath string, minSize int64) (bool, int64, error)
```

#### [NEW] `internal/app/backup/verify/reader.go`

Helper untuk membuka reader pipeline (reuse pattern dari [mysql.go](file:///home/dbado/sfDBTools/internal/app/restore/helpers/mysql.go#L120-L157)).

```go
// OpenVerifyReader membuka file dan menyiapkan decrypt → decompress reader
// Digunakan untuk header/footer validation
// Reuse pattern dari restore/helpers.OpenAndPrepareReader
func OpenVerifyReader(filePath string, encryptionKey string) (io.Reader, []io.Closer, error)
```

#### [NEW] `internal/app/backup/verify/report.go`

Display hasil verifikasi ke terminal.

```go
// DisplayResult menampilkan hasil verifikasi dalam format tabel
func DisplayResult(result *types_backup.VerificationResult, filePath string)

// DisplayBatchResults menampilkan hasil verifikasi batch (directory scan)
func DisplayBatchResults(results map[string]*types_backup.VerificationResult)
```

Output contoh:
```
📋 Hasil Verifikasi Backup
--------------------------------------------------------------------------------
┌──────────────────┬───────────────────────────────────────────────────────┐
│ Parameter        │ Nilai                                                 │
├──────────────────┼───────────────────────────────────────────────────────┤
│ File             │ dbsaas_host_20260424_153025_dbserver1.sql.zst.enc     │
│ Status           │ ✓ PASSED                                             │
│ Checksum (SHA256)│ a3f2b1c4d5e6...                                      │
│ File Size        │ 152.3 MB (valid, min: 1 KB)                          │
│ SQL Header       │ ✓ Valid (-- MariaDB dump 11.4.5)                     │
│ SQL Footer       │ ✓ Valid (-- Dump completed on 2026-04-24 15:30:25)   │
│ Verified At      │ 2026-04-24 15:31:02                                  │
└──────────────────┴───────────────────────────────────────────────────────┘
```

---

### Component 4: CLI Command

#### [NEW] `cmd/backup/verify.go`

Registers `sfdbtools backup verify` dengan subcommands dan flags.

```bash
# Verify single file (full check: checksum + size + header/footer)
sfdbtools backup verify <backup-file>

# Verify semua backup di directory
sfdbtools backup verify --dir /media/ArchiveDB/20260424/

# Verify backup terbaru
sfdbtools backup verify --latest --profile localhost_3306.cnf.enc

# Hanya checksum (cepat, tanpa decrypt/decompress)
sfdbtools backup verify --checksum-only <backup-file>

# Compare dengan expected hash (untuk validasi setelah transfer)
sfdbtools backup verify --expected-hash <sha256hex> <backup-file>

# Output format JSON (untuk scripting/monitoring)
sfdbtools backup verify --format json <backup-file>
```

Flags:
| Flag | Default | Deskripsi |
|------|---------|-----------|
| `--dir` | - | Verify semua file di directory |
| `--latest` | false | Verify backup terbaru |
| `--profile` | - | Profile untuk `--latest` lookup |
| `--checksum-only` | false | Hanya generate checksum (tanpa header/footer) |
| `--expected-hash` | - | Expected hash untuk comparison |
| `--format` | table | Output format: `table` atau `json` |
| `--algo` | sha256 | Override checksum algorithm |

---

### Component 5: Integration ke Backup Engine

#### [MODIFY] [builder.go](file:///home/dbado/sfDBTools/internal/app/backup/execution/builder.go)

Setelah backup selesai dan metadata dihasilkan, jalankan post-backup verification jika `post_backup_check: true`.

```diff
 func (e *Engine) buildRealBackupInfo(...) types_backup.DatabaseBackupInfo {
     status := determineBackupStatus(writeResult, cfg, e.Log)
     duration := timer.Elapsed()
     endTime := time.Now()
     meta := e.generateBackupMetadata(cfg, writeResult, duration, startTime, endTime, status, dbVersion)

+    // Post-backup verification
+    if e.Config.Backup.Verification.PostBackupCheck && status != consts.BackupStatusFailed {
+        verifyResult := verify.PostBackupCheck(cfg.OutputPath, e.Config.Backup.Verification, e.Log)
+        meta.Verification = verifyResult
+    }

     manifestPath := ""
     if e.Config.Backup.Output.SaveBackupInfo {
         manifestPath = metadata.TrySaveBackupMetadata(meta, ...)
     }
     ...
 }
```

#### [NEW] `internal/app/backup/verify/post_backup.go`

Convenience function khusus untuk post-backup check (subset dari full verify).

```go
// PostBackupCheck menjalankan verifikasi ringan setelah backup selesai
// Default: checksum + size check saja (tanpa header/footer)
func PostBackupCheck(filePath string, cfg appconfig.VerificationConfig, logger applog.Logger) *types_backup.VerificationResult
```

#### [MODIFY] [main.go](file:///home/dbado/sfDBTools/cmd/backup/main.go)

Register command baru:
```diff
 func init() {
     ...
     CmdBackupMain.AddCommand(CmdBackupSchedule)
+    CmdBackupMain.AddCommand(CmdBackupVerify)
 }
```

#### [MODIFY] `MarshalJSON` dan `UnmarshalJSON` di [results.go](file:///home/dbado/sfDBTools/internal/app/backup/model/types_backup/results.go)

Tambahkan `Verification` group di output JSON `.meta.json`:
```diff
 metaJSON := struct {
     ...
     Warnings    []string               `json:"warnings,omitempty"`
+    Verification *types_backup.VerificationResult `json:"verification,omitempty"`
 }{
     ...
     Warnings: b.Warnings,
+    Verification: b.Verification,
 }
```

---

## File Structure Summary

```
internal/app/backup/verify/
├── checker.go          # Core verification orchestrator (Check, CheckFromMetadata)
├── checksum.go         # SHA-256/MD5 streaming hash generation
├── header_validator.go # SQL header/footer validation (rolling tail buffer)
├── size_validator.go   # File size sanity check + ParseMinFileSize
├── reader.go           # Reader pipeline helper (decrypt → decompress)
├── post_backup.go      # Convenience function untuk auto-verify post-backup
└── report.go           # Display verification results (table/JSON)

cmd/backup/
├── verify.go           # CLI: sfdbtools backup verify
```

## Verification Plan

### Automated Tests

1. **Checksum accuracy:**
   - Buat file test, generate checksum dengan `sha256sum` (OS tool), bandingkan dengan output `verify.GenerateChecksum()`
   - Modify 1 byte di file, pastikan checksum mismatch terdeteksi

2. **Header/Footer validation:**
   - Buat `.sql` sederhana dengan header/footer valid → pastikan pass
   - Buat `.sql` tanpa footer (simulasi backup terpotong) → pastikan fail
   - Buat `.sql.zst` (compressed) dengan content valid → pastikan pass setelah decompress

3. **Size validation:**
   - File 0 bytes → fail
   - File < min threshold → fail
   - File > min threshold → pass

4. **Post-backup integration:**
   - Jalankan `sfdbtools backup single --db dbsaas_host`
   - Baca `.meta.json`, pastikan field `verification.checksum_hash` terisi
   - Pastikan `verification.verify_status` = `"passed"`

### Manual Verification

1. **CLI verify:**
   ```bash
   sfdbtools backup verify /backup/test/dbsaas_host_20260424.sql.zst.enc
   ```
   Pastikan output tabel muncul dengan benar.

2. **Checksum comparison setelah transfer:**
   ```bash
   sfdbtools backup verify --expected-hash $(cat backup.sha256) backup.sql.zst.enc
   ```

3. **Directory scan:**
   ```bash
   sfdbtools backup verify --dir /backup/test/
   ```
   Pastikan semua file diproses dan hasilnya ditampilkan dalam batch summary.
