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

	"sfDBTools/internal/types"
	"sfDBTools/pkg/profilehelper"
	"sfDBTools/pkg/ui"
)

// ExecuteRestoreCommand adalah entry point untuk menjalankan restore command
func (s *Service) ExecuteRestoreCommand(ctx context.Context, restoreConfig types.RestoreEntryConfig) error {
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

	// Verify backup file (basic validation)
	s.Log.Info("Memverifikasi backup file...")
	if err := s.verifyBackupFile(ctx); err != nil {
		return fmt.Errorf("verifikasi backup file gagal: %w", err)
	}

	// Show options dan minta konfirmasi jika ShowOptions=true dan Force=false
	if s.RestoreEntry.ShowOptions && !s.RestoreOptions.Force {
		proceed, err := s.DisplayRestoreOptions()
		if err != nil {
			return fmt.Errorf("gagal menampilkan opsi restore: %w", err)
		}
		if !proceed {
			return types.ErrUserCancelled
		}
	}

	// Koneksi ke target database
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
		result, err = s.executeRestoreMulti(ctx)
	default:
		return fmt.Errorf("mode restore tidak valid: %s", restoreConfig.RestoreMode)
	}

	if err != nil {
		return fmt.Errorf("restore gagal: %w", err)
	}

	// Display results
	s.DisplayRestoreResult(result)

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

	// Untuk mode multi, source harus direktori
	if s.RestoreEntry.RestoreMode == "multi" {
		if !fileInfo.IsDir() {
			return fmt.Errorf("source harus directory untuk restore multi: %s", sourceFile)
		}
		s.Log.Infof("✓ Source directory validated: %s", sourceFile)
		return nil
	}

	// Untuk mode single dan all, source harus file
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
	s.Log.Infof("Loading target profile...")

	// Gunakan profilehelper untuk load profile dengan fallback ke env vars
	profile, err := profilehelper.LoadTargetProfile(
		s.RestoreOptions.TargetProfile,
		s.RestoreOptions.TargetProfileKey,
	)
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
	s.Log.Infof("Connecting to target database...")

	// Gunakan profilehelper untuk koneksi yang konsisten
	client, err := profilehelper.ConnectWithTargetProfile(s.TargetProfile, "mysql")
	if err != nil {
		return fmt.Errorf("gagal connect ke target database: %w", err)
	}

	s.Client = client
	s.Log.Info("✓ Connected to target database")

	return nil
}
