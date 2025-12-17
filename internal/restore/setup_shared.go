// File : internal/restore/setup_shared.go
// Deskripsi : Shared setup functions untuk restore operations
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-17
// Last Modified : 2025-12-17

package restore

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/helper"
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/profilehelper"
	"sfDBTools/pkg/ui"
	"strings"
	"time"
)

// resolveBackupFile resolve lokasi file backup
func (s *Service) resolveBackupFile(filePath *string) error {
	if *filePath == "" {
		defaultDir := s.Config.Backup.Output.BaseDirectory
		if defaultDir == "" {
			defaultDir = "."
		}

		validExtensions := []string{
			".sql",
			".sql.gz", ".sql.gz.enc",
			".sql.xz", ".sql.xz.enc",
			".sql.zst", ".sql.zst.enc",
			".sql.zlib", ".sql.zlib.enc",
			".enc",
		}

		selectedFile, err := input.SelectFileInteractive(defaultDir, "Masukkan path directory atau file backup", validExtensions)
		if err != nil {
			return fmt.Errorf("gagal memilih file backup: %w", err)
		}
		*filePath = selectedFile
	}

	if _, err := os.Stat(*filePath); os.IsNotExist(err) {
		return fmt.Errorf("file backup tidak ditemukan: %s", *filePath)
	}

	absPath, err := filepath.Abs(*filePath)
	if err != nil {
		return fmt.Errorf("gagal mendapatkan absolute path: %w", err)
	}
	*filePath = absPath

	s.Log.Infof("File backup: %s", *filePath)
	return nil
}

// resolveEncryptionKey resolve encryption key untuk decrypt file
func (s *Service) resolveEncryptionKey(filePath string, encryptionKey *string) error {
	isEncrypted := strings.HasSuffix(strings.ToLower(filePath), ".enc")

	if !isEncrypted {
		s.Log.Debug("File backup tidak terenkripsi")
		return nil
	}

	if *encryptionKey == "" {
		key, err := input.PromptPassword("Masukkan encryption key untuk decrypt file backup")
		if err != nil {
			return fmt.Errorf("gagal mendapatkan encryption key: %w", err)
		}
		*encryptionKey = key
	}

	s.Log.Debug("Encryption key berhasil di-resolve")
	return nil
}

// resolveTargetProfile resolve profile database target
func (s *Service) resolveTargetProfile(profileInfo *types.ProfileInfo) error {
	configDir := s.Config.ConfigDir.DatabaseProfile

	loadedProfile, err := profilehelper.ResolveAndLoadProfile(profilehelper.ProfileLoadOptions{
		ConfigDir:         configDir,
		ProfilePath:       profileInfo.Path,
		ProfileKey:        profileInfo.EncryptionKey,
		EnvProfilePath:    consts.ENV_TARGET_PROFILE,
		EnvProfileKey:     consts.ENV_TARGET_PROFILE_KEY,
		RequireProfile:    true,
		ProfilePurpose:    "target",
		AllowInteractive:  true,
		InteractivePrompt: "Pilih target profile untuk restore:",
	})
	if err != nil {
		return fmt.Errorf("gagal load profile: %w", err)
	}

	*profileInfo = *loadedProfile
	s.Profile = loadedProfile
	s.Log.Infof("Target profile: %s (%s:%d)",
		filepath.Base(profileInfo.Path),
		s.Profile.DBInfo.Host,
		s.Profile.DBInfo.Port)

	return nil
}

// connectToTargetDatabase membuat koneksi ke database target
func (s *Service) connectToTargetDatabase(ctx context.Context) error {
	s.Log.Info("Menghubungkan ke database target...")

	cfg := database.Config{
		Host:                 s.Profile.DBInfo.Host,
		Port:                 s.Profile.DBInfo.Port,
		User:                 s.Profile.DBInfo.User,
		Password:             s.Profile.DBInfo.Password,
		AllowNativePasswords: true,
		ParseTime:            true,
		Database:             "",
	}

	client, err := database.NewClient(ctx, cfg, 10*time.Second, 10, 5, 5*time.Minute)
	if err != nil {
		return fmt.Errorf("koneksi database target gagal: %w", err)
	}

	s.TargetClient = client
	s.Log.Info("Koneksi ke database target berhasil")

	return nil
}

// resolveTicketNumber resolve ticket number
func (s *Service) resolveTicketNumber(ticket *string) error {
	if *ticket == "" {
		result, err := input.AskTicket("restore")
		if err != nil {
			return fmt.Errorf("gagal mendapatkan ticket number: %w", err)
		}
		*ticket = result
	}

	s.Log.Infof("Ticket number: %s", *ticket)
	return nil
}

