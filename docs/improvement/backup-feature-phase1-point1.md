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

> [!IMPORTANT]
> **3. Dual-Mode: Interaktif & Non-Interaktif**
> Command `backup verify` mendukung dua mode:
>
> - **Interaktif (default):** Wizard dengan `survey` prompt — DBA dipandu memilih target, algorithm, dan opsi verifikasi.
> - **Non-interaktif (flags/args):** Semua parameter lewat flags — cocok untuk scripting, cron, pipeline, dan scheduler systemd.
>
> Pattern: jika `args == 0` dan tidak ada flags yang di-set → mode interaktif. Jika ada `args[0]` (file path) atau `--dir`/`--latest` → mode non-interaktif. Konsisten dengan command lain (`backup filter`, `profile create`, dll).

> [!IMPORTANT]
> **4. Go Best Practice: Thin `cmd/` Layer**
> File `cmd/backup/verify.go` harus **tipis** — hanya berisi:
> - Flag registration (`init()`)
> - Mode detection (interaktif vs non-interaktif)
> - Survey prompts (karena ini UI concern, boleh di `cmd/`)
> - Delegasi ke `internal/app/backup/verify/` untuk semua business logic
>
> **Yang TIDAK boleh ada di `cmd/`:**
> - Directory scanning / file filtering (`os.ReadDir`, extension check)
> - Option building / config mapping
> - Result aggregation / batch processing
>
> Semua itu harus di package `internal/app/backup/verify/`. Pattern ini konsisten dengan command lain seperti `backup filter` yang mendelegasikan ke `internal/app/backup`.
>
> ```
> cmd/backup/verify.go           → parse flags, detect mode, call internal
> internal/app/backup/verify/     → ALL business logic
> ```

> [!WARNING]
> **5. Dry-Run Restore ditunda ke iterasi berikutnya**
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
// Menggunakan io.CopyBuffer dari file ke hash.Hash untuk memory-efficient streaming
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
// DisplayResult menampilkan hasil verifikasi dalam format tabel atau JSON
func DisplayResult(result *types_backup.VerificationResult, filePath string, format string)

// DisplayBatchResults menampilkan hasil verifikasi batch (directory scan)
func DisplayBatchResults(results map[string]*types_backup.VerificationResult, format string)
```

Output contoh (format table):
```
📋 Hasil Verifikasi Backup
────────────────────────────────────────────────────────────────────────────
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

Output contoh (format json):
```json
{
  "file": "dbsaas_host_20260424_153025_dbserver1.sql.zst.enc",
  "verify_status": "passed",
  "checksum_algo": "sha256",
  "checksum_hash": "a3f2b1c4d5e6...",
  "header_valid": true,
  "footer_valid": true,
  "size_valid": true,
  "file_size_bytes": 159621120,
  "verified_at": "2026-04-24T15:31:02+07:00"
}
```

Batch output contoh (format table):
```
📋 Hasil Verifikasi Batch — /backup/test/
────────────────────────────────────────────────────────────────────────────
┌─────────────────────────────────────────┬────────┬──────────┬────────┬────────┐
│ FILE                                    │ STATUS │ CHECKSUM │ HEADER │ FOOTER │
├─────────────────────────────────────────┼────────┼──────────┼────────┼────────┤
│ dbsaas_host_20260424.sql.zst.enc        │ ✓ PASS │ a3f2b1.. │ ✓      │ ✓      │
│ dbsf_biznet_20260424.sql.zst.enc        │ ✓ PASS │ 7e9c4d.. │ ✓      │ ✓      │
│ mysql_20260424.sql.zst.enc              │ ✗ FAIL │ -        │ ✓      │ ✗      │
└─────────────────────────────────────────┴────────┴──────────┴────────┴────────┘

Summary: 2/3 passed, 1 failed
```

---

### Component 4: CLI Command (Dual-Mode: Interaktif & Non-Interaktif)

#### [REFACTOR] `cmd/backup/verify.go`

> [!WARNING]
> **Refactoring Required:** File `verify.go` saat ini mencampur business logic (directory scanning, file extension filtering, option building, result aggregation) langsung di `cmd/` layer. Ini harus di-refactor agar `cmd/` hanya menjadi **thin wrapper**.

##### Tanggung Jawab `cmd/backup/verify.go` (seharusnya):

```go
// cmd/backup/verify.go — THIN LAYER ONLY
//
// ✅ Yang boleh ada di sini:
//   - Flag registration (init)
//   - Mode detection (interaktif vs non-interaktif)
//   - Survey prompts (UI concern)
//   - Build CheckOptions dari flags/prompts
//   - Panggil verify.CheckSingle() / verify.CheckBatch()
//   - Panggil verify.DisplayResult() / verify.DisplayBatchResults()
//   - Exit code handling
//
// ❌ Yang TIDAK boleh ada di sini:
//   - os.ReadDir / filepath.Walk (→ pindah ke verify.CheckBatch)
//   - File extension filtering (→ pindah ke verify.IsBackupFile)
//   - Result map aggregation (→ pindah ke verify.CheckBatch)
//   - getLatestBackup logic (→ pindah ke verify atau catalog package)
```

