// File : internal/restore/setup_types.go
// Deskripsi : Common types untuk setup restore operations
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-30
// Last Modified : 2026-01-05
package restore

import (
	"sfDBTools/internal/domain"
)

// basicSetupOptions contains common setup options for all modes
type basicSetupOptions struct {
	file          *string
	encryptionKey *string
	profile       *domain.ProfileInfo
	interactive   bool
}

// postDatabaseSetupOptions contains post-database setup options
type postDatabaseSetupOptions struct {
	ticket       *string
	dropTarget   *bool
	skipBackup   *bool
	skipGrants   *bool
	grantsFile   *string
	backupFile   string
	stopOnError  bool
	includeDmart bool
	interactive  bool
}
