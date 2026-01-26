// File : internal/app/dbcopy/model/types.go
// Deskripsi : Type definitions untuk db-copy operations
// Author : Hadiyatna Muflihun
// Tanggal : 26 Januari 2026
// Last Modified : 26 Januari 2026
package model

// CommonCopyOptions berisi field yang sama untuk semua mode copy
type CommonCopyOptions struct {
	// Profile & Authentication
	SourceProfile    string
	SourceProfileKey string
	TargetProfile    string
	TargetProfileKey string

	// Audit & Control
	Ticket string

	// Behavior Flags
	SkipConfirm     bool
	ContinueOnError bool
	DryRun          bool
	ExcludeData     bool
	IncludeDmart    bool
	PrebackupTarget bool

	// Working Directory
	Workdir string
}

// P2POptions untuk Primary to Primary copy
type P2POptions struct {
	CommonCopyOptions

	// Rule-based mode
	ClientCode       string
	TargetClientCode string

	// Explicit mode (override rule-based)
	SourceDB string
	TargetDB string
}

// P2SOptions untuk Primary to Secondary copy
type P2SOptions struct {
	CommonCopyOptions

	// Rule-based mode
	ClientCode string
	Instance   string

	// Explicit mode (override rule-based)
	SourceDB string
	TargetDB string
}

// S2SOptions untuk Secondary to Secondary copy
type S2SOptions struct {
	CommonCopyOptions

	// Rule-based mode
	ClientCode     string
	SourceInstance string
	TargetInstance string

	// Explicit mode (override rule-based)
	SourceDB string
	TargetDB string
}

// CopyMode menentukan jenis copy operation
type CopyMode string

const (
	ModeP2P CopyMode = "p2p" // Primary to Primary
	ModeP2S CopyMode = "p2s" // Primary to Secondary
	ModeS2S CopyMode = "s2s" // Secondary to Secondary
)

// CopyResult menyimpan hasil copy operation
type CopyResult struct {
	Success         bool
	SourceDB        string
	TargetDB        string
	CompanionCopied bool
	Message         string
	Error           error
}
