// File : internal/restore/restore_entry.go
// Deskripsi : Entry point untuk restore command execution
// Author : Hadiyatna Muflihun
// Tanggal : 2025-11-05
// Last Modified : 2025-11-05

package restore

import (
	"context"
	"fmt"
	"os"
	"sfDBTools/internal/profileselect"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/helper"
	"sfDBTools/pkg/ui"
	"time"
)

// ExecuteRestoreCommand adalah entry point untuk menjalankan restore command
func (s *Service) ExecuteRestoreCommand(ctx context.Context, restoreConfig types.RestoreEntryConfig) error {
	startTime := time.Now()

	// Set restore entry config
	s.RestoreEntry = &restoreConfig

	// Show header
	ui.Headers(restoreConfig.HeaderTitle)

	s.Log.Infof("%s Memulai restore database...", restoreConfig.LogPrefix)

	// Validate source file exists
	if err := s.validateSourceFile(); err != nil {
		return fmt.Errorf("validasi source file gagal: %w", err)
	}

	// Resolve target profile
	if err := s.resolveTargetProfile(ctx); err != nil {
		return fmt.Errorf("resolve target profile gagal: %w", err)
	}

	// Show options jika diminta
	if restoreConfig.ShowOptions {
		s.displayRestoreOptions()
	}

	// Verify backup file (decrypt, decompress, checksum)
	if s.RestoreOptions.VerifyChecksum {
		s.Log.Info("Memverifikasi backup file...")
		if err := s.verifyBackupFile(ctx); err != nil {
			return fmt.Errorf("verifikasi backup file gagal: %w", err)
		}
	}

	// Connect to target database
	if err := s.connectToTargetDatabase(ctx); err != nil {
		return fmt.Errorf("koneksi ke target database gagal: %w", err)
	}
	defer s.Client.Close()

	// Execute restore based on mode
	var result types.RestoreResult
	var err error

	switch restoreConfig.RestoreMode {
	case "single":
		result, err = s.executeRestoreSingle(ctx)
	case "all":
		result, err = s.executeRestoreAll(ctx)
	case "multi":
		return fmt.Errorf("restore multi belum diimplementasikan")
	default:
		return fmt.Errorf("mode restore tidak valid: %s", restoreConfig.RestoreMode)
	}

	if err != nil {
		return fmt.Errorf("restore gagal: %w", err)
	}

	// Display results
	s.displayRestoreResults(result, time.Since(startTime))

	s.Log.Info(restoreConfig.SuccessMsg)
	return nil
}

// validateSourceFile memvalidasi keberadaan file backup source
func (s *Service) validateSourceFile() error {
	sourceFile := s.RestoreOptions.SourceFile

	if sourceFile == "" {
		return fmt.Errorf("source file tidak boleh kosong")
	}

	fileInfo, err := os.Stat(sourceFile)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file backup tidak ditemukan: %s", sourceFile)
		}
		return fmt.Errorf("gagal mengakses file backup: %w", err)
	}

	if fileInfo.IsDir() {
		return fmt.Errorf("source harus file, bukan directory: %s", sourceFile)
	}

	if fileInfo.Size() == 0 {
		return fmt.Errorf("file backup kosong: %s", sourceFile)
	}

	s.Log.Infof("✓ Source file validated: %s (size: %d bytes)", sourceFile, fileInfo.Size())
	return nil
}