##### Refactoring: Pindahkan logic ke `internal/`

**Sebelum (❌ current — semua di `cmd/`):**
```go
// cmd/backup/verify.go — terlalu banyak logic
func runBatchVerify(dirPath string) {
    files, _ := os.ReadDir(dirPath)           // ❌ business logic
    for _, f := range files {
        ext := filepath.Ext(f.Name())          // ❌ business logic  
        if ext == ".sql" || ext == ".zst" ... { // ❌ business logic
            res, _ := verify.Check(filePath, opts, logger)
            results[filePath] = res             // ❌ aggregation logic
        }
    }
    verify.DisplayBatchResults(results, format)
}
```

**Sesudah (✅ thin cmd layer):**
```go
// cmd/backup/verify.go — thin wrapper
func runBatchVerify(dirPath string) {
    opts := buildCheckOptionsFromFlags()  // ✅ flag → opts mapping
    results, err := verify.CheckBatch(dirPath, opts, getLogger())  // ✅ delegasi
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }
    verify.DisplayBatchResults(results, verifyFormat)  // ✅ delegasi
    if verify.HasFailures(results) {
        os.Exit(1)
    }
}
```

##### Fungsi baru di `internal/app/backup/verify/`:

```go
// checker.go — tambahkan batch verification

// CheckBatch melakukan verifikasi pada semua backup files di directory
// Handles: directory scanning, file filtering, result aggregation
func CheckBatch(dirPath string, opts CheckOptions, logger applog.Logger) (map[string]*VerificationResult, error)

// IsBackupFile mengecek apakah file adalah backup file berdasarkan extension
// Mendukung: .sql, .gz, .zst, .enc, .zip, dan kombinasinya
func IsBackupFile(filename string) bool

// HasFailures mengecek apakah ada failure di batch results
func HasFailures(results map[string]*VerificationResult) bool
```

##### Struktur akhir `cmd/backup/verify.go`:

```go
var CmdBackupVerify = &cobra.Command{
    Run: func(cmd *cobra.Command, args []string) {
        // 1. Mode detection
        if verifyDir != "" {
            runBatchVerify(verifyDir)
            return
        }
        if len(args) == 0 && !verifyLatest {
            runInteractiveVerify()  // survey prompts → build opts → delegasi
            return
        }
        // 2. Resolve target file
        targetFile := resolveTargetFile(args)
        // 3. Build opts dari flags → delegasi ke internal
        runSingleVerify(targetFile)
    },
}

// runSingleVerify — thin: build opts + call verify.Check + display
func runSingleVerify(targetFile string) { ... }

// runBatchVerify — thin: call verify.CheckBatch + display
func runBatchVerify(dirPath string) { ... }

// runInteractiveVerify — survey prompts → build opts → call verify.Check
func runInteractiveVerify() { ... }

// buildCheckOptionsFromFlags — map CLI flags ke verify.CheckOptions
func buildCheckOptionsFromFlags() verify.CheckOptions { ... }
```

Registers `sfdbtools backup verify` — **Verifikasi integritas file backup.**

Mode detection logic:
- `args[0]` ada (file path) → **non-interaktif**, single file verify
- `--dir` flag di-set → **non-interaktif**, batch verify
- `--latest` flag di-set → **non-interaktif**, latest backup verify
- Tidak ada args dan tidak ada flags → **interaktif**, wizard

##### Non-Interaktif (flags/args)

```bash
# Verify single file (full check: checksum + size + header/footer)
sfdbtools backup verify /backup/test/dbsaas_host_20260424.sql.zst.enc

# Verify semua backup di directory
sfdbtools backup verify --dir /media/ArchiveDB/20260424/

# Verify backup terbaru
sfdbtools backup verify --latest --profile localhost_3306.cnf.enc

# Hanya checksum (cepat, tanpa decrypt/decompress)
sfdbtools backup verify --checksum-only /backup/test/dbsaas_host.sql.zst.enc

# Compare dengan expected hash (untuk validasi setelah rsync/scp)
sfdbtools backup verify --expected-hash a3f2b1c4d5e6... /backup/test/dbsaas_host.sql.zst.enc

# Output format JSON (untuk scripting/monitoring)
sfdbtools backup verify --format json /backup/test/dbsaas_host.sql.zst.enc

# Encryption key untuk header/footer check pada file .enc
sfdbtools backup verify --encryption-key mySecretKey /backup/test/dbsaas_host.sql.zst.enc

# Pipeline: verify dan pipe JSON ke monitoring tool
sfdbtools backup verify --format json --dir /backup/test/ | jq '.[] | select(.verify_status == "failed")'

# Cron job: verify terbaru dan exit code non-zero jika gagal
sfdbtools backup verify --latest --profile localhost_3306.cnf.enc --checksum-only
```

##### Interaktif (wizard)

Jika dijalankan tanpa arguments dan tanpa flags:

