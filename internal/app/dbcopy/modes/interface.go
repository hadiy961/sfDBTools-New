// File : internal/app/dbcopy/modes/interface.go
// Deskripsi : Interface definitions untuk copy modes (ISP-compliant)
// Author : Hadiyatna Muflihun
// Tanggal : 26 Januari 2026
// Last Modified : 26 Januari 2026
package modes

import (
	"context"

	"sfdbtools/internal/app/dbcopy/model"
)

// Executor adalah interface untuk semua mode copy operations
type Executor interface {
	// Execute menjalankan copy operation dengan options yang sesuai
	Execute(ctx context.Context) (*model.CopyResult, error)
}

// ModeRegistry menyimpan factory functions untuk membuat executors
type ModeRegistry interface {
	// GetExecutor mengembalikan executor yang sesuai berdasarkan mode
	GetExecutor(mode model.CopyMode, opts interface{}) (Executor, error)
}
