package notifycmd

import (
	"fmt"
	appdeps "sfdbtools/internal/cli/deps"
	"github.com/spf13/cobra"
)

var CmdNotifyDeleteWebhook = &cobra.Command{
	Use:   "delete-webhook",
	Short: "Hapus webhook Telegram yang aktif agar bisa menggunakan polling (get-chat-id)",
	Run: func(cmd *cobra.Command, args []string) {
		svc := appdeps.Deps.NotifyService
		if svc == nil {
			fmt.Println("Error: notification service tidak diinisialisasi")
			return
		}

		fmt.Println("⏳ Menghapus Telegram webhook...")
		err := svc.DeleteWebhook()
		if err != nil {
			fmt.Printf("❌ Gagal menghapus webhook: %v\n", err)
			return
		}

		fmt.Println("✅ Webhook berhasil dihapus! Sekarang Anda bisa menjalankan 'get-chat-id'.")
	},
}
