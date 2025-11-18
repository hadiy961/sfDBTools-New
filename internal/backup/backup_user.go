// File : internal/backup/backup_user.go
// Deskripsi : Fungsi untuk export dan menyimpan user grants
// Author : Hadiyatna Muflihun
// Tanggal : 2024-11-18
// Last Modified : 2024-11-18

package backup

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sfDBTools/pkg/helper"
	"sfDBTools/pkg/ui"
	"strings"
)

// ExportAndSaveUserGrants mengambil user grants dari database dan menyimpannya ke file.
// File akan disimpan dengan nama yang sama dengan backup file tetapi dengan suffix _user.sql
func (s *Service) ExportAndSaveUserGrants(ctx context.Context, backupFilePath string) error {
	// Jika ExcludeUser = true, skip export
	if s.BackupDBOptions.ExcludeUser {
		s.Log.Debug("ExcludeUser: flag diaktifkan, skip export user grants")
		return nil
	}

	s.Log.Info("Memulai export user grants...")
	timer := helper.NewTimer()

	// Export user grants dari database
	s.Log.Debug("Mengambil user grants dari database...")
	userGrantsSQL, err := s.Client.ExportAllUserGrants(ctx)
	if err != nil {
		s.Log.Errorf("Gagal mendapatkan user grants: %v", err)
		return fmt.Errorf("gagal mendapatkan user grants: %w", err)
	}

	// Generate nama file user grants berdasarkan nama backup file
	userFilePath := generateUserFilePath(backupFilePath)
	s.Log.Infof("File user grants akan disimpan di: %s", userFilePath)

	// Tulis ke file
	s.Log.Debugf("Menulis user grants ke file: %s", userFilePath)
	err = os.WriteFile(userFilePath, []byte(userGrantsSQL), 0644)
	if err != nil {
		s.Log.Errorf("Gagal menulis file user grants: %v", err)
		return fmt.Errorf("gagal menulis file user grants: %w", err)
	}

	duration := timer.Elapsed()
	s.Log.Infof("✓ User grants berhasil disimpan ke: %s (durasi: %v)", userFilePath, duration)

	// Print ke UI juga
	ui.PrintSuccess(fmt.Sprintf("File user grants berhasil dibuat: %s", userFilePath))

	return nil
}

// generateUserFilePath menghasilkan path file user grants dari path backup file.
// Contoh: all_databases_20251118_094529_localhost.zst -> all_databases_20251118_094529_localhost_user.sql
func generateUserFilePath(backupFilePath string) string {
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

	// Tambahkan suffix _user.sql
	userFilename := baseFilename + "_user.sql"

	return filepath.Join(dir, userFilename)
}
