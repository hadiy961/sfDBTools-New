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
)

// ExecuteRestoreSingle menjalankan restore single database
func (s *Service) ExecuteRestoreSingle(ctx context.Context) (*types.RestoreResult, error) {
	executor, err := modes.GetExecutor("single", s)
	if err != nil {
		return nil, err
	}
	return executor.Execute(ctx)
}

// ExecuteRestorePrimary menjalankan restore primary database
func (s *Service) ExecuteRestorePrimary(ctx context.Context) (*types.RestoreResult, error) {
	executor, err := modes.GetExecutor("primary", s)
	if err != nil {
		return nil, err
	}
	return executor.Execute(ctx)
}
