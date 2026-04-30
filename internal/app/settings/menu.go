package settings

import (
	"database/sql"
	"fmt"
	"sfdbtools/internal/shared/database"
	"sfdbtools/internal/ui/prompt"
	"sort"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/fatih/color"
)

func (s *Service) MaintenanceMenu() {
	for {
		options := []string{
			"1. Reset All to Defaults",
			"2. Update/Repair Systemd Timer",
			"0. Back",
		}
		
		sel, _, err := prompt.SelectOne("Maintenance:", options, 0)
		if err != nil || strings.Contains(sel, "Back") {
			return
		}

		switch sel[0:1] {
		case "1":
			s.ResetSettings()
		case "2":
			s.UpdateSystemdTimer()
		}
	}
}

func (s *Service) EditSettingsMenu() {
	db, err := database.GetSQLite()
	if err != nil {
		fmt.Println(color.RedString("Error DB: %v", err))
		return
	}

	for {
		// Ambil semua kategori unik
		rows, err := db.Query("SELECT DISTINCT category FROM app_settings")
		if err != nil {
			fmt.Println(color.RedString("Gagal mengambil kategori: %v", err))
			return
		}
		
		var categories []string
		for rows.Next() {
			var cat string
			rows.Scan(&cat)
			categories = append(categories, cat)
		}
		rows.Close()
		sort.Strings(categories)
		categories = append(categories, "Cancel")

		selectedCat, _, err := prompt.SelectOne("Pilih Kategori:", categories, 0)
		if err != nil || selectedCat == "Cancel" {
			return
		}

		// Route based on category
		switch selectedCat {
		case "notify", "notification":
			s.NotifyMenu(db)
		case "backup":
			s.BackupMenu(db)
		case "cloud":
			s.CloudMenu(db)
		case "security":
			s.SecurityMenu(db)
		default:
			s.GenericCategoryMenu(db, selectedCat)
		}
	}
}

func (s *Service) GenericCategoryMenu(db *sql.DB, category string) {
	rows, _ := db.Query("SELECT key, value, is_locked FROM app_settings WHERE category = ?", category)
	var keys []string
	valMap := make(map[string]string)
	lockedMap := make(map[string]bool)
	for rows.Next() {
		var k, v string
		var l int
		rows.Scan(&k, &v, &l)
		keys = append(keys, k)
		valMap[k] = v
		lockedMap[k] = l == 1
	}
	rows.Close()
	sort.Strings(keys)

	if len(keys) == 0 {
		fmt.Println(color.YellowString("Kategori '%s' belum memiliki parameter.", category))
		return
	}

	boolPtrs := make(map[string]*bool)
	strPtrs := make(map[string]*string)

	var fields []huh.Field

	for _, key := range keys {
		val := valMap[key]
		isLocked := lockedMap[key]
		k := key
		
		title := k
		if isLocked {
			title = fmt.Sprintf("%s %s", k, color.RedString("[LOCKED BY ADMIN]"))
		}
		
		if val == "true" || val == "false" {
			b := val == "true"
			boolPtrs[k] = &b
			f := huh.NewConfirm().Title(title).Value(boolPtrs[k])
			if isLocked {
				f = f.ReadOnly(true)
			}
			fields = append(fields, f)
		} else {
			str := val
			strPtrs[k] = &str
			
			isSecret := false
			if strings.Contains(k, "key") || strings.Contains(k, "token") || strings.Contains(k, "password") || strings.Contains(k, "secret") {
				isSecret = true
			}
			
			input := huh.NewInput().Title(title).Value(strPtrs[k])
			if isSecret {
				input = input.EchoMode(huh.EchoModePassword)
			}
			if isLocked {
				input = input.ReadOnly(true)
			}
			fields = append(fields, input)
		}
	}

	form := huh.NewForm(
		huh.NewGroup(fields...).Title(fmt.Sprintf("Edit Kategori: %s", category)),
	)

	err := form.Run()
	if err != nil {
		fmt.Println(color.YellowString("\n[INTERRUPT] Operasi dibatalkan."))
		return
	}

	changes := 0
	for _, key := range keys {
		if lockedMap[key] {
			continue // Jangan simpan jika di-lock
		}
		oldVal := valMap[key]
		var newVal string
		if bPtr, ok := boolPtrs[key]; ok {
			newVal = "false"
			if *bPtr {
				newVal = "true"
			}
		} else if sPtr, ok := strPtrs[key]; ok {
			newVal = *sPtr
		}

		if newVal != oldVal {
			s.saveSetting(db, key, newVal, category)
			changes++
		}
	}

	if changes > 0 {
		fmt.Println(color.GreenString("Berhasil memperbarui %d parameter!", changes))
		
		// [SYSTEMD HOOK]
		if category == "cloud" {
			s.UpdateSystemdTimer()
		}
	} else {
		fmt.Println("Tidak ada perubahan yang disimpan.")
	}
}

func (s *Service) NotifyMenu(db *sql.DB) {
	for {
		options := []string{
			"🚀 [WIZARD] Setup Telegram Notification",
			"🔧 Edit Discord Settings",
			"🔧 Edit All Notify Settings (Manual)",
			"Back",
		}
		
		sel, _, err := prompt.SelectOne("Notification Settings:", options, 0)
		if err != nil || sel == "Back" {
			return
		}

		switch sel {
		case "🚀 [WIZARD] Setup Telegram Notification":
			s.SetupTelegramWizard()
		case "🔧 Edit Discord Settings":
			s.GenericCategoryMenu(db, "notify") // Ini butuh filter key LIKE 'discord%' tapi untuk simpelnya panggil generic dulu
		case "🔧 Edit All Notify Settings (Manual)":
			s.GenericCategoryMenu(db, "notify")
		}
	}
}

func (s *Service) BackupMenu(db *sql.DB) {
	for {
		options := []string{
			"🚀 [WIZARD] Quick Backup Engine Setup",
			"🔧 Edit All Backup Parameters (Manual)",
			"Back",
		}
		sel, _, err := prompt.SelectOne("Backup Settings:", options, 0)
		if err != nil || sel == "Back" {
			return
		}

		switch sel {
		case "🚀 [WIZARD] Quick Backup Engine Setup":
			s.SetupBackupWizard()
		case "🔧 Edit All Backup Parameters (Manual)":
			s.GenericCategoryMenu(db, "backup")
		}
	}
}
