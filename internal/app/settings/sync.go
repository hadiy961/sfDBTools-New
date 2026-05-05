package settings

import (
	"context"
	"database/sql"
	"fmt"
	"sfdbtools/internal/app/sync"
	appconfig "sfdbtools/internal/services/config"
	"sfdbtools/internal/shared/connectivity"
	"sfdbtools/internal/shared/database"
	"sfdbtools/internal/ui/print"
	"sfdbtools/internal/ui/prompt"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/fatih/color"
)

// SyncConfig mengupdate field di Config dengan nilai dari SQLite.
func SyncConfig(cfg *appconfig.Config) {
	if cfg == nil {
		return
	}

	// Jika SQLite belum inisialisasi, skip
	db, err := database.GetSQLite()
	if err != nil {
		return
	}
	_ = db

	// 1. Notification - Telegram
	cfg.Notify.Telegram.Enabled = database.GetSettingBool("telegram_enabled")
	if val := database.GetSetting("telegram_bot_token"); val != "" {
		// Use local crypto package for decoding if needed
		cfg.Notify.Telegram.BotToken = val
	}
	if val := database.GetSetting("telegram_chat_id"); val != "" {
		cfg.Notify.Telegram.ChatID = val
	}
	
	// 2. Cloud & Update (Partial sync to config struct if needed)
	// Add other fields as necessary based on appconfig.Config structure
}

func (s *Service) CloudMenu(db *sql.DB) {
	for {
		options := []string{
			"1. Manual Sync Now (Push/Pull)",
			"2. Configure Hub Credentials (Host, User, Pass)",
			"3. Sync Behavior (Mode, Auto, Jobs)",
			"4. Update & Auto-Update Settings",
			"0. Back",
		}
		
		sel, _, err := prompt.SelectOne("Cloud Hub Control:", options, 0)
		if err != nil || strings.Contains(sel, "Back") {
			return
		}

		switch sel[0:1] {
		case "1":
			s.RunManualSync()
		case "2":
			s.CredentialWizard(db)
		case "3":
			s.SyncBehaviorWizard(db)
		case "4":
			s.UpdateSettingsWizard(db)
		}
	}
}

func (s *Service) CredentialWizard(db *sql.DB) {
	keys := []string{"sync_type", "sync_host", "sync_port", "sync_user", "sync_password", "sync_database"}
	s.SpecificSettingsForm(db, "Hub Credentials", keys, "cloud")
}

func (s *Service) SyncBehaviorWizard(db *sql.DB) {
	keys := []string{"sync_enabled", "sync_auto", "sync_mode", "sync_interval", "sync_jobs_enabled", "heartbeat_enabled"}
	s.SpecificSettingsForm(db, "Sync Behavior", keys, "cloud")
}

func (s *Service) UpdateSettingsWizard(db *sql.DB) {
	keys := []string{"auto_update_enabled", "check_internet_startup"}
	s.SpecificSettingsForm(db, "Update Settings", keys, "cloud")
}

