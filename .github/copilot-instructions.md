<!-- Last updated: 2026-01-23 -->

# sfdbtools — Copilot coding instructions

sfdbtools adalah CLI Go untuk operasi MySQL/MariaDB (backup/restore/db-scan/cleanup/crypto/profile). Fokus utama: **streaming pipeline** (hemat RAM), **safety**, dan **otomasi**.

## Big picture (mulai baca dari sini)
- Entrypoint: [main.go](../main.go) → bootstrap runtime flags → (opsional) auto-update → load config → `cmd.Execute(deps)`.
- Root CLI (Cobra): [cmd/root.go](../cmd/root.go) menjalankan `PersistentPreRunE`, set runtime mode (`--quiet/-q`), lalu log `argv` yang sudah dimasking (lihat [cmd/args_sanitize.go](../cmd/args_sanitize.go)).
- Dependency injection: `*internal/cli/deps.Dependencies` dibuat di `main.go` lalu disimpan global via `cmd.Execute()`; command/service membaca lewat `internal/cli/deps`.

## Struktur folder & boundary
- `cmd/`: definisi command + parsing flags (tipis). Contoh: [cmd/backup/main.go](../cmd/backup/main.go), [cmd/restore/main.go](../cmd/restore/main.go).
- `internal/app/`: orkestrasi workflow per fitur (backup/restore/profile/dbscan/cleanup/script).
- `internal/services/`: implementasi service (config/logger/crypto/dll).
- `pkg/`: library reusable (compress/encrypt/consts/helper/runtimecfg/validation/dll).

## Konfigurasi (zero-config first run)
- Loader: [internal/services/config/appconfig_loaders.go](../internal/services/config/appconfig_loaders.go) membaca `SFDB_APPS_CONFIG`; jika kosong pakai default `/etc/sfDBTools/config.yaml` (lihat [pkg/consts/consts_paths.go](../pkg/consts/consts_paths.go)).
- Jika config belum ada, tool akan auto-generate default; kalau tidak bisa menulis ke `/etc/...` (non-root), fallback ke `XDG_CONFIG_HOME/sfdbtools/config.yaml` atau `~/.config/sfdbtools/config.yaml` (lihat [internal/services/config/appconfig_defaults.go](../internal/services/config/appconfig_defaults.go)).

## Pola penting (jangan ubah arah)
- Streaming backup: `mariadb-dump` (fallback `mysqldump`) → (opsional) `compress.Writer` → (opsional) `encrypt.Writer` → file. Referensi utama: [internal/app/backup/writer/engine.go](../internal/app/backup/writer/engine.go). Jangan buffer seluruh dump ke memori.
- Backup modes via factory + interface kecil: [internal/app/backup/modes/interface.go](../internal/app/backup/modes/interface.go), [internal/app/backup/modes/factory.go](../internal/app/backup/modes/factory.go).
- Restore companion `_dmart`: auto-detect / pilih file + aturan non-interaktif `--force` dan `--continue-on-error`: [internal/app/restore/companion_helpers.go](../internal/app/restore/companion_helpers.go).

## Output bersih & logging aman
- `completion`, `version`, `update` harus bisa jalan tanpa config dan tanpa noise (lihat [main.go](../main.go) dan [cmd/root.go](../cmd/root.go)).
- Jika menambah flag/arg sensitif baru, pastikan term-nya ikut ter-mask di [cmd/args_sanitize.go](../cmd/args_sanitize.go) (pattern `password|token|secret|key` dll).

## Workflow developer (yang dipakai repo ini)
- Unit test: `go test ./...`
- Build+run ke `/usr/bin/sfdbtools` (butuh root): `sudo bash bash/build_run.sh -- --help` (script: [bash/build_run.sh](../bash/build_run.sh)).
- Installer/uninstaller ada di folder `scripts/` (lihat [README.md](../README.md), [scripts/install.sh](../scripts/install.sh), [scripts/uninstall.sh](../scripts/uninstall.sh)).

## Env var penting
- Source of truth: [pkg/consts/consts_env.go](../pkg/consts/consts_env.go)
- Umum: `SFDB_APPS_CONFIG`, `SFDB_QUIET`, `SFDB_NO_AUTO_UPDATE`, `SFDB_BACKUP_ENCRYPTION_KEY`.

