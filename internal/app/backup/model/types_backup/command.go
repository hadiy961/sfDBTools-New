// File : internal/app/backup/model/types_backup/command.go
// Deskripsi : Command execution config structs
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-05
// Last Modified : 2026-01-05

package types_backup

// ExecutionConfig menyimpan konfigurasi untuk execution
type ExecutionConfig struct {
	Mode        string
	HeaderTitle string
	LogPrefix   string
	SuccessMsg  string
}
