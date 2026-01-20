// File : internal/app/backup/cleanup_critical.go
// Deskripsi : Critical cleanup functions untuk force exit scenario
// Author : Hadiyatna Muflihun
// Tanggal : 2026-01-20
// Last Modified : 2026-01-20
package backup

import (
	"os"
	applog "sfdbtools/internal/services/log"
)

// criticalCleanup melakukan cleanup minimal yang HARUS dilakukan saat force exit.
// Dipanggil sebelum os.Exit() untuk mencegah resource leaks yang kritis.
//
// Issue #58: os.Exit() bypasses deferred cleanup, menyebabkan:
// - Stale lock files (deadlock pada backup selanjutnya)
// - Partial backup files tidak ter-cleanup (disk space waste)
// - File handles tidak ter-flush (corrupted files)
//
// Prinsip critical cleanup:
// 1. HANYA cleanup resources yang KRITIS (stale locks, partial files)
// 2. Cepat dan tidak boleh block (no network calls)
// 3. Best-effort (ignore errors, log saja)
// 4. Skip cleanup yang bisa auto-recover (connection pool timeout sendiri)
func criticalCleanup(state *BackupExecutionState, logger applog.Logger) {
	if state == nil {
		return
	}

	logger.Debug("Menjalankan critical cleanup sebelum force exit...")

	// 1. Cleanup partial backup file jika ada
	// Ini penting untuk mencegah disk space waste dan file corrupt
	currentFile, hasFile := state.GetCurrentBackupFile()
	if hasFile && currentFile != "" {
		logger.Debugf("Menghapus partial backup file: %s", currentFile)
		if err := os.Remove(currentFile); err != nil {
			// Best-effort: log saja jika gagal, jangan block
			logger.Debugf("Gagal menghapus partial file (non-critical): %v", err)
		} else {
			logger.Debugf("✓ Partial backup file terhapus: %s", currentFile)
		}
		state.ClearCurrentBackupFile()
	}

	// 2. Cleanup additional resources via state's Cleanup method
	// Ini akan cleanup resources yang di-register via EnableCleanup()
	// Contoh: lock files, temp directories, dll
	state.Cleanup()

	logger.Debug("Critical cleanup selesai")
}
