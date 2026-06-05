package settings

import (
	"fmt"
	"sfdbtools/internal/shared/database"
	"sfdbtools/internal/ui/table"
	"sort"
	"strings"

	"github.com/fatih/color"
)

func (s *Service) ViewAllSettings() {
	db, err := database.GetSQLite()
	if err != nil {
		fmt.Println(color.RedString("Error: %v", err))
		return
	}

	rows, err := db.Query("SELECT key, value, category FROM app_settings ORDER BY category, key")
	if err != nil {
		fmt.Println(color.RedString("Error: %v", err))
		return
	}
	defer rows.Close()

	settingsByCat := make(map[string][][]string)
	var categories []string

	for rows.Next() {
		var key, val, cat string
		rows.Scan(&key, &val, &cat)

		// Beautify Value
		displayVal := val
		if val == "true" || val == "enabled" || val == "active" {
			displayVal = color.GreenString("✔ %s", val)
		} else if val == "false" || val == "disabled" || val == "inactive" {
			displayVal = color.RedString("✘ %s", val)
		} else {
			displayVal = color.CyanString("%s", val)
		}

		if _, ok := settingsByCat[cat]; !ok {
			categories = append(categories, cat)
		}
		settingsByCat[cat] = append(settingsByCat[cat], []string{key, displayVal})
	}
	sort.Strings(categories)

	categoryDisplayNames := map[string]string{
		"general":      "General Settings",
		"cloud":        "Cloud & Remote Sync",
		"log":          "Logging & Debugging",
		"notification": "Notification Services",
		"database":     "Default/Local Database (Fallback)",
		"backup":       "Backup & Retention Policies",
		"config":       "System Path Configurations",
		"crypto":       "Security & Encryption",
		"notify":       "Notification Services",
	}

	categoryDescriptions := map[string]string{
		"general":      "Pengaturan dasar aplikasi dan perilaku sistem.",
		"cloud":        "Sinkronisasi ke storage remote atau database remote.",
		"log":          "Level pencatatan aktivitas dan lokasi log.",
		"notification": "Pengaturan pengiriman notifikasi (Telegram, Discord, dll).",
		"database":     "Koneksi default jika profile tidak ditentukan (Localhost).",
		"backup":       "Aturan retensi dan kompresi untuk proses backup.",
		"config":       "Lokasi direktori untuk profil dan daftar database.",
		"crypto":       "Kunci enkripsi data sensitive.",
		"notify":       "Pengaturan pengiriman notifikasi (Telegram, Discord, dll).",
	}

	fmt.Println(color.HiWhiteString("\n  DETAIL PENGATURAN SFDBTOOLS"))
	fmt.Println(color.HiBlackString("  =================================================="))

	for _, cat := range categories {
		displayName := categoryDisplayNames[cat]
		if displayName == "" {
			displayName = strings.ToUpper(cat)
		}
		description := categoryDescriptions[cat]

		fmt.Printf("\n  %s %s\n", s.getCategoryIcon(cat), color.New(color.Bold, color.FgHiYellow).Sprint(displayName))
		if description != "" {
			fmt.Printf("    \033[3;90m%s\033[0m\n", description) // Italic Gray description
		}
		table.Render([]string{"Parameter", "Value"}, settingsByCat[cat])
	}
	fmt.Println()
}

func (s *Service) getCategoryIcon(cat string) string {
	switch strings.ToLower(cat) {
	case "notification", "notify":
		return "🔔"
	case "sync", "cloud":
		return "☁️  "
	case "log":
		return "📝"
	case "database", "storage":
		return "📦"
	case "general":
		return "⚙️ "
	default:
		return "🔹"
	}
}
