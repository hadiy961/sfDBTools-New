package settings

import (
	"fmt"
	"sfdbtools/internal/crypto"
	"sfdbtools/internal/shared/database"

	"github.com/charmbracelet/huh"
	"github.com/fatih/color"
)

func (s *Service) SetupTelegramWizard() {
	db, err := database.GetSQLite()
	if err != nil {
		fmt.Println(color.RedString("Error: %v", err))
		return
	}

	fmt.Println(color.HiYellowString("\n--- WIZARD KONFIGURASI TELEGRAM ---"))
	fmt.Println("Gunakan bot @BotFather untuk membuat bot dan @userinfobot untuk mendapatkan ChatID.")

	var currentEnabled, currentToken, currentChatID string
	db.QueryRow("SELECT value FROM app_settings WHERE key = 'telegram_enabled'").Scan(&currentEnabled)
	db.QueryRow("SELECT value FROM app_settings WHERE key = 'telegram_bot_token'").Scan(&currentToken)
	db.QueryRow("SELECT value FROM app_settings WHERE key = 'telegram_chat_id'").Scan(&currentChatID)

	decToken, _, _ := crypto.DecodeEnvSecret(currentToken)

	token := decToken
	chatID := currentChatID
	enabled := currentEnabled == "true"

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Aktifkan Notifikasi Telegram?").
				Value(&enabled),
			huh.NewInput().
				Title("Bot Token:").
				EchoMode(huh.EchoModePassword).
				Value(&token),
			huh.NewInput().
				Title("Chat ID:").
				Value(&chatID),
		).Title("Telegram Settings"),
	)

	err = form.Run()
	if err != nil {
		fmt.Println(color.YellowString("\n[INTERRUPT] Operasi dibatalkan."))
		return
	}

	if token != "" {
		encToken, err := crypto.EncodeEnvSecret(token)
		if err == nil {
			s.saveSetting(db, "telegram_bot_token", encToken, "notify")
		} else {
			fmt.Println(color.RedString("Gagal mengenkripsi token: %v", err))
		}
	} else {
		// allow clearing the token
		s.saveSetting(db, "telegram_bot_token", "", "notify")
	}

	s.saveSetting(db, "telegram_chat_id", chatID, "notify")
	s.saveSetting(db, "telegram_enabled", fmt.Sprintf("%t", enabled), "notify")

	fmt.Println(color.GreenString("\n[SUCCESS] Konfigurasi Telegram berhasil disimpan!"))
}
