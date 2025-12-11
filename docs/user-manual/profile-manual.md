# Panduan Pengguna: Manajemen Profil (`profile`)

Fitur `profile` memungkinkan Anda untuk menyimpan, mengelola, dan menggunakan kembali kredensial koneksi database secara aman. Sebuah "profil" adalah file konfigurasi terenkripsi yang menyimpan informasi seperti host, port, user, dan password.

Dengan menggunakan profil, Anda tidak perlu mengetik ulang detail koneksi setiap kali menjalankan perintah seperti `backup` atau `restore`.

## 1. Konsep Dasar

- **File Profil**: Setiap profil disimpan sebagai file `.cnf.enc` di dalam direktori konfigurasi aplikasi.
- **Enkripsi**: Isi file profil dienkripsi menggunakan AES-256. Anda akan memerlukan "kunci enkripsi" (sebuah password) untuk membuat atau menggunakan profil. Kunci ini bisa diatur melalui _environment variable_ `SFDB_SOURCE_PROFILE_KEY` atau dimasukkan saat diminta.
- **Nama Profil**: Anda merujuk ke sebuah profil menggunakan namanya (misalnya, `produksidb`, `staging-server`).

---

## 2. Perintah yang Tersedia

### A. `profile create`

Gunakan perintah ini untuk membuat profil koneksi baru.

**Mode Interaktif (Disarankan):**
Jalankan perintah dengan flag `--interactive` atau `-i`. Aplikasi akan memandu Anda langkah demi langkah untuk mengisi setiap detail.

```bash
sfdbtools profile create --interactive
```

**Mode Non-Interaktif:**
Anda dapat menyediakan semua detail koneksi melalui _flag_. Ini berguna untuk skrip otomatis.

```bash
# Contoh membuat profil bernama 'dev-db'
sfdbtools profile create --profil dev-db --host 127.0.0.1 --port 3307 --user developer
```
- `--profil <nama>`: Nama untuk profil baru Anda.
- `--host <alamat>`, `--port <nomor>`, `--user <nama>`: Detail koneksi database.
- `--key <kunci>`: (Opsional) Kunci enkripsi. Jika tidak diberikan, akan diminta atau dibaca dari _environment variable_.

---

### B. `profile show`

Gunakan perintah ini untuk melihat detail dari profil yang sudah ada.

```bash
sfdbtools profile show --file dev-db
```
- `--file <nama>`: Nama profil yang ingin Anda lihat.

Secara _default_, _password_ tidak akan ditampilkan. Untuk melihat _password_ (akan memerlukan input kunci enkripsi lagi untuk verifikasi), gunakan _flag_ `--reveal-password`.

```bash
sfdbtools profile show --file dev-db --reveal-password
```

---

### C. `profile edit`

Gunakan perintah ini untuk mengubah profil yang sudah ada. Anda harus selalu menentukan profil mana yang akan diedit menggunakan _flag_ `--file`.

**Mode Interaktif (Disarankan):**
Mode interaktif akan menampilkan menu untuk memilih profil, lalu memandu Anda untuk mengubah setiap _field_.

```bash
sfdbtools profile edit --interactive
```

**Mode Non-Interaktif:**
Ubah _field_ tertentu dengan menyediakannya sebagai _flag_.

```bash
# Mengubah host pada profil 'dev-db'
sfdbtools profile edit --file dev-db --host 192.168.1.100

# Mengganti nama profil dari 'dev-db' menjadi 'development'
sfdbtools profile edit --file dev-db --new-name development
```

---

### D. `profile delete`

Gunakan perintah ini untuk menghapus satu atau beberapa profil.

**Mode Interaktif:**
Jalankan perintah tanpa _flag_ apa pun. Aplikasi akan menampilkan daftar profil yang ada dan memungkinkan Anda memilih mana yang akan dihapus (bisa lebih dari satu).

```bash
sfdbtools profile delete
```
Anda akan diminta konfirmasi sebelum file benar-benar dihapus.

**Mode Non-Interaktif:**
Hapus profil tertentu secara langsung. Gunakan `--force` atau `-F` untuk melewati konfirmasi.

```bash
# Menghapus 'old-profile' tanpa konfirmasi
sfdbtools profile delete --file old-profile --force
```
