% Review: internal/profile

Ringkasan singkat
-----------------
Modul `internal/profile` menyediakan Service untuk membuat, mengedit, menghapus, dan menampilkan profil konfigurasi database yang tersimpan sebagai file terenkripsi (format .cnf.enc). Modul ini menggunakan paket internal lain (profileselect, types) serta banyak utilitas dari `pkg/*` (fsops, encrypt, input, ui, validation, helper, database, dll.). Secara umum alur logika terlihat modular dan mengikuti pemisahan tanggung jawab antara prompt/input, validasi, enkripsi, I/O, dan tampilan.

Temuan fungsional & logika
-------------------------
- Create/Edit/Show/Delete dipecah dengan jelas ke dalam method yang spesifik.
- Flow interaktif vs non-interaktif ditangani (flags vs wizard interactive). Logika fallback, preservasi password saat edit, dan rename flow pada save ditangani.
- Sebagian besar error dibungkus atau dioperkan kembali, dan banyak pesan user-friendly (ui.PrintError/Warning/Info) digunakan.
- Terdapat cek koneksi DB sebelum menyimpan (database.ConnectionTest) dan opsi untuk menyimpan walau koneksi gagal di interactive mode.

Area yang perlu diperbaiki / catatan
----------------------------------
1) Inisialisasi `Service.Config` pada NewService
	 - NewService menerima cfg dan logs tetapi pada path default kode hanya meng-assign Config jika ada; di kode sekarang Config sudah di-assign, namun ada komentar "Perbaikan" yang menunjukkan riwayat perubahan. Pastikan semua pemanggil memang memberikan `cfg` yang valid; fungsi NewService sudah meng-assign Config tapi jika nil tidak ada fallback.

2) Penanganan error yang terlalu umum / pesan yang bocor
	 - Beberapa error dikembalikan langsung ke pemanggil tanpa pembungkusan (`fmt.Errorf("gagal menyimpan file konfigurasi: %v", err)`) — lebih konsisten gunakan `%w` untuk membungkus jika ingin preserve error chain.
	 - Pada beberapa tempat user-visible errors memuat detail internal seperti path atau error asli dari OS; itu kadang diperlukan, tapi pastikan tidak mengekspos data sensitif.

3) Validasi dan aturan nama file
	 - Fungsi `buildFileName` dan helper `TrimProfileSuffix` membantu menghindari double suffix. Baik.
	 - Namun ada cek pada `promptDBConfigName` yang membolehkan default name "my_database_config" — pastikan ini sesuai kebutuhan produksi.

4) Concurrency
	 - Modul ini tidak menampilkan goroutine eksplisit atau akses bersama state lintas goroutine. Tidak ada isu konkurensi yang jelas.

5) Testing
	 - Tidak ada test unit yang menyertai fungsi-fungsi ini. Rekomendasi: tambahkan unit test untuk:
		 * buildFileName / filePathInConfigDir
		 * formatConfigToINI
		 * flow SaveProfile (mock fsops + encrypt + database)

6) Penggunaan environment variable dan keamanan password
	 - Untuk create flow, jika ENV SFDB_DB_PASSWORD diset maka password otomatis diambil — ini praktis.
	 - Untuk edit flow, repo memilih untuk tidak otomatis menimpa password dari env (baik). Namun saat menampilkan `reveal-password` ada prompt ulang encryption key dan file dibuka; pastikan proses ini tidak menuliskan password ke log.

7) Logging
	 - Banyak penggunaan `s.Log.Info/Warn/Error`. Pastikan logger yang disuntikkan (types.Deps.Logger) tidak menulis password/secret.

8) Dependensi input interaktif
	 - Beberapa fungsi mengandalkan paket `survey` via `pkg/input` wrapper. Pastikan saat menjalankan non-interactive environment (CI) jalur non-interaktif benar-benar tidak memanggil prompt.

Gaya Go & praktik
------------------
- Struktur package `profile` rapi, dengan Service struct yang memegang state.
- Penamaan umumnya mengikuti idiom Go.
- Receiver method menggunakan pointer receiver pada `Service`, tepat karena method memodifikasi state.
- Beberapa fungsi mengembalikan `fmt.Errorf("... %v", err)` — di Go modern disarankan menggunakan `%w` untuk wrapping (`fmt.Errorf("...: %w", err)`) ketika ingin membiarkan caller melakukan errors.Is/As.

