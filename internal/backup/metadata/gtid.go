// File : internal/backup/metadata/gtid.go
// Deskripsi : Fungsi untuk capture dan menyimpan informasi GTID
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-05
// Last Modified : 2025-12-05

package metadata

import (
	"context"
	"fmt"
	"os"
	"sfDBTools/internal/applog"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/helper"
	"sfDBTools/pkg/ui"
	"strings"
	"time"

	backuphelper "sfDBTools/internal/backup/helper"
)

// SaveGTIDToFile menyimpan informasi GTID yang sudah diambil ke file.
// File GTID akan disimpan dengan nama yang sama dengan backup file tetapi dengan suffix _gtid.txt
func SaveGTIDToFile(logger applog.Logger, gtidInfo *database.GTIDInfo, backupFilePath string) error {
	if gtidInfo == nil {
		logger.Warn("GTID info is nil, skip saving")
		return nil
	}

	logger.Info("Menyimpan informasi GTID ke file...")
	logger.Debugf("Backup file path: %s", backupFilePath)
	timer := helper.NewTimer()

	// Generate nama file GTID berdasarkan nama backup file
	gtidFilePath := backuphelper.GenerateGTIDFilePath(backupFilePath)
	logger.Infof("File GTID akan disimpan di: %s", gtidFilePath)

	// Format konten file GTID
	content := formatGTIDContent(gtidInfo)

	// Tulis ke file
	logger.Debugf("Menulis konten GTID ke file: %s", gtidFilePath)
	err := os.WriteFile(gtidFilePath, []byte(content), 0644)
	if err != nil {
		logger.Errorf("Gagal menulis file GTID: %v", err)
		return fmt.Errorf("gagal menulis file GTID: %w", err)
	}

	duration := timer.Elapsed()
	logger.Infof("✓ Informasi GTID berhasil disimpan ke: %s (durasi: %v)", gtidFilePath, duration)

	// Print ke UI juga
	ui.PrintSuccess(fmt.Sprintf("File GTID berhasil dibuat: %s", gtidFilePath))

	return nil
}

// CaptureAndSaveGTID mengambil informasi GTID dari database dan menyimpannya ke file.
// File GTID akan disimpan dengan nama yang sama dengan backup file tetapi dengan suffix _gtid.txt
func CaptureAndSaveGTID(ctx context.Context, client *database.Client, logger applog.Logger, backupFilePath string, captureGTID bool) error {
	if !captureGTID {
		logger.Debug("CaptureGTID: flag tidak diaktifkan, skip capture GTID")
		return nil // Skip jika capture GTID tidak diaktifkan
	}

	logger.Info("Memulai capture informasi GTID...")
	logger.Debugf("Backup file path: %s", backupFilePath)
	timer := helper.NewTimer()

	// Dapatkan informasi GTID dari database
	logger.Debug("Mengambil informasi GTID dari database...")
	gtidInfo, err := client.GetFullGTIDInfo(ctx)
	if err != nil {
		logger.Errorf("Gagal mendapatkan informasi GTID: %v", err)
		return fmt.Errorf("gagal mendapatkan informasi GTID: %w", err)
	}
	logger.Debugf("GTID Info: File=%s, Pos=%d, GTID=%s", gtidInfo.MasterLogFile, gtidInfo.MasterLogPos, gtidInfo.GTIDBinlog)

	// Generate nama file GTID berdasarkan nama backup file
	gtidFilePath := backuphelper.GenerateGTIDFilePath(backupFilePath)
	logger.Infof("File GTID akan disimpan di: %s", gtidFilePath)

	// Format konten file GTID
	content := formatGTIDContent(gtidInfo)

	// Tulis ke file
	logger.Debugf("Menulis konten GTID ke file: %s", gtidFilePath)
	err = os.WriteFile(gtidFilePath, []byte(content), 0644)
	if err != nil {
		logger.Errorf("Gagal menulis file GTID: %v", err)
		return fmt.Errorf("gagal menulis file GTID: %w", err)
	}

	duration := timer.Elapsed()
	logger.Infof("✓ Informasi GTID berhasil disimpan ke: %s (durasi: %v)", gtidFilePath, duration)

	// Print ke UI juga
	ui.PrintSuccess(fmt.Sprintf("File GTID berhasil dibuat: %s", gtidFilePath))

	return nil
}

// formatGTIDContent memformat informasi GTID ke dalam string untuk disimpan ke file.
func formatGTIDContent(info *database.GTIDInfo) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	var builder strings.Builder
	builder.WriteString("# GTID Information\n")
	builder.WriteString(fmt.Sprintf("# Generated at: %s\n", timestamp))
	builder.WriteString("\n")
	builder.WriteString(fmt.Sprintf("MASTER_LOG_FILE = %s\n", info.MasterLogFile))
	builder.WriteString(fmt.Sprintf("MASTER_LOG_POS = %d\n", info.MasterLogPos))

	if info.GTIDBinlog != "" {
		builder.WriteString(fmt.Sprintf("gtid_binlog = %s\n", info.GTIDBinlog))
	} else {
		builder.WriteString("gtid_binlog = (not available)\n")
	}

	return builder.String()
}