func (s *Service) SpecificSettingsForm(db *sql.DB, title string, keys []string, category string) {
	valMap := make(map[string]string)
	lockedMap := make(map[string]bool)
	for _, k := range keys {
		valMap[k] = database.GetSetting(k)
		// Check is_locked for each key
		var l int
		_ = db.QueryRow("SELECT is_locked FROM app_settings WHERE key = ?", k).Scan(&l)
		lockedMap[k] = l == 1
	}

	boolPtrs := make(map[string]*bool)
	strPtrs := make(map[string]*string)
	var fields []huh.Field

	for _, k := range keys {
		val := valMap[k]
		isLocked := lockedMap[k]
		key := k
		
		fTitle := key
		if isLocked {
			fTitle = fmt.Sprintf("%s %s", key, color.RedString("[LOCKED BY ADMIN]"))
		}
		
		if val == "true" || val == "false" || val == "" && (strings.Contains(k, "enabled") || strings.Contains(k, "auto")) {
			b := val == "true"
			boolPtrs[key] = &b
			if isLocked {
				fields = append(fields, huh.NewNote().Title(fTitle).Description(fmt.Sprintf("Current value: %v", b)))
			} else {
				fields = append(fields, huh.NewConfirm().Title(fTitle).Value(boolPtrs[key]))
			}
		} else {
			str := val
			strPtrs[key] = &str
			isSecret := strings.Contains(k, "password") || strings.Contains(k, "key")
			if isLocked {
				displayVal := str
				if isSecret {
					displayVal = "********"
				}
				fields = append(fields, huh.NewNote().Title(fTitle).Description(fmt.Sprintf("Current value: %s", displayVal)))
			} else {
				input := huh.NewInput().Title(fTitle).Value(strPtrs[key])
				if isSecret {
					input = input.EchoMode(huh.EchoModePassword)
				}
				fields = append(fields, input)
			}
		}
	}

	form := huh.NewForm(huh.NewGroup(fields...).Title("Edit: " + title))
	if err := form.Run(); err != nil {
		return
	}

	for _, k := range keys {
		if lockedMap[k] {
			continue
		}
		var newVal string
		if bPtr, ok := boolPtrs[k]; ok {
			newVal = "false"
			if *bPtr {
				newVal = "true"
			}
		} else if sPtr, ok := strPtrs[k]; ok {
			newVal = *sPtr
		}

		if newVal != valMap[k] {
			s.saveSetting(db, k, newVal, category)
		}
	}
	
	if category == "cloud" {
		s.UpdateSystemdTimer()
	}
	fmt.Println(color.GreenString("Pengaturan berhasil disimpan!"))
}

func (s *Service) RunManualSync() {
	print.PrintAppHeader("Manual Sync")
	fmt.Println("Sedang mencoba sinkronisasi ke remote hub...")

	if !connectivity.IsInternetAvailable(context.Background(), 3*time.Second) {
		fmt.Println(color.RedString("[ERROR] Tidak ada koneksi internet."))
		return
	}

	// 1. Get Settings
	syncEnabled := database.GetSettingBool("sync_enabled")
	if !syncEnabled {
		fmt.Println(color.YellowString("[WARN] Sinkronisasi dinonaktifkan di pengaturan."))
		return
	}

	// 2. Connect to Remote
	remote, err := database.ConnectToRemoteHub()
	if err != nil {
		fmt.Println(color.RedString("[ERROR] Gagal terhubung ke Remote Hub: %v", err))
		return
	}
	defer remote.Close()

	local, _ := database.GetSQLite()
	clientCode := s.deps.Config.General.ClientCode
	
	syncSvc := sync.NewService(local, remote, clientCode)
	ctx := context.Background()

	// 3. Perform Sync Actions
	fmt.Println(color.CyanString("Menghubungkan ke Remote Hub..."))
	
	// Ensure remote tables
	_ = syncSvc.MigrateRemoteHub(ctx)

	fmt.Print("- Syncing Settings...")
	if database.GetSetting("sync_mode") == "two-way" {
		syncSvc.PullSettings(ctx)
	}
	syncSvc.PushSettings(ctx)
	fmt.Println(color.GreenString(" OK"))

	fmt.Print("- Syncing Profiles...")
	if database.GetSetting("sync_mode") == "two-way" {
		syncSvc.PullProfiles(ctx)
	}
	syncSvc.PushProfiles(ctx)
	fmt.Println(color.GreenString(" OK"))

	fmt.Print("- Syncing Jobs...")
	if database.GetSetting("sync_mode") == "two-way" {
		syncSvc.PullJobs(ctx)
	}
	syncSvc.PushJobs(ctx)
	fmt.Println(color.GreenString(" OK"))

	fmt.Print("- Sending Heartbeat...")
	syncSvc.SendHeartbeat(ctx, database.GetSetting("backup_output_base_directory"))
	fmt.Println(color.GreenString(" OK"))

	fmt.Println(color.GreenString("\n[SUCCESS] Sinkronisasi selesai!"))
}