```
=== Interactive Backup Verification ===

? Pilih target verifikasi: (Use arrow keys)
  > Single File
    Directory (Batch)

? Masukkan path file/directory: /backup/test/dbsaas_host_20260424.sql.zst.enc

? Lakukan validasi struktur SQL (Header/Footer)? (Y/n) Yes

? Apakah backup ini dienkripsi (.enc)? (y/N) Yes
? Masukkan Encryption Key: ********

? Pilih Algoritma Checksum: (Use arrow keys)
  > sha256
    md5
    skip (tidak generate checksum)

Memulai verifikasi...

📋 Hasil Verifikasi Backup
────────────────────────────────────────────────────────────────────────────
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

##### Flags

| Flag | Default | Deskripsi |
|------|---------|-----------|
| `--dir` | - | Verify semua file di directory |
| `--latest` | false | Verify backup terbaru |
| `--profile` | - | Profile untuk `--latest` lookup |
| `--checksum-only` | false | Hanya generate checksum (tanpa header/footer) |
| `--expected-hash` | - | Expected hash untuk comparison |
| `--format` | table | Output format: `table` atau `json` |
| `--algo` | sha256 | Override checksum algorithm |
| `--encryption-key` | - | Kunci dekripsi untuk header/footer check pada file .enc |
| `--min-size` | 0 | Override minimum file size (bytes) |

##### Exit Code Behavior

| Scenario | Exit Code | Gunanya |
|----------|-----------|---------|
| Semua check passed | 0 | Scripting: `if sfdbtools backup verify ...; then ...` |
| Ada check yang failed | 1 | Monitoring: alert jika backup korup |
| File tidak ditemukan / error fatal | 1 | Error handling |

> [!TIP]
> **Exit code non-zero pada failure** memungkinkan integrasi dengan monitoring tools (Nagios, Zabbix, cron mailto) tanpa perlu parsing output.

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

Convenience function khusus untuk post-backup check (subset dari full verify). Ini selalu **non-interaktif** karena dipanggil dari engine.

```go
// PostBackupCheck menjalankan verifikasi ringan setelah backup selesai
// Default: checksum + size check saja (tanpa header/footer, kecuali config enable)
// Selalu non-interaktif (dipanggil dari engine, bukan CLI)
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
├── post_backup.go      # Convenience function untuk auto-verify post-backup (non-interaktif)
└── report.go           # Display verification results (table/JSON)

cmd/backup/
├── verify.go           # CLI: sfdbtools backup verify (interaktif + non-interaktif)
```

## Interaktif vs Non-Interaktif — Summary

| Scenario | Mode | Trigger |
|----------|------|---------|
| `sfdbtools backup verify` (tanpa args) | **Interaktif** | Wizard: pilih target → header/footer? → encryption key → algo |
| `sfdbtools backup verify <file>` | **Non-interaktif** | Full verify langsung, flags optional |
| `sfdbtools backup verify --dir <path>` | **Non-interaktif** | Batch verify, semua file di directory |
| `sfdbtools backup verify --latest --profile <p>` | **Non-interaktif** | Verify backup terbaru |
| `sfdbtools backup verify --checksum-only <file>` | **Non-interaktif** | Hanya checksum (cepat) |
| Post-backup auto-verify (engine) | **Non-interaktif** | Otomatis, config-driven |

> [!TIP]
> **Pattern Detection:** Mode interaktif hanya aktif jika `len(args) == 0` DAN tidak ada flags `--dir` / `--latest` yang di-set. Ini berbeda dari pattern `--quiet` yang dipakai di `filter.go` karena verify command sudah punya argumen posisional (file path) sebagai indikator non-interaktif.

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

5. **Exit code:**
   - Verify file valid → exit 0
   - Verify file korup → exit 1

### Manual Verification

1. **Interaktif — verify wizard:**
   ```bash
   sfdbtools backup verify
   # Pastikan wizard muncul: pilih target → header/footer → algo
   ```

2. **Non-interaktif — single file:**
   ```bash
   sfdbtools backup verify /backup/test/dbsaas_host_20260424.sql.zst.enc
   # Pastikan output tabel muncul tanpa prompt
   ```

3. **Non-interaktif — batch directory:**
   ```bash
   sfdbtools backup verify --dir /backup/test/
   # Pastikan semua file diproses dan batch summary ditampilkan
   ```

4. **Non-interaktif — checksum comparison (post-transfer):**
   ```bash
   sfdbtools backup verify --expected-hash $(cat backup.sha256) backup.sql.zst.enc
   # Pastikan match/mismatch terdeteksi
   ```

5. **Non-interaktif — JSON output (scripting):**
   ```bash
   sfdbtools backup verify --format json /backup/test/dbsaas_host.sql.zst.enc
   # Pastikan output JSON valid dan parseable
   ```

6. **Post-backup auto-verify:**
   ```bash
   sfdbtools backup single --db dbsaas_host --profile localhost_3306
   cat /backup/test/*.meta.json | jq '.verification'
   # Pastikan checksum_hash dan verify_status terisi
   ```
