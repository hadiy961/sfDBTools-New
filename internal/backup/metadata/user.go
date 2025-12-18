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

	backuphelper "sfDBTools/internal/backup/filehelper"
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

	// Tulis ke file
	logger.Debugf("Menulis user grants ke file: %s", userFilePath)
	err = os.WriteFile(userFilePath, []byte(userGrantsSQL), 0644)
	if err != nil {
		logger.Errorf("Gagal menulis file user grants: %v", err)
		return "", fmt.Errorf("gagal menulis file user grants: %w", err)
	}

	duration := timer.Elapsed()
	logger.Infof("✓ User grants berhasil disimpan ke: %s (durasi: %v)", userFilePath, duration)

	return userFilePath, nil
}

// ExportUserGrantsIfNeededWithLogging adalah wrapper untuk export user grants dengan logging
// Return: path file yang berhasil dibuat, atau empty string jika gagal/tidak ada user
func ExportUserGrantsIfNeededWithLogging(ctx context.Context, client *database.Client, logger applog.Logger, referenceBackupFile string, excludeUser bool, databases []string) string {
	if excludeUser {
		return ""
	}

	if referenceBackupFile == "" {
		logger.Warn("Tidak ada backup file untuk export user grants")
		return ""
	}

	// Ping database connection untuk memastikan masih valid
	// Jika koneksi invalid setelah backup lama, ini akan memicu reconnect otomatis
	if err := client.Ping(ctx); err != nil {
		logger.Warnf("Koneksi database tidak valid, mencoba reconnect: %v", err)
		// Ping akan trigger reconnect secara otomatis melalui connection pool
		// Coba ping sekali lagi setelah jeda singkat
		// time.Sleep(100 * time.Millisecond)
		if err := client.Ping(ctx); err != nil {
			logger.Errorf("Gagal reconnect ke database: %v", err)
			return ""
		}
		logger.Info("Reconnect ke database berhasil")
	}

	logger.Info("Export user grants ke file...")
	filePath, err := ExportAndSaveUserGrants(ctx, client, logger, referenceBackupFile, excludeUser, databases)
	if err != nil {
		logger.Errorf("Gagal export user grants: %v", err)
		return ""
	}
	return filePath
}