Error handling
--------------
- Sebagian besar error ditangani, terutama input error yang dipetakan lewat `validation.HandleInputError`.
- Perlu konsistensi penggunaan wrapping (`%w`) untuk memungkinkan pemeriksaan error di pemanggil.

Keamanan & performa
--------------------
- Enkripsi menggunakan `encrypt.EncryptAES` — pastikan implementasi encrypt menggunakan GCM atau mode yang aman (tidak hanya AES-ECB).
- File disimpan sebagai `.cnf.enc`. Pastikan file permission (fsops.WriteFile) membuat file dengan permission yang ketat (600) untuk mencegah akses tidak sah.
- Saat membaca ukuran dan last modified digunakan `os.Stat` — baik.

Maintainability & dokumentasi
-----------------------------
- Fungsi dan file relatif singkat dan terpisah tugasnya.
- Beberapa fungsi memiliki komentar yang menjelaskan tujuan; dapat ditingkatkan lagi dengan godoc pada fungsi-fungsi publik (walau package internal, tetap berguna).

Rekomendasi perbaikan (prioritas)
--------------------------------
1. Tambahkan unit test untuk fungsional kritikal (formatConfigToINI, buildFileName, SaveProfile dengan mock fsops/encrypt/db).
2. Konsistensi wrapping error: gunakan `%w` saat meneruskan error.
3. Pastikan enkripsi menggunakan mode aman dan fsops.WriteFile men-set permission file yang aman.
4. Tambah validasi dan tes untuk rename flow pada SaveProfile (skenario: write succeeds, remove fails).
5. Verifikasi bahwa logger tidak menulis data sensitif (password/encryption key).

Daftar paket yang digunakan oleh modul ini
---------------------------------------
Catatan: saya hanya mencatat paket yang dipakai oleh file-file dalam `internal/profile` dan `cmd/cmd_profile`.

- internal/profileselect
	- Peran: memilih dan memuat file konfigurasi yang sudah ada; parsing konten terenkripsi.
	- Catatan: dependensi internal, periksa error wrapping dan handling saat kunci enkripsi salah.

- internal/types
	- Peran: definisi struct options dan injeksi dependency (types.Deps).

- pkg/input
	- Peran: wrapper untuk interaksi pengguna (ask string, password, yes/no, multiselect). Menggunakan `survey` di balik layar.
	- Catatan: pastikan wrapper dapat di-mock dan tidak mem-block pada mode non-interactive.

- pkg/ui
	- Peran: mencetak headers, tables, warn/info/error ke terminal.

- pkg/validation
	- Peran: validasi input dan konstanta error (ErrConnectionFailedRetry, ErrUserCancelled).

- pkg/fsops
	- Peran: operasi filesystem (ReadDirFiles, Exists, WriteFile, RemoveFile, CreateDirIfNotExist, CheckDirExists).
	- Catatan: penting untuk memastikan WriteFile mengatur permission aman.

- pkg/consts
	- Peran: konstanta env var names (ENV_PROFILE_ENC_KEY, ENV_TARGET_DB_PASSWORD).

- pkg/helper
	- Peran: helper util seperti ResolveEncryptionKey, ResolveConfigPath, TrimProfileSuffix.

- pkg/database
	- Peran: ConnectionTest dan util koneksi; penting untuk memastikan ConnectionTest cepat dan tidak men-block panjang.

- pkg/encrypt
	- Peran: enkripsi/dekripsi AES. HARUS diperiksa implementasinya (GCM vs CBC/ECB) dan penanganan IV/nonce.

- pkg/parsing
	- Peran: mem-parse flags menjadi option structs.

- pkg/flags
	- Peran: helper untuk men-attach flags ke cobra commands.

- github.com/AlecAivazis/survey/v2
	- Peran: library prompt interaktif. Catatan: survey dapat bermasalah di beberapa terminal/headless — `pkg/input` harus menyediakan fallback.

- github.com/spf13/cobra
	- Peran: CLI command handling.

Catatan akhir
------------
Secara keseluruhan modul `internal/profile` terlihat matang dari sisi pembagian tanggung jawab dan flow alur kerja. Rekomendasi terbesar adalah: tambahkan test, konsisten membungkus error, periksa detail implementasi enkripsi dan pengaturan permission file, serta pastikan logger dan output tidak mengekspose data sensitif.

Jika Anda mau, saya bisa:
- Membuat beberapa unit test skeleton untuk fungsi-fungsi kecil (buildFileName, formatConfigToINI).
- Menambahkan pembungkusan error (mengganti beberapa fmt.Errorf ke %w) dalam patch terpisah.

Selesai.