// resolveTargetProfile me-resolve target profile untuk restore
func (s *Service) resolveTargetProfile(ctx context.Context) error {
	profilePath := s.RestoreOptions.TargetProfile
	profileKey := s.RestoreOptions.TargetProfileKey

	// Resolve dari environment variable jika tidak ada
	if profilePath == "" {
		profilePath = helper.GetEnvOrDefault(consts.ENV_TARGET_PROFILE, "")
		if profilePath == "" {
			return fmt.Errorf("target profile tidak tersedia, gunakan flag --profile atau env %s", consts.ENV_TARGET_PROFILE)
		}
	}

	if profileKey == "" {
		profileKey = helper.GetEnvOrDefault(consts.ENV_TARGET_PROFILE_KEY, "")
	}

	s.Log.Infof("Loading target profile: %s", profilePath)

	// Load profile menggunakan profileselect
	profile, err := profileselect.LoadAndParseProfile(profilePath, profileKey)
	if err != nil {
		return fmt.Errorf("gagal load target profile: %w", err)
	}

	s.TargetProfile = profile
	s.Log.Infof("✓ Target profile loaded: %s@%s:%d",
		profile.DBInfo.User, profile.DBInfo.Host, profile.DBInfo.Port)

	return nil
}

// connectToTargetDatabase membuat koneksi ke database target
func (s *Service) connectToTargetDatabase(ctx context.Context) error {
	s.Log.Info("Connecting to target database...")

	creds := types.DestinationDBConnection{
		DBInfo: s.TargetProfile.DBInfo,
	}

	client, err := database.ConnectToDestinationDatabase(creds)

	if err != nil {
		return fmt.Errorf("gagal connect ke target database: %w", err)
	}

	s.Client = client
	s.Log.Info("✓ Connected to target database")

	return nil
}

// displayRestoreOptions menampilkan opsi restore sebelum eksekusi
func (s *Service) displayRestoreOptions() {
	ui.PrintSubHeader("Restore Options")
	fmt.Printf("  Source File         : %s\n", s.RestoreOptions.SourceFile)
	fmt.Printf("  Target Profile      : %s\n", s.RestoreOptions.TargetProfile)
	fmt.Printf("  Target Database     : %s\n", s.RestoreOptions.TargetDB)
	fmt.Printf("  Verify Checksum     : %v\n", s.RestoreOptions.VerifyChecksum)
	fmt.Printf("  Backup Before       : %v\n", s.RestoreOptions.BackupBeforeRestore)
	fmt.Printf("  Mode                : %s\n", s.RestoreOptions.Mode)
	fmt.Printf("  Dry Run             : %v\n", s.RestoreOptions.DryRun)
	fmt.Println()

	if !s.RestoreOptions.Force {
		// Simple confirmation prompt
		fmt.Print("Lanjutkan restore? (y/n): ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			s.Log.Warn("Restore dibatalkan oleh user")
			os.Exit(0)
		}
	}
}

// displayRestoreResults menampilkan hasil restore
func (s *Service) displayRestoreResults(result types.RestoreResult, duration time.Duration) {
	ui.PrintSubHeader("Restore Summary")

	// Show pre-backup info if available
	if result.PreBackupFile != "" {
		fmt.Printf("  Safety Backup       : %s\n", result.PreBackupFile)
	}

	fmt.Printf("  Total Databases     : %d\n", result.TotalDatabases)

	// Hitung skipped databases untuk dry-run
	skippedCount := 0
	for _, info := range result.RestoreInfo {
		if info.Status == "skipped" {
			skippedCount++
		}
	}

	if skippedCount > 0 {
		fmt.Printf("  Skipped (Dry-Run)   : %d\n", skippedCount)
	} else {
		fmt.Printf("  Successful Restore  : %d\n", result.SuccessfulRestore)
		fmt.Printf("  Failed Restore      : %d\n", result.FailedRestore)
	}

	fmt.Printf("  Total Time          : %s\n", duration.String())
	fmt.Println()

	// Show restore details
	if len(result.RestoreInfo) > 0 {
		ui.PrintSubHeader("Restore Details")
		for _, info := range result.RestoreInfo {
			status := "✓"
			if info.Status == "failed" {
				status = "✗"
			} else if info.Status == "skipped" {
				status = "⊙"
			}
			fmt.Printf("  %s %s -> %s (%s)\n",
				status, info.DatabaseName, info.TargetDatabase, info.Duration)

			if info.ErrorMessage != "" {
				fmt.Printf("     Error: %s\n", info.ErrorMessage)
			}
		}
		fmt.Println()
	}
}
