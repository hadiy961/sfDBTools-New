# Dokumentasi `pkg` Direktori

Direktori `pkg` berisi paket-paket bersama yang dapat digunakan kembali di seluruh bagian aplikasi `sfDBTools`. Paket-paket ini dirancang untuk menjadi independen dari logika bisnis utama dan menyediakan fungsionalitas umum seperti interaksi database, operasi file system, enkripsi, dan UI.

## Indeks Paket

- [`backuphelper`](#pkgbackuphelper)
- [`compress`](#pkgcompress)
- [`consts`](#pkgconsts)
- [`cryptohelper`](#pkgcryptohelper)
- [`database`](#pkgdatabase)
- [`defaultval`](#pkgdefaultval)
- [`encrypt`](#pkgencrypt)
- [`errorlog`](#pkgerrorlog)
- [`flags`](#pkgflags)
- [`fsops`](#pkgfsops)
- [`global`](#pkglobal)
- [`helper`](#pkghelper)
- [`input`](#pkginput)
- [`parsing`](#pkgparsing)
- [`process`](#pkgprocess)
- [`profilehelper`](#pkgprofilehelper)
- [`servicehelper`](#pkgservicehelper)
- [`ui`](#pkgui)
- [`validation`](#pkgvalidation)

---

### `pkg/backuphelper`

**Tujuan:** Menyediakan fungsi-fungsi bantuan (_helper_) yang murni dan tanpa ketergantungan state untuk operasi backup. Logika di dalam paket ini bersifat portabel dan berfokus pada tugas-tugas spesifik seperti membangun argumen `mysqldump` dan memfilter daftar database.

**Files:**
- `logic.go`: Berisi logika untuk memfilter kandidat database berdasarkan mode backup (`primary`, `secondary`, `single`) dan mengekstrak versi `mysqldump` dari output.
- `mysqldump.go`: Berisi fungsi untuk membangun _command-line arguments_ yang akan dieksekusi oleh `mysqldump`. Ini juga mencakup logika untuk mendeteksi _error_ fatal dan menyembunyikan _password_ dari _log_.

**Fungsi Utama:**

- `BuildMysqldumpArgs(...) []string`: Fungsi inti yang merakit _slice_ dari argumen baris perintah untuk `mysqldump`. Fungsi ini secara dinamis menambahkan flag seperti `--host`, `--port`, `--databases`, atau `--all-databases` berdasarkan parameter koneksi dan filter yang diberikan.
- `FilterCandidatesByMode(...) []string`: Menerima _slice_ nama database dan mode backup (`primary`, `secondary`), lalu mengembalikan daftar database yang telah difilter sesuai dengan aturan penamaan yang berlaku untuk mode tersebut.
- `ExtractMysqldumpVersion(...) string`: Mengekstrak string versi `mysqldump` dari output _stderr_, yang berguna untuk pencatatan metadata.
- `IsFatalMysqldumpError(...) bool`: Menganalisis pesan _error_ dari `mysqldump` untuk menentukan apakah _error_ tersebut bersifat fatal (misalnya, "access denied") atau hanya peringatan yang bisa diabaikan.
- `MaskPasswordInArgs(...) []string`: Menerima _slice_ argumen `mysqldump` dan mengembalikan salinannya dengan nilai _password_ yang telah disamarkan (`********`) untuk keamanan logging.

---
### `pkg/compress`

**Tujuan:** Menyediakan antarmuka terpadu untuk operasi kompresi dan dekompresi data menggunakan berbagai algoritma. Paket ini dirancang untuk performa tinggi dengan memanfaatkan kompresi paralel (_multi-threaded_) jika memungkinkan (misalnya, `pgzip` dan `zstd`).

**Files:**
- `compress_const.go`: Mendefinisikan tipe dan konstanta utama seperti `CompressionType` (e.g., `Gzip`, `Zstd`) dan `CompressionLevel` (1-9).
- `compress_writer.go`: Berisi implementasi `factory` untuk membuat _writer_ yang sesuai dengan algoritma kompresi yang dipilih. Ini adalah komponen _low-level_ yang menangani penulisan data terkompresi secara _on-the-fly_.
- `compress_decompress.go`: Menyediakan `factory` untuk membuat _reader_ yang dapat mendekompilasi data dari berbagai format. Juga berisi fungsi untuk mendeteksi tipe kompresi dari ekstensi file.
- `compress_main.go`: Menyediakan fungsi _high-level_ seperti `CompressFile` untuk mengompresi seluruh file dan `GetFileExtension` untuk mendapatkan ekstensi file yang benar.
- `compress_validation.go`: Berisi fungsi validasi untuk memastikan tipe dan level kompresi yang diberikan didukung.
- `compress_ratio_data.go`: Menyimpan data empiris tentang rasio kompresi yang diharapkan untuk setiap algoritma dan level. Data ini dapat digunakan untuk estimasi ukuran file.

**Fitur Utama:**

- **Dukungan Algoritma:** Mendukung berbagai algoritma kompresi populer, termasuk `gzip`, `pgzip` (paralel), `zstd`, `zlib`, dan `xz`.
- **Streaming:** Didesain untuk bekerja dengan `io.Writer` dan `io.Reader`, memungkinkan kompresi/dekompresi data secara _streaming_ tanpa perlu memuat seluruh file ke dalam memori.
- **Performa Tinggi:** Menggunakan implementasi paralel (multi-threaded) untuk `pgzip` dan `zstd` untuk memaksimalkan _throughput_ pada mesin dengan banyak CPU.
- **Antarmuka Terpadu:**
    - `NewCompressingWriter(...)`: Membuat `io.WriteCloser` yang secara transparan mengompresi semua data yang ditulis kepadanya.
    - `NewDecompressingReader(...)`: Membuat `io.ReadCloser` yang secara transparan mendekompilasi data saat dibaca.
- **Fungsi Utilitas:**
    - `DetectCompressionTypeFromFile(path string)`: Mendeteksi jenis kompresi dari ekstensi file (misalnya, `.gz`, `.zst`).
    - `ValidateCompressionType(type string)`: Memvalidasi apakah string tipe kompresi didukung.

---
### `pkg/consts`

**Tujuan:** Menyediakan lokasi terpusat untuk semua konstanta global yang digunakan di seluruh aplikasi, terutama nama-nama _environment variable_. Ini membantu menghindari _"magic strings"_ dan memudahkan pengelolaan konfigurasi.

**Files:**
- `consts_env.go`: Mendefinisikan semua nama _environment variable_ yang dikenali oleh `sfDBTools`. Variabel-variabel ini digunakan untuk mengkonfigurasi koneksi database, kunci enkripsi, dan mode operasi dari luar aplikasi.
- `consts_backup.go`: Saat ini kosong, tetapi dimaksudkan untuk menampung konstanta yang spesifik untuk operasi _backup_ dan _restore_.

**Contoh Konstanta:**
- `ENV_DB_HOST`: Nama _environment variable_ untuk _host_ database (`SFDB_DB_HOST`).
- `ENV_DB_USER`: Nama _environment variable_ untuk _user_ database (`SFDB_DB_USER`).
- `ENV_DB_PASSWORD`: Nama _environment variable_ untuk _password_ database (`SFDB_DB_PASSWORD`).
- `ENV_ENCRYPTION_KEY`: Nama _environment variable_ untuk kunci enkripsi generik (`SFDB_ENCRYPTION_KEY`).
- `ENV_QUIET`: Nama _environment variable_ untuk mengaktifkan _quiet mode_ (`SFDB_QUIET`).

---
### `pkg/cryptohelper`

**Tujuan:** Menyediakan fungsi-fungsi bantuan untuk menangani berbagai mode input pengguna yang diperlukan oleh perintah-perintah kriptografi (`encrypt`, `decrypt`, `base64`, dll.). Paket ini mengabstraksi logika untuk membaca data dari _flag_, _standard input (stdin)_, atau melalui _prompt_ interaktif.

**Files:**
- `cryptohelper_input.go`: Berisi fungsi dasar untuk membaca data dari _flag_ atau _pipe_ (stdin).
- `cryptohelper_interactive.go`: Berisi fungsi yang lebih canggih yang akan mencoba membaca dari _flag/pipe_ terlebih dahulu, dan jika tidak ada, akan beralih (_fallback_) ke _prompt_ interaktif yang meminta pengguna untuk memasukkan data atau _path_ file.
- `cryptohelper_quiet.go`: Menyediakan fungsi untuk mengatur _"quiet mode"_.

**Fitur Utama:**

- **Input Fleksibel**:
  - `GetInputBytes(flagVal string)` dan `GetInputString(flagVal string)`: Membaca input dari _flag_ `--input` atau dari _stdin_ jika ada data yang di-_pipe_. Ideal untuk skrip otomatis.
- **Mode Interaktif**:
  - `GetInteractiveInputBytes(prompt string)`: Menampilkan _prompt_ kepada pengguna dan menerima input multi-baris hingga pengguna menekan `Ctrl+D`.
- **Fallback Cerdas**:
  - `GetInputBytesOrInteractive(flagVal, prompt string)` dan `GetInputStringOrInteractive(flagVal, prompt string)`: Menggabungkan dua mode di atas. Fungsi ini adalah "pilihan terbaik" yang pertama-tama memeriksa _flag_ dan _pipe_, lalu beralih ke mode interaktif jika tidak ada input lain. Ini memberikan pengalaman pengguna yang mulus baik untuk penggunaan manual maupun otomatis.
  - `GetFilePathOrInteractive(...)`: Fungsi serupa yang khusus untuk mendapatkan _path_ file, dengan validasi bawaan untuk memeriksa apakah file tersebut ada.
- **Quiet Mode**:
  - `SetupQuietMode(logger)`: Memeriksa _environment variable_ `SFDB_QUIET`. Jika diatur, semua _log_ akan dialihkan ke `stderr`, sehingga `stdout` tetap bersih dan hanya berisi output data murni (misalnya, teks terenkripsi). Ini sangat penting agar hasil perintah dapat di-_pipe_ ke perintah lain.

---
### `pkg/database`

**Tujuan:** Menyediakan lapisan abstraksi yang kuat dan terpusat untuk semua interaksi dengan database MySQL/MariaDB. Paket ini mengenkapsulasi logika koneksi, eksekusi _query_, pemfilteran, dan pengumpulan data, sehingga paket-paket lain tidak perlu berinteraksi langsung dengan driver `sql`.

**Struktur Utama:**
- `Client`: Struct utama yang memegang _connection pool_ (`*sql.DB`) dan menjadi _receiver_ untuk semua metode interaksi database.
- `Config`: Struct yang menyimpan parameter koneksi (host, user, password, dll.) dan digunakan untuk membuat `DSN` (Data Source Name).

**Files & Fungsionalitas:**

1.  **Koneksi & Konfigurasi (`database_config.go`, `database_connection.go`)**
    - **`NewClient(...)`**: Fungsi _factory_ utama untuk membuat `Client` baru. Fungsi ini membuka koneksi, mengkonfigurasi _connection pool_ (misalnya, `MaxOpenConns`, `ConnMaxLifetime`), dan melakukan `Ping` untuk memvalidasi koneksi.
    - **`ConnectToSourceDatabase(...)`**, **`ConnectToDestinationDatabase(...)`**: Fungsi _helper_ tingkat tinggi yang membungkus `NewClient` dengan konfigurasi standar untuk koneksi "sumber" dan "tujuan".
    - **`DSN()`**: Metode pada `Config` yang menghasilkan _string_ DSN yang kompatibel dengan driver `go-sql-driver/mysql`.

2.  **Pengambilan Data & Metrik (`database_count.go`, `database_gtid.go`, `database_user.go`)**
    - **`GetFullGTIDInfo(...)`**: Mengambil informasi `MASTER STATUS` dan posisi GTID, yang krusial untuk _backup_ dan replikasi.
    - **`ExportAllUserGrants(...)`**: Mengambil `SHOW GRANTS` untuk semua pengguna di server dan memformatnya sebagai skrip SQL yang dapat dieksekusi.
    - **`GetDatabaseSize(...)`**, **`GetTableCount(...)`**, dll.: Kumpulan fungsi untuk mengambil metrik spesifik dari sebuah database (ukuran, jumlah tabel, _view_, _procedure_, dll.). Fungsi-fungsi ini dioptimalkan untuk kecepatan, misalnya dengan menggunakan `SHOW TABLES` daripada _query_ ke `information_schema`.

3.  **Pengumpulan Detail Database Secara Concurrent (`database_detail_collector.go`, `database_dbscan_query.go`)**
    - **`CollectDatabaseDetailsWithOptions(...)`**: Fungsi canggih yang menggunakan _worker pool_ (goroutines) untuk mengumpulkan detail (ukuran, jumlah tabel, dll.) dari banyak database secara bersamaan.
    - **Pola Worker**: Menggunakan _channels_ (`jobs`, `results`) dan `sync.WaitGroup` untuk mendistribusikan pekerjaan ke beberapa _worker_, secara signifikan mempercepat proses `db-scan`.
    - **`SaveDatabaseDetail(...)`**: Menyimpan detail database yang terkumpul ke dalam tabel di "database aplikasi" menggunakan _stored procedure_ `sp_insert_database_detail`.

4.  **Pemfilteran Database (`database_filter.go`, `database_filter_helper.go`)**
    - **`FilterDatabases(...)`**: Logika inti untuk memfilter daftar database. Mendukung penyertaan (_whitelist_) dan pengecualian (_blacklist_) baik dari argumen baris perintah maupun dari file.
    - **`IsSystemDatabase(...)`**: Fungsi _helper_ untuk memeriksa apakah sebuah database termasuk database sistem (misalnya, `mysql`, `information_schema`).

5.  **Manajemen Variabel Server (`database_global_var.go`, `database_session_var.go`)**
    - Menyediakan fungsi untuk mendapatkan dan mengatur variabel server seperti `max_statement_time`.
    - **`WithGlobalMaxStatementTime(...)`**: Fungsi penting yang memungkinkan pengaturan variabel `GLOBAL` (memerlukan hak akses `SUPER`) dan secara otomatis mengembalikannya ke nilai semula setelah selesai. Ini penting untuk memastikan `mysqldump` tidak _timeout_.

---
### `pkg/defaultval`

**Tujuan:** Menyediakan fungsi-fungsi _factory_ untuk membuat dan menginisialisasi struct opsi (_options struct_) dengan nilai-nilai _default_. Paket ini bertindak sebagai jembatan antara konfigurasi aplikasi statis (dari `config.yaml`), _environment variable_, dan nilai _default_ yang aman, yang kemudian digunakan untuk mengisi _flag_ pada _command-line interface_ (CLI).

**Files:**
- `default_backup.go`: Membuat `types_backup.BackupDBOptions` dengan nilai _default_ dari konfigurasi `backup`.
- `default_cleanup.go`: Membuat `types.CleanupOptions` dari konfigurasi `backup.cleanup`.
- `default_dbscan.go`: Membuat `types.ScanOptions` dari konfigurasi `backup.include` dan _environment variable_ terkait `db-scan`.
- `default_profile.go`: Membuat opsi _default_ untuk perintah-perintah yang berhubungan dengan profil (`profile create`, `profile show`).
- `default_dbinfo.go`: Menyediakan nilai _default_ dasar untuk koneksi database.

**Alur Kerja:**
1.  Sebuah fungsi di paket ini (misalnya, `DefaultBackupOptions`) dipanggil.
2.  Fungsi tersebut memuat konfigurasi utama aplikasi dari `internal/appconfig`.
3.  Ia juga membaca _environment variable_ yang relevan (menggunakan `pkg/helper.GetEnvOrDefault`).
4.  Nilai-nilai dari konfigurasi dan _environment variable_ digunakan untuk mengisi _field-field_ dalam sebuah _struct_ opsi (misalnya, `BackupDBOptions`).
5.  _Struct_ yang sudah terisi ini dikembalikan dan digunakan oleh `pkg/flags` untuk mendefinisikan nilai _default_ dari setiap _flag_ CLI.

**Fitur Utama:**
- **Pusat Konfigurasi Default**: Menyatukan logika pembuatan nilai _default_ di satu tempat, sehingga CLI tetap konsisten dengan `config.yaml`.
- **Aman dari Kegagalan**: Fungsi-fungsi ini dirancang untuk tidak _panic_ jika `config.yaml` tidak ditemukan. Mereka akan mengembalikan _struct_ dengan nilai-nilai yang paling aman (misalnya, fitur dinonaktifkan) untuk mencegah perilaku yang tidak diinginkan.
- **Dinamis**: Beberapa nilai _default_, seperti `OutputDir` pada `DefaultBackupOptions`, dibuat secara dinamis dengan memproses _template_ tanggal/waktu, memberikan contoh nyata kepada pengguna saat mereka menjalankan `--help`.

---
### `pkg/encrypt`

**Tujuan:** Menyediakan fungsionalitas enkripsi dan dekripsi menggunakan AES-256-GCM yang kompatibel dengan format `openssl enc -pbkdf2`. Paket ini mendukung operasi pada data di memori (_in-memory_) maupun secara _streaming_, yang sangat penting untuk menangani file besar seperti _backup_ database.

**Files & Fungsionalitas:**

1.  **Logika Enkripsi Inti (`encrypt_aes.go`)**
    - **`EncryptAES(plaintext, passphrase)`**: Menerima _slice byte_ dan _passphrase_, lalu mengembalikan data terenkripsi dalam format OpenSSL (`Salted__` + 8-byte _salt_ + _ciphertext_).
    - **`DecryptAES(encryptedPayload, passphrase)`**: Menerima data terenkripsi dan _passphrase_, lalu mengembalikan _plaintext_ asli. Fungsi ini menangani parsing header `Salted__` dan _salt_.
    - **`deriveKey(...)`**: Menggunakan `PBKDF2` dengan `SHA256` dan 100,000 iterasi untuk menghasilkan kunci 32-byte (AES-256) dari _passphrase_ dan _salt_.

2.  **Operasi Streaming (`encrypt_writer.go`, `encrypt_reader.go`)**
    - **`NewEncryptingWriter(...)`**: Membuat `io.WriteCloser` yang mengenkripsi data secara _on-the-fly_ saat data ditulis kepadanya. Data dibagi menjadi _chunk-chunk_ (64KB) untuk efisiensi memori. Sangat cocok untuk _pipeline_ backup (`mysqldump | encrypt | ...`).
    - **`NewDecryptingReader(...)`**: Membuat `io.Reader` yang mendekripsi data secara _on-the-fly_ saat data dibaca darinya. Ini memungkinkan _restore_ file terenkripsi yang besar tanpa harus memuat seluruh file ke memori.

3.  **Operasi File (`encrypt_file.go`)**
    - **`EncryptFile(inputPath, outputPath, ...)`**: Fungsi _high-level_ untuk mengenkripsi seluruh file.
    - **`DecryptFile(inputPath, outputPath, ...)`**: Fungsi _high-level_ untuk mendekripsi seluruh file.
    - **`IsEncryptedFile(filePath)`**: Memeriksa 8 byte pertama sebuah file untuk melihat apakah file tersebut memiliki _header_ `Salted__`, yang menandakan file terenkripsi.

4.  **Input Pengguna (`encrypt_prompt.go`)**
    - **`EncryptionPrompt(...)`**: Fungsi _helper_ untuk mendapatkan _password_ enkripsi. Fungsi ini pertama-tama akan memeriksa _environment variable_ (misalnya, `SFDB_ENCRYPTION_KEY`). Jika tidak ada, ia akan menampilkan _prompt_ interaktif yang meminta pengguna untuk memasukkan _password_.

---
### `pkg/errorlog`

**Tujuan:** Menyediakan utilitas untuk mencatat (_logging_) _error_ yang terjadi selama operasi ke dalam file log yang terpisah dan terstruktur. Ini membantu dalam _debugging_ dan audit jejak kegagalan.

**Fitur Utama:**
- **`NewErrorLogger(...)`**: Membuat _logger_ baru yang dikonfigurasi untuk fitur tertentu (misalnya, "backup", "restore").
- **`LogWithOutput(...)`**: Mencatat pesan _error_ beserta _output_ detail (seperti _stderr_ dari `mysqldump`) ke dalam file log.
- **Rotasi File Harian**: Nama file log mencakup tanggal (misalnya, `sfDBTools_backup_error_2025-12-12.log`), secara efektif memisahkan log untuk setiap hari.
- **Format Terstruktur**: Setiap entri log berisi _timestamp_, detail konteks (misalnya, nama database), pesan _error_, dan _output_ tambahan, membuatnya mudah dibaca dan dianalisis.

---
### `pkg/flags`

**Tujuan:** Memusatkan definisi semua _command-line flags_ yang digunakan oleh berbagai perintah `cobra`. Paket ini memastikan konsistensi nama dan deskripsi _flag_ di seluruh aplikasi.

**Struktur:**
- Setiap file (misalnya, `flags_backup.go`, `flags_restore.go`) berisi fungsi-fungsi untuk menambahkan _flag_ yang relevan ke sebuah `*cobra.Command`.
- Fungsi-fungsi ini menerima _pointer_ ke _struct_ opsi (misalnya, `*types.BackupDBOptions`).
- Nilai _default_ untuk _flag_ diambil dari _struct_ opsi tersebut, yang sebelumnya telah diisi oleh paket `pkg/defaultval`.

**Contoh Fungsi:**
- **`AddBackupFlags(...)`**: Menambahkan semua _flag_ yang umum untuk operasi _backup_, seperti `--profile`, `--compress-type`, `--output-dir`, dan filter `--exclude-db`.
- **`AddRestoreSingleFlags(...)`**: Menambahkan _flag_ yang diperlukan untuk perintah `restore single`, termasuk menandai _flag_ penting seperti `--source` dan `--profile` sebagai wajib diisi (`MarkFlagRequired`).

---
### `pkg/fsops` (File System Operations)

**Tujuan:** Menyediakan kumpulan fungsi bantuan untuk semua operasi yang berhubungan dengan _file system_. Paket ini mengabstraksi dan menyederhanakan tugas-tugas seperti memeriksa keberadaan file, membaca/menulis file, dan membuat direktori.

**Files & Fungsionalitas:**

1.  **Pemeriksaan (`fsops_check.go`)**
    - `FileExists(path)`: Memeriksa apakah sebuah _path_ ada dan merupakan file.
    - `DirExists(path)`: Memeriksa apakah sebuah _path_ ada dan merupakan direktori.

2.  **Membaca & Menulis (`fsops_read.go`, `fsops_write.go`)**
    - `ReadLinesFromFile(path)`: Membaca file teks dan mengembalikannya sebagai _slice of strings_.
    - `WriteFile(path, data)`: Menulis _slice byte_ ke sebuah file.
    - `EnsureDir(dir)` dan `CreateDirIfNotExist(dir)`: Memastikan sebuah direktori ada, dan membuatnya jika belum ada.

3.  **Utilitas (`fsops_helper.go`)**
    - **`BuildSubdirPathFromPattern(...)`**: Fungsi canggih untuk membuat nama subdirektori secara dinamis dari sebuah _pattern_ (misalnya, `"{year}/{month}/{day}"`). Mendukung token seperti `{year}`, `{month}`, `{day}`, dan `{timestamp}`.

---
### `pkg/global`

**Tujuan:** Menyediakan fungsi dan variabel global yang bersifat umum dan dapat diakses dari mana saja di seluruh aplikasi.

**Fitur Utama:**
- **Akses Global ke Dependensi**:
  - `GetLogger()`: Mengembalikan _instance logger_ global yang disimpan di `types.Deps`.
  - `GetConfig()`: Mengembalikan _instance_ konfigurasi aplikasi global.
- **Pemformatan**:
  - `FormatFileSize(bytes)`: Mengubah ukuran file dalam _byte_ menjadi format yang mudah dibaca manusia (misalnya, "1.2 MB") menggunakan `go-humanize`.
  - `FormatDuration(d)`: Mengubah `time.Duration` menjadi format jam, menit, dan detik.

---
### `pkg/helper`

**Tujuan:** Menyediakan kumpulan fungsi bantuan (_helper_) yang bersifat umum, tidak memiliki ketergantungan pada _state_ aplikasi, dan dapat digunakan kembali di berbagai paket. Paket ini berisi utilitas untuk manipulasi _string_, _slice_, _path_, dan _environment variable_.

**Files & Fungsionalitas:**

- **`helper_flag.go`**:
  - `GetStringFlagOrEnv(...)` dan varian lainnya (`Int`, `Bool`, `StringSlice`): Fungsi-fungsi ini adalah inti dari sistem konfigurasi. Mereka mencoba mengambil nilai dari _flag_ `cobra` terlebih dahulu. Jika _flag_ tidak diatur, mereka akan beralih (_fallback_) untuk membaca nilai dari _environment variable_ yang sesuai. Ini memberikan fleksibilitas konfigurasi yang tinggi kepada pengguna.

- **`helper_list.go`**:
  - `ListTrimNonEmpty(...)`: Membersihkan _slice of strings_ dari spasi berlebih dan entri kosong.
  - `StringSliceContainsFold(...)`: Memeriksa keberadaan sebuah _string_ dalam _slice_ (tanpa mempedulikan besar/kecil huruf).
  - `ListUnique(...)`: Menghapus entri duplikat dari sebuah _slice_.
  - `ListSubtract(...)`: Mengembalikan elemen-elemen dari _slice_ A yang tidak ada di _slice_ B.

- **`helper_path.go` & `helper_ext.go`**:
  - `GenerateBackupFilename(...)`: Membuat nama file backup yang terstandardisasi berdasarkan _pattern_ yang mencakup nama database, _timestamp_, dan _hostname_.
  - `GenerateBackupDirectory(...)`: Membuat _path_ direktori backup berdasarkan _pattern_ (misalnya, `{year}{month}{day}`).
  - `StripAllBackupExtensions(...)`: Menghapus semua ekstensi yang dikenal (`.sql`, `.gz`, `.enc`) dari nama file untuk mendapatkan nama dasar database.
  - `ExpandPath(...)`: Mengubah _tilde_ (`~`) menjadi _path_ direktori _home_ pengguna.

- **`helper_env.go`**:
  - `GetEnvOrDefault(...)`: Mengambil nilai dari _environment variable_ atau mengembalikan nilai _default_ jika tidak ada.

- **`helper_timer.go`**:
  - `NewTimer()`: Menyediakan objek _timer_ sederhana untuk mengukur durasi eksekusi sebuah operasi, dengan fungsi `Elapsed()` untuk mendapatkan hasilnya.

- **Lainnya**: `helper_encrypt.go` dan `helper_profile.go` berisi fungsi-fungsi kecil yang mendukung operasi enkripsi dan profil.

---
### `pkg/input`

**Tujuan:** Mengelola semua bentuk input interaktif dari pengguna. Paket ini membungkus _library_ `survey/v2` untuk menyediakan antarmuka yang konsisten dan ramah pengguna untuk _prompt_, menu pilihan, dan input _password_.

**Files & Fungsionalitas:**
- **`input_general.go`**:
  - `ShowMenu(...)` & `ShowMultiSelect(...)`: Menampilkan menu pilihan (_single_ atau _multi-choice_) kepada pengguna dan mengembalikan pilihan mereka.
  - `AskPassword(...)`: Menampilkan _prompt_ untuk memasukkan _password_ dengan input yang disamarkan.
  - `AskString(...)`, `AskInt(...)`, `AskYesNo(...)`: Menampilkan _prompt_ untuk berbagai tipe data dengan dukungan untuk nilai _default_ dan validasi.

- **`input_validator.go`**:
  - Menyediakan sekumpulan fungsi validator yang siap pakai untuk `survey`.
  - `ValidatePort(...)`: Memastikan input adalah nomor _port_ yang valid (1-65535).
  - `ValidateFilename(...)`: Memastikan input adalah nama file yang aman.
  - `ValidateNonEmpty(...)`: Memastikan input tidak kosong.
  - `ComposeValidators(...)`: Menggabungkan beberapa validator menjadi satu.

---
### `pkg/parsing`

**Tujuan:** Berfungsi sebagai lapisan "parser" yang mengambil _input mentah_ dari _flag-flag_ `cobra` dan _environment variables_ lalu mengubahnya menjadi _struct_ opsi yang terstruktur dan tervalidasi (misalnya, `types.BackupDBOptions`). Paket ini adalah langkah terakhir sebelum _struct_ opsi diserahkan ke _service layer_ untuk dieksekusi.

**Alur Kerja:**
1.  Sebuah perintah `cobra` dijalankan (misalnya, `sfdbtools backup all --exclude-system`).
2.  Fungsi yang relevan di `pkg/parsing` (misalnya, `ParsingBackupOptions`) dipanggil.
3.  Fungsi ini menggunakan `pkg/helper` (`GetStringFlagOrEnv`, dll.) untuk membaca nilai dari setiap _flag_ dan/atau _environment variable_.
4.  Ia mengisi _struct_ opsi (misalnya, `types_backup.BackupDBOptions`) dengan nilai-nilai yang telah dibaca dan dinormalisasi.
5.  Untuk kasus yang kompleks seperti `db-scan filter`, paket ini juga melakukan logika tambahan seperti menggabungkan daftar _include_ dan _exclude_, lalu melakukan operasi "subtract" untuk mendapatkan daftar final database yang akan diproses.
6.  _Struct_ opsi yang sudah lengkap dikembalikan dan siap digunakan.

**Files:**
- Setiap file `parsing_*.go` dikhususkan untuk mem-_parsing_ _flag_ dari satu atau sekelompok perintah tertentu (misalnya, `parsing_backup.go` untuk perintah-perintah _backup_, `parsing_profile_flags.go` untuk perintah-perintah _profile_).
- `parsing_ini.go`: Menyediakan parser sederhana untuk file format `.ini`, yang digunakan untuk membaca konfigurasi dari file profil.

---
### `pkg/process`

**Tujuan:** Mengelola eksekusi proses eksternal, terutama untuk menjalankan aplikasi dalam mode _daemon_ (latar belakang).

**Files & Fungsionalitas:**
- **`proses_daemon.go`**:
  - `SpawnDaemon(...)`: Fungsi inti untuk memulai proses baru di latar belakang. Fungsi ini mengalihkan `stdout` dan `stderr` ke file log, dan secara opsional menulis ID proses (PID) ke dalam `pidfile`.
  - `CheckAndReadPID(...)`: Memeriksa apakah `pidfile` ada, membaca PID di dalamnya, dan memverifikasi apakah proses dengan PID tersebut masih berjalan. Ini digunakan untuk mencegah menjalankan _daemon_ yang sama lebih dari satu kali.

---
### `pkg/profilehelper`

**Tujuan:** Menyediakan fungsi-fungsi bantuan tingkat tinggi untuk memuat, me-resolve, dan menggunakan profil database. Paket ini menggabungkan logika dari `internal/profileselect` dan `pkg/database` untuk menyederhanakan proses koneksi ke database menggunakan profil.

**Files & Fungsionalitas:**
- **`profilehelper_loader.go`**:
  - **`ResolveAndLoadProfile(...)`**: Fungsi pusat yang sangat penting. Ia me-resolve _path_ profil dari berbagai sumber dengan urutan prioritas: _flag_ CLI -> _environment variable_ -> _prompt_ interaktif. Setelah _path_ ditemukan, ia memuat dan mendekripsi file profil.
  - `LoadTargetProfile(...)` dan `LoadSourceProfile(...)`: _Wrapper_ yang lebih spesifik di sekitar `ResolveAndLoadProfile` yang telah dikonfigurasi untuk skenario "target" (restore) dan "sumber" (backup/scan).

- **`profilehelper_connection.go`**:
  - **`ConnectWithProfile(...)`**: Fungsi _helper_ yang menerima _struct_ `ProfileInfo` yang sudah dimuat, lalu memanggil `database.ConnectToSourceDatabase` untuk membuat koneksi. Ini menyembunyikan detail pembuatan kredensial.

---
### `pkg/servicehelper`

**Tujuan:** Menyediakan blok bangunan dasar untuk _service layer_ di dalam aplikasi. Paket ini menawarkan fungsionalitas umum seperti penanganan _graceful shutdown_ dan _mutex locking_.

**Files & Fungsionalitas:**
- **`servicehelper_base.go`**:
  - **`BaseService`**: Sebuah _struct_ yang dapat disematkan (_embedded_) ke dalam _struct service_ lain.
  - `SetCancelFunc(...)` & `Cancel()`: Menyimpan dan memanggil `context.CancelFunc`, yang merupakan mekanisme utama untuk _graceful shutdown_ di seluruh aplikasi.
  - `WithLock(fn)`: Menjalankan sebuah fungsi di dalam _mutex lock_, memastikan operasi tersebut aman dari _race conditions_ (_thread-safe_).

- **`servicehelper_progress.go`**:
  - **`TrackProgress(tracker)`**: Fungsi _helper_ cerdas yang menggunakan `defer` untuk secara otomatis mengatur status "in progress" menjadi `true` di awal operasi dan `false` di akhir, bahkan jika terjadi _panic_.

---
### `pkg/ui`

**Tujuan:** Mengelola semua output ke _terminal_ pengguna. Paket ini bertanggung jawab untuk menampilkan _header_, tabel, _spinner_, pesan berwarna, dan elemen antarmuka pengguna lainnya secara konsisten.

**Files & Fungsionalitas:**
- **`ui_formatting.go`**:
  - `PrintHeader(...)`, `PrintSubHeader(...)`: Mencetak judul dengan _border_ dan pemisah.
  - `PrintSuccess(...)`, `PrintError(...)`, `PrintWarning(...)`: Mencetak pesan dengan warna dan ikon yang sesuai (hijau/✅, merah/❌, kuning).
  - `FormatTable(...)`: Menggunakan `olekukonko/tablewriter` untuk merender data dalam format tabel yang rapi.

- **`ui_spinner.go`**:
  - `NewSpinnerWithElapsed(...)`: Membuat _spinner_ animasi (`briandowns/spinner`) yang secara otomatis menampilkan durasi waktu yang telah berlalu (misalnya, "... (1m 23s)").

- **`ui_terminal.go`**:
  - `ClearScreen()`: Membersihkan layar _terminal_ dengan perintah yang sesuai untuk sistem operasi (misalnya, `clear` di Linux, `cls` di Windows).
  - `GetTerminalSize()`: Mendeteksi lebar dan tinggi _terminal_.

- **`ui_filter_stats.go`**:
  - `DisplayFilterStats(...)`: Komponen UI yang dapat digunakan kembali untuk menampilkan statistik hasil pemfilteran database dalam format tabel yang informatif.

---
### `pkg/validation`

**Tujuan:** Menyediakan fungsi-fungsi validasi kecil yang dapat digunakan kembali.

**Files & Fungsionalitas:**
- **`validation_backup_dir.go`**:
  - `ValidateSubdirPattern(...)`: Memvalidasi sintaks dari _pattern_ subdirektori backup, memastikan tidak ada token yang tidak dikenal atau _path traversal_.

- **`validation_ext.go`**:
  - `ProfileExt(...)`: Memastikan nama file profil selalu memiliki ekstensi `.cnf.enc`.

- **`validation_handler.go`**:
  - `HandleInputError(...)`: Mengubah _error_ `terminal.InterruptErr` (dari `Ctrl+C`) menjadi _error_ kustom `ErrUserCancelled` untuk penanganan yang lebih konsisten.
