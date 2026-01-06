// File : internal/restore/executor.go
// Deskripsi : Factory/Wrapper untuk membuat executor berdasarkan mode restore
// Author : Hadiyatna Muflihun
// Tanggal : 17 Desember 2025
// Last Modified : 5 Januari 2026
package restore

import (
	"context"
	restoremodel "sfdbtools/internal/app/restore/model"
	"sfdbtools/internal/app/restore/modes"
	"sfdbtools/pkg/consts"
)

// ExecuteRestoreSingle menjalankan restore single database
func (s *Service) ExecuteRestoreSingle(ctx context.Context) (*restoremodel.RestoreResult, error) {
	executor, err := modes.GetExecutor(consts.ModeSingle, s)
	if err != nil {
		return nil, err
	}
	return executor.Execute(ctx)
}

// ExecuteRestorePrimary menjalankan restore primary database
func (s *Service) ExecuteRestorePrimary(ctx context.Context) (*restoremodel.RestoreResult, error) {
	executor, err := modes.GetExecutor(consts.ModePrimary, s)
	if err != nil {
		return nil, err
	}
	return executor.Execute(ctx)
}

// ExecuteRestoreSecondary menjalankan restore secondary database
func (s *Service) ExecuteRestoreSecondary(ctx context.Context) (*restoremodel.RestoreResult, error) {
	executor, err := modes.GetExecutor(consts.ModeSecondary, s)
	if err != nil {
		return nil, err
	}
	return executor.Execute(ctx)
}

// ExecuteRestoreAll menjalankan restore all databases dengan streaming filtering
func (s *Service) ExecuteRestoreAll(ctx context.Context) (*restoremodel.RestoreResult, error) {
	executor, err := modes.GetExecutor(consts.ModeAll, s)
	if err != nil {
		return nil, err
	}
	return executor.Execute(ctx)
}

// ExecuteRestoreSelection menjalankan restore selection berbasis CSV
func (s *Service) ExecuteRestoreSelection(ctx context.Context) (*restoremodel.RestoreResult, error) {
	executor, err := modes.GetExecutor(consts.ModeSelection, s)
	if err != nil {
		return nil, err
	}
	return executor.Execute(ctx)
}

// ExecuteRestoreCustom menjalankan restore custom (SFCola account detail)
func (s *Service) ExecuteRestoreCustom(ctx context.Context) (*restoremodel.RestoreResult, error) {
	executor, err := modes.GetExecutor(consts.ModeCustom, s)
	if err != nil {
		return nil, err
	}
	return executor.Execute(ctx)
}
