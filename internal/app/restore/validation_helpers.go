// File : internal/restore/validation_helpers.go
// Deskripsi : Helper functions untuk validation operations
// Author : Hadiyatna Muflihun
// Tanggal : 19 Desember 2025
// Last Modified : 5 Januari 2026
package restore

import (
	"context"
	"fmt"
	"sfdbtools/internal/domain"
	"sfdbtools/internal/crypto"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/ui/print"
	"sfdbtools/internal/ui/prompt"
	"strings"
)

// validateApplicationPassword memvalidasi password aplikasi sebelum restore primary
func (s *Service) validateApplicationPassword() error {
	s.Log.Info("Meminta password aplikasi untuk validasi restore primary")

	var password string
	var err error

	if !s.IsInteractive() {
		// Non-interactive mode: attempt to read from ENV, expecting encrypted format if used.
		// Use crypto.ResolveKey which will automatically decrypt SFDBTOOLS:... formatted strings
		resolvedKey, _, resolveErr := crypto.ResolveKey("", consts.ENV_APPS_PASSWORD, true)
		if resolveErr != nil || strings.TrimSpace(resolvedKey) == "" {
			return fmt.Errorf("mode non-interaktif memerlukan environment variable %s untuk konfirmasi", consts.ENV_APPS_PASSWORD)
		}
		password = resolvedKey
	} else {
		// Interactive mode: Prompt password
		password, err = prompt.PromptPassword("Masukkan password aplikasi untuk melanjutkan restore primary:")
		if err != nil {
			return fmt.Errorf("gagal membaca password: %w", err)
		}
	}

	// Validasi password dengan ENV_PASSWORD_APP dari consts
	if password != consts.ENV_PASSWORD_APP {
		s.Log.Error("Password aplikasi tidak valid")
		return fmt.Errorf("password aplikasi tidak valid")
	}

	s.Log.Info("Password aplikasi valid, melanjutkan restore")
	print.PrintSuccess("✓ Password aplikasi valid")

	return nil
}

// DropAllDatabases menghapus semua database non-sistem
func (s *Service) DropAllDatabases(ctx context.Context) error {
	s.Log.Info("Mengambil daftar database untuk drop all...")

	dbList, err := s.TargetClient.GetDatabaseList(ctx)
	if err != nil {
		return fmt.Errorf("gagal mengambil daftar database: %w", err)
	}

	droppedCount := 0

	for _, dbName := range dbList {
		// Skip system databases
		if _, isSystem := domain.SystemDatabases[strings.ToLower(dbName)]; isSystem {
			continue
		}

		s.Log.Infof("Dropping database: %s", dbName)
		if err := s.TargetClient.DropDatabase(ctx, dbName); err != nil {
			s.Log.Errorf("Gagal drop database %s: %v", dbName, err)
			return fmt.Errorf("gagal drop database %s: %w", dbName, err)
		}
		droppedCount++
	}

	s.Log.Infof("Berhasil drop %d database", droppedCount)
	return nil
}
