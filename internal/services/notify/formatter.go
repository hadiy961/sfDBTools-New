// File: internal/services/notify/formatter.go
package notify

import (
	"fmt"
	"strings"
)

// levelEmoji mengembalikan emoji sesuai Level untuk Telegram
func levelEmoji(l Level) string {
	switch l {
	case LevelCritical:
		return "🔴"
	case LevelWarning:
		return "🟡"
	case LevelSuccess:
		return "✅"
	default:
		return "ℹ️"
	}
}

// levelIcon mengembalikan icon ASCII untuk email (tidak support emoji)
func levelIcon(l Level) string {
	switch l {
	case LevelCritical:
		return "[CRITICAL]"
	case LevelWarning:
		return "[WARNING]"
	case LevelSuccess:
		return "[SUCCESS]"
	default:
		return "[INFO]"
	}
}

// FormatTelegramMessage memformat Message untuk Telegram (HTML)
// Gunakan tag HTML: <b>, <i>, <code>, <pre>
func FormatTelegramMessage(msg Message) string {
	var sb strings.Builder
	emoji := levelEmoji(msg.Level)

	sb.WriteString(fmt.Sprintf("%s <b>%s</b>\n\n", emoji, escapeHTML(msg.Title)))
	if msg.Feature != "" {
		sb.WriteString(fmt.Sprintf("📦 <b>Feature:</b> <code>%s</code>\n\n", escapeHTML(msg.Feature)))
	}
	sb.WriteString(msg.Body)
	sb.WriteString("\n\n<i>— sfDBTools Notification</i>")
	return sb.String()
}

// FormatEmailMessage memformat Message untuk email (plain text)
func FormatEmailMessage(msg Message) string {
	var sb strings.Builder
	icon := levelIcon(msg.Level)

	sb.WriteString(fmt.Sprintf("%s %s\n", icon, msg.Title))
	sb.WriteString(strings.Repeat("─", 60) + "\n\n")
	if msg.Feature != "" {
		sb.WriteString(fmt.Sprintf("Feature : %s\n\n", msg.Feature))
	}
	sb.WriteString(msg.Body)
	sb.WriteString("\n\n" + strings.Repeat("─", 60))
	sb.WriteString("\nDikirim oleh sfDBTools Notification Service")
	return sb.String()
}

// escapeHTML mem-escape karakter HTML agar tidak merusak parse_mode Telegram
func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}
