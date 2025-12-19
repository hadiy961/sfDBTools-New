// File : internal/restore/validation_helpers.go
// Deskripsi : Helper functions untuk validation operations
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-19
// Last Modified : 2025-12-19

package restore

import (
	"context"
	"fmt"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/ui"
	"strings"
)

// validateApplicationPassword memvalidasi password aplikasi sebelum restore primary
func (s *Service) validateApplicationPassword() error {
	s.Log.Info("Meminta password aplikasi untuk validasi restore primary")

	// Prompt password
	password, err := input.PromptPassword("Masukkan password aplikasi untuk melanjutkan restore primary:")
	if err != nil {
		return fmt.Errorf("gagal membaca password: %w", err)
	}

	// Validasi password dengan ENV_PASSWORD_APP dari consts
	if password != consts.ENV_PASSWORD_APP {
		s.Log.Error("Password aplikasi tidak valid")
		return fmt.Errorf("password aplikasi tidak valid")
	}

	s.Log.Info("Password aplikasi valid, melanjutkan restore")
	ui.PrintSuccess("✓ Password aplikasi valid")

	return nil
}

// DropAllDatabases menghapus semua database non-sistem
func (s *Service) DropAllDatabases(ctx context.Context) error {
	s.Log.Info("Mengambil daftar database untuk drop all...")

	dbList, err := s.TargetClient.GetDatabaseList(ctx)
	if err != nil {
		return fmt.Errorf("gagal mengambil daftar database: %w", err)
	}

	systemDBs := []string{"mysql", "sys", "information_schema", "performance_schema"}
	droppedCount := 0

	for _, dbName := range dbList {
		// Skip system databases
		isSystem := false
		for _, sys := range systemDBs {
			if strings.EqualFold(dbName, sys) {
				isSystem = true
				break
			}
		}
		if isSystem {
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
