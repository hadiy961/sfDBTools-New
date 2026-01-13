// File : internal/app/profile/executor/interfaces.go
// Deskripsi : ISP-compliant interfaces untuk Executor (Interface Segregation Principle)
// Author : Hadiyatna Muflihun
// Tanggal : 14 Januari 2026
// Last Modified : 14 Januari 2026
package executor

import "sfdbtools/internal/domain"

// WizardRunner menangani wizard flow operations
type WizardRunner interface {
	RunWizard(mode string) error
	PromptCreateRetrySelectedFields() error
}

// ProfileDisplay menangani tampilan profile details
type ProfileDisplay interface {
	DisplayProfileDetails()
}

// ProfileValidator menangani validasi profile
type ProfileValidator interface {
	CheckConfigurationNameUnique(mode string) error
	ValidateProfileInfo(p *domain.ProfileInfo) error
}

// ProfileFormatter menangani formatting profile ke berbagai format
type ProfileFormatter interface {
	FormatConfigToINI() string
}

// ProfileSelector menangani pemilihan profile interaktif
type ProfileSelector interface {
	PromptSelectExistingConfig() error
}

// ProfileLoader menangani loading profile snapshot
type ProfileLoader interface {
	LoadSnapshotFromPath(absPath string) (*domain.ProfileInfo, error)
}

// KeyResolver menangani resolusi encryption key
type KeyResolver interface {
	ResolveProfileEncryptionKey(existing string, allowPrompt bool) (key string, source string, err error)
}
