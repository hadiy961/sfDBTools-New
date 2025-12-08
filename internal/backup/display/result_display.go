// File : internal/backup/display/result_display.go
// Deskripsi : Display logic untuk backup results
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-05
// Last Modified : 2025-12-05

package display

import (
	"fmt"
	"sfDBTools/internal/types"
	"sfDBTools/internal/types/types_backup"
	"sfDBTools/pkg/ui"
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
	ui.PrintSubHeader("Hasil Backup Database")

	d.displaySummary()
	d.displaySuccessDetails()
	d.displayFailures()
}

// displaySummary menampilkan ringkasan statistik
func (d *ResultDisplayer) displaySummary() {
	data := [][]string{
		{"Total Database Ditemukan", fmt.Sprintf("%d", d.result.TotalDatabases)},
		{"Total Database Dibackup", ui.ColorText(fmt.Sprintf("%d", d.result.SuccessfulBackups), ui.ColorGreen)},
		{"Total Gagal Dibackup", ui.ColorText(fmt.Sprintf("%d", d.result.FailedBackups), ui.ColorRed)},
		{"Total Waktu Proses", ui.ColorText(d.result.TotalTimeTaken.String(), ui.ColorCyan)},
	}
	ui.FormatTable([]string{"Kategori", "Jumlah"}, data)
}

// displaySuccessDetails menampilkan detail backup yang berhasil
func (d *ResultDisplayer) displaySuccessDetails() {
	if len(d.result.BackupInfo) == 0 {
		return
	}

	ui.PrintSubHeader("Detail Backup yang Berhasil")

	for _, info := range d.result.BackupInfo {
		fmt.Println()
		d.displayBackupInfo(info)
	}
}

// displayBackupInfo menampilkan detail satu backup info
func (d *ResultDisplayer) displayBackupInfo(info types.DatabaseBackupInfo) {
	data := [][]string{
		{"Database", ui.ColorText(info.DatabaseName, ui.ColorCyan)},
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

	ui.FormatTable([]string{"Parameter", "Nilai"}, data)

	// Tampilkan warning jika ada
	if info.Warnings != "" {
		ui.PrintWarning("\n⚠ Warning dari mysqldump:")
		fmt.Println(ui.ColorText(info.Warnings, ui.ColorYellow))
	}
}

// buildMetadataRows builds metadata rows jika tersedia
func (d *ResultDisplayer) buildMetadataRows(info types.DatabaseBackupInfo) [][]string {
	rows := [][]string{}

	if info.BackupID != "" {
		rows = append(rows, []string{"Backup ID", ui.ColorText(info.BackupID, ui.ColorYellow)})
	}
	if !info.StartTime.IsZero() {
		rows = append(rows, []string{"Start Time", ui.ColorText(info.StartTime.String(), ui.ColorCyan)})
	}
	if !info.EndTime.IsZero() {
		rows = append(rows, []string{"End Time", ui.ColorText(info.EndTime.String(), ui.ColorCyan)})
	}
	if info.ThroughputMBps > 0 {
		rows = append(rows, []string{"Throughput", ui.ColorText(fmt.Sprintf("%.2f MB/s", info.ThroughputMBps), ui.ColorGreen)})
	}
	if info.ManifestFile != "" {
		rows = append(rows, []string{"Manifest", ui.ColorText(info.ManifestFile, ui.ColorPurple)})
	}

	return rows
}

// buildCompressionRow builds compression info row
func (d *ResultDisplayer) buildCompressionRow() []string {
	if d.compressionEnabled {
		return []string{"Kompresi", ui.ColorText("Enabled ("+d.compressionType+")", ui.ColorGreen)}
	}
	return []string{"Kompresi", ui.ColorText("Disabled", ui.ColorYellow)}
}

// buildEncryptionRow builds encryption info row
func (d *ResultDisplayer) buildEncryptionRow() []string {
	if d.encryptionEnabled {
		return []string{"Enkripsi", ui.ColorText("Enabled", ui.ColorGreen)}
	}
	return []string{"Enkripsi", ui.ColorText("Disabled", ui.ColorYellow)}
}

// displayFailures menampilkan daftar database yang gagal
func (d *ResultDisplayer) displayFailures() {
	if d.result.FailedBackups == 0 {
		return
	}

	ui.PrintSubHeader("Daftar Database Gagal Dibackup")

	if len(d.result.FailedDatabaseInfos) > 0 {
		for i, failedInfo := range d.result.FailedDatabaseInfos {
			fmt.Printf("%d. %s\n", i+1, ui.ColorText(failedInfo.DatabaseName, ui.ColorRed))
			fmt.Printf("   Error: %s\n", failedInfo.Error)
			fmt.Println()
		}
	} else if len(d.result.FailedDatabases) > 0 {
		for dbName, errMsg := range d.result.FailedDatabases {
			ui.PrintError("- " + dbName + ": " + errMsg)
		}
	}
}

// formatStatus formats backup status dengan warna
func (d *ResultDisplayer) formatStatus(status string) string {
	if status == "success_with_warnings" {
		return ui.ColorText("Sukses dengan Warning", ui.ColorYellow)
	}
	return ui.ColorText(status, ui.ColorGreen)
}
