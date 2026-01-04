# sfDBTools

sfDBTools adalah CLI utility untuk operasi MySQL/MariaDB: backup, restore, db-scan, cleanup, crypto, dan manajemen profil koneksi.

Target utama: penggunaan di environment server (Linux) dengan fokus pada **streaming** (hemat RAM), **safety**, dan **otomasi**.

## Instalasi (Linux amd64)

### One-click install (recommended)

Install sebagai root (akan auto pilih `.deb` / `.rpm` / tar sesuai OS):

```bash
curl -fsSL https://raw.githubusercontent.com/hadiy961/sfDBTools-New/main/scripts/install.sh | sudo bash
```

Install tanpa root (tar ke `~/.local/bin`):

```bash
curl -fsSL https://raw.githubusercontent.com/hadiy961/sfDBTools-New/main/scripts/install.sh | bash
```

Verifikasi:

```bash
sfdbtools version
sfdbtools --help
```

### Uninstall

```bash
curl -fsSL https://raw.githubusercontent.com/hadiy961/sfDBTools-New/main/scripts/uninstall.sh | sudo bash
```

Uninstall + hapus config user (HATI-HATI):

```bash
curl -fsSL https://raw.githubusercontent.com/hadiy961/sfDBTools-New/main/scripts/uninstall.sh | sudo bash -s -- --purge
```

## Requirements (dependensi runtime)

- `mysqldump` wajib tersedia untuk fitur backup (`db-backup`).
- `mysql` CLI wajib tersedia untuk fitur restore (`db-restore`).
  - Biasanya tersedia dari paket `mysql-client` atau `mariadb-client` (nama paket tergantung distro).
- Akses network ke server database + user DB yang punya privilege sesuai operasi.
- Untuk beberapa fitur `script`, diperlukan `bash`.

## Konfigurasi (auto)

Saat pertama kali dijalankan, jika file konfigurasi belum ada, sfDBTools akan membuat config default otomatis.

- Jika `SFDB_APPS_CONFIG` diset, file config dibuat di path tersebut.
- Jika tidak diset:
  - akan mencoba `/etc/sfDBTools/config.yaml` (jika punya permission)
  - fallback ke `~/.config/sfDBTools/config.yaml` (atau `XDG_CONFIG_HOME/sfDBTools/config.yaml`)

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

### 2) Backup

Backup single database:

```bash
sfdbtools db-backup single \
  --profile ./configs/prod-db.cnf.enc \
  --profile-key "my-secret-key" \
  --database "target_db" \
  --ticket "TICKET-123"
```

Backup semua database (misal exclude system DB):

```bash
sfdbtools db-backup all \
  --profile ./configs/prod-db.cnf.enc \
  --profile-key "my-secret-key" \
  --exclude-system \
  --ticket "TICKET-123"
```

Jika ingin enkripsi output backup:

```bash
export SFDB_BACKUP_ENCRYPTION_KEY="my-backup-key"
sfdbtools db-backup single --profile ./configs/prod-db.cnf.enc --profile-key "my-secret-key" --database "target_db" --ticket "TICKET-123"
```

### 3) Restore

Restore single database (wajib `--ticket`):

```bash
sfdbtools db-restore single \
  --file "/path/to/backup.sql.gz.enc" \
  --encryption-key "my-backup-key" \
  --ticket "TICKET-123"
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

## Lisensi

Internal Tool DataOn.
