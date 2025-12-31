## Pola Duplikasi yang Perlu Dipusatkan (Paket Backup)

- **Alias mode separated/separate tersebar**  
  - **Lokasi**: Kondisi `ModeSeparated || ModeSeparate` di `internal/backup/mode_config.go`, `display/options_helpers.go`, `execution/engine.go`, `path_helpers.go`, `modes/factory.go`, `metadata/generator.go`, `setup/session.go`.  
  - **Alasan**: Alias ganda memaksa OR check di banyak tempat; normalisasi mode sekali (helper/enum) menghindari ketidakkonsistenan dan mempercepat penambahan mode baru.

- **Deteksi single/primary/secondary bercabang manual**  
  - **Lokasi**: `setup/session.go` menentukan perlakuan interaktif (ticket, edit loop) lewat OR per mode, sementara helper `modes.IsSingleModeVariant` hanya dipakai sebagian (`path_helpers.go`, `modes/iterative.go`).  
  - **Alasan**: Kebutuhan “single-variant behavior” terduplikasi; satu helper yang dipakai konsisten akan menyatukan validasi dan flow kontrol.

- **Filter suffix (_dmart/_temp/_archive/_secondary) terduplikasi**  
  - **Lokasi**: `selection/filtering_logic.go` dan `selection/selector.go` masing-masing melakukan hardcode pengecualian suffix pada beberapa cabang (mode primary/secondary/single, client/instance filter, companion expansion) dengan pola `if strings.HasSuffix(..., SuffixTemp/Archive/Dmart) { continue }`.  
  - **Alasan**: Aturan nama companion/temp/archive/secondary tersebar; satu utilitas klasifikasi nama DB (primary/secondary/companion/terlarang) bisa dipakai ulang agar perubahan suffix hanya di satu titik.

- **Skip system DB ganda**  
  - **Lokasi**: `selection/selector.go` dan `selection/filtering_logic.go` sama-sama membaca `types.SystemDatabases` langsung, terpisah dari filter sistem di `pkg/database`.  
  - **Alasan**: Dua sumber kebenaran untuk system DB di dalam paket backup; pusatkan ke satu helper filter agar penambahan/ubah system DB tidak terlewat.

- **Prompt ticket per-mode terduplikasi**  
  - **Lokasi**: `setup/session.go` memiliki blok serupa untuk ALL, SINGLE, PRIMARY, SECONDARY, dan combined/separated.  
  - **Alasan**: Perbedaan hanya pada mode target; ekstrak ke fungsi `ensureTicket(mode string, interactive bool)` agar validasi & default ticket konsisten dan mengurangi copy-paste.

- **Fallback hostname/host diduplikasi**  
  - **Lokasi**: `setup/session.go` menetapkan `HostName` dari server atau fallback `Host`, sementara `backup/path_helpers.go` mengulang fallback ketika membentuk path/filename.  
  - **Alasan**: Dua titik normalisasi hostname berpotensi divergen; lakukan satu kali di setup dan gunakan nilai final untuk seluruh path generation.

- **Kondisi ekspor user grants per mode berulang**  
  - **Lokasi**: `execution/engine.go` memeriksa `ModeSeparated/ModeSeparate/ModeSingle` sebelum `ExportUserGrantsIfNeeded`; logika pemilihan file utama per mode di `modes/iterative.go` menggunakan kondisi serupa.  
  - **Alasan**: Mode yang membutuhkan grants sebaiknya ditentukan dari satu tabel/flag sehingga ekspor/grants update tidak perlu diubah di banyak file.
