// File: internal/services/notify/builder.go
package notify

import (
	"fmt"
	"os"
	"strings"
	"time"

	types_backup "sfdbtools/internal/app/backup/model/types_backup"
	types_cleanup "sfdbtools/internal/app/cleanup/model"
	restoremodel "sfdbtools/internal/app/restore/model"
	"sfdbtools/internal/ui/text"
)

// Helper internal untuk mendapatkan hostname
func getHostname() string {
	h, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return h
}

// Helper untuk format waktu
func formatTimestamp(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// Helper internal untuk konversi duration
func formatDuration(d time.Duration) string {
	return d.Round(time.Second).String()
}

// ============================================================================
// Backup Notifications
// ============================================================================

// BuildBackupMessage menyusun pesan notifikasi untuk hasil backup.
func BuildBackupMessage(res *types_backup.BackupResult, err error, mode string, ticket string, profile string) Message {
	hostname := getHostname()
	timestamp := formatTimestamp(time.Now())

	if err != nil {
		return Message{
			Title:   fmt.Sprintf("Backup Gagal (%s)", mode),
			Level:   LevelCritical,
			Feature: "backup",
			Body:    fmt.Sprintf("<b>Host:</b> <code>%s</code>\n<b>Time:</b> <code>%s</code>\n<b>Ticket:</b> <code>%s</code>\n<b>Profile:</b> <code>%s</code>\n\n<b>Error:</b>\n<pre><code>%s</code></pre>", escapeHTML(hostname), escapeHTML(timestamp), escapeHTML(ticket), escapeHTML(profile), escapeHTML(err.Error())),
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<b>Host:</b> <code>%s</code>\n", escapeHTML(hostname)))
	sb.WriteString(fmt.Sprintf("<b>Time:</b> <code>%s</code>\n", escapeHTML(timestamp)))
	if ticket != "" {
		sb.WriteString(fmt.Sprintf("<b>Ticket:</b> <code>%s</code>\n", escapeHTML(ticket)))
	}
	if profile != "" {
		sb.WriteString(fmt.Sprintf("<b>Profile:</b> <code>%s</code>\n", escapeHTML(profile)))
	}
	sb.WriteString(fmt.Sprintf("<b>Durasi Total:</b> <code>%s</code>\n", formatDuration(res.TotalTimeTaken)))
	sb.WriteString(fmt.Sprintf("<b>Berhasil:</b> <code>%d</code> | <b>Gagal:</b> <code>%d</code>\n\n", res.SuccessfulBackups, res.FailedBackups))

	if len(res.BackupInfo) > 0 {
		sb.WriteString("<b>✅ Database Berhasil:</b>\n")
		for _, info := range res.BackupInfo {
			sizeMB := float64(info.FileSize) / 1024 / 1024
			sb.WriteString(fmt.Sprintf("• <code>%s</code> (%.2f MB)\n", escapeHTML(info.DatabaseName), sizeMB))
		}
		sb.WriteString("\n")
	}

	if len(res.FailedDatabaseInfos) > 0 {
		sb.WriteString("<b>❌ Database Gagal:</b>\n")
		for _, info := range res.FailedDatabaseInfos {
			sb.WriteString(fmt.Sprintf("• <code>%s</code>: %s\n", escapeHTML(info.DatabaseName), escapeHTML(info.Error)))
		}
	}

	level := LevelSuccess
	if res.FailedBackups > 0 {
		level = LevelWarning
		if res.SuccessfulBackups == 0 {
			level = LevelCritical
		}
	}

	return Message{
		Title:   fmt.Sprintf("Hasil Backup (%s)", mode),
		Level:   level,
		Feature: "backup",
		Body:    sb.String(),
	}
}

// ============================================================================
// Restore Notifications
// ============================================================================

// BuildRestoreMessage menyusun pesan notifikasi untuk hasil restore.
func BuildRestoreMessage(res *restoremodel.RestoreResult, err error, targetHost string, ticket string) Message {
	hostname := getHostname()
	timestamp := formatTimestamp(time.Now())

	if err != nil {
		return Message{
			Title:   "Restore Gagal",
			Level:   LevelCritical,
			Feature: "restore",
			Body:    fmt.Sprintf("<b>Host:</b> <code>%s</code>\n<b>Time:</b> <code>%s</code>\n<b>Ticket:</b> <code>%s</code>\n<b>Target Host:</b> <code>%s</code>\n\n<b>Error:</b>\n<pre><code>%s</code></pre>", escapeHTML(hostname), escapeHTML(timestamp), escapeHTML(ticket), escapeHTML(targetHost), escapeHTML(err.Error())),
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<b>Host:</b> <code>%s</code>\n", escapeHTML(hostname)))
	sb.WriteString(fmt.Sprintf("<b>Time:</b> <code>%s</code>\n", escapeHTML(timestamp)))
	if ticket != "" {
		sb.WriteString(fmt.Sprintf("<b>Ticket:</b> <code>%s</code>\n", escapeHTML(ticket)))
	}
	sb.WriteString(fmt.Sprintf("<b>Target Host:</b> <code>%s</code>\n", escapeHTML(targetHost)))
	if res.TargetDB != "" {
		sb.WriteString(fmt.Sprintf("<b>Target DB:</b> <code>%s</code>\n", escapeHTML(res.TargetDB)))
	}
	sb.WriteString(fmt.Sprintf("<b>Durasi:</b> <code>%s</code>\n", escapeHTML(res.Duration)))
	sb.WriteString(fmt.Sprintf("<b>Source File:</b>\n<code>%s</code>\n\n", escapeHTML(res.SourceFile)))

	level := LevelSuccess
	if res.SQLErrors > 0 || res.SQLWarnings > 0 {
		level = LevelWarning
		sb.WriteString(fmt.Sprintf("⚠ Diselesaikan dengan <b>%d Error</b> dan <b>%d Warning</b> SQL.\n", res.SQLErrors, res.SQLWarnings))
	} else {
		sb.WriteString("Restore selesai tanpa error SQL.\n")
	}

	return Message{
		Title:   "Hasil Restore",
		Level:   level,
		Feature: "restore",
		Body:    sb.String(),
	}
}

// ============================================================================
// Copy Notifications
// ============================================================================

// CopyDBInfo represents a single copy result for notification building.
type CopyDBInfo struct {
	SourceDB string
	TargetDB string
	Duration time.Duration
	Error    error
}

// BuildCopyDBMessage menyusun pesan notifikasi untuk kloning banyak DB.
func BuildCopyDBMessage(results []CopyDBInfo, err error, ticket string) Message {
	hostname := getHostname()
	timestamp := formatTimestamp(time.Now())

	if err != nil {
		return Message{
			Title:   "Copy Database Gagal",
			Level:   LevelCritical,
			Feature: "copy",
			Body:    fmt.Sprintf("<b>Host:</b> <code>%s</code>\n<b>Time:</b> <code>%s</code>\n<b>Ticket:</b> <code>%s</code>\n\n<b>Error:</b>\n<pre><code>%s</code></pre>", escapeHTML(hostname), escapeHTML(timestamp), escapeHTML(ticket), escapeHTML(err.Error())),
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<b>Host:</b> <code>%s</code>\n", escapeHTML(hostname)))
	sb.WriteString(fmt.Sprintf("<b>Time:</b> <code>%s</code>\n", escapeHTML(timestamp)))
	if ticket != "" {
		sb.WriteString(fmt.Sprintf("<b>Ticket:</b> <code>%s</code>\n", escapeHTML(ticket)))
	}
	success := 0
	failed := 0
	var totalDuration time.Duration

	for _, r := range results {
		totalDuration += r.Duration
		if r.Error != nil {
			failed++
		} else {
			success++
		}
	}

	sb.WriteString(fmt.Sprintf("<b>Durasi Total:</b> <code>%s</code>\n", formatDuration(totalDuration)))
	sb.WriteString(fmt.Sprintf("<b>Berhasil:</b> <code>%d</code> | <b>Gagal:</b> <code>%d</code>\n\n", success, failed))

	for _, r := range results {
		if r.Error != nil {
			sb.WriteString(fmt.Sprintf("❌ <code>%s</code> -> <code>%s</code>\n   Error: %s\n", escapeHTML(r.SourceDB), escapeHTML(r.TargetDB), escapeHTML(r.Error.Error())))
		} else {
			sb.WriteString(fmt.Sprintf("✅ <code>%s</code> -> <code>%s</code> (%s)\n", escapeHTML(r.SourceDB), escapeHTML(r.TargetDB), formatDuration(r.Duration)))
		}
	}

	level := LevelSuccess
	if failed > 0 {
		level = LevelWarning
		if success == 0 {
			level = LevelCritical
		}
	}

	return Message{
		Title:   "Hasil Copy Database",
		Level:   level,
		Feature: "copy",
		Body:    sb.String(),
	}
}

// BuildCopyTableMessage menyusun pesan notifikasi untuk penyalinan tabel.
func BuildCopyTableMessage(sourceDB, targetDB string, tables []string, duration time.Duration, err error, ticket string) Message {
	hostname := getHostname()
	timestamp := formatTimestamp(time.Now())

	if err != nil {
		return Message{
			Title:   "Copy Tabel Gagal",
			Level:   LevelCritical,
			Feature: "copy",
			Body:    fmt.Sprintf("<b>Host:</b> <code>%s</code>\n<b>Time:</b> <code>%s</code>\n<b>Ticket:</b> <code>%s</code>\n<b>Source:</b> <code>%s</code>\n<b>Target:</b> <code>%s</code>\n\n<b>Error:</b>\n<pre><code>%s</code></pre>", escapeHTML(hostname), escapeHTML(timestamp), escapeHTML(ticket), escapeHTML(sourceDB), escapeHTML(targetDB), escapeHTML(err.Error())),
		}
	}

	return Message{
		Title:   "Copy Tabel Selesai",
		Level:   LevelSuccess,
		Feature: "copy",
		Body:    fmt.Sprintf("<b>Host:</b> <code>%s</code>\n<b>Time:</b> <code>%s</code>\n<b>Ticket:</b> <code>%s</code>\n<b>Source:</b> <code>%s</code>\n<b>Target:</b> <code>%s</code>\n<b>Total Tabel:</b> <code>%d</code>\n<b>Durasi:</b> <code>%s</code>\n", escapeHTML(hostname), escapeHTML(timestamp), escapeHTML(ticket), escapeHTML(sourceDB), escapeHTML(targetDB), len(tables), formatDuration(duration)),
	}
}

// ============================================================================
// Cleanup Notifications
// ============================================================================

// BuildCleanupMessage menyusun pesan notifikasi untuk proses cleanup.
func BuildCleanupMessage(mode string, res *types_cleanup.CleanupResult, err error, ticket string) Message {
	hostname := getHostname()
	timestamp := formatTimestamp(time.Now())

	if err != nil {
		return Message{
			Title:   "Cleanup Gagal",
			Level:   LevelCritical,
			Feature: "cleanup",
			Body:    fmt.Sprintf("<b>Host:</b> <code>%s</code>\n<b>Time:</b> <code>%s</code>\n<b>Ticket:</b> <code>%s</code>\n\n<b>Error:</b>\n<pre><code>%s</code></pre>", escapeHTML(hostname), escapeHTML(timestamp), escapeHTML(ticket), escapeHTML(err.Error())),
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<b>Host:</b> <code>%s</code>\n", escapeHTML(hostname)))
	sb.WriteString(fmt.Sprintf("<b>Time:</b> <code>%s</code>\n", escapeHTML(timestamp)))
	if ticket != "" {
		sb.WriteString(fmt.Sprintf("<b>Ticket:</b> <code>%s</code>\n", escapeHTML(ticket)))
	}
	sb.WriteString(fmt.Sprintf("Penghapusan file backup lama (Mode: <code>%s</code>) telah selesai dieksekusi tanpa error.\n\n", escapeHTML(mode)))
	
	if res != nil {
		sb.WriteString(fmt.Sprintf("<b>File Dihapus:</b> <code>%d</code>\n", res.DeletedCount))
		sb.WriteString(fmt.Sprintf("<b>Ruang Dibebaskan:</b> <code>%s</code>\n", escapeHTML(text.FormatFileSize(res.TotalFreedSize))))
	}

	return Message{
		Title:   "Cleanup Selesai",
		Level:   LevelInfo,
		Feature: "cleanup",
		Body:    sb.String(),
	}
}
