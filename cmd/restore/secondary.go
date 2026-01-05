package restorecmd

import (
	"fmt"
	appdeps "sfDBTools/internal/cli/deps"
	"sfDBTools/internal/cli/flags"
	"sfDBTools/internal/restore"

	"github.com/spf13/cobra"
)

// CmdRestoreSecondary adalah command untuk restore secondary database
var CmdRestoreSecondary = &cobra.Command{
	Use:   "secondary",
	Short: "Restore database Secondary (dari file atau dari Primary)",
	Long: `Restore database secondary untuk kebutuhan training/testing atau instance turunan.

Sumber restore dapat berasal dari:
  - file  : restore dari file backup dump
  - primary : backup primary di target server, lalu restore hasil backup ke secondary

Companion (dmart):
	- Secara default tool akan mencoba ikut restore companion database (_dmart).
	- Untuk mode --from file, gunakan --dmart-file atau auto-detect (dmart-detect).
	- Untuk mode --from primary, tool akan mencari database primary_dmart dan ikut memulihkan ke secondary_dmart (jika ada).

Target secondary dibentuk dari client-code dan instance:
  dbsf_nbc_{client-code}_secondary_{instance}
  dbsf_biznet_{client-code}_secondary_{instance}`,
	Example: `  # 1) Restore secondary dari file
  sfdbtools db-restore secondary --from file --file "backup.sql.gz" --client-code "adaro" --instance "training" --ticket "TICKET-123"

  # 2) Restore secondary dari primary (backup primary dulu)
	sfdbtools db-restore secondary --from primary --client-code "adaro" --instance "training" --ticket "TICKET-123"`,
	Run: func(cmd *cobra.Command, args []string) {
		if appdeps.Deps == nil {
			fmt.Println("✗ Dependencies tidak tersedia. Pastikan aplikasi diinisialisasi dengan benar.")
			return
		}

		if err := restore.ExecuteRestoreSecondaryCommand(cmd, appdeps.Deps); err != nil {
			appdeps.Deps.Logger.Error("restore secondary gagal: " + err.Error())
		}
	},
}

func init() {
	flags.AddRestoreSecondaryAllFlags(CmdRestoreSecondary)
}
