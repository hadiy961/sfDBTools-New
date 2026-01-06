# sfdbtools

sfdbtools adalah CLI utility untuk operasi MySQL/MariaDB: backup, restore, db-scan, cleanup, crypto, dan manajemen profil koneksi.

Target utama: penggunaan di environment server (Linux) dengan fokus pada **streaming** (hemat RAM), **safety**, dan **otomasi**.

## Instalasi (Linux amd64)

### One-click install (recommended)

Install sebagai root (akan auto pilih `.deb` / `.rpm` / tar sesuai OS):

```bash
curl -fsSL https://raw.githubusercontent.com/hadiy961/sfdbtools-New/main/scripts/install.sh | sudo bash
```

Install tanpa root (tar ke `~/.local/bin`):

```bash
curl -fsSL https://raw.githubusercontent.com/hadiy961/sfdbtools-New/main/scripts/install.sh | bash
```

Verifikasi:

```bash
sfdbtools version
sfdbtools --help
```

## Auto Update

sfdbtools bisa melakukan auto-update dari GitHub Releases.

- Default: auto-update aktif saat startup.
- Disable paksa (override): `SFDB_NO_AUTO_UPDATE=1`
- Manual update: `sfdbtools update`

Catatan:
- Auto-update saat ini hanya untuk `linux/amd64` (sesuai workflow release).
- Jika binary terpasang di `/usr/bin`, jalankan dengan `sudo` agar bisa overwrite.
- Saat startup, sfdbtools akan cek koneksi internet dulu. Jika tidak ada internet, proses update akan di-skip.
- Saat mode non-quiet, proses cek update menampilkan spinner singkat (output ke stderr).

### Uninstall

```bash
curl -fsSL https://raw.githubusercontent.com/hadiy961/sfdbtools-New/main/scripts/uninstall.sh | sudo bash
```

Uninstall + hapus config user (HATI-HATI):

```bash
curl -fsSL https://raw.githubusercontent.com/hadiy961/sfdbtools-New/main/scripts/uninstall.sh | sudo bash -s -- --purge
```

## Requirements (dependensi runtime)

- `mysqldump` wajib tersedia untuk fitur backup (`db-backup`).
- `mysql` CLI wajib tersedia untuk fitur restore (`db-restore`).
  - Biasanya tersedia dari paket `mysql-client` atau `mariadb-client` (nama paket tergantung distro).
- Akses network ke server database + user DB yang punya privilege sesuai operasi.
- Untuk beberapa fitur `script`, diperlukan `bash`.

## Konfigurasi (auto)

Saat pertama kali dijalankan, jika file konfigurasi belum ada, sfdbtools akan membuat config default otomatis.

- Jika `SFDB_APPS_CONFIG` diset, file config dibuat di path tersebut.
- Jika tidak diset:
  - akan mencoba `/etc/sfDBTools/config.yaml` (jika punya permission)
  - fallback ke `~/.config/sfdbtools/config.yaml` (atau `XDG_CONFIG_HOME/sfdbtools/config.yaml`)

## Quickstart

### 1) Buat profile koneksi (wizard)

Jalankan tanpa flag untuk wizard interaktif:

```bash
sfdbtools profile create
```

Atau one-liner (contoh):

```bash
sfdbtools profile create \
  --profile "prod-db" \
  --host "10.0.0.5" \
  --port 3306 \
  --user "admin" \
  --password "s3cr3t" \
  --profile-key "my-secret-key"
```

Lihat profile:

```bash
sfdbtools profile show
```

### 2) Backup Database

#### Backup Single Database

Backup satu database dengan kompresi dan enkripsi:

```bash
sfdbtools db-backup single \
  --profile ./configs/prod-db.cnf.enc \
  --profile-key "my-secret-key" \
  --database "myapp_db" \
  --ticket "TICKET-123" \
  --compression zstd \
  --encryption-key "backup-secret"
```

#### Backup Semua Database (Exclude System DB)

```bash
sfdbtools db-backup all \
  --profile ./configs/prod-db.cnf.enc \
  --profile-key "my-secret-key" \
  --exclude-system \
  --compression gzip \
  --ticket "BACKUP-DAILY-001"
```

#### Backup Primary Databases (Pattern-Based)

Backup semua database yang match pattern primary (contoh: `dbsf_nbc_*`):