// resolveGrantsFile resolve lokasi file user grants
func (s *Service) resolveGrantsFile(skipGrants bool, grantsFile *string, backupFile string) error {
	if skipGrants {
		s.Log.Info("Skip restore user grants")
		return nil
	}

	if *grantsFile != "" {
		if _, err := os.Stat(*grantsFile); os.IsNotExist(err) {
			return fmt.Errorf("file grants tidak ditemukan: %s", *grantsFile)
		}
		s.Log.Infof("File grants: %s", *grantsFile)
		return nil
	}

	// Auto-detect
	autoGrantsFile := helper.AutoDetectGrantsFile(backupFile)
	if autoGrantsFile != "" {
		*grantsFile = autoGrantsFile
		s.Log.Infof("✓ Grants file ditemukan: %s", filepath.Base(autoGrantsFile))
		ui.PrintSuccess(fmt.Sprintf("✓ Grants file ditemukan: %s", filepath.Base(autoGrantsFile)))
		return nil
	}

	// Not found
	backupDir := filepath.Dir(backupFile)
	matches, err := filepath.Glob(filepath.Join(backupDir, "*_users.sql"))
	if err == nil && len(matches) > 0 {
		s.Log.Infof("Ditemukan %d file user grants di folder: %s", len(matches), backupDir)
		ui.PrintInfo(fmt.Sprintf("📁 Ditemukan %d file user grants di folder yang sama", len(matches)))

		options := []string{
			"⏭️  [ Skip restore user grants ]",
			"📁 [ Browse file grants secara manual ]",
		}
		for _, match := range matches {
			options = append(options, filepath.Base(match))
		}

		selected, err := input.SelectSingleFromList(options, "Pilih file user grants untuk di-restore")
		if err != nil {
			s.Log.Warnf("Gagal memilih file grants: %v, skip restore grants", err)
			return nil
		}

		if selected == "⏭️  [ Skip restore user grants ]" {
			s.Log.Info("User grants tidak akan di-restore (user memilih skip)")
			return nil
		}

		if selected != "📁 [ Browse file grants secara manual ]" {
			for _, match := range matches {
				if filepath.Base(match) == selected {
					*grantsFile = match
					s.Log.Infof("File grants dipilih: %s", match)
					return nil
				}
			}
		}
	}

	// Manual browse
	ui.PrintInfo("💡 File user grants tidak ditemukan atau Anda ingin pilih file lain")
	confirmed, err := input.PromptConfirm("Apakah Anda ingin memilih file user grants secara manual?")
	if err != nil || !confirmed {
		s.Log.Info("Skip restore user grants")
		return nil
	}

	validExtensions := []string{".sql"}
	selectedFile, err := input.SelectFileInteractive(backupDir, "Masukkan path directory atau file user grants", validExtensions)
	if err != nil {
		s.Log.Warnf("Gagal memilih file grants: %v, skip restore grants", err)
		return nil
	}

	*grantsFile = selectedFile
	s.Log.Infof("File grants dipilih secara manual: %s", selectedFile)

	return nil
}

// setupBackupOptions setup opsi backup pre-restore
func (s *Service) setupBackupOptions(backupOpts *types.RestoreBackupOptions, encryptionKey string) {
	if backupOpts.OutputDir == "" {
		defaultDir := s.Config.Backup.Output.BaseDirectory
		if defaultDir == "" {
			defaultDir = "./backups"
		}

		fmt.Println()
		fmt.Println("💾 Backup pre-restore akan dilakukan sebelum restore database")
		fmt.Printf("   Default directory: %s\n", defaultDir)
		fmt.Println()

		backupDir, err := input.AskString("Masukkan direktori untuk backup pre-restore (kosongkan untuk default)", defaultDir, nil)
		if err != nil {
			s.Log.Warnf("Gagal mendapatkan input direktori backup, menggunakan default: %v", err)
			backupDir = defaultDir
		}

		backupDir = strings.TrimSpace(backupDir)
		if backupDir == "" {
			backupDir = defaultDir
		}

		backupOpts.OutputDir = backupDir
		s.Log.Infof("Direktori backup pre-restore: %s", backupDir)
	}

	if !backupOpts.Compression.Enabled {
		backupOpts.Compression = types.CompressionOptions{
			Enabled: s.Config.Backup.Compression.Enabled,
			Type:    s.Config.Backup.Compression.Type,
			Level:   s.Config.Backup.Compression.Level,
		}
	}

	if !backupOpts.Encryption.Enabled {
		backupOpts.Encryption = types.EncryptionOptions{
			Enabled: s.Config.Backup.Encryption.Enabled,
			Key:     encryptionKey,
		}
	}
}


