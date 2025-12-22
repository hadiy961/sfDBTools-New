// File : internal/restore/executor.go
// Deskripsi : Factory/Wrapper untuk membuat executor berdasarkan mode restore
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-17
// Last Modified : 2025-12-17

package restore

import (
	"context"
	"sfDBTools/internal/restore/modes"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/consts"
)

// ExecuteRestoreSingle menjalankan restore single database
func (s *Service) ExecuteRestoreSingle(ctx context.Context) (*types.RestoreResult, error) {
	executor, err := modes.GetExecutor(consts.ModeSingle, s)
	if err != nil {
		return nil, err
	}
	return executor.Execute(ctx)
}

// ExecuteRestorePrimary menjalankan restore primary database
func (s *Service) ExecuteRestorePrimary(ctx context.Context) (*types.RestoreResult, error) {
	executor, err := modes.GetExecutor(consts.ModePrimary, s)
	if err != nil {
		return nil, err
	}
	return executor.Execute(ctx)
}

// ExecuteRestoreAll menjalankan restore all databases dengan streaming filtering
func (s *Service) ExecuteRestoreAll(ctx context.Context) (*types.RestoreResult, error) {
	executor, err := modes.GetExecutor(consts.ModeAll, s)
	if err != nil {
		return nil, err
	}
	return executor.Execute(ctx)
}

// ExecuteRestoreSelection menjalankan restore selection berbasis CSV
func (s *Service) ExecuteRestoreSelection(ctx context.Context) (*types.RestoreResult, error) {
	executor, err := modes.GetExecutor(consts.ModeSelection, s)
	if err != nil {
		return nil, err
	}
	return executor.Execute(ctx)
}
