// File : cmd/backup/run_helper.go
// Deskripsi : Helper eksekusi command backup untuk menjaga alur konsisten
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-30
// Last Modified : 2025-12-30

package backupcmd

import (
	"fmt"
	"sfDBTools/internal/backup"
	appdeps "sfDBTools/internal/deps"

	"github.com/spf13/cobra"
)

// runBackupCommand mengeksekusi backup dengan pengecekan dependency dan resolver mode terpadu.
func runBackupCommand(cmd *cobra.Command, modeResolver func() (string, error)) {
	if appdeps.Deps == nil {
		fmt.Println("✗ Dependencies tidak tersedia. Pastikan aplikasi diinisialisasi dengan benar.")
		return
	}

	backupMode, err := modeResolver()
	if err != nil {
		if appdeps.Deps.Logger != nil {
			appdeps.Deps.Logger.Error(err.Error())
		}
		return
	}

	if err := backup.ExecuteBackup(cmd, appdeps.Deps, backupMode); err != nil {
		return
	}
}
