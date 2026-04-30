package notifycmd

import (
	"fmt"
	appdeps "sfdbtools/internal/cli/deps"
	"sfdbtools/internal/services/notify"

	"github.com/spf13/cobra"
)

var (
	testChannel string // flag: --channel telegram|email (kosong = semua)
)

var CmdNotifyTest = &cobra.Command{
	Use:   "test",
	Short: "Kirim test notification untuk validasi konfigurasi",
	Example: `  sfdbtools notify test
  sfdbtools notify test --channel telegram
  sfdbtools notify test --channel email`,
	Run: func(cmd *cobra.Command, args []string) {
		svc := appdeps.Deps.NotifyService
		if svc == nil {
			fmt.Println("Error: notification service tidak diinisialisasi")
			return
		}

		msg := notify.Message{
			Title:   "Test Notification",
			Body:    "Ini adalah pesan test dari sfDBTools.\nJika Anda menerima pesan ini, konfigurasi notifikasi sudah benar.",
			Level:   notify.LevelInfo,
			Feature: "system",
		}

		// Override channel jika flag diisi
		if testChannel != "" {
			msg.Channels = []notify.Channel{notify.Channel(testChannel)}
		}

		report := svc.Send(msg)

		// Tampilkan hasil
		fmt.Println("\n=== Hasil Test Notification ===")
		if len(report.Results) == 0 {
			fmt.Println("⚠ Tidak ada channel yang aktif. Periksa config notify.default_channels")
			fmt.Println("  atau aktifkan telegram/email di config.yaml")
			return
		}
		for _, r := range report.Results {
			if r.Success {
				fmt.Printf("  ✅ %-10s → Berhasil\n", r.Channel)
			} else {
				fmt.Printf("  ❌ %-10s → Gagal: %v\n", r.Channel, r.Err)
			}
		}
	},
}

func init() {
	CmdNotifyTest.Flags().StringVar(&testChannel, "channel", "", "Channel target (telegram|email). Default: semua channel enabled")
}
