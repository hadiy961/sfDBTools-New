package notifycmd

import (
	"fmt"
	appdeps "sfdbtools/internal/cli/deps"
	"github.com/spf13/cobra"
)

var CmdNotifyGetChatID = &cobra.Command{
	Use:   "get-chat-id",
	Short: "Cari Telegram Chat ID dari pesan terbaru yang masuk ke bot",
	Long: `Membantu mencari Chat ID (User atau Group).
Caranya:
1. Pastikan bot_token sudah diisi di config.yaml
2. Kirim pesan apa saja ke bot Anda (lewat private chat atau grup)
3. Jalankan perintah ini untuk melihat Chat ID dari pengirim pesan.`,
	Run: func(cmd *cobra.Command, args []string) {
		svc := appdeps.Deps.NotifyService
		if svc == nil {
			fmt.Println("Error: notification service tidak diinisialisasi")
			return
		}

		fmt.Println("🔍 Mencari pesan terbaru dari Telegram Bot API...")
		updates, err := svc.GetUpdates()
		if err != nil {
			fmt.Printf("❌ Gagal mengambil updates: %v\n", err)
			return
		}

		if len(updates) == 0 {
			fmt.Println("\nℹ️ Tidak ada pesan terbaru.")
			fmt.Println("Tips: Kirim pesan ke bot Anda sekarang, lalu jalankan perintah ini lagi.")
			return
		}

		fmt.Printf("\nFound %d recent messages:\n", len(updates))
		fmt.Println("--------------------------------------------------------------------------------")
		fmt.Printf("%-15s | %-15s | %-15s | %s\n", "Type", "Chat ID", "Sender", "Text")
		fmt.Println("--------------------------------------------------------------------------------")
		
		for _, u := range updates {
			chatType := u.Message.Chat.Type
			chatID := u.Message.Chat.ID
			sender := u.Message.From.Username
			if sender == "" {
				sender = "unknown"
			}
			text := u.Message.Text
			if len(text) > 30 {
				text = text[:27] + "..."
			}

			fmt.Printf("%-15s | %-15d | %-15s | %s\n", chatType, chatID, sender, text)
		}
		fmt.Println("--------------------------------------------------------------------------------")
		fmt.Println("\n✅ Silakan copy Chat ID di atas dan masukkan ke config.yaml (notify.telegram.chat_id)")
	},
}
