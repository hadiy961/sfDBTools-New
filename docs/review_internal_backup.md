# Review Paket `internal/backup` (sfDBTools)

Tanggal: 2026-01-02

Dokumen ini merangkum hasil review paket `internal/backup` beserta sub-packagenya. Fokusnya adalah: arsitektur, alur eksekusi, tanggung jawab tiap subpackage, dan poin penting terkait safety (fail-fast, cleanup, enkripsi/kompresi, metadata).

---

## 1) Gambaran Umum

`internal/backup` adalah implementasi fitur **backup MySQL/MariaDB** yang berorientasi CLI, dengan pola orkestrasi yang cukup jelas:

1. **Cmd layer** memanggil entrypoint unified `ExecuteBackup()`.
2. `backup.Service` melakukan **setup** (profile, ticket, output dir, validasi encryption/compression/filter).
3. Service memilih **mode executor** via `modes.GetExecutor()`.
4. Mode executor menjalankan backup melalui **execution engine** (membangun args, menjalankan mysqldump, retry, metadata).
5. Execution engine memakai **writer engine** untuk streaming output ke file melalui pipeline `encrypt`/`compress`.
6. (Opsional) Export **user grants**, capture **GTID**, dan generate/update **metadata manifest** (`.meta.json`).
7. Hasil diekspose sebagai `types_backup.BackupResult` dan ditampilkan via `display.ResultDisplayer`.

Karakter desain yang menonjol:

- **Clean-ish separation**: selection/filtering, setup, execution, writer, metadata dibuat terpisah.
- **ISP** (Interface Segregation Principle) diterapkan di `internal/backup/modes/interface.go` untuk menahan coupling ke `Service`.
- **Streaming pipeline**: tidak buffer besar; `mysqldump stdout → (optional) compression → (optional) encryption → file`.
- **Safety**: state tracking untuk cleanup file partial saat cancel/signal, validasi minimal untuk encryption key, dan atomic write metadata.

---

## 2) Entry Points & Alur Panggilan

### 2.1 Entry point dari cmd layer

- `ExecuteBackup(cmd, deps, mode)` (package `backup`) adalah entrypoint unified untuk semua mode.
- Memanggil `GetExecutionConfig(mode)` untuk mendapatkan:
  - header title
  - log prefix
  - success message

### 2.2 Orkestrasi utama

- `Service.ExecuteBackupCommand(ctx, config)`:
  1. `PrepareBackupSession(...)` (package `setup`): load profile, konek DB, filter DB, generate output dir/filename preview, (opsional) edit loop interaktif.
  2. `Service.ExecuteBackup(ctx, client, dbFiltered, mode)`:
     - `SetupBackupExecution()` (setup common: ticket, output dir, encryption/compression config)
     - `executeBackupByMode()` → `modes.GetExecutor()` → `executor.Execute(...)`
  3. Tampilkan hasil via `display.NewResultDisplayer(...).Display()`.

### 2.3 Graceful shutdown

- `ExecuteBackup()` memasang handler SIGINT/SIGTERM.
- `Service.HandleShutdown()` melakukan rollback: bila backup sedang berjalan, file output yang sedang dibuat akan dihapus (best-effort).

---

## 3) Paket Inti: `backup` (root)

### 3.1 `Service`

`backup.Service` memegang dependency runtime utama:

- `Config` (`appconfig.Config`)
- `Log` (`applog.Logger`)
- `ErrorLog` (`errorlog.ErrorLogger`)
- `BackupDBOptions` (parsed options dari cmd)
- `Client` (`database.Client`) saat session berjalan
- state untuk shutdown cleanup:
  - `currentBackupFile`, `backupInProgress`

Service juga mengimplement `modes.BackupService` (composite interface), dan menyediakan *bridge methods* agar mode executor tidak perlu mengimpor subpackage terlalu banyak.

### 3.2 Path generation

