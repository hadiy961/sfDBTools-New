package spinnerguard

// File : pkg/spinnerguard/spinner_guard.go
// Deskripsi : Guard untuk suspend output spinner saat ada write ke console
// Author : Hadiyatna Muflihun
// Tanggal : 2 Januari 2026
// Last Modified : 2 Januari 2026

import "sync"

var (
	mu        sync.RWMutex
	suspender func(action func())
)

// SetSuspender mendaftarkan handler untuk men-suspend spinner saat menulis ke console.
// Jika tidak diset, RunWithSuspendedSpinner akan langsung menjalankan action.
func SetSuspender(s func(action func())) {
	mu.Lock()
	suspender = s
	mu.Unlock()
}

// RunWithSuspendedSpinner menjalankan action sambil men-suspend spinner aktif (jika ada).
func RunWithSuspendedSpinner(action func()) {
	mu.RLock()
	s := suspender
	mu.RUnlock()
	if s == nil {
		action()
		return
	}
	s(action)
}
