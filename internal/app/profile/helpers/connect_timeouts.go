package helpers

import "time"

// Default timeout untuk operasi connect (DB/SSH) pada fitur profile.
// Dibuat terpusat supaya konsisten di seluruh flow.
//
// Catatan: saat ini belum dibuat configurable lewat env/flag agar tetap sederhana.
const (
	defaultProfileConnectTimeout = 15 * time.Second
)

func ProfileConnectTimeout() time.Duration {
	return defaultProfileConnectTimeout
}
