// File : internal/backup/display/result_display.go
// Deskripsi : Display logic untuk backup results
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-05
// Last Modified : 2025-12-05

package display

import (
	"fmt"
	"sfdbtools/internal/app/backup/model/types_backup"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/ui/print"
	"sfdbtools/internal/ui/table"
	"sfdbtools/internal/ui/text"
)

// ResultDisplayer handles display of backup results
type ResultDisplayer struct {
	result             *types_backup.BackupResult
	compressionEnabled bool
	compressionType    string
	encryptionEnabled  bool
}

// NewResultDisplayer creates new result displayer
func NewResultDisplayer(result *types_backup.BackupResult, compressionEnabled bool, compressionType string, encryptionEnabled bool) *ResultDisplayer {
	return &ResultDisplayer{
		result:             result,
		compressionEnabled: compressionEnabled,
		compressionType:    compressionType,
		encryptionEnabled:  encryptionEnabled,
	}
}

// Display menampilkan hasil backup
func (d *ResultDisplayer) Display() {
	print.PrintSubHeader("Hasil Backup Database")

	d.displaySummary()
	d.displaySuccessDetails()
	d.displayFailures()
}

// displaySummary menampilkan ringkasan statistik
func (d *ResultDisplayer) displaySummary() {
	data := [][]string{
		{"Total Database Ditemukan", fmt.Sprintf("%d", d.result.TotalDatabases)},
		{"Total Database Dibackup", text.ColorText(fmt.Sprintf("%d", d.result.SuccessfulBackups), consts.UIColorGreen)},
		{"Total Gagal Dibackup", text.ColorText(fmt.Sprintf("%d", d.result.FailedBackups), consts.UIColorRed)},
		{"Total Waktu Proses", text.ColorText(d.result.TotalTimeTaken.String(), consts.UIColorCyan)},
	}
	table.Render([]string{"Kategori", "Jumlah"}, data)
}

// displaySuccessDetails menampilkan detail backup yang berhasil
func (d *ResultDisplayer) displaySuccessDetails() {
	if len(d.result.BackupInfo) == 0 {
		return
	}

	print.PrintSubHeader("Detail Backup yang Berhasil")

	for _, info := range d.result.BackupInfo {
		fmt.Println()
		d.displayBackupInfo(info)
	}
}

// displayBackupInfo menampilkan detail satu backup info
func (d *ResultDisplayer) displayBackupInfo(info types_backup.DatabaseBackupInfo) {
	data := [][]string{
		{"Database", text.ColorText(info.DatabaseName, consts.UIColorCyan)},
		{"Status", d.formatStatus(info.Status)},
		{"File Output", info.OutputFile},
		{"Ukuran File", info.FileSizeHuman},
		{"Durasi Backup", info.Duration},
	}

	// Tambahkan metadata jika tersedia
	data = append(data, d.buildMetadataRows(info)...)

	// Tambahkan info kompresi dan enkripsi
	data = append(data, d.buildCompressionRow())
	data = append(data, d.buildEncryptionRow())

	table.Render([]string{"Parameter", "Nilai"}, data)

	// Tampilkan warning jika ada
	if info.Warnings != "" {
		print.PrintWarn("\n⚠ Warning dari mysqldump:")
		fmt.Println(text.ColorText(info.Warnings, consts.UIColorYellow))
	}
}

// buildMetadataRows builds metadata rows jika tersedia
func (d *ResultDisplayer) buildMetadataRows(info types_backup.DatabaseBackupInfo) [][]string {
	rows := [][]string{}

	if info.BackupID != "" {
		rows = append(rows, []string{"Backup ID", text.ColorText(info.BackupID, consts.UIColorYellow)})
	}
	if !info.StartTime.IsZero() {
		rows = append(rows, []string{"Start Time", text.ColorText(info.StartTime.String(), consts.UIColorCyan)})
	}
	if !info.EndTime.IsZero() {
		rows = append(rows, []string{"End Time", text.ColorText(info.EndTime.String(), consts.UIColorCyan)})
	}
	if info.ThroughputMBps > 0 {
		rows = append(rows, []string{"Throughput", text.ColorText(fmt.Sprintf("%.2f MB/s", info.ThroughputMBps), consts.UIColorGreen)})
	}
	if info.ManifestFile != "" {
		rows = append(rows, []string{"Manifest", text.ColorText(info.ManifestFile, consts.UIColorPurple)})
	}

	return rows
}

// buildCompressionRow builds compression info row
func (d *ResultDisplayer) buildCompressionRow() []string {
	if d.compressionEnabled {
		return []string{"Kompresi", text.ColorText("Enabled ("+d.compressionType+")", consts.UIColorGreen)}
	}
	return []string{"Kompresi", text.ColorText("Disabled", consts.UIColorYellow)}
}

// buildEncryptionRow builds encryption info row
func (d *ResultDisplayer) buildEncryptionRow() []string {
	if d.encryptionEnabled {
		return []string{"Enkripsi", text.ColorText("Enabled", consts.UIColorGreen)}
	}
	return []string{"Enkripsi", text.ColorText("Disabled", consts.UIColorYellow)}
}

// displayFailures menampilkan daftar database yang gagal
func (d *ResultDisplayer) displayFailures() {
	if d.result.FailedBackups == 0 {
		return
	}

	print.PrintSubHeader("Daftar Database Gagal Dibackup")

	if len(d.result.FailedDatabaseInfos) > 0 {
		for i, failedInfo := range d.result.FailedDatabaseInfos {
			fmt.Printf("%d. %s\n", i+1, text.ColorText(failedInfo.DatabaseName, consts.UIColorRed))
			fmt.Printf("   Error: %s\n", failedInfo.Error)
			fmt.Println()
		}
	} else if len(d.result.FailedDatabases) > 0 {
		for dbName, errMsg := range d.result.FailedDatabases {
			print.PrintError("- " + dbName + ": " + errMsg)
		}
	}
}

// formatStatus formats backup status dengan warna
func (d *ResultDisplayer) formatStatus(status string) string {
	if status == consts.BackupStatusSuccessWithWarnings {
		return text.ColorText("Sukses dengan Warning", consts.UIColorYellow)
	}
	return text.ColorText(status, consts.UIColorGreen)
}