- `Service.GenerateFullBackupPath(dbName, mode)` menghasilkan full path final berdasarkan:
  - hostname (`DBInfo.HostName` fallback ke `DBInfo.Host`)
  - compression type
  - encryption enabled
  - exclude-data

- `Service.GenerateBackupPaths(ctx, client, dbFiltered)`:
  - generate output directory (pattern/config)
  - generate filename preview
  - untuk mode single/primary/secondary: lakukan selection + companion expansion
  - untuk mode all/combined: dukung custom base filename (`--filename`) dan tetap menjaga extension chain (mis. `.sql.gz.enc`).

### 3.3 Mode config

- `GetExecutionConfig(mode)` memusatkan label UI/log/success message per mode.

---

## 4) Subpackage `internal/backup/modes`

### 4.1 Interface & dependency boundaries

- `ModeExecutor.Execute(ctx, databases) BackupResult`.
- `BackupService` adalah gabungan dari sub-interface kecil:
  - `BackupContext`: logger & options
  - `BackupExecutor`: `ExecuteAndBuildBackup` & `ExecuteBackupLoop`
  - `BackupPathProvider`: `GenerateFullBackupPath`
  - `BackupMetadata`: GTID + user grants hooks
  - `BackupResultConverter`: loop result → result

Tujuannya: executor mode bergantung pada kontrak yang kecil dan jelas.

### 4.2 Factory

- `GetExecutor(mode, svc)`:
  - `all` + `combined` → `CombinedExecutor`
  - `single`/`primary`/`secondary`/`separated` → `IterativeExecutor`

### 4.3 Combined mode

- Semua DB dibackup dalam **1 file**.
- Capture GTID sebelum backup (best-effort, warning-only).
- Untuk user grants:
  - mode `all`: export semua grants (databases filter = `nil`/kosong)
  - mode `filter` single-file (dikendalikan flag internal `Filter.IsFilterCommand`): export grants yang relevan terhadap DB terpilih.
- Update metadata dengan path user grants aktual.

### 4.4 Iterative mode

Dipakai untuk:

- `separated`: multi file (per DB)
- `single`: single db (tanpa companion temp/archive)
- `primary`/`secondary`: db utama + companion (`_dmart` bila enabled)

Ciri penting:

- Eksekusi dilakukan dengan `ExecuteBackupLoop`.
- Mode `primary`/`secondary` melakukan:
  - export user grants sekali (berdasarkan list db)
  - update metadata pada file utama
  - membuat **combined metadata** untuk semua database yang dibackup
  - menghapus metadata individual untuk companion database
- Ada fallback error bila semua gagal pada variant single.

---

## 5) Subpackage `internal/backup/setup`

`setup` mengelola preflight dan flow interaktif/non-interaktif.

### 5.1 Load profile & koneksi

- `CheckAndSelectConfigFile()`:
  - Non-interaktif: wajib `--profile` dan `--profile-key` (atau env key).
  - Interaktif: boleh pilih profile bila mode mengizinkan dan `profile path` belum diberikan.

- `PrepareBackupSession(...)`:
  - Menjamin `Ticket` diminta sebelum koneksi DB dibuat (untuk mode interaktif).
  - Connect menggunakan `profilehelper.ConnectWithProfile`.
  - Mengambil hostname dari server dan menyimpan ke `Profile.DBInfo.HostName` (fallback ke config bila gagal).

### 5.2 Filtering DB

- `GetFilteredDatabases(...)`:
  - Bila ini adalah perintah filter dan belum ada include flags, tampilkan multi-select.
  - Selain itu gunakan `database.FilterDatabases` berdasarkan exclude/include/system list.

### 5.3 Edit loop interaktif

Untuk mode yang mengizinkan interaktif (mis. `all`, `single`, `primary`, `secondary`, dll), session bisa:

- render tabel opsi (`display.OptionsDisplayer.Render()`)
- prompt aksi: `Lanjutkan | Ubah opsi | Batalkan`
- validasi minimal sebelum lanjut:
  - Ticket wajib
  - jika encryption enabled, backup key wajib tersedia

