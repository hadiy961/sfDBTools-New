# Panduan Review Kode Go Umum

### I. Fungsionalitas & Logika Bisnis (The "What")

Fokus utama adalah memastikan kode tersebut menyelesaikan masalah yang dimaksudkan dengan andal.

1.  **Kebenaran Logika (*Correctness*):**
    * Apakah kode mengimplementasikan persyaratan fungsionalitas yang diminta dengan benar?
    * Apakah ada *bug* atau logika yang salah dalam alur utama (*happy path*)?
2.  **Penanganan *Edge Cases*:**
    * Apakah nilai-nilai masukan khusus ditangani, seperti input kosong (`nil`, *slice* kosong, *string* kosong)?
    * Apakah kasus batas (*boundary conditions*) (misalnya, batas atas/bawah, perhitungan nol) ditangani dengan benar?

---

### II. Idiomatik Go & Gaya (Go Style & Idioms)

Kode Go harus terlihat seperti kode Go. Kepatuhan terhadap gaya komunitas sangat penting untuk keterbacaan.

1.  **Pemformatan & *Linting*:**
    * Apakah kode sudah diotomatisasi dengan `gofmt` dan `goimports`?
    * Apakah *linter* statis (seperti `staticcheck` atau `revive`) dilewati tanpa pelanggaran serius?
2.  **Penamaan (*Naming*):**
    * Apakah nama *package*, fungsi, variabel, dan *struct* jelas, ringkas, dan deskriptif, mengikuti konvensi Go (misalnya, nama pendek di cakupan lokal, nama yang lebih panjang di cakupan paket)?
3.  **Struktur Kode Go:**
    * **Penerima *Method* (*Receiver*):** Apakah sudah memilih antara *value receiver* atau *pointer receiver* dengan benar? (Penting: Gunakan *pointer receiver* jika *method* memutasi *struct* atau *struct* berisi *field* sinkronisasi seperti `sync.Mutex`).
    * **Penggunaan *Interface*:** Apakah mengikuti prinsip "menerima *interface*, mengembalikan *struct*" (*Accept interfaces, return structs*)?
    * **Struktur Paket:** Apakah *package* memiliki tanggung jawab tunggal dan nama *package* sudah menghindari nama generik seperti `util` atau `common`?

---

### III. Penanganan *Error* (Error Handling)

Penanganan *error* adalah pilar dalam Go.

1.  ***Error* Tidak Diabaikan:**
    * Apakah semua nilai `error` yang dikembalikan ditangani secara eksplisit (diperiksa, dicatat, atau dikembalikan)?
2.  **Pembungkusan *Error* (*Error Wrapping*):**
    * Untuk *error* yang dilewatkan ke pemanggil, apakah *error* dibungkus dengan benar menggunakan `%w` di `fmt.Errorf` untuk menjaga konteks dan memungkinkan pemeriksaan dengan `errors.Is` atau `errors.As`?
3.  **Alur Kontrol:**
    * Apakah kode memprioritaskan penanganan *error* terlebih dahulu (*indent error flow*) untuk mengurangi *nesting* dan menjaga alur normal di tingkat indentasi minimal?
    * Apakah *panic* dihindari, kecuali untuk kesalahan program yang tidak dapat dipulihkan?

---

### IV. Konkurensi & *Goroutine*

Untuk kode yang melibatkan konkurensi, keamanan dan efisiensi sangat penting.

1.  ***Race Conditions*:**
    * Apakah ada data bersama (*shared state*) yang diakses dan dimodifikasi oleh beberapa *goroutine* tanpa perlindungan yang tepat (menggunakan `sync.Mutex`, `sync.RWMutex`, atau komunikasi melalui *channel*)?
2.  ***Goroutine Leaks*:**
    * Apakah setiap *goroutine* memiliki rencana yang jelas tentang cara dan kapan ia akan berakhir?
    * Apakah `context` dan `select` digunakan dengan benar untuk membatalkan operasi atau menghentikan *goroutine* yang *blocking*?
3.  **Sinkronisasi:**
    * Apakah `sync.WaitGroup` digunakan dengan benar untuk menunggu sekelompok *goroutine* selesai?
    * Apakah *channel* digunakan dengan benar untuk komunikasi antar *goroutine*?

---

### V. Keamanan & Kinerja (*Security* & *Performance*)

1.  **Keamanan:**
    * **Validasi Input:** Apakah semua input eksternal dari pengguna divalidasi dan disanitasi sebelum digunakan, terutama pada operasi yang sensitif?
    * **Database/I/O:** Apakah menggunakan *parameterized statements* untuk mencegah *SQL Injection*?
    * **Pencatatan Data:** Apakah data sensitif (kredensial, kunci API) tidak dicatat (*logged*) atau terekspos dalam pesan *error*?
2.  **Kinerja:**
    * **Alokasi Berlebihan:** Apakah ada alokasi memori yang tidak perlu di dalam *loop* yang berjalan sering (*hot loops*)?
    * ***Slice* dan *Map*:** Apakah `make` digunakan dengan kapasitas yang sesuai saat membuat *slice* atau *map* yang ukurannya diketahui, untuk menghindari *reallocation* yang tidak perlu?
    * **Penggunaan Kembali *Buffer*:** Untuk operasi I/O intensif, apakah *buffer* digunakan kembali (misalnya, dengan `sync.Pool`) jika memungkinkan?

---

### VI. Keterbacaan & Kemudahan Pemeliharaan (*Maintainability*)

1.  **Komentar dan Dokumentasi:**
    * Apakah semua fungsi/variabel yang **diekspor** memiliki komentar dokumentasi yang baik (`godoc`)?
    * Apakah ada komentar yang menjelaskan **alasan** di balik logika yang kompleks atau non-jelas?
2.  **DRY (*Don't Repeat Yourself*):**
    * Apakah ada duplikasi kode (*code duplication*) yang dapat diekstraksi menjadi fungsi pembantu atau *method* baru?
3.  **Keterbacaan Kode:**
    * Apakah fungsi/metode terlalu panjang dan memiliki lebih dari satu tanggung jawab? Kode yang baik harus mudah dipahami secara sekilas.
    * Apakah *dependencies* (*import*) dikelola dan tidak ada *import* yang tidak terpakai?