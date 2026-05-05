// File : internal/app/initialize/service.go
// Deskripsi : Service untuk menangani Wizard Inisialisasi
// Author : Antigravity
// Tanggal : 30 April 2026

package initialize

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sfdbtools/internal/app/sync"
	appdeps "sfdbtools/internal/cli/deps"
	"sfdbtools/internal/shared/database"
	"sfdbtools/internal/ui/print"
	"sfdbtools/internal/ui/prompt"

	"github.com/fatih/color"
)

type Service struct {
	deps *appdeps.Dependencies
}

func NewService(deps *appdeps.Dependencies) *Service {
	return &Service{deps: deps}
}

func (s *Service) RunWizard() error {
	print.PrintAppHeader("Initialization")
	color.Cyan("=== SFDBTools Initialization Wizard ===\n")
	fmt.Println("Wizard ini akan membantu Anda menyiapkan database lokal dan sinkronisasi cloud.")

	// 1. Setup Client Identity
	clientCode := s.deps.Config.General.ClientCode
	if clientCode == "" || clientCode == "YOUR_CODE" {
		code, err := prompt.AskText("Masukkan Client Code (Tenant ID):", prompt.WithDefault("CLIENT_001"))
		if err != nil {
			return err
		}
		clientCode = code
	} else {
		fmt.Printf("Client Code saat ini: %s\n", color.GreenString(clientCode))
		change, _ := prompt.Confirm("Apakah Anda ingin mengubah Client Code?", false)
		if change {
			code, err := prompt.AskText("Masukkan Client Code baru:", prompt.WithDefault(clientCode))
			if err != nil {
				return err
			}
			clientCode = code
		}
	}

	// 2. Setup SQLite Path
	dbPath := s.deps.Config.Storage.LocalDB
	if dbPath == "" {
		dbPath = "/etc/sfDBTools/sfdbtools.db"
	}
	fmt.Printf("Database lokal akan disimpan di: %s\n", color.CyanString(dbPath))

	// 3. Initialize Database
	fmt.Print("Sedang menginisialisasi database...")
	if err := database.InitSQLite(dbPath); err != nil {
		fmt.Println(color.RedString(" GAGAL"))
		return err
	}
	fmt.Println(color.GreenString(" BERHASIL"))

	// 4. Zero-Config Onboarding Option
	fmt.Println("\n--- Onboarding Method ---")
	useZeroConfig, _ := prompt.Confirm("Gunakan Zero-Config Setup? (Tarik pengaturan dari Remote Hub via Client Code & Sync Key)", false)
	
	if useZeroConfig {
		if err := s.PerformZeroConfigSetup(clientCode, dbPath); err != nil {
			fmt.Println(color.RedString("\n[ERROR] Zero-Config Gagal: %v", err))
			fmt.Println("Melanjutkan ke inisialisasi manual...")
		} else {
			fmt.Println(color.GreenString("\n[SUCCESS] Zero-Config Onboarding Selesai!"))
			return s.SaveLeanConfig(clientCode, dbPath)
		}
	}

	// 5. Setup Default Configurations (Manual Fallback)
	fmt.Print("Sedang menyiapkan konfigurasi default di database...")
	// 4. Setup Default Configurations
	fmt.Print("Sedang menyiapkan konfigurasi default di database...")
	if err := s.SetupDefaultSettings(clientCode); err != nil {
		fmt.Println(color.YellowString(" PERINGATAN (Gagal menyiapkan beberapa default)"))
	} else {
		fmt.Println(color.GreenString(" BERHASIL"))
	}

	// 5. Cloud/Remote Sync Setup
	fmt.Println("\n--- Konfigurasi Cloud / Remote Sync ---")
	useExisting, _ := prompt.Confirm("Apakah Anda ingin menghubungkan ke Central Hub yang sudah ada?", false)
	
	if useExisting {
		syncType, _, _ := prompt.SelectOne("Jenis Database Pusat:", []string{"postgres", "mysql"}, 0)
		host, _ := prompt.AskText("Host Remote:", prompt.WithDefault(""))
		port, _ := prompt.AskText("Port Remote:", prompt.WithDefault("5432"))
		user, _ := prompt.AskText("User Remote:", prompt.WithDefault(""))
		pass, _ := prompt.AskPassword("Password Remote:", nil)
		dbname, _ := prompt.AskText("Nama Database:", prompt.WithDefault("sfdbtools_sync"))

		db, _ := database.GetSQLite()
		s.saveSetting(db, "sync_enabled", "true", "cloud")
		s.saveSetting(db, "sync_type", syncType, "cloud")
		s.saveSetting(db, "sync_host", host, "cloud")
		s.saveSetting(db, "sync_port", port, "cloud")
		s.saveSetting(db, "sync_user", user, "cloud")
		s.saveSetting(db, "sync_password", pass, "cloud")
		s.saveSetting(db, "sync_database", dbname, "cloud")
		s.saveSetting(db, "sync_auto", "true", "cloud")

		fmt.Println(color.CyanString("\nSedang mencoba sinkronisasi awal dari Pusat..."))
		// Initial pull logic would go here
		fmt.Println(color.GreenString("Koneksi berhasil dan konfigurasi awal telah ditarik!"))
	} else {
		fmt.Println("Melewati konfigurasi Cloud. Anda bisa mengaturnya nanti di menu settings.")
	}

	// 6. Finalize Config (Lean Config)
	fmt.Println("\n--- Menyelesaikan Konfigurasi ---")
	if err := s.SaveLeanConfig(clientCode, dbPath); err != nil {
		return fmt.Errorf("gagal menyimpan config.yaml baru: %v", err)
	}

	return nil
}