File lain di `setup/` (`edit_menu_engine.go`, `backup_file_changes.go`, `backup_option_changes.go`, dll) mengimplement menu perubahan opsi (output file, encryption, compression, filter, dsb) dan menerapkan fail-safe (contoh: mematikan encryption jika key kosong ketika user batal input key).

---

## 6) Subpackage `internal/backup/selection`

`selection` menangani pemilihan database (terutama untuk mode filter dan mode single-variant).

### 6.1 Multi-select untuk filter command

- `GetFilteredDatabasesWithMultiSelect(...)`:
  - mengambil semua DB
  - menghapus system DB
  - **tidak mendukung** DB ber-suffix `_temp` dan `_archive`
  - menampilkan multi-select (survey)
  - persist hasil sebagai `Filter.IncludeDatabases` agar session loop tidak minta ulang

### 6.2 Single/Primary/Secondary selection + companion

- `SelectDatabaseAndBuildList(...)` memilih 1 DB (auto-select bila bisa) dan menambahkan companion:
  - untuk primary/secondary: saat `IncludeDmart` true, akan mencoba menambahkan `<db>_dmart` jika ada.

- `FilterCandidatesByMode(...)` dan helper lain menerapkan aturan naming:
  - mode primary: harus prefix tertentu (contoh `dbsf_nbc_`), bukan secondary, bukan dmart/temp/archive
  - mode secondary: harus mengandung suffix `secondary`, bukan dmart/temp/archive

---

## 7) Subpackage `internal/backup/execution`

`execution.Engine` adalah orkestrator backup per file:

- build mysqldump args
- menjalankan mysqldump via writer
- retry untuk error umum
- cleanup file gagal
- generate metadata

### 7.1 Build args

- `BuildMysqldumpArgs(baseArgs, dbInfo, filter, dbFiltered, singleDB, totalDBFound)`:
  - menambahkan host/port/user/password
  - memasukkan `baseDumpArgs` dari config
  - `--no-data` jika exclude-data
  - menentukan apakah memakai:
    - `--all-databases` (no filter + jumlah DB sama dengan total)
    - `--databases <db...>` untuk multi
    - `<db>` untuk single

- `MaskPasswordInArgs` tersedia untuk logging aman (password masking).

### 7.2 Retry strategies

- Jika stderr mengindikasikan SSL mismatch (client require SSL, server tidak support): tambahkan `--skip-ssl`.
- Jika stderr mengindikasikan `unknown option/variable`: coba hapus 1 opsi yang bermasalah.

### 7.3 Loop multi-DB

- `ExecuteBackupLoop(...)` menjalankan `ExecuteAndBuildBackup` per database.
- Untuk mode `separated` dan `single`, user grants bisa diexport per DB dan metadata di-update (jika `SaveBackupInfo` aktif).

### 7.4 Metadata generation

- `generateBackupMetadata(...)` membentuk `types_backup.BackupMetadata`.
- Catatan penting: metadata untuk `separated` tidak menyimpan GTID/replication/source host info (tidak relevan per desain).

---

## 8) Subpackage `internal/backup/writer`

`writer.Engine` bertanggung jawab menjalankan proses `mysqldump` dan streaming output ke file.

### 8.1 Pipeline

Urutan layer saat menulis:

- base: `bufio.Writer` ke file output
- Layer 1 (dekat file): **Encryption** (`encrypt.NewEncryptingWriter`)
- Layer 2: **Compression** (`compress.NewCompressingWriter`)

Sehingga alur akhirnya:

`mysqldump stdout → (compress) → (encrypt) → file`

### 8.2 Key resolution

- Bila encryption enabled, kunci diambil via resolver `ResolveEncryptionKey`:
  - dari opsi (`--backup-key`) atau
  - dari env `SFDB_BACKUP_ENCRYPTION_KEY`

### 8.3 Error handling

