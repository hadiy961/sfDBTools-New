# sfDBTools - Dokumentasi Backup All

## Daftar Isi
- [sfDBTools - Dokumentasi Backup All](#sfdbtools---dokumentasi-backup-all)
  - [Daftar Isi](#daftar-isi)
  - [Deskripsi](#deskripsi)
  - [Flags / Parameter](#flags--parameter)
  - [Aturan parameter](#aturan-parameter)
  - [Urutan Proses (Mode Interaktif)](#urutan-proses-mode-interaktif)
  - [Urutan Proses (Mode Non-Interaktif)](#urutan-proses-mode-non-interaktif)
  - [Contoh Penggunaan](#contoh-penggunaan)
    - [Backup Interaktif (Default)](#backup-interaktif-default)
    - [Backup Non-Interaktif (untuk Automation/Scripting)](#backup-non-interaktif-untuk-automationscripting)
    - [Menggunakan Environment Variables](#menggunakan-environment-variables)
  - [Output Files](#output-files)
    - [1. File Backup Utama](#1-file-backup-utama)
    - [2. File Metadata (JSON)](#2-file-metadata-json)
    - [3. File User Grants (Optional)](#3-file-user-grants-optional)
    - [Contoh Struktur Output Directory](#contoh-struktur-output-directory)
  - [Fitur-Fitur Khusus](#fitur-fitur-khusus)
    - [1. Capture GTID (Global Transaction ID)](#1-capture-gtid-global-transaction-id)
    - [2. Export User Grants](#2-export-user-grants)
    - [3. Filter dan Exclude Options](#3-filter-dan-exclude-options)
      - [Exclude System Databases](#exclude-system-databases)
      - [Exclude Empty Databases](#exclude-empty-databases)
      - [Exclude Data (Schema Only)](#exclude-data-schema-only)
    - [4. Compression](#4-compression)
    - [5. Encryption](#5-encryption)
    - [6. Dry Run Mode](#6-dry-run-mode)
  - [Metadata File](#metadata-file)
  - [Troubleshooting](#troubleshooting)
    - [Error: "ticket wajib diisi pada mode non-interaktif"](#error-ticket-wajib-diisi-pada-mode-non-interaktif)
    - [Error: "profile-key wajib diisi pada mode non-interaktif"](#error-profile-key-wajib-diisi-pada-mode-non-interaktif)
    - [Error: "backup-key wajib diisi saat enkripsi aktif pada mode non-interaktif"](#error-backup-key-wajib-diisi-saat-enkripsi-aktif-pada-mode-non-interaktif)
    - [Warning: "SHOW MASTER STATUS tidak mengembalikan hasil"](#warning-show-master-status-tidak-mengembalikan-hasil)
    - [Error: "gagal koneksi ke database"](#error-gagal-koneksi-ke-database)
    - [Backup Terlalu Lambat](#backup-terlalu-lambat)
    - [File Backup Terlalu Besar](#file-backup-terlalu-besar)
  - [Best Practices](#best-practices)

## Deskripsi
Perintah ini digunakan untuk mencadangkan semua database dari server.

## Flags / Parameter
* `-x, --exclude-data`: Mengecualikan data dari pencadangan (schema only) (default: dari file konfigurasi).
* `-S, --exclude-system`: Mengecualikan sistem database (default: dari file konfigurasi).
* `-E, --exclude-empty`: Mengecualikan database kosong (default: dari file konfigurasi).
* `-n, --non-interactive`: Tidak melakukan interaksi (default: false).
* `-C, --skip-compress`: Melewati proses kompresi pada file backup (default: dari file konfigurasi).
* `-p, --profile`: Menentukan profil koneksi database sumber yang akan digunakan untuk pencadangan.
* `-k, --profile-key`: Kunci enkripsi untuk profil database (jika diperlukan) (default: env `SFDB_SOURCE_PROFILE_KEY`).
* `-t, --ticket`: Menentukan tiket atau referensi untuk pencadangan (**Wajib diisi**).
* `-o, --backup-dir`: Direktori output untuk menyimpan file backup (default: dari file konfigurasi).
* `-d, --dry-run`: Menjalankan perintah tanpa melakukan pencadangan sebenarnya (default: false).
* `-f, --filename`: Menentukan nama file untuk backup (opsional) (Tanpa ekstensi) (default: auto dari config/pattern).
* `-l, --compress-level`: Menentukan level kompresi (1-9) (default: dari file konfigurasi) *Jika `--skip-compress=false`*.
* `-c, --compress`: Menentukan jenis kompresi (gzip, zstd, xz, zlib, pgzip, none) (default: dari file konfigurasi).
* `--skip-encrypt`: Melewati proses enkripsi pada file backup (default: dari file konfigurasi).
* `-K, --backup-key`: Kunci enkripsi untuk file backup *jika `--skip-encrypt=false`*.

## Aturan parameter
`--non-interactive` harus diatur ke true jika digunakan dalam skrip otomatisasi.

Jika `--non-interactive=true`:
- Semua input yang tidak bisa diprompt harus sudah tersedia, termasuk: `--ticket`, `--profile`, `--profile-key`, dan `--backup-key` (jika enkripsi aktif dan `--skip-encrypt=false`).
- Jika ada yang kosong, perintah akan gagal (fail-fast).

Jika `--non-interactive=false` (interaktif):
- Jika `--ticket` tidak diberikan, aplikasi akan meminta input.
- Jika `--profile` tidak diberikan, aplikasi akan meminta pemilihan profile.
- Jika profile terenkripsi dan `--profile-key` tidak diberikan, aplikasi akan meminta input.
- Jika enkripsi aktif dan `--backup-key` kosong, aplikasi akan meminta input.

Pada mode interaktif, setelah ringkasan opsi ditampilkan, user dapat memilih **"Ubah opsi"**.
Saat memilih **"Ubah opsi"**, aplikasi akan menampilkan submenu **"Pilih opsi yang ingin diubah"** (mirip modul restore), sehingga user bisa memilih hanya opsi tertentu yang ingin diubah. Submenu ini termasuk opsi untuk **memilih ulang profile** (dan aplikasi akan reconnect ke database source berdasarkan profile baru).

Default `--skip-encrypt` mengacu ke konfigurasi `backup.encryption.enabled`.
Default `--skip-compress` mengacu ke konfigurasi `backup.compression.enabled`.
`--compress-level` dan `--compress` hanya berlaku jika kompresi diaktifkan.
`--backup-dir` harus diatur jika ingin menimpa direktori output default dari file konfigurasi.
`--filename` opsional, jika tidak diisi maka nama file akan dihasilkan secara otomatis berdasarkan pola penamaan default.

## Urutan Proses (Mode Interaktif)
1. Pemilihan profil database sumber (jika tidak diberikan melalui flag).
2. Koneksi ke database server dan mengambil hostname server.
3. Jika `--ticket` tidak diberikan, aplikasi akan meminta input ticket number.
4. Aplikasi mengambil daftar database dan menerapkan filter (exclude-data / exclude-system / exclude-empty) sesuai default konfigurasi atau override dari flag.
5. Aplikasi menghasilkan preview direktori output dan nama file backup (sesuai config/pattern, kompresi, enkripsi, dan mode).
6. Aplikasi menampilkan ringkasan opsi ("Opsi Backup") + statistik hasil filtering, lalu user memilih aksi: **Lanjutkan / Ubah opsi / Batalkan**.
7. Jika user memilih **Ubah opsi**, aplikasi menampilkan submenu **"Pilih opsi yang ingin diubah"** dengan pilihan:
   - Profile (pilih ulang profile + reconnect ke database)
   - Ticket number
   - Capture GTID
   - Export user grants
   - Exclude system databases
   - Exclude empty databases
   - Exclude data (schema only)
   - Backup directory
   - Filename
   - Encryption (toggle + input backup key)
   - Compression (toggle + pilih type + level)
   - Kembali
   
   Setelah opsi diubah, aplikasi kembali ke ringkasan opsi (langkah 6) dengan nilai yang sudah di-update dan re-filter database sesuai perubahan.
8. Jika user memilih **Lanjutkan**, aplikasi melakukan validasi minimal sebelum eksekusi (ticket wajib terisi; jika enkripsi aktif maka backup-key wajib tersedia).
9. Memulai proses backup dengan menampilkan progress.
10. Menampilkan hasil backup setelah selesai (summary statistik + detail backup).

## Urutan Proses (Mode Non-Interaktif)
1. Validasi semua flag dan parameter yang diberikan.
2. Memulai proses backup.
3. Menampilkan hasil backup setelah selesai.

## Contoh Penggunaan

### Backup Interaktif (Default)
```bash
# Backup semua database dengan prompt interaktif
sfdbtools db-backup all

# Backup dengan profile tertentu (akan prompt untuk opsi lain)
sfdbtools db-backup all --profile /etc/sfDBTools/profiles/db_prod.cnf.enc
```

### Backup Non-Interaktif (untuk Automation/Scripting)
```bash
# Backup lengkap dengan semua parameter wajib
sfdbtools db-backup all \
  --non-interactive \
  --profile /etc/sfDBTools/profiles/db_prod.cnf.enc \
  --profile-key "your-profile-encryption-key" \
  --ticket "TICK-2025-001" \
  --backup-key "your-backup-encryption-key" \
  --backup-dir /backup/daily

# Backup tanpa enkripsi dan kompresi (raw SQL dump)
sfdbtools db-backup all \
  --non-interactive \
  --profile /etc/sfDBTools/profiles/db_prod.cnf.enc \
  --profile-key "your-profile-encryption-key" \
  --ticket "TICK-2025-001" \
  --skip-encrypt \
  --skip-compress \
  --backup-dir /backup/raw

# Backup schema only (tanpa data)
sfdbtools db-backup all \
  --non-interactive \
  --exclude-data \
  --profile /etc/sfDBTools/profiles/db_prod.cnf.enc \
  --profile-key "your-profile-encryption-key" \
  --ticket "TICK-2025-001" \
  --backup-key "your-backup-encryption-key"

# Backup dengan custom compression type dan level
sfdbtools db-backup all \
  --non-interactive \
  --compress zstd \
  --compress-level 9 \
  --profile /etc/sfDBTools/profiles/db_prod.cnf.enc \
  --profile-key "your-profile-encryption-key" \
  --ticket "TICK-2025-001" \
  --backup-key "your-backup-encryption-key"
```

### Menggunakan Environment Variables
```bash
# Set environment variables untuk credentials
export SFDB_SOURCE_PROFILE_KEY="your-profile-encryption-key"
export SFDB_BACKUP_ENCRYPTION_KEY="your-backup-encryption-key"

# Run backup dengan environment variables
sfdbtools db-backup all \
  --non-interactive \
  --profile /etc/sfDBTools/profiles/db_prod.cnf.enc \
  --ticket "TICK-2025-001" \
  --backup-dir /backup/daily
```

## Output Files

Backup ALL menghasilkan beberapa file output:

### 1. File Backup Utama
File SQL dump dengan nama otomatis berdasarkan pattern:
- Format: `all_<hostname>_<timestamp>_<jumlah-db>db.sql[.kompresi][.enc]`
- Format (exclude-data): `all_nodata_<hostname>_<timestamp>_<jumlah-db>db.sql[.kompresi][.enc]`
- Contoh tanpa enkripsi/kompresi: `all_prod-server_20251230_143022_23db.sql`
- Contoh dengan kompresi zstd: `all_prod-server_20251230_143022_23db.sql.zst`
- Contoh dengan kompresi + enkripsi: `all_prod-server_20251230_143022_23db.sql.zst.enc`
- Contoh exclude-data: `all_nodata_prod-server_20251230_143022_23db.sql.zst.enc`

### 2. File Metadata (JSON)
File metadata berisi informasi lengkap tentang backup:
- Format: `<nama-file-backup>.meta.json`
- Contoh: `all_prod-server_20251230_143022_23db.sql.zst.enc.meta.json`
- Isi metadata:
  - Informasi backup (waktu, durasi, ukuran file)
  - Daftar database yang di-backup
  - Daftar database yang di-exclude
  - Informasi GTID (jika capture GTID aktif)
  - Informasi kompresi dan enkripsi
  - Versi mysqldump dan MariaDB/MySQL
  - Ticket number
  - Warning/errors (jika ada)

### 3. File User Grants (Optional)
File SQL berisi user grants untuk database yang di-backup:
- Format: `<nama-file-backup-tanpa-ekstensi>_users.sql`
- Contoh: `all_prod-server_20251230_143022_23db_users.sql`
- Dihasilkan jika opsi "Export user grants" aktif (default: true, kecuali `ExcludeUser=true`)
- Berisi `CREATE USER` dan `GRANT` statements untuk semua user yang memiliki akses ke database yang di-backup

### Contoh Struktur Output Directory
```
/backup/daily/
├── all_prod-server_20251230_143022_23db.sql.zst.enc            # File backup utama (encrypted + compressed)
├── all_prod-server_20251230_143022_23db.sql.zst.enc.meta.json  # Metadata
└── all_prod-server_20251230_143022_23db_users.sql              # User grants
```

## Fitur-Fitur Khusus

### 1. Capture GTID (Global Transaction ID)
GTID digunakan untuk replikasi dan point-in-time recovery di MariaDB/MySQL.

**Cara Kerja:**
- Saat backup dimulai, aplikasi mengambil GTID position dari server (menggunakan `SHOW MASTER STATUS` dan `BINLOG_GTID_POS`)
- GTID position disimpan di metadata file
- Berguna untuk setup replication slave atau restore dengan GTID consistency

**Konfigurasi:**
- Default: diambil dari config file (`backup.capture_gtid`)
- Override: via flag atau interactive prompt
- Informasi GTID disimpan di metadata meliputi:
  - `master_log_file`: Nama binlog file
  - `master_log_pos`: Posisi di binlog
  - `gtid_binlog`: GTID position (format MariaDB)

**Catatan:**
- Binary logging harus aktif di server database
- Jika binary logging tidak aktif, GTID capture akan di-skip dengan warning

### 2. Export User Grants
Backup user grants untuk semua user yang memiliki akses ke database yang di-backup.

**Cara Kerja:**
- Aplikasi query user yang memiliki privileges ke database yang di-backup
- Generate `CREATE USER` statements dengan password hash original
- Generate `GRANT` statements untuk semua privileges
- Simpan ke file terpisah (`*_users.sql`)

**Konfigurasi:**
- Default: aktif (export user grants)
- Disable: set flag di config atau via interactive prompt
- File grants dicatat di metadata (`user_grants_file`)

**Kegunaan:**
- Restore user permissions saat restore database
- Migrasi user accounts ke server baru
- Backup security configuration

### 3. Filter dan Exclude Options

#### Exclude System Databases
- Mengecualikan database sistem: `mysql`, `information_schema`, `performance_schema`, `sys`
- Default: dari config file
- Override: `--exclude-system` atau `-S`

#### Exclude Empty Databases
- Mengecualikan database yang tidak memiliki tabel
- Berguna untuk menghemat space dan waktu backup
- Default: dari config file
- Override: `--exclude-empty` atau `-E`

#### Exclude Data (Schema Only)
- Backup hanya struktur database (CREATE TABLE, etc.) tanpa data
- Berguna untuk:
  - Backup schema untuk development/testing
  - Dokumentasi struktur database
  - Template database baru
- Default: dari config file
- Override: `--exclude-data` atau `-x`

### 4. Compression

**Jenis Kompresi yang Didukung:**
- `zstd` (Zstandard) - **Recommended** - balance antara speed dan compression ratio
- `gzip` - Compatible, moderate compression
- `pgzip` - Parallel gzip (lebih cepat dari gzip di multi-core)
- `xz` - Compression ratio tertinggi, tapi lambat
- `zlib` - Compatible dengan banyak tools
- `none` - Tanpa kompresi

**Compression Level:**
- Range: 1-9
- 1 = fastest, compression ratio terendah
- 9 = slowest, compression ratio tertinggi
- Default: dari config file (biasanya 6)
- Rekomendasi:
  - Level 1-3: untuk backup cepat, space tidak terlalu masalah
  - Level 6: balanced (default)
  - Level 9: untuk backup archival, prioritas space saving

**Skip Compression:**
- Flag: `--skip-compress` atau `-C`
- Hasilkan raw SQL file tanpa kompresi
- Berguna untuk debugging atau direct inspection

### 5. Encryption

**Algoritma Enkripsi:**
- AES-256-GCM (Galois/Counter Mode) dengan PBKDF2 key derivation (100,000 iterations)
- Compatible dengan OpenSSL (format: `Salted__` header + salt + ciphertext)
- Streaming encryption (tidak load seluruh file ke memory)
- Salt size: 8 bytes (random per file)

**Cara Kerja:**
- Backup di-encrypt menggunakan key yang diberikan dengan PBKDF2 key derivation (100,000 iterations)
- Salt random 8 bytes ditambahkan untuk setiap file
- Format: `Salted__` (8 bytes) + salt (8 bytes) + encrypted data
- Dekripsi: gunakan tools sfDBTools atau OpenSSL dengan format yang sama

**Skip Encryption:**
- Flag: `--skip-encrypt`
- Hasilkan file tanpa enkripsi
- **Perhatian:** File backup berisi data sensitif, hindari skip encrypt di production

**Backup Key Management:**
- Bisa via flag: `--backup-key`
- Bisa via ENV: `SFDB_BACKUP_ENCRYPTION_KEY`
- Bisa via interactive prompt (mode interaktif)
- **Best Practice:** Gunakan key management system atau secrets vault

### 6. Dry Run Mode
- Flag: `--dry-run` atau `-d`
- Jalankan semua validasi tanpa eksekusi backup sebenarnya
- Berguna untuk:
  - Testing command dan parameter
  - Validasi profile dan koneksi
  - Preview output filename dan directory
  - Estimasi waktu dan resource

## Metadata File

File metadata (`.meta.json`) berisi informasi lengkap tentang backup. Contoh struktur:

```json
{
  "backup_file": "all_prod-server_20251230_143022_23db.sql.zst.enc",
  "backup_type": "all",
  "database_names": ["app_db", "users_db", "logs_db", ...],
  "excluded_databases": ["mysql", "information_schema", "sys", "performance_schema"],
  "hostname": "prod-server.example.com",
  "backup_start_time": "2025-12-30T14:30:22Z",
  "backup_end_time": "2025-12-30T14:35:48Z",
  "backup_duration": "5m26s",
  "file_size": 1234567890,
  "file_size_human": "1.15 GB",
  "compressed": true,
  "compression_type": "zstd",
  "encrypted": true,
  "exclude_data": false,
  "backup_status": "success",
  "warnings": [],
  "generated_by": "sfDBTools",
  "generated_at": "2025-12-30T14:35:48Z",
  "ticket": "TICK-2025-001",
  "user_grants_file": "all_prod-server_20251230_143022_23db_users.sql",
  "mysqldump_version": "10.19",
  "mariadb_version": "10.11.6-MariaDB",
  "gtid_info": {
    "master_log_file": "mysql-bin.000123",
    "master_log_pos": 12345678,
    "gtid_binlog": "0-1-12345"
  },
  "source_host": "prod-db-01.internal",
  "source_port": 3306
}
```

## Troubleshooting

### Error: "ticket wajib diisi pada mode non-interaktif"
**Penyebab:** Flag `--ticket` tidak diberikan saat menggunakan `--non-interactive`
**Solusi:** Tambahkan `--ticket "TICKET-NUMBER"` ke command

### Error: "profile-key wajib diisi pada mode non-interaktif"
**Penyebab:** Profile terenkripsi tapi key tidak tersedia
**Solusi:** 
- Tambahkan `--profile-key "KEY"` ke command, atau
- Set environment variable `SFDB_SOURCE_PROFILE_KEY="KEY"`

### Error: "backup-key wajib diisi saat enkripsi aktif pada mode non-interaktif"
**Penyebab:** Enkripsi aktif tapi backup key tidak tersedia
**Solusi:**
- Tambahkan `--backup-key "KEY"` ke command, atau
- Set environment variable `SFDB_BACKUP_ENCRYPTION_KEY="KEY"`, atau
- Gunakan `--skip-encrypt` jika tidak perlu enkripsi

### Warning: "SHOW MASTER STATUS tidak mengembalikan hasil"
**Penyebab:** Binary logging tidak aktif di database server
**Dampak:** GTID capture akan di-skip
**Solusi:** Aktifkan binary logging di MySQL/MariaDB config jika GTID diperlukan

### Error: "gagal koneksi ke database"
**Penyebab:** Profile salah atau database tidak reachable
**Solusi:**
- Validasi profile file bisa di-decrypt dengan key yang benar
- Test koneksi ke database secara manual
- Cek network connectivity dan firewall

### Backup Terlalu Lambat
**Solusi:**
- Gunakan compression level lebih rendah (1-3 instead of 9)
- Gunakan compression type yang lebih cepat (`zstd` atau `pgzip`)
- Pertimbangkan `--exclude-data` jika hanya butuh schema
- Gunakan `--exclude-empty` untuk skip empty databases

### File Backup Terlalu Besar
**Solusi:**
- Gunakan compression level lebih tinggi (9)
- Gunakan compression type dengan ratio tinggi (`xz`)
- Pertimbangkan backup per-database (mode `single` atau `filter`) alih-alih `all`
- Gunakan `--exclude-data` jika data tidak perlu di-backup

## Best Practices

1. **Selalu gunakan encryption** di production environment
2. **Store backup keys securely** (gunakan vault/secrets manager)
3. **Gunakan ticket system** untuk tracking backup operations
4. **Test restore procedure** secara berkala
5. **Monitor backup file size** dan duration untuk detect anomalies
6. **Verify metadata file** setelah backup untuk memastikan integrity
7. **Gunakan `--dry-run`** untuk testing command sebelum production run
8. **Schedule regular backups** dengan automation (cron/systemd timer)
9. **Keep multiple backup generations** (daily, weekly, monthly)
10. **Document restore procedures** dan simpan bersama backup files