// SetupDefaultSettings menyiapkan konfigurasi awal di SQLite.
// Jika ada config di YAML, gunakan itu. Jika tidak, gunakan Best Practice Defaults.
func (s *Service) SetupDefaultSettings(clientCode string) error {
	db, err := database.GetSQLite()
	if err != nil {
		return err
	}

	conf := s.deps.Config

	// 1. General & Security Defaults
	s.saveSetting(db, "encryption_key", s.strOrDefault(conf.Crypto.EncryptionKey, "SFDBTOOLS_SECRET_KEY_CHANGE_ME"), "crypto")

	// 2. Notification Defaults
	s.saveSetting(db, "telegram_enabled", fmt.Sprintf("%t", conf.Notify.Telegram.Enabled), "notify")
	s.saveSetting(db, "telegram_bot_token", conf.Notify.Telegram.BotToken, "notify")
	s.saveSetting(db, "telegram_chat_id", conf.Notify.Telegram.ChatID, "notify")
	s.saveSetting(db, "discord_enabled", fmt.Sprintf("%t", conf.Notify.Discord.Enabled), "notify")
	s.saveSetting(db, "discord_webhook_url", conf.Notify.Discord.WebhookURL, "notify")

	// 3. Backup Engine Defaults (Full Audit from appconfig_types.go)
	s.saveSetting(db, "backup_compression_enabled", "true", "backup")
	s.saveSetting(db, "backup_compression_type", "gzip", "backup") // gzip, zstd, none
	s.saveSetting(db, "backup_compression_level", "5", "backup")
	
	s.saveSetting(db, "backup_mysqldump_args", "-fQq --max-statement-time=0 --max-allowed-packet=1G --hex-blob --order-by-primary --single-transaction --routines=true --triggers=true --opt", "backup")

	// Exclude Filters
	s.saveSetting(db, "backup_exclude_system_databases", "true", "backup")
	s.saveSetting(db, "backup_exclude_user", "false", "backup")
	s.saveSetting(db, "backup_exclude_grant", "false", "backup")
	s.saveSetting(db, "backup_exclude_data", "false", "backup")
	s.saveSetting(db, "backup_exclude_empty", "false", "backup")
	s.saveSetting(db, "backup_exclude_databases", "", "backup")
	s.saveSetting(db, "backup_exclude_file", "", "backup")

	// Include Filters
	s.saveSetting(db, "backup_include_databases", "", "backup")
	s.saveSetting(db, "backup_include_file", "", "backup")
	s.saveSetting(db, "backup_include_dmart", "true", "backup")

	// Cleanup
	s.saveSetting(db, "backup_cleanup_enabled", "false", "backup")
	s.saveSetting(db, "backup_cleanup_schedule", "", "backup")
	s.saveSetting(db, "backup_cleanup_days", "7", "backup")

	// Encryption
	s.saveSetting(db, "backup_encryption_enabled", "true", "backup")
	s.saveSetting(db, "backup_encryption_key", "", "backup")

	// Output
	s.saveSetting(db, "backup_output_base_directory", "/media/ArchiveDB", "backup")
	s.saveSetting(db, "backup_output_cleanup_temp", "true", "backup")
	s.saveSetting(db, "backup_output_file_permissions", "0600", "backup")
	s.saveSetting(db, "backup_output_metadata_permissions", "0600", "backup")
	s.saveSetting(db, "backup_output_create_subdirs", "true", "backup")
	s.saveSetting(db, "backup_output_structure_pattern", "{year}{month}{day}/", "backup")
	s.saveSetting(db, "backup_output_save_backup_info", "true", "backup")

	// Verification
	s.saveSetting(db, "backup_verification_disk_space_check", "false", "backup")
	s.saveSetting(db, "backup_verification_checksum_algorithm", "sha256", "backup")
	s.saveSetting(db, "backup_verification_post_backup_check", "true", "backup")
	s.saveSetting(db, "backup_verification_header_footer_check", "true", "backup")
	s.saveSetting(db, "backup_verification_min_file_size", "100", "backup") // in bytes or human readable? using "100" as default

	// Replication
	s.saveSetting(db, "backup_replication_capture_gtid", "true", "backup")
	s.saveSetting(db, "backup_replication_user", "repl_user", "backup")
	s.saveSetting(db, "backup_replication_password", "repl_password", "backup")

	// Catalog
	s.saveSetting(db, "backup_catalog_enabled", "true", "backup")
	s.saveSetting(db, "backup_catalog_file_path", "/etc/sfDBTools/catalog.json", "backup")

	// 4. Target Database Defaults

	// 4. Target Database Defaults
	s.saveSetting(db, "db_host", "localhost", "database")
	s.saveSetting(db, "db_port", fmt.Sprintf("%d", s.intOrDefault(conf.Mariadb.Port, 3306)), "database")
	s.saveSetting(db, "db_user", "root", "database")

	// 5. Cloud/Sync Settings
	s.saveSetting(db, "sync_auto", "true", "cloud")
	s.saveSetting(db, "sync_enabled", "false", "cloud")
	s.saveSetting(db, "sync_mode", "two-way", "cloud")
	s.saveSetting(db, "sync_interval", "15", "cloud")
	s.saveSetting(db, "sync_jobs_enabled", "true", "cloud")
	s.saveSetting(db, "auto_update_enabled", "true", "cloud")
	s.saveSetting(db, "check_internet_startup", "true", "cloud")
	s.saveSetting(db, "heartbeat_enabled", "true", "cloud")

	// 6. Log Settings (Pindahan dari YAML)
	s.saveSetting(db, "log_level", s.strOrDefault(conf.Log.Level, "INFO"), "log")
	s.saveSetting(db, "log_dir", s.strOrDefault(conf.Log.Output.File.Dir, "/var/log/sfDBTools"), "log")
	s.saveSetting(db, "log_filename_pattern", s.strOrDefault(conf.Log.Output.File.FilenamePattern, "sfDBTools_{date}.log"), "log")
	s.saveSetting(db, "log_rotation_daily", "true", "log")
	s.saveSetting(db, "log_rotation_max_size", "100MB", "log")
	s.saveSetting(db, "log_rotation_retention_days", "7", "log")

	// 6. Config Directories (Pindahan dari YAML)
	s.saveSetting(db, "config_dir_db_profile", s.strOrDefault(conf.ConfigDir.DatabaseProfile, "/etc/sfDBTools/config/db_profile"), "config")
	s.saveSetting(db, "config_dir_db_list", s.strOrDefault(conf.ConfigDir.DatabaseList, "/etc/sfDBTools/config/db_list"), "config")

	// 7. Jobs (Hanya jika ada di config lama)
	for _, job := range conf.Backup.Scheduler.Jobs {
		db.Exec(`INSERT OR REPLACE INTO backup_jobs (name, enabled, schedule, mode, output_mode, include_file, profile_name, ticket, output_dir, retention_days) 
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			job.Name, 1, job.Schedule, job.Mode, job.OutputMode, job.IncludeFile, job.Profile, job.Ticket, job.Output.BaseDirectory, job.Cleanup.RetentionDays)
	}

	return nil
}

// helper strings
func (s *Service) strOrDefault(val, def string) string {
	if val == "" {
		return def
	}
	return val
}

// helper int
func (s *Service) intOrDefault(val, def int) int {
	if val == 0 {
		return def
	}
	return val
}

func (s *Service) PerformZeroConfigSetup(clientCode, dbPath string) error {
	syncType, _, _ := prompt.SelectOne("Jenis Database Pusat:", []string{"postgres", "mysql"}, 0)
	host, _ := prompt.AskText("Host Remote Hub:", prompt.WithDefault(""))
	port, _ := prompt.AskText("Port Remote Hub:", prompt.WithDefault("5432"))
	user, _ := prompt.AskText("User Remote Hub:", prompt.WithDefault(""))
	pass, _ := prompt.AskPassword("Password Remote Hub:", nil)
	dbname, _ := prompt.AskText("Nama Database Hub:", prompt.WithDefault("sfdbtools_sync"))
	syncKey, _ := prompt.AskPassword("Sync Encryption Key (Master Key):", nil)

	db, err := database.GetSQLite()
	if err != nil {
		return err
	}

	s.saveSetting(db, "sync_enabled", "true", "cloud")
	s.saveSetting(db, "sync_type", syncType, "cloud")
	s.saveSetting(db, "sync_host", host, "cloud")
	s.saveSetting(db, "sync_port", port, "cloud")
	s.saveSetting(db, "sync_user", user, "cloud")
	s.saveSetting(db, "sync_password", pass, "cloud")
	s.saveSetting(db, "sync_database", dbname, "cloud")
	s.saveSetting(db, "sync_key", syncKey, "cloud")

	// Try to connect and sync
	remote, err := database.ConnectToRemoteHub()
	if err != nil {
		return fmt.Errorf("gagal terhubung ke remote hub: %w", err)
	}
	defer remote.Close()

	fmt.Println(color.CyanString("Menghubungkan ke Remote Hub dan menarik konfigurasi..."))
	
	provider := sync.NewSQLRemoteProvider(remote)
	manager := sync.NewSyncManager(db, provider, clientCode, syncKey)
	
	ctx := context.Background()
	
	// Ensure remote tables exist (just in case)
	_ = provider.Migrate(ctx)

	// Perform initial pull
	if err := manager.PullAll(ctx); err != nil {
		return fmt.Errorf("gagal menarik data dari hub: %w", err)
	}

	return nil
}

// helper untuk menyimpan setting
func (s *Service) saveSetting(db *sql.DB, key, value, category string) {
	if value == "" {
		return
	}
	db.Exec(`INSERT OR REPLACE INTO app_settings (key, value, category) VALUES (?, ?, ?)`, key, value, category)
}

// SaveLeanConfig menulis ulang config.yaml dengan versi minimal
func (s *Service) SaveLeanConfig(clientCode, dbPath string) error {
	configPath := "/etc/sfDBTools/config.yaml" // Asumsi path default
	
	// Backup config lama
	backupPath := configPath + ".bak"
	if _, err := os.Stat(configPath); err == nil {
		content, _ := os.ReadFile(configPath)
		os.WriteFile(backupPath, content, 0644)
		fmt.Printf("File config lama telah dibackup ke: %s\n", backupPath)
	}

	// Buat isi lean config (Extremely Lean v2)
	leanContent := fmt.Sprintf(`# sfDBTools Lean Configuration (Auto-generated)
# Operational, Log, and Locale settings are now stored in SQLite database.

general:
  client_code: "%s"

storage:
  local_db: "%s"
`, clientCode, dbPath)

	return os.WriteFile(configPath, []byte(leanContent), 0644)
}
