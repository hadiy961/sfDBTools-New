# sfDBTools

**sfDBTools** adalah CLI (Command Line Interface) utility yang komprehensif untuk manajemen operasi database MySQL/MariaDB. Alat ini dirancang untuk lingkungan production dengan fokus pada keamanan data, kemudahan penggunaan, dan otomatisasi.

Dibangun menggunakan **Go (Golang)**, alat ini mendukung operasi backup dan restore dengan fitur enkripsi (AES-256), kompresi multi-format (Zstd, Gzip, XZ), dan manajemen profil koneksi yang aman.

## 🚀 Fitur Utama

### 🛡️ Backup Database
- **Multi-Mode**:
  - `single`: Backup satu database spesifik.
  - `all`: Backup seluruh database dalam server.
  - `filter`: Backup beberapa database terpilih (interactive/flag).
  - `primary` & `secondary`: Mode khusus untuk topologi database spesifik (mendukung companion DBs seperti `_dmart`).
- **Keamanan**: Enkripsi AES-256-GCM (OpenSSL compatible) untuk file output.
- **Efisiensi**: Streaming pipeline (mysqldump -> compress -> encrypt -> file) tanpa buffering memori besar.
- **Kompresi**: Mendukung `zstd` (recommended), `gzip`, `pgzip` (parallel gzip), `xz`, `zlib`.
- **Metadata**: Menyertakan file `.meta.json` dan `.users.sql` (grants) untuk setiap backup.

### 🔄 Restore Database
- **Safety First**: Validasi ketat untuk mencegah restore ke database primary secara tidak sengaja.
- **Auto-Detection**: Mendeteksi format kompresi dan enkripsi secara otomatis.
- **Smart Restore**:
  - Otomatis mencari dan me-restore *companion databases* (misal: `db_name` dan `db_name_dmart`).
  - **Pre-Restore Backup**: Opsi otomatis mem-backup target database sebelum ditimpa.
  - **User Grants**: Restore hak akses user aplikasi secara otomatis setelah data dipulihkan.

### 🔐 Manajemen Profil
- Menyimpan kredensial database dalam file terenkripsi (`.cnf.enc`).
- Mencegah *hardcoding* password dalam script atau command history.

### 📊 Monitoring & Utility
- **DB Scan**: Mengumpulkan metrik database (ukuran, jumlah tabel, prosedur, dll).
- **Cleanup**: Rotasi file backup otomatis berdasarkan usia file.
- **Crypto Utils**: Enkripsi/dekripsi file atau teks ad-hoc.

---

## 🛠️ Instalasi & Build

Pastikan Anda memiliki **Go 1.25+** terinstal.

```bash
# Clone repository
git clone https://your-repo/sfDBTools.git
cd sfDBTools

# Download dependencies
go mod tidy

# Build binary
go build -o bin/sfdbtools main.go

# Atau gunakan script helper (Linux)
./scripts/build_run.sh -- help
```

---

## 🚢 Release

Checklist rilis (contoh versi `1.0.0`):

```bash
# Pastikan working tree bersih
git status

# Buat tag dan push
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# Build (akan inject metadata version/commit/build-date otomatis)
./scripts/build_run.sh --skip-run

# Verifikasi versi
go run . version
```

Alternatif (sekali jalan, otomatis trigger GitHub Release via tag):

```bash
bash ./scripts/release_github.sh 1.0.0
```

---

## 📖 Panduan Penggunaan

### 1. Konfigurasi Profil (Langkah Awal)
Sebelum melakukan operasi database, buat profil koneksi terlebih dahulu:

```bash
# Membuat profil baru (Interaktif)
sfdbtools profile create --interactive

# List profil yang tersedia
sfdbtools profile show --file my_prod_db.cnf.enc
```

### 2. Melakukan Backup

**Format Umum:**
`sfdbtools backup [mode] --ticket [TICKET_NO] [flags]`

```bash
# Backup satu database dengan kompresi ZSTD
sfdbtools backup single --db my_app_db --ticket TICKET-123 --compress-type zstd

# Backup semua database, kecuali system DB
sfdbtools backup all --exclude-system --ticket TICKET-123

# Backup dengan enkripsi
sfdbtools backup single --db sensitive_db --backup-key "RAHASIA" --ticket TICKET-123
```

### 3. Melakukan Restore

**Format Umum:**
`sfdbtools restore [mode] --ticket [TICKET_NO] [flags]`

```bash
# Restore ke database baru (Single Mode)
sfdbtools restore single --file /backups/my_app_db_2025.sql.zst --target-db my_app_dev --ticket TICKET-123

# Restore Primary (Otomatis handle _dmart dan user app)
sfdbtools restore primary --file /backups/dbsf_nbc_client.sql.zst --client-code client --ticket TICKET-123
```

### 4. Database Scanning
```bash
# Scan ukuran dan metadata semua database lokal
sfdbtools db-scan all-local --profile my_prod.cnf.enc
```

---

## 🏗️ Arsitektur Kode

Project ini menerapkan prinsip **Clean Architecture** dan **Design Patterns** untuk memastikan *maintainability*.

### Struktur Direktori
*   `cmd/`: Entry point untuk Cobra CLI commands.
*   `internal/`: Logika bisnis inti.
    *   `backup/`: Menggunakan **Strategy Pattern** (`CombinedExecutor`, `IterativeExecutor`) untuk menangani berbagai mode backup.
    *   `restore/`: Menggunakan struktur serupa dengan backup untuk konsistensi, memisahkan logika `Single` dan `Primary`.
*   `pkg/`: Library pendukung yang *reusable* (enkripsi, kompresi, database wrapper, UI).

### Prinsip Desain
1.  **KISS (Keep It Simple, Stupid)**: Alur *streaming* linear untuk data backup/restore.
2.  **DRY (Don't Repeat Yourself)**: Penggunaan shared helpers di `pkg/` untuk validasi, koneksi, dan file operations.
3.  **Fail-Fast**: Validasi ketat di awal (koneksi, keberadaan file, aturan safety) sebelum proses berat dimulai.
4.  **Security**: Enkripsi *at rest* dan penanganan password yang aman (masking di logs).

---

## 📝 Lisensi

Copyright © 2025. Internal Tool.
