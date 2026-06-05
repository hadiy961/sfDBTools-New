package settings

import (
	"fmt"
	"sfdbtools/internal/crypto"
	"sfdbtools/internal/shared/database"

	"github.com/charmbracelet/huh"
	"github.com/fatih/color"
)

func (s *Service) SetupBackupWizard() {
	db, err := database.GetSQLite()
	if err != nil {
		fmt.Println(color.RedString("Error: %v", err))
		return
	}

	// --- Ambil semua keys backup dari SQLite ---
	allKeys := []string{
		// Compression
		"backup_compression_enabled", "backup_compression_type", "backup_compression_level",
		// Filters
		"backup_exclude_system_databases", "backup_include_dmart",
		"backup_exclude_user", "backup_exclude_grant",
		"backup_exclude_data", "backup_exclude_empty",
		"backup_exclude_file", "backup_include_file",
		// Cleanup
		"backup_cleanup_enabled", "backup_cleanup_days", "backup_cleanup_schedule",
		// Encryption
		"backup_encryption_enabled", "backup_encryption_key",
		// Output
		"backup_output_base_directory", "backup_output_structure_pattern",
		"backup_output_cleanup_temp", "backup_output_create_subdirs",
		"backup_output_file_permissions", "backup_output_metadata_permissions",
		"backup_output_save_backup_info",
		// Catalog
		"backup_catalog_enabled", "backup_catalog_file_path",
		// Verification
		"backup_verification_post_backup_check", "backup_verification_header_footer_check",
		"backup_verification_checksum_algorithm", "backup_verification_disk_space_check",
		"backup_verification_min_file_size",
		// mysqldump
		"backup_mysqldump_args",
		// Replication
		"backup_replication_capture_gtid", "backup_replication_user", "backup_replication_password",
	}

	vals := make(map[string]string)
	for _, k := range allKeys {
		var v string
		db.QueryRow("SELECT value FROM app_settings WHERE key = ?", k).Scan(&v)
		vals[k] = v
	}

	// Decrypt sensitive fields
	decEncKey, _, _ := crypto.DecodeEnvSecret(vals["backup_encryption_key"])
	decReplPass, _, _ := crypto.DecodeEnvSecret(vals["backup_replication_password"])

	// --- Bind variables ---
	var (
		// Compression
		compEn   = vals["backup_compression_enabled"] == "true"
		compType = vals["backup_compression_type"]
		compLev  = vals["backup_compression_level"]

		// Filters
		filterSys    = vals["backup_exclude_system_databases"] == "true"
		filterDmart  = vals["backup_include_dmart"] == "true"
		filterUser   = vals["backup_exclude_user"] == "true"
		filterGrant  = vals["backup_exclude_grant"] == "true"
		filterData   = vals["backup_exclude_data"] == "true"
		filterEmpty  = vals["backup_exclude_empty"] == "true"
		excFile      = vals["backup_exclude_file"]
		incFile      = vals["backup_include_file"]

		// Cleanup
		clEn       = vals["backup_cleanup_enabled"] == "true"
		clDays     = vals["backup_cleanup_days"]
		clSchedule = vals["backup_cleanup_schedule"]

		// Encryption
		encEn  = vals["backup_encryption_enabled"] == "true"
		encKey = decEncKey

		// Output
		dir         = vals["backup_output_base_directory"]
		pattern     = vals["backup_output_structure_pattern"]
		cleanTemp   = vals["backup_output_cleanup_temp"] == "true"
		createSubs  = vals["backup_output_create_subdirs"] == "true"
		filePerm    = vals["backup_output_file_permissions"]
		metaPerm    = vals["backup_output_metadata_permissions"]
		saveInfo    = vals["backup_output_save_backup_info"] == "true"

		// Catalog
		catEn       = vals["backup_catalog_enabled"] == "true"
		catFilePath = vals["backup_catalog_file_path"]

		// Verification
		verifPost     = vals["backup_verification_post_backup_check"] == "true"
		verifHeader   = vals["backup_verification_header_footer_check"] == "true"
		verifChecksum = vals["backup_verification_checksum_algorithm"]
		verifDisk     = vals["backup_verification_disk_space_check"] == "true"
		verifMinSize  = vals["backup_verification_min_file_size"]

		// mysqldump
		dumpArgs = vals["backup_mysqldump_args"]

		// Replication
		replGTID = vals["backup_replication_capture_gtid"] == "true"
		replUser = vals["backup_replication_user"]
		replPass = decReplPass
	)

	fmt.Println(color.CyanString("\n⚙️  Backup Settings Wizard  (Tab/Arrow = Navigate, Ctrl+C = Cancel)\n"))

	form := huh.NewForm(
		// === GROUP 1: COMPRESSION ===
		huh.NewGroup(
			huh.NewConfirm().
				Title("Aktifkan Kompresi?").
				Description("Kompres output backup untuk menghemat disk space.").
				Value(&compEn),
			huh.NewSelect[string]().
				Title("Tipe Kompresi:").
				Options(
					huh.NewOption("gzip (default)", "gzip"),
					huh.NewOption("zstd (lebih cepat)", "zstd"),
					huh.NewOption("none (tanpa kompresi)", "none"),
				).
				Value(&compType),
			huh.NewInput().
				Title("Compression Level (1-9):").
				Description("1 = cepat, 9 = kompresi maksimal.").
				Value(&compLev),
		).Title("💾 Compression"),

		// === GROUP 2: FILTER DATABASE ===
		huh.NewGroup(
			huh.NewConfirm().Title("Exclude System Databases?").
				Description("Lewati DB sistem: information_schema, performance_schema, mysql, sys.").
				Value(&filterSys),
			huh.NewConfirm().Title("Exclude User & Grants?").
				Description("Jangan backup user grants MySQL.").
				Value(&filterUser),
			huh.NewConfirm().Title("Exclude Grant Statements?").
				Description("Skip GRANT statements di dalam SQL dump.").
				Value(&filterGrant),
			huh.NewConfirm().Title("Exclude Data (Schema Only)?").
				Description("Hanya backup struktur tabel, tanpa data.").
				Value(&filterData),
			huh.NewConfirm().Title("Exclude Empty Databases?").
				Description("Skip database yang tidak punya tabel.").
				Value(&filterEmpty),
			huh.NewConfirm().Title("Include _dmart Databases?").
				Description("Sertakan database dengan prefix _dmart.").
				Value(&filterDmart),
			huh.NewInput().Title("Exclude Databases (comma-separated):").
				Description("Contoh: db_test,db_dev").
				Value(&excFile),
			huh.NewInput().Title("Include Databases (comma-separated):").
				Description("Kosongkan untuk backup semua.").
				Value(&incFile),
		).Title("🗂️  Database Filters"),

		// === GROUP 3: MYSQLDUMP ===
		huh.NewGroup(
			huh.NewInput().
				Title("mysqldump Extra Args:").
				Description("Flag tambahan untuk mysqldump. Biarkan jika tidak yakin.").
				Value(&dumpArgs),
		).Title("🔧 mysqldump Arguments"),

		// === GROUP 4: CLEANUP ===
		huh.NewGroup(
			huh.NewConfirm().Title("Aktifkan Auto Cleanup?").
				Description("Hapus otomatis backup lama berdasarkan retention.").
				Value(&clEn),
			huh.NewInput().Title("Retention (hari):").
				Description("Backup lebih tua dari N hari akan dihapus.").
				Value(&clDays),
			huh.NewInput().Title("Cleanup Schedule (cron):").
				Description("Contoh: 0 2 * * * (setiap jam 2 pagi). Kosongkan jika tidak pakai cron.").
				Value(&clSchedule),
			huh.NewConfirm().Title("Aktifkan Enkripsi Backup?").
				Description("Enkripsi file backup setelah dibuat.").
				Value(&encEn),
			huh.NewInput().Title("Encryption Key:").
				EchoMode(huh.EchoModePassword).
				Description("Key akan dienkripsi sebelum disimpan ke DB.").
				Value(&encKey),
		).Title("🧹 Cleanup & Security"),

		// === GROUP 5: OUTPUT ===
		huh.NewGroup(
			huh.NewInput().Title("Base Directory:").
				Description("Direktori utama penyimpanan backup.").
				Value(&dir),
			huh.NewInput().Title("Folder Structure Pattern:").
				Description("Contoh: {year}{month}{day}/ — variabel: {year},{month},{day},{host}.").
				Value(&pattern),
			huh.NewConfirm().Title("Buat Subdirectory Otomatis?").
				Description("Buat subfolder jika belum ada.").
				Value(&createSubs),
			huh.NewConfirm().Title("Cleanup Temp Files setelah Backup?").
				Description("Hapus file .tmp setelah proses selesai.").
				Value(&cleanTemp),
			huh.NewConfirm().Title("Simpan Backup Info File?").
				Description("Simpan metadata backup (ukuran, waktu, checksum) ke file .info.").
				Value(&saveInfo),
			huh.NewInput().Title("File Permissions (octal):").
				Description("Contoh: 0600").
				Value(&filePerm),
			huh.NewInput().Title("Metadata Permissions (octal):").
				Description("Contoh: 0600").
				Value(&metaPerm),
		).Title("📂 Output & Filesystem"),

		// === GROUP 6: CATALOG & VERIFICATION ===
		huh.NewGroup(
			huh.NewConfirm().Title("Aktifkan Backup Catalog?").
				Description("Catat setiap backup ke file catalog JSON.").
				Value(&catEn),
			huh.NewInput().Title("Path Catalog File:").
				Description("Contoh: /etc/sfDBTools/catalog.json").
				Value(&catFilePath),
			huh.NewConfirm().Title("Post-Backup Verification?").
				Description("Verifikasi backup segera setelah dibuat.").
				Value(&verifPost),
			huh.NewConfirm().Title("Verify Header & Footer?").
				Description("Cek apakah file dump diawali/diakhiri dengan benar.").
				Value(&verifHeader),
			huh.NewSelect[string]().
				Title("Checksum Algorithm:").
				Options(
					huh.NewOption("sha256 (rekomendasi)", "sha256"),
					huh.NewOption("md5 (lebih cepat)", "md5"),
					huh.NewOption("none", "none"),
				).
				Value(&verifChecksum),
			huh.NewConfirm().Title("Disk Space Check?").
				Description("Cek ketersediaan disk sebelum mulai backup.").
				Value(&verifDisk),
			huh.NewInput().Title("Min File Size (bytes):").
				Description("Backup di bawah ukuran ini dianggap invalid.").
				Value(&verifMinSize),
		).Title("📋 Catalog & Verification"),

		// === GROUP 7: REPLICATION ===
		huh.NewGroup(
			huh.NewConfirm().Title("Capture GTID?").
				Description("Simpan informasi GTID replication saat backup.").
				Value(&replGTID),
			huh.NewInput().Title("Replication User:").
				Value(&replUser),
			huh.NewInput().Title("Replication Password:").
				EchoMode(huh.EchoModePassword).
				Value(&replPass),
		).Title("🔄 Replication"),
	)

	err = form.Run()
	if err != nil {
		fmt.Println(color.YellowString("\n[INTERRUPT] Operasi dibatalkan. Tidak ada yang disimpan."))
		return
	}

	// === SIMPAN KE SQLITE ===
	// Compression
	s.saveSetting(db, "backup_compression_enabled", fmt.Sprintf("%t", compEn), "backup")
	s.saveSetting(db, "backup_compression_type", compType, "backup")
	s.saveSetting(db, "backup_compression_level", compLev, "backup")

	// Filters (note: we bind excFile/incFile to the DB/Filter fields, not file fields)
	s.saveSetting(db, "backup_exclude_system_databases", fmt.Sprintf("%t", filterSys), "backup")
	s.saveSetting(db, "backup_include_dmart", fmt.Sprintf("%t", filterDmart), "backup")
	s.saveSetting(db, "backup_exclude_user", fmt.Sprintf("%t", filterUser), "backup")
	s.saveSetting(db, "backup_exclude_grant", fmt.Sprintf("%t", filterGrant), "backup")
	s.saveSetting(db, "backup_exclude_data", fmt.Sprintf("%t", filterData), "backup")
	s.saveSetting(db, "backup_exclude_empty", fmt.Sprintf("%t", filterEmpty), "backup")
	s.saveSetting(db, "backup_exclude_file", excFile, "backup")
	s.saveSetting(db, "backup_include_file", incFile, "backup")

	// mysqldump
	s.saveSetting(db, "backup_mysqldump_args", dumpArgs, "backup")

	// Cleanup
	s.saveSetting(db, "backup_cleanup_enabled", fmt.Sprintf("%t", clEn), "backup")
	s.saveSetting(db, "backup_cleanup_days", clDays, "backup")
	s.saveSetting(db, "backup_cleanup_schedule", clSchedule, "backup")

	// Encryption
	s.saveSetting(db, "backup_encryption_enabled", fmt.Sprintf("%t", encEn), "backup")
	if encKey != "" {
		encVal, encErr := crypto.EncodeEnvSecret(encKey)
		if encErr == nil {
			s.saveSetting(db, "backup_encryption_key", encVal, "backup")
		} else {
			fmt.Println(color.RedString("Gagal mengenkripsi kunci: %v", encErr))
		}
	} else {
		s.saveSetting(db, "backup_encryption_key", "", "backup")
	}

	// Output
	s.saveSetting(db, "backup_output_base_directory", dir, "backup")
	s.saveSetting(db, "backup_output_structure_pattern", pattern, "backup")
	s.saveSetting(db, "backup_output_cleanup_temp", fmt.Sprintf("%t", cleanTemp), "backup")
	s.saveSetting(db, "backup_output_create_subdirs", fmt.Sprintf("%t", createSubs), "backup")
	s.saveSetting(db, "backup_output_file_permissions", filePerm, "backup")
	s.saveSetting(db, "backup_output_metadata_permissions", metaPerm, "backup")
	s.saveSetting(db, "backup_output_save_backup_info", fmt.Sprintf("%t", saveInfo), "backup")

	// Catalog
	s.saveSetting(db, "backup_catalog_enabled", fmt.Sprintf("%t", catEn), "backup")
	s.saveSetting(db, "backup_catalog_file_path", catFilePath, "backup")

	// Verification
	s.saveSetting(db, "backup_verification_post_backup_check", fmt.Sprintf("%t", verifPost), "backup")
	s.saveSetting(db, "backup_verification_header_footer_check", fmt.Sprintf("%t", verifHeader), "backup")
	s.saveSetting(db, "backup_verification_checksum_algorithm", verifChecksum, "backup")
	s.saveSetting(db, "backup_verification_disk_space_check", fmt.Sprintf("%t", verifDisk), "backup")
	s.saveSetting(db, "backup_verification_min_file_size", verifMinSize, "backup")

	// Replication
	s.saveSetting(db, "backup_replication_capture_gtid", fmt.Sprintf("%t", replGTID), "backup")
	s.saveSetting(db, "backup_replication_user", replUser, "backup")
	if replPass != "" {
		encReplPass, encErr := crypto.EncodeEnvSecret(replPass)
		if encErr == nil {
			s.saveSetting(db, "backup_replication_password", encReplPass, "backup")
		}
	} else {
		s.saveSetting(db, "backup_replication_password", "", "backup")
	}

	fmt.Println(color.GreenString("\n ✅  Semua konfigurasi backup berhasil disimpan!"))
}