## Konvensi repo
- `--quite` adalah alias deprecated untuk `--quiet` (lihat [cmd/root.go](../cmd/root.go)).
- Gunakan bahasa Indonesia untuk komentar/dokumentasi, dan update header `Last Modified` jika file Go memilikinya.

---

## Filosofi Desain Go (KRITIKAL)

Ikuti prinsip-prinsip spesifik ini saat menulis atau melakukan refaktorisasi kode untuk proyek ini:

### **DRY vs. Dependensi ("The Go Way")**

* **Prinsip**: "Sedikit penyalinan lebih baik daripada sedikit dependensi."
* **Panduan**: Jangan membuat *library* bersama yang raksasa hanya untuk memenuhi prinsip DRY. Lebih baik menduplikasi beberapa baris logika sederhana di dua tempat daripada menghubungkan keduanya ke fungsi bersama yang menciptakan ketergantungan kompleks.
* **Tujuan**: Kemandirian kode diprioritaskan di atas deduplikasi yang ketat.

### **KISS (Keep It Simple, Stupid)**

* **Prinsip**: Kode harus "membosankan" dan eksplisit.
* **Panduan**: Hindari *Generics* yang kompleks (kecuali benar-benar diperlukan), *Reflection*, atau kode satu baris yang "cerdas". Jika Anda perlu membuka 5 file untuk memahami satu fungsi, berarti kode tersebut terlalu kompleks.
* **Batasan**: Go tidak memiliki operator ternary; jangan mencoba menirunya dengan logika yang rumit.

### **YAGNI (You Ain't Gonna Need It)**

* **Prinsip**: Jangan mendesain untuk masa depan yang hipotetis.
* **Panduan**:
* **JANGAN** buat *Interface* jika saat ini hanya ada satu implementasi.
* **JANGAN** buat struktur folder yang dalam untuk "ekspansi masa depan."
* Refaktorisasi di Go itu mudah; bangunlah untuk kebutuhan *saat ini*.

### **Adaptasi SOLID**

* **SRP (Single Responsibility)**: Sebuah paket harus memiliki satu tujuan yang jelas (contoh: `net/http`).
* **ISP (Interface Segregation)**: **Sangat Penting**. Buat *interface* sekecil mungkin. *Interface* dengan 1 metode (seperti `io.Reader`) jauh lebih baik daripada satu *interface* dengan 10 metode.
* **DIP (Dependency Inversion)**: Fungsi harus menerima *interface* tetapi (umumnya) mengembalikan *concrete struct*.

---

## Konsistensi & Pemeliharaan
Ikuti panduan berikut untuk menjaga kualitas dan keberlanjutan proyek:

* **Audit Prinsip Desain**: Lakukan peninjauan kode secara berkala untuk memastikan kepatuhan terhadap prinsip desain Go yang telah ditetapkan.
* **Refaktorisasi Berkala**: Lakukan refaktorisasi pada *service* dan *mode* secara rutin guna menjaga kejelasan serta kesederhanaan logika.
* **Dokumentasi Arsitektur**: Perbarui dokumentasi dan komentar kode untuk mencerminkan keputusan arsitektur serta pola (*pattern*) yang digunakan.
* **Standardisasi Fitur Baru**: Pastikan setiap fitur baru mengikuti pola yang sudah ada demi kemudahan pemeliharaan jangka panjang.
* **Minimalisir Dependensi**: Jaga agar dependensi tetap minimal dan hanya yang relevan untuk menghindari *bloat* (pembengkakan) pada sistem.
* **Review Filosofis**: Dorong proses *code review* yang berfokus pada kepatuhan terhadap filosofi desain dan kualitas kode.
* **Bahasa Dokumentasi**: Gunakan **Bahasa Indonesia** untuk penulisan komentar di dalam kode dan dokumentasi teknis lainnya.
* **Metadata File**: Perbarui tanggal modifikasi terakhir (*last modified date*) pada bagian komentar *header* di setiap file setiap kali melakukan perubahan.
* **No unit test**: Proyek ini tidak menggunakan unit test; fokus pada integrasi dan pengujian fungsional.