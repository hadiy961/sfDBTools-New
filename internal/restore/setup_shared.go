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
	profilehelper "sfDBTools/pkg/helper/profile"
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/ui"
	"strings"
	"time"
)

// resolveBackupFile resolve lokasi file backup
func (s *Service) resolveBackupFile(filePath *string, allowInteractive bool) error {
	if strings.TrimSpace(*filePath) == "" {
		if !allowInteractive {
			return fmt.Errorf("file backup wajib diisi (--file) pada mode non-interaktif (--force)")
		}
		defaultDir := s.Config.Backup.Output.BaseDirectory
		if defaultDir == "" {
			defaultDir = "."
		}

		validExtensions := helper.ValidBackupFileExtensionsForSelection()

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

// resolveSelectionCSV resolve lokasi file CSV untuk restore selection.
// Jika kosong dan allowInteractive=true, user akan diprompt memilih file .csv.
func (s *Service) resolveSelectionCSV(csvPath *string, allowInteractive bool) error {
	if csvPath == nil {
		return fmt.Errorf("path CSV tidak tersedia")
	}

	if strings.TrimSpace(*csvPath) == "" {
		if !allowInteractive {
			return fmt.Errorf("path CSV wajib diisi (--csv) pada mode non-interaktif (--force)")
		}

		defaultDir := s.Config.Backup.Output.BaseDirectory
		if defaultDir == "" {
			defaultDir = "."
		}

		selectedFile, err := input.SelectFileInteractive(defaultDir, "Masukkan path CSV selection", []string{".csv"})
		if err != nil {
			return fmt.Errorf("gagal memilih file CSV: %w", err)
		}
		*csvPath = selectedFile
	}

	if _, err := os.Stat(*csvPath); os.IsNotExist(err) {
		return fmt.Errorf("file CSV tidak ditemukan: %s", *csvPath)
	}

	absPath, err := filepath.Abs(*csvPath)
	if err != nil {
		return fmt.Errorf("gagal mendapatkan absolute path: %w", err)
	}
	*csvPath = absPath

	s.Log.Infof("CSV file: %s", *csvPath)
	return nil
}

// resolveEncryptionKey resolve encryption key untuk decrypt file
func (s *Service) resolveEncryptionKey(filePath string, encryptionKey *string, allowInteractive bool) error {
	isEncrypted := helper.IsEncryptedFile(filePath)

	if !isEncrypted {
		s.Log.Debug("File backup tidak terenkripsi")
		return nil
	}

	if *encryptionKey == "" {
		if !allowInteractive {
			return fmt.Errorf("file backup terenkripsi; encryption key wajib diisi (--enc-key atau env) pada mode non-interaktif (--force)")
		}
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
func (s *Service) resolveTargetProfile(profileInfo *types.ProfileInfo, allowInteractive bool) error {
	configDir := s.Config.ConfigDir.DatabaseProfile

	loadedProfile, err := profilehelper.ResolveAndLoadProfile(profilehelper.ProfileLoadOptions{
		ConfigDir:         configDir,
		ProfilePath:       profileInfo.Path,
		ProfileKey:        profileInfo.EncryptionKey,
		EnvProfilePath:    consts.ENV_TARGET_PROFILE,
		EnvProfileKey:     consts.ENV_TARGET_PROFILE_KEY,
		RequireProfile:    true,
		ProfilePurpose:    "target",
		AllowInteractive:  allowInteractive,
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

	// Konsisten dengan fitur backup: ambil hostname dari server (SELECT @@hostname)
	// agar penamaan file memakai hostname (mis. dev12-dbdev-125) bukan IP.
	serverHostname, err := client.GetServerHostname(ctx)
	if err != nil {
		s.Log.Warnf("gagal mendapatkan hostname dari server: %v, menggunakan dari config", err)
		if s.Profile != nil && s.Profile.DBInfo.HostName == "" {
			s.Profile.DBInfo.HostName = s.Profile.DBInfo.Host
		}
	} else {
		if s.Profile != nil {
			s.Profile.DBInfo.HostName = serverHostname
		}
		s.Log.Infof("menggunakan hostname dari server: %s", serverHostname)
	}
	s.Log.Info("Koneksi ke database target berhasil")

	return nil
}

// resolveTicketNumber resolve ticket number
func (s *Service) resolveTicketNumber(ticket *string, allowInteractive bool) error {
	if strings.TrimSpace(*ticket) == "" {
		if !allowInteractive {
			return fmt.Errorf("ticket number wajib diisi (--ticket) pada mode non-interaktif (--force)")
		}
		result, err := input.AskTicket(consts.FeatureRestore)
		if err != nil {
			return fmt.Errorf("gagal mendapatkan ticket number: %w", err)
		}
		*ticket = result
	}

	s.Log.Infof("Ticket number: %s", *ticket)
	return nil
}

// resolveInteractiveSafetyOptions memberikan opsi interaktif untuk backup pre-restore dan drop target.
// Hanya aktif jika allowInteractive=true (tanpa --force).
func (s *Service) resolveInteractiveSafetyOptions(dropTarget *bool, skipBackup *bool, allowInteractive bool) error {
	if !allowInteractive {
		return nil
	}

	// Backup pre-restore
	backupDefault := true
	if skipBackup != nil {
		backupDefault = !*skipBackup
	}
	shouldBackup, err := input.AskYesNo("Lakukan backup sebelum restore?", backupDefault)
	if err != nil {
		return fmt.Errorf("gagal mendapatkan pilihan backup pre-restore: %w", err)
	}
	if skipBackup != nil {
		*skipBackup = !shouldBackup
	}

	// Drop target
	dropDefault := true
	if dropTarget != nil {
		dropDefault = *dropTarget
	}
	shouldDrop, err := input.AskYesNo("Drop target database sebelum restore?", dropDefault)
	if err != nil {
		return fmt.Errorf("gagal mendapatkan pilihan drop target: %w", err)
	}
	if dropTarget != nil {
		*dropTarget = shouldDrop
	}

	var dtVal interface{} = "<nil>"
	var sbVal interface{} = "<nil>"
	if dropTarget != nil {
		dtVal = *dropTarget
	}
	if skipBackup != nil {
		sbVal = *skipBackup
	}
	s.Log.Infof("Pilihan interaktif: drop-target=%v, skip-backup=%v", dtVal, sbVal)
	return nil
}

// resolveGrantsFile resolve lokasi file user grants
func (s *Service) resolveGrantsFile(skipGrants *bool, grantsFile *string, backupFile string, allowInteractive bool, stopOnError bool) error {
	if skipGrants != nil && *skipGrants {
		s.Log.Info("Skip restore user grants (--skip-grants)")
		return nil
	}

	if strings.TrimSpace(backupFile) == "" {
		s.Log.Info("Grants file: source file kosong, skip pencarian grants")
		return nil
	}

	s.Log.Infof("Mencari file user grants untuk source: %s", filepath.Base(backupFile))

	if *grantsFile != "" {
		if _, err := os.Stat(*grantsFile); os.IsNotExist(err) {
			missing := *grantsFile
			s.Log.Warnf("File grants dari flag tidak ditemukan: %s", missing)
			ui.PrintWarning(fmt.Sprintf("⚠️  File grants tidak ditemukan: %s", missing))

			if !allowInteractive {
				if stopOnError {
					return fmt.Errorf("file grants tidak ditemukan: %s", missing)
				}
				s.Log.Warn("Mode non-interaktif: skip restore user grants (file grants flag invalid)")
				return nil
			}

			// interactive fallback
			*grantsFile = ""
		} else {
			s.Log.Infof("File grants (user-specified): %s", *grantsFile)
			return nil
		}
	}

	// Auto-detect
	expected := helper.GenerateGrantsFilename(filepath.Base(backupFile))
	s.Log.Infof("Auto-detect grants rule: cari file '%s' di folder yang sama (%s)", expected, filepath.Dir(backupFile))
	autoGrantsFile := helper.AutoDetectGrantsFile(backupFile)
	if autoGrantsFile != "" {
		s.Log.Infof("✓ Grants file ditemukan: %s", filepath.Base(autoGrantsFile))

		if !allowInteractive {
			*grantsFile = autoGrantsFile
			return nil
		}

		options := []string{
			fmt.Sprintf("✅ [ Pakai grants file terdeteksi: %s ]", filepath.Base(autoGrantsFile)),
			"📁 [ Browse / pilih file grants lain ]",
			"⏭️  [ Skip restore user grants ]",
		}
		selected, err := input.SelectSingleFromList(options, "Grants file ditemukan. Gunakan file ini atau pilih yang lain?")
		if err == nil && selected == options[0] {
			*grantsFile = autoGrantsFile
			s.Log.Infof("Grants file dipakai (auto-detect): %s", filepath.Base(autoGrantsFile))
			return nil
		}
		if err == nil && selected == options[2] {
			if skipGrants != nil {
				*skipGrants = true
			}
			*grantsFile = ""
			s.Log.Info("User grants tidak akan di-restore (user memilih skip)")
			return nil
		}

		// User memilih browse / atau prompt gagal: lanjut ke flow interaktif (list/browse)
		*grantsFile = ""
	}

	if !allowInteractive {
		s.Log.Info("Mode non-interaktif: user grants tidak ditemukan, skip restore user grants")
		return nil
	}

	// Not found
	backupDir := filepath.Dir(backupFile)
	s.Log.Infof("Grants file tidak ditemukan via auto-detect; mencari di folder: %s", backupDir)
	matches, err := filepath.Glob(filepath.Join(backupDir, "*"+consts.UsersSQLSuffix))
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
			if skipGrants != nil {
				*skipGrants = true
			}
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
	s.Log.Infof("Tidak ada file user grants terdeteksi di folder: %s", backupDir)
	ui.PrintInfo("💡 File user grants tidak ditemukan atau Anda ingin pilih file lain")
	confirmed, err := input.PromptConfirm("Apakah Anda ingin memilih file user grants secara manual?")
	if err != nil || !confirmed {
		if skipGrants != nil {
			*skipGrants = true
		}
		s.Log.Info("Skip restore user grants (tidak ada file grants)")
		return nil
	}

	validExtensions := []string{consts.ExtSQL}
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
func (s *Service) setupBackupOptions(backupOpts *types.RestoreBackupOptions, encryptionKey string, allowInteractive bool) {
	if backupOpts.OutputDir == "" {
		defaultDir := s.Config.Backup.Output.BaseDirectory
		if defaultDir == "" {
			defaultDir = "./backups"
		}

		if !allowInteractive {
			backupOpts.OutputDir = defaultDir
			s.Log.Infof("Direktori backup pre-restore (non-interaktif): %s", backupOpts.OutputDir)
			goto finalize
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

finalize:

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
