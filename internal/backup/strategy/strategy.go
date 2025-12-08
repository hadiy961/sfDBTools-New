// File : internal/backup/strategy/strategy.go
// Deskripsi : Strategy pattern untuk berbagai mode backup
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-05
// Last Modified : 2025-12-05

package strategy

import (
	"context"
	"sfDBTools/internal/types/types_backup"
)

// BackupStrategy mendefinisikan interface untuk berbagai strategi backup
type BackupStrategy interface {
	// Execute menjalankan backup dengan strategi tertentu
	Execute(ctx context.Context, dbList []string) types_backup.BackupResult
}

// BackupMode enumeration untuk jenis-jenis backup
type BackupMode string

const (
	BackupModeSingle    BackupMode = "single"
	BackupModeSeparated BackupMode = "separated"
	BackupModeCombined  BackupMode = "combined"
	BackupModePrimary   BackupMode = "primary"
	BackupModeSecondary BackupMode = "secondary"
)

// StrategyFactory membuat strategy berdasarkan mode
// Ini memudahkan untuk menambah mode baru tanpa mengubah executor logic
// Contoh penggunaan:
//
//	strategy, err := StrategyFactory(BackupModeSingle, service)
//	if err != nil {
//	  // handle error
//	}
//	result := strategy.Execute(ctx, dbList)
type StrategyFactory func(mode BackupMode, svc BackupServiceProvider) (BackupStrategy, error)

// BackupServiceProvider adalah interface untuk service yang diperlukan strategy
// Ini memungkinkan strategy untuk mengakses dependencies tanpa circular imports
type BackupServiceProvider interface {
	// ExecuteBackupSingle menjalankan backup single mode
	ExecuteBackupSingle(ctx context.Context, dbList []string) types_backup.BackupResult

	// ExecuteBackupSeparated menjalankan backup separated mode
	ExecuteBackupSeparated(ctx context.Context, dbFiltered []string) types_backup.BackupResult

	// ExecuteBackupCombined menjalankan backup combined mode
	ExecuteBackupCombined(ctx context.Context, dbFiltered []string) types_backup.BackupResult
}