```bash
sfdbtools db-backup primary \
  --profile ./configs/prod-db.cnf.enc \
  --profile-key "my-secret-key" \
  --compression zstd \
  --ticket "PROD-BACKUP-20260105"
```

#### Backup dengan Custom Output Directory

```bash
sfdbtools db-backup single \
  --profile ./configs/prod-db.cnf.enc \
  --profile-key "my-secret-key" \
  --database "customer_data" \
  --output-dir "/backups/custom/location" \
  --ticket "CUSTOM-001"
```

#### Backup Tanpa Data (Schema Only)

```bash
sfdbtools db-backup single \
  --profile ./configs/prod-db.cnf.enc \
  --profile-key "my-secret-key" \
  --database "myapp_db" \
  --no-data \
  --ticket "SCHEMA-BACKUP-001"
```

#### Backup dengan User Grants

```bash
sfdbtools db-backup single \
  --profile ./configs/prod-db.cnf.enc \
  --profile-key "my-secret-key" \
  --database "myapp_db" \
  --dump-user-grants \
  --ticket "FULL-BACKUP-001"
```

### 3) Restore Database

#### Restore Single Database

Restore database dari backup file (wajib `--ticket`):

```bash
sfdbtools db-restore single \
  --file "/backups/myapp_db_20260105_120000.sql.gz.enc" \
  --encryption-key "backup-secret" \
  --ticket "RESTORE-123"
```

#### Restore ke Target Database Berbeda

```bash
sfdbtools db-restore single \
  --file "/backups/production_db.sql.gz.enc" \
  --target-database "staging_db" \
  --encryption-key "backup-secret" \
  --ticket "RESTORE-TO-STAGING"
```

#### Restore dengan Auto-Detect Companion Database

Restore primary database beserta companion-nya (_dmart, _temp, dll):

```bash
sfdbtools db-restore primary \
  --file "/backups/dbsf_nbc_client_20260105.sql.gz.enc" \
  --auto-detect-dmart \
  --encryption-key "backup-secret" \
  --ticket "RESTORE-WITH-COMPANION"
```

#### Restore Selection (Interactive)

Restore dengan memilih database secara interaktif:

```bash
sfdbtools db-restore selection \
  --file "/backups/all_databases_20260105.sql.gz.enc" \
  --encryption-key "backup-secret" \
  --ticket "SELECTIVE-RESTORE"
```

### 4) Database Scan

#### Scan Semua Database

Lihat metadata semua database di server:

```bash
sfdbtools db-scan all \
  --profile ./configs/prod-db.cnf.enc \
  --profile-key "my-secret-key"
```

#### Scan Database Lokal

Scan database dari file backup lokal:

```bash
sfdbtools db-scan all-local \
  --backup-dir "/backups/2026/01/05"
```

#### Scan dengan Filter

Scan database yang match pattern tertentu:

```bash
sfdbtools db-scan filter \
  --profile ./configs/prod-db.cnf.enc \
  --profile-key "my-secret-key" \
  --filter "dbsf_*"
```

### 5) Cleanup Old Backups

#### Cleanup Otomatis Berdasarkan Retention

```bash
sfdbtools cleanup auto \
  --backup-dir "/backups" \
  --retention-days 30 \
  --dry-run
```

Setelah yakin, jalankan tanpa `--dry-run`:

```bash
sfdbtools cleanup auto \
  --backup-dir "/backups" \
  --retention-days 30
```

#### Cleanup Manual (Interactive)

```bash
sfdbtools cleanup manual \
  --backup-dir "/backups/2025"
```

### 6) Profile Management

#### Create Profile (Interactive Wizard)

```bash
sfdbtools profile create
```

#### Show All Profiles

```bash
sfdbtools profile show
```

#### Show Specific Profile

```bash
sfdbtools profile show --file ./configs/prod-db.cnf.enc
```

#### Edit Profile

```bash
sfdbtools profile edit --file ./configs/prod-db.cnf.enc
```

#### Delete Profile

```bash
sfdbtools profile delete --file ./configs/prod-db.cnf.enc
```

### 7) Crypto Utilities

#### Encrypt File

```bash
sfdbtools crypto encrypt-file \
  --input /path/to/sensitive-data.txt \
  --output /path/to/sensitive-data.txt.enc \
  --key "my-encryption-key"
```

#### Decrypt File

```bash
sfdbtools crypto decrypt-file \
  --input /path/to/sensitive-data.txt.enc \
  --output /path/to/sensitive-data.txt \
  --key "my-encryption-key"
```

