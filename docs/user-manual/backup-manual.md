# Panduan Pengguna: Modul Backup Database (`db-backup`)

Modul ini adalah alat utama di **sfDBTools** untuk melakukan pencadangan (backup) database MySQL/MariaDB. Alat ini dirancang untuk keamanan dan efisiensi, dengan fitur enkripsi, kompresi, dan pencatatan metadata yang lengkap untuk memastikan integritas data.

## Perintah Dasar

Semua perintah backup diawali dengan `db-backup`.
```bash
sfdbtools db-backup <mode> [opsi]
```
*Alias yang bisa digunakan: `backup`, `dbbackup`, `dump`.*

---

## 1. Mode-Mode Backup

Pilih mode yang paling sesuai dengan kebutuhan Anda.

### A. `all`
**Backup semua database ke dalam satu file.** Mode ini akan mengambil seluruh database yang ada di server (kecuali database sistem) dan menggabungkannya menjadi satu file dump tunggal.

```bash
# Backup semua database
sfdbtools backup all

# Backup semua, tetapi lewati database 'test' dan 'dev'
sfdbtools backup all --exclude-db test --exclude-db dev
```

### B. `filter`
**Mode interaktif dan paling fleksibel.** Jika dijalankan tanpa flag `--db`, mode ini akan menampilkan menu interaktif untuk memilih database. Anda juga dapat menentukan daftar database secara manual.

Mode `filter` memiliki dua cara penyimpanan:
*   `--mode multi-file` (default): Menyimpan setiap database ke dalam file terpisah.
*   `--mode single-file`: Menggabungkan semua database yang dipilih ke dalam satu file.

```bash
# Buka menu interaktif untuk memilih database
sfdbtools backup filter

# Backup database 'db_a' dan 'db_b' ke file terpisah
sfdbtools backup filter --db db_a --db db_b

# Gabungkan 'db_a' dan 'db_b' ke dalam satu file
sfdbtools backup filter --db db_a --db db_b --mode single-file
```

### C. `single`
**Backup satu database spesifik.** Digunakan ketika Anda hanya ingin mem-backup satu database beserta file-file pendukungnya.

```bash
sfdbtools backup single --db dbsf_nbc_transaksi
```

### D. `primary` & `secondary`
**Mode otomatis untuk lingkungan terstandardisasi.** Perintah ini secara otomatis menemukan dan mem-backup database berdasarkan pola penamaan `dbsf_nbc_` tanpa perlu menentukannya satu per satu.

*   **`primary`**: Mencari dan mem-backup semua database "utama" (diawali `dbsf_nbc_` dan tidak mengandung kata `_secondary`, `_dmart`, dll.).
*   **`secondary`**: Mencari dan mem-backup semua database "sekunder" (yang namanya mengandung `_secondary`).

```bash
# Jalankan backup harian untuk semua database utama
sfdbtools backup primary

# Jalankan backup untuk semua database sekunder
sfdbtools backup secondary
```

---

## 2. Opsi Global (Flags)

Tambahkan opsi berikut pada perintah di atas untuk kustomisasi lebih lanjut.

| Opsi | Deskripsi | Contoh |
|:---|:---|:---|
| `--compress` | Mengaktifkan kompresi (ZSTD/GZIP) untuk memperkecil ukuran file. | `--compress` |
| `--encrypt` | Mengenkripsi file hasil backup menggunakan AES-256. | `--encrypt` |
| `--capture-gtid` | Mengambil informasi GTID (posisi replikasi) saat backup. | `--capture-gtid` |
| `--out <path>` | Menentukan folder tujuan untuk menyimpan hasil backup. | `--out /mnt/backup/daily` |
| `--exclude-db` | Melewati nama database tertentu saat backup. | `--exclude-db sys` |
| `--dry-run` | Melakukan simulasi proses backup tanpa membuat file. | `--dry-run` |

---

## 3. Struktur File Hasil Backup

Setiap proses backup yang sukses akan menghasilkan kumpulan file berikut di folder tujuan:

1.  **File Dump Database** (`.sql.gz.enc` atau format sejenis)
    *   File utama berisi data dan struktur (schema) dari database Anda.
2.  **File Metadata** (`.meta.json`)
    *   Laporan teknis berisi: waktu mulai/selesai, durasi, ukuran file, versi server. Jika flag `--capture-gtid` digunakan, informasi GTID juga akan disertakan di sini.
3.  **File User Grants** (`_users.sql`)
    *   Berisi perintah SQL (`GRANT USAGE...`) untuk memulihkan hak akses (permissions) user yang terkait dengan database.

---

## 4. Fitur Unggulan

*   **Streaming Langsung**: Data diproses secara *on-the-fly* dari database ke file terkompresi/terenkripsi, meminimalkan penggunaan RAM dan storage sementara.
*   **Graceful Shutdown**: Jika proses dibatalkan (`Ctrl+C`), aplikasi akan otomatis membersihkan file yang belum selesai dibuat untuk mencegah file korup dan memenuhi disk.
*   **Validasi Otomatis**: Aplikasi akan memeriksa ketersediaan `mysqldump` dan koneksi database sebelum memulai proses.