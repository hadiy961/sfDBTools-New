## Ringkasan Pola yang Perlu Dipusatkan

- **Klasifikasi nama database (primary/secondary/dmart/temp/archive)**  
  - **Lokasi**: `internal/backup/selection/filtering_logic.go` dan `internal/backup/selection/selector.go` memfilter berdasarkan suffix/prefix; `internal/restore/helpers/validation.go` serta berbagai executor di `internal/restore/modes` melakukan pengecekan serupa secara manual.  
  - **Alasan**: Aturan penamaan diulang dengan kombinasi `SuffixDmart`, `SuffixTemp`, `SuffixArchive`, dan `SecondarySuffix`. Satu helper deterministik akan menyamakan kriteria primary/secondary/dmart di seluruh backup & restore sehingga tidak ada perbedaan perilaku ketika aturan berubah.

- **Penyaringan database sistem**  
  - **Lokasi**: Banyak blok langsung membaca `types.SystemDatabases` (mis. `internal/backup/selection/selector.go`, `internal/backup/selection/filtering_logic.go`, `internal/restore/validation_helpers.go`, `internal/restore/validation_helpers.go` DropAll, setup restore) padahal `pkg/database` sudah memiliki `IsSystemDatabase/GetNonSystemDatabases`.  
  - **Alasan**: Duplikasi list dan pengecekan membuka peluang inkonsistensi bila daftar system DB berubah. Satu gateway filter (mis. di `pkg/database`) cukup untuk semua alur.

- **Factory executor mode backup vs restore**  
  - **Lokasi**: `internal/backup/modes/factory.go` dan `internal/restore/modes/factory.go` sama-sama switch mode → `NewXExecutor`.  
  - **Alasan**: Pola identik; registrasi map/tabel mode→constructor bisa dipakai ulang sehingga penambahan mode baru tidak perlu menyentuh dua switch terpisah.

- **Kerangka eksekusi restore (single/primary/secondary)**  
  - **Lokasi**: `internal/restore/modes/single.go`, `primary.go`, `secondary.go` mengulangi alur yang sama: start timer, set in-progress, handle dry-run, jalankan `commonRestoreFlow`, restore grants, post-restore, finalize result.  
  - **Alasan**: Blok pembuka/penutup identik mempersulit perubahan (mis. menambah logging/metrics) karena harus diubah di tiga file. Satu runner/templating flow dapat dipakai dengan hook per-mode.

- **Penanganan companion (_dmart) yang tersebar**  
  - **Lokasi**: Pembuatan nama companion dan penetapan `CompanionDB/CompanionFile` muncul di executor `primary.go`, `secondary.go`, helper `companion_detect.go` (`buildCompanionDBName`), serta flow `companionRestoreFlow`.  
  - **Alasan**: Logika penentuan companion DB/file (tambahkan suffix, skip ketika tidak ada) diulang; satu utilitas untuk “resolve companion (DB+file) dari target+flag+metadata” akan menyatukan keputusan skip/stop dan mencegah perbedaan hasil.

- **Validasi ekstensi file backup**  
  - **Lokasi**: `internal/restore/companion_helpers.go` memiliki `isValidBackupFileExtension` lokal, sementara `helper.ValidBackupFileExtensionsForSelection` sudah tersedia; `pkg/helper/services.go` juga membuat wrapper identik ke `pkg/helper/file`.  
  - **Alasan**: Dua sumber truth untuk ekstensi valid dan adanya wrapper tipis membuat risiko divergen saat format baru ditambah. Satu validator terpusat yang dipakai di semua path (termasuk companion/non-companion) menghilangkan duplikasi dan wrapper.

- **Definisi Cobra command backup per-mode**  
  - **Lokasi**: `cmd/backup/single.go`, `primary.go`, `secondary.go`, `all.go` memiliki pola sama: `DefaultBackupOptions`, `flags.AddBackupFlgs`, lalu `runBackupCommand` yang hanya mengembalikan mode.  
  - **Alasan**: Empat file berisi pola identik berbeda di string mode saja; builder/helper pendaftar command bisa mereduksi duplikasi tanpa menambah wrapper kosong, sehingga penambahan mode baru cukup menambah entri konfigurasi.