#### Encrypt Text (Interactive)

```bash
sfdbtools crypto encrypt-text
```

#### Base64 Encode

```bash
echo "Hello World" | sfdbtools crypto base64-encode
```

#### Base64 Decode

```bash
echo "SGVsbG8gV29ybGQ=" | sfdbtools crypto base64-decode
```

### 8) Automation & Pipeline Usage

#### Quiet Mode untuk CI/CD

```bash
export SFDB_QUIET=1  # Suppress banners dan spinner
sfdbtools db-backup single \
  --profile ./configs/prod-db.cnf.enc \
  --profile-key "my-secret-key" \
  --database "myapp_db" \
  --ticket "AUTO-BACKUP-$(date +%Y%m%d)" 2>&1 | tee backup.log
```

#### Environment Variables untuk Automation

```bash
# Set environment variables untuk password
export SFDB_SOURCE_PROFILE_KEY="profile-encryption-key"
export SFDB_BACKUP_ENCRYPTION_KEY="backup-encryption-key"

# Jalankan tanpa perlu specify key di CLI
sfdbtools db-backup single \
  --profile ./configs/prod-db.cnf.enc \
  --database "myapp_db" \
  --ticket "AUTO-BACKUP-001"
```

#### Cron Job Example

```cron
# Daily backup jam 2 pagi
0 2 * * * cd /opt/sfdbtools && SFDB_QUIET=1 sfdbtools db-backup all --profile /etc/sfDBTools/config/db_profile/prod.cnf.enc --ticket "DAILY-$(date +\%Y\%m\%d)" >> /var/log/sfdbtools-backup.log 2>&1

# Weekly cleanup retention 30 hari
0 3 * * 0 sfdbtools cleanup auto --backup-dir /backups --retention-days 30 >> /var/log/sfdbtools-cleanup.log 2>&1
```

## Ringkasan Command

- `sfdbtools db-backup`: backup database (subcommand: `all`, `filter`, `single`, `primary`, `secondary`)
- `sfdbtools db-restore`: restore database (subcommand: `single`, `primary`, `secondary`, `all`, `selection`, `custom`)
- `sfdbtools db-scan`: scan metadata database (subcommand: `all`, `all-local`, `filter`)
- `sfdbtools profile`: create/show/edit/delete profile koneksi
- `sfdbtools cleanup`: housekeeping file backup
- `sfdbtools crypto`: encrypt/decrypt file/text + base64 utils
- `sfdbtools script`: encrypt/extract/info/run bundle script
- `sfdbtools completion`: generate shell completion
- `sfdbtools version`: tampilkan versi

Untuk detail flag tiap command, gunakan:

```bash
sfdbtools <command> --help
sfdbtools <command> <subcommand> --help
```

## Environment Variables (yang sering dipakai)

- `SFDB_APPS_CONFIG`: override lokasi config YAML.
- `SFDB_QUIET=1`: suppress banner/spinner (cocok untuk pipeline).
- `SFDB_BACKUP_ENCRYPTION_KEY`: default key untuk enkripsi backup.
- `SFDB_ENCRYPTION_KEY`: default key untuk beberapa perintah `crypto`.
- `SFDB_SCRIPT_KEY`: key untuk bundle `script`.

## Catatan Pengembang (UI Facade)

Mulai UI-1, seluruh code internal wajib lewat facade `internal/ui/*`:

- Gunakan `internal/ui/print`, `internal/ui/prompt`, `internal/ui/table`, `internal/ui/progress`, `internal/ui/text`, `internal/ui/style`.
- Per UI-2, implementasi UI output sudah berada di `internal/ui/*` (tidak ada lagi pemakaian paket UI legacy).
- Per UI-3, engine prompt dipisah ke `internal/ui/input` dan public API tetap di `internal/ui/prompt`.

Pembagian tanggung jawab:

- `internal/ui/input`: wrapper low-level untuk `survey` + validator + selector (primitive interaksi).
- `internal/ui/prompt`: facade/high-level API yang dipakai feature (mis. `AskText`, `AskPassword`, `SelectOne/Many`) dan menjaga konsistensi UX.

Rencana deprecation bertahap:

- UI-3: pindahkan prompt/validator ke `internal/ui/input` + `internal/ui/prompt`, lalu phase-out pemakaian legacy sepenuhnya.

## Lisensi

Internal Tool DataOn.
