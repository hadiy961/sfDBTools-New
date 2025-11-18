// File : internal/backup/backup_gtid.go
// Deskripsi : Fungsi untuk capture dan menyimpan informasi GTID
// Author : Hadiyatna Muflihun
// Tanggal : 2024-11-18
// Last Modified : 2024-11-18

package backup

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/helper"
	"sfDBTools/pkg/ui"
	"strings"
	"time"
)

// SaveGTIDToFile menyimpan informasi GTID yang sudah diambil ke file.
// File GTID akan disimpan dengan nama yang sama dengan backup file tetapi dengan suffix _gtid.txt
func (s *Service) SaveGTIDToFile(gtidInfo *database.GTIDInfo, backupFilePath string) error {
	if gtidInfo == nil {
		s.Log.Warn("GTID info is nil, skip saving")
		return nil
	}

	s.Log.Info("Menyimpan informasi GTID ke file...")
	s.Log.Debugf("Backup file path: %s", backupFilePath)
	timer := helper.NewTimer()

	// Generate nama file GTID berdasarkan nama backup file
	gtidFilePath := generateGTIDFilePath(backupFilePath)
	s.Log.Infof("File GTID akan disimpan di: %s", gtidFilePath)

	// Format konten file GTID
	content := formatGTIDContent(gtidInfo)

	// Tulis ke file
	s.Log.Debugf("Menulis konten GTID ke file: %s", gtidFilePath)
	err := os.WriteFile(gtidFilePath, []byte(content), 0644)
	if err != nil {
		s.Log.Errorf("Gagal menulis file GTID: %v", err)
		return fmt.Errorf("gagal menulis file GTID: %w", err)
	}

	duration := timer.Elapsed()
	s.Log.Infof("✓ Informasi GTID berhasil disimpan ke: %s (durasi: %v)", gtidFilePath, duration)

	// Print ke UI juga
	ui.PrintSuccess(fmt.Sprintf("File GTID berhasil dibuat: %s", gtidFilePath))

	return nil
}

// CaptureGTID mengambil informasi GTID dari database dan menyimpannya ke file.
// File GTID akan disimpan dengan nama yang sama dengan backup file tetapi dengan suffix _gtid.txt
// DEPRECATED: Gunakan SaveGTIDToFile setelah mengambil GTID dengan GetFullGTIDInfo
func (s *Service) CaptureGTID(ctx context.Context, backupFilePath string) error {
	if !s.BackupDBOptions.CaptureGTID {
		s.Log.Debug("CaptureGTID: flag tidak diaktifkan, skip capture GTID")
		return nil // Skip jika capture GTID tidak diaktifkan
	}

	s.Log.Info("Memulai capture informasi GTID...")
	s.Log.Debugf("Backup file path: %s", backupFilePath)
	timer := helper.NewTimer()

	// Dapatkan informasi GTID dari database
	s.Log.Debug("Mengambil informasi GTID dari database...")
	gtidInfo, err := s.Client.GetFullGTIDInfo(ctx)
	if err != nil {
		s.Log.Errorf("Gagal mendapatkan informasi GTID: %v", err)
		return fmt.Errorf("gagal mendapatkan informasi GTID: %w", err)
	}
	s.Log.Debugf("GTID Info: File=%s, Pos=%d, GTID=%s", gtidInfo.MasterLogFile, gtidInfo.MasterLogPos, gtidInfo.GTIDBinlog)

	// Generate nama file GTID berdasarkan nama backup file
	gtidFilePath := generateGTIDFilePath(backupFilePath)
	s.Log.Infof("File GTID akan disimpan di: %s", gtidFilePath)

	// Format konten file GTID
	content := formatGTIDContent(gtidInfo)

	// Tulis ke file
	s.Log.Debugf("Menulis konten GTID ke file: %s", gtidFilePath)
	err = os.WriteFile(gtidFilePath, []byte(content), 0644)
	if err != nil {
		s.Log.Errorf("Gagal menulis file GTID: %v", err)
		return fmt.Errorf("gagal menulis file GTID: %w", err)
	}

	duration := timer.Elapsed()
	s.Log.Infof("✓ Informasi GTID berhasil disimpan ke: %s (durasi: %v)", gtidFilePath, duration)

	// Print ke UI juga
	ui.PrintSuccess(fmt.Sprintf("File GTID berhasil dibuat: %s", gtidFilePath))

	return nil
}

// generateGTIDFilePath menghasilkan path file GTID dari path backup file.
// Contoh: all_databases_20251118_094529_localhost.zst -> all_databases_20251118_094529_localhost_gtid.txt
func generateGTIDFilePath(backupFilePath string) string {
	// Dapatkan direktori dan nama file
	dir := filepath.Dir(backupFilePath)
	filename := filepath.Base(backupFilePath)

	// Hapus ekstensi dari nama file
	// Bisa ada multiple extensions seperti .sql.gz atau .sql.zst.enc
	baseFilename := filename
	for {
		ext := filepath.Ext(baseFilename)
		if ext == "" {
			break
		}
		baseFilename = strings.TrimSuffix(baseFilename, ext)
	}

	// Tambahkan suffix _gtid.txt
	gtidFilename := baseFilename + "_gtid.txt"

	return filepath.Join(dir, gtidFilename)
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
