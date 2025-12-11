// File : internal/backup/metadata/user.go
// Deskripsi : Fungsi untuk export dan menyimpan user grants
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

	backuphelper "sfDBTools/internal/backup/helper"
)

// ExportAndSaveUserGrants mengambil user grants dari database dan menyimpannya ke file.
// File akan disimpan dengan nama yang sama dengan backup file tetapi dengan suffix _user.sql
// Jika databases tidak kosong, hanya export user yang memiliki grants ke database tersebut.
// Jika databases kosong, export semua user grants.
// Return: path file yang berhasil dibuat, atau empty string jika gagal/tidak ada user
func ExportAndSaveUserGrants(ctx context.Context, client *database.Client, logger applog.Logger, backupFilePath string, excludeUser bool, databases []string) (string, error) {
	// Jika ExcludeUser = true, skip export
	if excludeUser {
		logger.Debug("ExcludeUser: flag diaktifkan, skip export user grants")
		return "", nil
	}

	logger.Info("Memulai export user grants...")
	timer := helper.NewTimer()

	// Export user grants dari database
	var userGrantsSQL string
	var err error

	if len(databases) == 0 {
		// Export semua user grants (mode all)
		logger.Debug("Mengambil semua user grants dari database...")
		userGrantsSQL, err = client.ExportAllUserGrants(ctx)
	} else {
		// Export hanya user dengan grants ke database tertentu (mode filter, primary, secondary)
		logger.Debugf("Mengambil user grants untuk database: %v", databases)
		userGrantsSQL, err = client.ExportUserGrantsForDatabases(ctx, databases)
	}

	if err != nil {
		logger.Errorf("Gagal mendapatkan user grants: %v", err)
		return "", fmt.Errorf("gagal mendapatkan user grants: %w", err)
	}

	// Generate nama file user grants berdasarkan nama backup file
	userFilePath := backuphelper.GenerateUserFilePath(backupFilePath)
	logger.Infof("File user grants akan disimpan di: %s", userFilePath)

	// Tulis ke file
	logger.Debugf("Menulis user grants ke file: %s", userFilePath)
	err = os.WriteFile(userFilePath, []byte(userGrantsSQL), 0644)
	if err != nil {
		logger.Errorf("Gagal menulis file user grants: %v", err)
		return "", fmt.Errorf("gagal menulis file user grants: %w", err)
	}

	duration := timer.Elapsed()
	logger.Infof("✓ User grants berhasil disimpan ke: %s (durasi: %v)", userFilePath, duration)

	// Print ke UI juga
	ui.PrintSuccess(fmt.Sprintf("File user grants berhasil dibuat: %s", userFilePath))

	return userFilePath, nil
}
