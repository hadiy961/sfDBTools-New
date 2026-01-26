// File : internal/app/dbcopy/modes/service_interface.go
// Deskripsi : Service interface untuk dependency injection ke executors
// Author : Hadiyatna Muflihun
// Tanggal : 26 Januari 2026
// Last Modified : 26 Januari 2026
package modes

import (
	"context"

	"sfdbtools/internal/app/dbcopy/model"
	"sfdbtools/internal/domain"
	"sfdbtools/internal/shared/database"
)

// CopyService adalah interface yang menyediakan semua capabilities untuk copy executors
// Mengikuti ISP (Interface Segregation Principle)
type CopyService interface {
	// Profile & Connection Management
	SetupProfiles(opts *model.CommonCopyOptions, allowInteractive bool) (*domain.ProfileInfo, *domain.ProfileInfo, error)
	SetupConnections(srcProfile *domain.ProfileInfo) (*database.Client, error)
	SetupWorkdir(opts *model.CommonCopyOptions) (workdir string, cleanup func(), err error)

	// Database Operations
	ResolvePrimaryDB(ctx context.Context, client interface{}, clientCode string) (string, error)
	CheckCompanionDatabase(ctx context.Context, client *database.Client, primaryDB string, includeDmart bool) (companionName string, exists bool, err error)
	ValidateNotCopyToSelf(srcProfile, tgtProfile *domain.ProfileInfo, sourceDB, targetDB string, mode string) error

	// Backup Operations
	ResolveBackupEncryptionKey() (string, error)
	BackupSingleDB(ctx context.Context, profile *domain.ProfileInfo, client *database.Client, dbName, ticket, workdir string, excludeData bool) (string, error)

	// Restore Operations
	RestorePrimary(ctx context.Context, profile *domain.ProfileInfo, file, companionFile, targetDB, ticket, encryptionKey string, includeDmart, dropTarget, skipBackup, skipGrants, continueOnError, nonInteractive bool) error
	RestoreSecondary(ctx context.Context, profile *domain.ProfileInfo, file, companionFile, ticket, clientCode, instance, encryptionKey string, includeDmart, dropTarget, skipBackup, continueOnError, nonInteractive bool) error
	RestoreSingle(ctx context.Context, profile *domain.ProfileInfo, file, targetDB, ticket, encryptionKey string, dropTarget, skipBackup, skipGrants, continueOnError, nonInteractive bool) error
}