- stderr dicapture penuh.
- Error fatal vs non-fatal ditentukan oleh heuristik substring (mis. `access denied`, `can't connect` dianggap fatal; beberapa warning tertentu dianggap non-fatal).
- Untuk fatal error: fungsi return error + excerpt stderr.
- Untuk non-fatal: return sukses tetapi mencatat warning.

### 8.4 Progress monitor

- `databaseMonitorWriter` mem-parsing marker `-- Current Database: ` dari output mysqldump untuk menampilkan progres per DB pada mode combined.

---

## 9) Subpackage `internal/backup/metadata` & `internal/backup/grants` & `internal/backup/gtid`

### 9.1 Metadata

- `GenerateBackupMetadata(cfg)` mengisi field penting:
  - file info, ukuran, waktu, status, warning
  - compression/encryption flags
  - ticket
  - (kondisional) GTID + replication credentials + source host/port

- Persist metadata:
  - `SaveBackupMetadata(...)` memakai *atomic write* (tmp file + rename)
  - `TrySaveBackupMetadata(...)` non-fatal (warning-only)

- Update metadata:
  - `UpdateMetadataUserGrantsFile(...)` menulis path user grants (atau `"none"` jika kosong)
  - `UpdateMetadataWithDatabaseDetails(...)` untuk primary/secondary mengisi `DatabaseDetails` (per DB file size/path)

### 9.2 User grants

- Implementasi query:
  - `GetUserList()` dari `mysql.user`
  - `SHOW GRANTS FOR 'user'@'host'`
  - export semua atau terfilter berdasarkan database list
- Wrapper `grants.ExportUserGrantsIfNeeded`:
  - skip saat dry-run
  - skip saat `ExcludeUser` true
  - ping DB sebelum export (reconnect best-effort)

### 9.3 GTID

- `gtid.Capture(...)` mengambil GTID (best-effort) hanya saat flag enabled.

---

## 10) Subpackage `internal/backup/display`

- `OptionsDisplayer` menampilkan konfigurasi backup dalam bentuk tabel dan (opsional) meminta konfirmasi.
- `ResultDisplayer` menampilkan ringkasan hasil dan detail file output, termasuk throughput dan manifest jika ada.

---

## 11) Catatan Teknis (Observasi Review)

Poin-poin berikut bukan “bug report final”, tapi hal yang layak dicatat untuk maintainability/safety.

- Heuristik fatal/non-fatal di `writer.Engine.isFatalMysqldumpError` berbasis substring; ini cukup pragmatis, tapi sensitif terhadap variasi pesan error antar versi MySQL/MariaDB.
- `databaseMonitorWriter.Write()` melakukan konversi `[]byte → string` untuk setiap chunk; untuk output besar ini bisa menambah overhead CPU. (Tetap aman secara fungsional.)
- `execution.Engine` dan `writer.Engine` sama-sama logging ke `ErrorLog` pada beberapa jalur; ini bagus untuk observability, tapi perlu dipastikan tidak menduplikasi log terlalu banyak.

---

## 12) Ringkasan Peran Tiap Subpackage

- `internal/backup` (root): entrypoints, `Service`, orchestration, path helpers, mode config, graceful shutdown.
- `internal/backup/setup`: preflight, profile/ticket, validasi, interactive edit loop.
- `internal/backup/selection`: multi-select & aturan pemilihan DB (primary/secondary), companion expansion.
- `internal/backup/modes`: mode strategy (combined vs iterative) + interface boundaries.
- `internal/backup/execution`: engine mysqldump args + retry + loop + metadata build.
- `internal/backup/writer`: eksekusi mysqldump + streaming pipeline compress/encrypt + progress monitor.
- `internal/backup/metadata`: generate/save/update manifest JSON + export user grants.
- `internal/backup/grants`: wrapper/hook untuk user grants.
- `internal/backup/gtid`: capture GTID sebelum backup.
- `internal/backup/display`: render opsi & hasil.
