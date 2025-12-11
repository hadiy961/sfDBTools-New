# Panduan Developer: `internal/profile`

Paket `profile` bertanggung jawab untuk mengelola seluruh siklus hidup profil koneksi database. Profil adalah file `.cnf.enc` yang berisi kredensial database (host, user, password, port) dalam format INI yang dienkripsi menggunakan AES-256-GCM.

## Arsitektur & Pola Desain

- **Service Pattern**: Logika utama dienkapsulasi dalam `profile.Service`. _Struct_ ini memegang _state_ (seperti opsi dari CLI) dan dependensi (seperti _logger_ dan konfigurasi aplikasi).
- **Constructor Tunggal**: `NewProfileService` berfungsi sebagai satu-satunya titik masuk untuk membuat _service_. Ia menerima `interface{}` dan menggunakan _type switch_ untuk menentukan mode operasi (`create`, `edit`, `show`, `delete`) dan menginisialisasi _struct_ opsi yang relevan.
- **Pemisahan Logika**:
  - **Handler Utama** (`profile_create.go`, `profile_edit.go`, dll.): Setiap file berisi fungsi _handler_ tingkat tinggi (misalnya, `CreateProfile`, `EditProfile`) yang mengorkestrasi alur kerja untuk satu perintah.
  - **Wizard & Prompt** (`profile_wizard.go`, `profile_prompt.go`): Logika untuk mode interaktif dipisahkan ke dalam `runWizard` yang memanggil fungsi-fungsi _prompt_ individual (misalnya, `promptDBConfigName`, `promptProfileInfo`).
  - **Validasi & Penyimpanan** (`profile_validator.go`, `profile_save.go`): Logika untuk memvalidasi input (misalnya, keunikan nama profil) dan proses penyimpanan (termasuk enkripsi dan penulisan file) diisolasi dalam fungsi-fungsinya sendiri.

## Alur Kerja Utama

### 1. Alur Pembuatan Profil (`CreateProfile`)

1.  **Inisialisasi**: `NewProfileService` dibuat dengan `types.ProfileCreateOptions`.
2.  **Mode Deteksi**: `CreateProfile` memeriksa _flag_ `Interactive`.
    - **Mode Interaktif**: `runWizard("create")` dipanggil.
      - Wizard akan memanggil serangkaian _prompt_ (`promptDBConfigName`, `promptProfileInfo`) untuk mengisi `s.ProfileInfo`.
      - Pengguna diminta konfirmasi sebelum melanjutkan.
      - Kunci enkripsi diminta dari pengguna atau dibaca dari `ENV`.
    - **Mode Non-Interaktif**:
      - Parameter dari _flag_ CLI divalidasi (`ValidateProfileInfo`).
      - Keunikan nama file profil diperiksa (`CheckConfigurationNameUnique`).
3.  **Validasi Koneksi**: `SaveProfile` dipanggil, yang pertama-tama menjalankan `database.ConnectionTest` untuk memverifikasi kredensial.
    - Jika koneksi gagal (dan dalam mode interaktif), pengguna ditanya apakah ingin tetap menyimpan profil.
4.  **Enkripsi & Penyimpanan**:
    - Konten profil diformat sebagai string INI (`formatConfigToINI`).
    - Konten dienkripsi menggunakan `encrypt.EncryptAES`.
    - Hasil enkripsi ditulis ke file `.cnf.enc` di direktori konfigurasi.

### 2. Alur Pengeditan Profil (`EditProfile`)

1.  **Inisialisasi**: `NewProfileService` dibuat dengan `types.ProfileEditOptions`.
2.  **Mode Deteksi**: `EditProfile` memeriksa _flag_ `Interactive`.
    - **Mode Interaktif**: `runWizard("edit")` dipanggil.
      - Wizard pertama-tama memanggil `promptSelectExistingConfig` untuk memungkinkan pengguna memilih profil yang akan diedit.
      - Profil yang ada dimuat dan didekripsi (`profileselect.LoadAndParseProfile`). Salinan datanya disimpan di `s.OriginalProfileInfo` untuk perbandingan.
      - Wizard kemudian menampilkan _prompt_ untuk setiap _field_, dengan nilai yang ada sebagai _default_.
    - **Mode Non-Interaktif**:
      - _Flag_ `--file` wajib ada.
      - Profil yang ada dimuat sebagai `OriginalProfileInfo`.
      - Nilai-nilai dari _flag_ CLI lainnya menimpa nilai yang ada.
3.  **Validasi & Penyimpanan**: Alurnya mirip dengan _create_, dengan logika tambahan dalam `SaveProfile` untuk:
    - Mendeteksi jika nama profil diubah.
    - Jika nama diubah, file lama akan dihapus setelah file baru berhasil dibuat (proses _rename_).

## Komponen Penting

- **`Service`**: _Struct_ pusat yang menampung semua dependensi dan _state_.
- **`types.Profile*Options`**: _Struct_ yang membawa data dari _layer_ `parsing` ke _service_.
- **`s.ProfileInfo`**: _State_ yang "hidup" yang diisi baik oleh _flag_ CLI maupun input interaktif.
- **`s.OriginalProfileInfo`**: Salinan _read-only_ dari profil yang dimuat dalam mode `edit` atau `show`, digunakan untuk perbandingan dan tampilan.
- **`profile_save.go`**: Berisi logika kritis untuk menguji koneeksi, mengenkripsi, dan menulis file, termasuk menangani _rename_ dan _overwrite_.
- **`pkg/input` & `pkg/ui`**: `profile` sangat bergantung pada paket-paket ini untuk semua interaksi dan tampilan pengguna.

## Ketergantungan Eksternal

- **`internal/profileselect`**: Digunakan untuk menampilkan menu pemilihan file profil.
- **`pkg/encrypt`**: Digunakan untuk enkripsi dan dekripsi file `.cnf.enc`.
- **`pkg/fsops`**: Untuk semua operasi baca/tulis file.
- **`pkg/input`**: Untuk semua _prompt_ interaktif.
- **`pkg/ui`**: Untuk menampilkan tabel ringkasan, _header_, dan pesan status.
