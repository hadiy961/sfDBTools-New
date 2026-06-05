// File: internal/services/notify/model.go
package notify

// Channel mendefinisikan nama provider notifikasi
type Channel string

const (
	ChannelTelegram Channel = "telegram"
	ChannelEmail    Channel = "email"
)

// Level menentukan tingkat urgensi notifikasi
type Level string

const (
	LevelInfo     Level = "INFO"
	LevelWarning  Level = "WARNING"
	LevelCritical Level = "CRITICAL"
	LevelSuccess  Level = "SUCCESS"
)

// Message adalah unit pesan yang akan dikirim
type Message struct {
	// Title adalah judul singkat notifikasi
	// contoh: "[sfDBTools] Backup Health Alert"
	Title string

	// Body adalah isi pesan, boleh berisi HTML jika target channel support
	Body string

	// Level menentukan urgensi (dipakai untuk emoji/warna prefix)
	Level Level

	// Feature adalah nama fitur pengirim (untuk context di pesan)
	// contoh: "backup-health", "scheduler", "restore"
	Feature string

	// Channels override channel default dari config.
	// Jika kosong, gunakan config.Notify.DefaultChannels
	Channels []Channel

	// ToEmails override daftar penerima email dari config.
	// Jika nil, gunakan config.Email.ToEmails
	ToEmails []string
}

// Result menyimpan hasil pengiriman notifikasi per channel
type Result struct {
	Channel Channel
	Success bool
	Err     error
}

// SendReport adalah kumpulan hasil dari semua channel
type SendReport struct {
	Message Message
	Results []Result
}

// AnySuccess mengembalikan true jika minimal satu channel berhasil
func (r *SendReport) AnySuccess() bool {
	for _, res := range r.Results {
		if res.Success {
			return true
		}
	}
	return false
}

// HasError mengembalikan true jika ada channel yang gagal
func (r *SendReport) HasError() bool {
	for _, res := range r.Results {
		if !res.Success {
			return true
		}
	}
	return false
}
