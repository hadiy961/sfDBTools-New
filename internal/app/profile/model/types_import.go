// File : internal/app/profile/model/types_import.go
// Deskripsi : Types dan constants untuk import profile functionality
// Author : Hadiyatna Muflihun
// Tanggal : 25 Januari 2026
// Last Modified : 25 Januari 2026

package types

import (
	"fmt"
	"strings"
)

// Import conflict resolution strategies
const (
	ImportConflictFail      = "fail"
	ImportConflictSkip      = "skip"
	ImportConflictOverwrite = "overwrite"
	ImportConflictRename    = "rename"
)

// Import plan actions
const (
	ImportPlanCreate    = "create"
	ImportPlanOverwrite = "overwrite"
	ImportPlanRename    = "rename"
	ImportPlanSkip      = "skip"
)

// Import skip reasons
const (
	ImportSkipReasonInvalid   = "invalid"
	ImportSkipReasonDuplicate = "duplicate"
	ImportSkipReasonConflict  = "conflict"
	ImportSkipReasonConnTest  = "conn_test"
	ImportSkipReasonUnknown   = "unknown"
)

// ImportCellError represents validation error untuk cell tertentu di import file
type ImportCellError struct {
	Row     int
	Column  string
	Message string
}

// String implements Stringer interface untuk ImportCellError
func (e ImportCellError) String() string {
	col := strings.TrimSpace(e.Column)
	if col == "" {
		col = "(unknown)"
	}
	return fmt.Sprintf("row %d, col %s: %s", e.Row, col, e.Message)
}

// ImportRow represents single row dari import source (XLSX/CSV)
type ImportRow struct {
	RowNum int

	// Database connection fields
	Name       string
	Host       string
	Port       int
	User       string
	Password   string
	ProfileKey string

	// SSH tunnel fields
	SSHEnabled  bool
	SSHHost     string
	SSHPort     int
	SSHUser     string
	SSHPassword string
	SSHIdentity string
	SSHLocal    int

	// Processing metadata
	Skip       bool   // Whether to skip this row
	SkipReason string // Reason for skipping (invalid|duplicate|conflict|conn_test|unknown)
	PlanAction string // Action to take (create|overwrite|rename|skip)

	// Derived fields
	PlannedName string // Final name setelah conflict resolution
	RenamedFrom string // Original name jika di-rename
}
