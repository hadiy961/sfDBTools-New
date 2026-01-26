// File : internal/app/dbcopy/setup.go
// Deskripsi : Common setup logic untuk copy operations
// Author : Hadiyatna Muflihun
// Tanggal : 26 Januari 2026
// Last Modified : 26 Januari 2026
package dbcopy

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"sfdbtools/internal/app/dbcopy/helpers"
	"sfdbtools/internal/app/dbcopy/model"
	"sfdbtools/internal/domain"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/database"
	"sfdbtools/internal/shared/runtimecfg"
)

// SetupContext menyiapkan context dengan cancellation dan signal handling
func (s *Service) SetupContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		s.log.Warn("Menerima signal interupsi, membatalkan operasi...")
		cancel()
	}()

	return ctx, cancel
}

// SetupProfiles memuat source dan target profiles
func (s *Service) SetupProfiles(opts *model.CommonCopyOptions, allowInteractive bool) (*domain.ProfileInfo, *domain.ProfileInfo, error) {
	// Load source profile
	srcProfile, err := s.LoadProfile(
		opts.SourceProfile,
		opts.SourceProfileKey,
		consts.ENV_SOURCE_PROFILE,
		consts.ENV_SOURCE_PROFILE_KEY,
		allowInteractive,
		"source",
	)
	if err != nil {
		return nil, nil, fmt.Errorf("gagal load source profile: %w", err)
	}

	// Resolve target profile (default: sama dengan source)
	targetPath, targetKey := helpers.ResolveTargetProfile(
		srcProfile,
		opts.TargetProfile,
		opts.TargetProfileKey,
	)

	// Load target profile
	tgtProfile, err := s.LoadProfile(
		targetPath,
		targetKey,
		consts.ENV_TARGET_PROFILE,
		consts.ENV_TARGET_PROFILE_KEY,
		allowInteractive,
		"target",
	)
	if err != nil {
		return nil, nil, fmt.Errorf("gagal load target profile: %w", err)
	}

	return srcProfile, tgtProfile, nil
}

// SetupWorkdir memastikan working directory tersedia
func (s *Service) SetupWorkdir(opts *model.CommonCopyOptions) (workdir string, cleanup func(), err error) {
	return s.EnsureWorkdir(opts.Workdir)
}

// SetupConnections membuat koneksi database ke source (target di-skip karena restore handle sendiri)
func (s *Service) SetupConnections(srcProfile *domain.ProfileInfo) (*database.Client, error) {
	srcClient, err := s.ConnectDB(srcProfile)
	if err != nil {
		return nil, fmt.Errorf("gagal connect ke source database: %w", err)
	}

	return srcClient, nil
}

// DetermineNonInteractiveMode menentukan apakah mode non-interaktif aktif
func DetermineNonInteractiveMode(opts *model.CommonCopyOptions) bool {
	return runtimecfg.IsQuiet() || opts.SkipConfirm
}

// CheckCompanionDatabase memeriksa apakah companion database (_dmart) ada
func (s *Service) CheckCompanionDatabase(ctx context.Context, client *database.Client, primaryDB string, includeDmart bool) (companionName string, exists bool, err error) {
	if !includeDmart {
		return "", false, nil
	}

	companion := strings.TrimSpace(primaryDB) + "_dmart"
	found, err := s.CheckDatabaseExists(ctx, client, companion)
	if err != nil {
		return companion, false, fmt.Errorf("gagal cek companion database: %w", err)
	}

	return companion, found, nil
}
