// File : pkg/servicehelper/servicehelper_progress.go
// Deskripsi : Helper untuk progress tracking dalam service operations
// Author : Hadiyatna Muflihun
// Tanggal : 2025-11-11
// Last Modified : 2025-11-11

package servicehelper

// ProgressTracker interface untuk service yang memiliki progress tracking
type ProgressTracker interface {
	SetRestoreInProgress(bool)
}

// TrackProgress menandai operasi sedang berlangsung dan otomatis clear saat selesai
// Menggunakan defer pattern untuk memastikan cleanup selalu terjadi
//
// Usage:
//
//	defer servicehelper.TrackProgress(service)()
//
// atau dengan variable:
//
//	cleanup := servicehelper.TrackProgress(service)
//	defer cleanup()
func TrackProgress(tracker ProgressTracker) func() {
	tracker.SetRestoreInProgress(true)
	return func() {
		tracker.SetRestoreInProgress(false)
	}
}
