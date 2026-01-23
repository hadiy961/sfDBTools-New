// File : internal/ui/menu/icons.go
// Deskripsi : Icon mapping untuk command
// Author : Hadiyatna Muflihun
// Tanggal : 23 Januari 2026
// Last Modified : 23 Januari 2026

package menu

func getCommandIcon(cmdName string) string {
	iconMap := map[string]string{
		"db-backup":  "💾",
		"backup":     "💾",
		"db-restore": "♻️",
		"restore":    "♻️",
		"db-scan":    "🔍",
		"dbscan":     "🔍",
		"profile":    "⚙️",
		"cleanup":    "🧹",
		"crypto":     "🔐",
		"script":     "📜",
		"version":    "ℹ️",
		"update":     "⬆️",
	}

	if icon, ok := iconMap[cmdName]; ok {
		return icon
	}
	return "▶️"
}
